package auth

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	"github.com/your-org/your-service/internal/config"
	"github.com/your-org/your-service/internal/modules/rbac"
	"github.com/your-org/your-service/internal/modules/tenant"
	"github.com/your-org/your-service/internal/modules/user"
	"github.com/your-org/your-service/internal/pkg/appctx"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/hash"
	"github.com/your-org/your-service/internal/pkg/jwt"
	"github.com/your-org/your-service/internal/pkg/tokens"
	"github.com/your-org/your-service/internal/queue"
)

const (
	defaultInviteTTLHours        = 72
	defaultPasswordResetTTLHours = 1
	failedLoginLockThreshold     = 10
	failedLoginLockMinutes       = 15
)

type Service struct {
	repo           *Repository
	tenantSvc      *tenant.Service
	userSvc        *user.Service
	rbacSvc        *rbac.Service
	signer         *jwt.Signer
	cache          cache.Cache
	producer       queue.Producer
	cfg            *config.Config
	log            *zap.Logger
	refreshTTL     time.Duration
}

func NewService(
	repo *Repository,
	tenantSvc *tenant.Service,
	userSvc *user.Service,
	rbacSvc *rbac.Service,
	signer *jwt.Signer,
	c cache.Cache,
	p queue.Producer,
	cfg *config.Config,
	log *zap.Logger,
) *Service {
	refreshTTL := time.Duration(cfg.Auth.RefreshTokenDays) * 24 * time.Hour
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &Service{
		repo:       repo,
		tenantSvc:  tenantSvc,
		userSvc:    userSvc,
		rbacSvc:    rbacSvc,
		signer:     signer,
		cache:      c,
		producer:   p,
		cfg:        cfg,
		log:        log,
		refreshTTL: refreshTTL,
	}
}

// ── discover ───────────────────────────────────────────────────────────────

// Discover returns the tenants the email belongs to. Always returns a non-nil
// slice; never reveals whether the email exists.
func (s *Service) Discover(ctx context.Context, email string) ([]DiscoveredTenant, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return []DiscoveredTenant{}, nil
	}
	ids, err := s.userSvc.ListTenantIDsByEmail(ctx, email)
	if err != nil {
		s.log.Warn("discover lookup failed", zap.Error(err))
		return []DiscoveredTenant{}, nil
	}
	out := make([]DiscoveredTenant, 0, len(ids))
	for _, id := range ids {
		t, err := s.tenantSvc.GetTenant(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, DiscoveredTenant{
			ID:           t.ID,
			Name:         t.Name,
			Slug:         t.Slug,
			LogoURL:      t.LogoURL,
			PrimaryColor: t.PrimaryColor,
		})
	}
	return out, nil
}

// ── login ──────────────────────────────────────────────────────────────────

func (s *Service) Login(ctx context.Context, req LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	u, err := s.userSvc.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperr.New(apperr.CodeInvalidCredentials, "invalid email or password", nil)
	}
	if u.Status == "suspended" || u.Status == "archived" {
		return nil, apperr.New(apperr.CodeForbidden, "user not allowed to sign in", nil)
	}
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return nil, apperr.New(apperr.CodeForbidden, "account temporarily locked", nil)
	}
	if !hash.ComparePassword(u.PasswordHash, req.Password) {
		_ = s.userSvc.RecordFailedLogin(ctx, u.ID)
		if u.FailedLoginCount+1 >= failedLoginLockThreshold {
			until := time.Now().Add(failedLoginLockMinutes * time.Minute)
			_ = s.lockUser(ctx, u.ID, until)
		}
		return nil, apperr.New(apperr.CodeInvalidCredentials, "invalid email or password", nil)
	}

	memberships, err := s.userSvc.ListMembershipsByUser(ctx, u.ID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "lookup memberships failed", err)
	}
	relevant := []user.Membership{}
	for _, m := range memberships {
		if m.TenantID == req.TenantID && (m.Status == "active" || m.Status == "invited") {
			relevant = append(relevant, m)
		}
	}
	if len(relevant) == 0 {
		return nil, apperr.New(apperr.CodeForbidden, "no membership in tenant", nil)
	}
	active := pickDefaultMembership(relevant)

	t, err := s.tenantSvc.GetTenant(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	resp, err := s.issueLoginPair(ctx, u, t, &active, ip, userAgent)
	if err != nil {
		return nil, err
	}
	if err := s.userSvc.RecordLogin(ctx, u.ID, ip, userAgent); err != nil {
		s.log.Warn("record login failed", zap.Error(err))
	}
	orgs, _ := s.collectOrgs(ctx, relevant)
	resp.Organizations = orgs
	return resp, nil
}

// SwitchOrg re-issues tokens with a different active organization.
func (s *Service) SwitchOrg(ctx context.Context, targetOrg uuid.UUID, ip, userAgent string) (*LoginResponse, error) {
	uid := appctx.UserID(ctx)
	tid := appctx.TenantID(ctx)
	if uid == uuid.Nil || tid == uuid.Nil {
		return nil, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil)
	}
	memberships, err := s.userSvc.ListMembershipsByUser(ctx, uid)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "lookup memberships failed", err)
	}
	var picked *user.Membership
	for i, m := range memberships {
		if m.TenantID == tid && m.OrganizationID == targetOrg && m.Status == "active" {
			picked = &memberships[i]
			break
		}
	}
	if picked == nil {
		return nil, apperr.New(apperr.CodeForbidden, "no active membership in target org", nil)
	}
	u, err := s.userSvc.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	t, err := s.tenantSvc.GetTenant(ctx, tid)
	if err != nil {
		return nil, err
	}
	return s.issueLoginPair(ctx, u, t, picked, ip, userAgent)
}

// ── refresh ────────────────────────────────────────────────────────────────

func (s *Service) Refresh(ctx context.Context, token, ip, userAgent string) (*LoginResponse, error) {
	tokenHash := hash.SHA256(token)
	rec, err := s.repo.GetRefreshByHash(ctx, tokenHash)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeInvalidToken, "invalid refresh token", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "refresh lookup failed", err)
	}
	if rec.RevokedAt != nil {
		// Reuse of a revoked token → revoke the entire family.
		_ = s.repo.RevokeRefreshFamily(ctx, rec.FamilyID, "reuse_detected")
		return nil, apperr.New(apperr.CodeInvalidToken, "refresh token reuse detected", nil)
	}
	if rec.ExpiresAt.Before(time.Now()) {
		return nil, apperr.New(apperr.CodeTokenExpired, "refresh token expired", nil)
	}
	u, err := s.userSvc.GetByID(ctx, rec.UserID)
	if err != nil {
		return nil, err
	}
	t, err := s.tenantSvc.GetTenant(ctx, rec.TenantID)
	if err != nil {
		return nil, err
	}
	var picked *user.Membership
	if rec.OrganizationID != nil {
		m, err := s.userSvc.ListMembershipsByUser(ctx, u.ID)
		if err == nil {
			for i := range m {
				if m[i].OrganizationID == *rec.OrganizationID {
					picked = &m[i]
					break
				}
			}
		}
	}
	resp, err := s.issueLoginPair(ctx, u, t, picked, ip, userAgent)
	if err != nil {
		return nil, err
	}
	// Rotate: revoke old, link new via family.
	_ = s.repo.RevokeRefresh(ctx, rec.ID, "rotated")
	_ = s.repo.TouchRefresh(ctx, rec.ID)
	return resp, nil
}

// ── logout / sessions ──────────────────────────────────────────────────────

func (s *Service) Logout(ctx context.Context, refresh string) error {
	if refresh == "" {
		return nil
	}
	rec, err := s.repo.GetRefreshByHash(ctx, hash.SHA256(refresh))
	if err != nil {
		return nil
	}
	return s.repo.RevokeRefresh(ctx, rec.ID, "logout")
}

func (s *Service) ListSessions(ctx context.Context) ([]RefreshToken, error) {
	uid := appctx.UserID(ctx)
	if uid == uuid.Nil {
		return nil, apperr.New(apperr.CodeUnauthorized, "not authenticated", nil)
	}
	return s.repo.ListActiveRefreshByUser(ctx, uid)
}

func (s *Service) RevokeSession(ctx context.Context, jti uuid.UUID) error {
	rec, err := s.repo.GetRefreshByJTI(ctx, jti)
	if err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "session not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "revoke session failed", err)
	}
	if rec.UserID != appctx.UserID(ctx) && !appctx.IsSuperAdmin(ctx) {
		return apperr.New(apperr.CodeForbidden, "cannot revoke another user's session", nil)
	}
	return s.repo.RevokeRefresh(ctx, jti, "user_revoked")
}

// ── register (new tenant + admin) ──────────────────────────────────────────

// Register creates a fresh tenant, default org, owner user, owner membership,
// and seeds system roles. Returns a login pair so the client lands logged-in.
// The new tenant has NO subscription — caller must onboard one before any
// gated feature works.
func (s *Service) Register(ctx context.Context, req RegisterRequest, ip, userAgent string) (*LoginResponse, error) {
	// 1. Tenant + default org.
	t, defaultOrg, err := s.tenantSvc.CreateTenant(ctx, tenant.CreateTenantRequest{
		Slug:           req.OrganizationSlug,
		Name:           req.OrganizationName,
		AdminEmail:     req.Email,
		AdminFirstName: req.FirstName,
		AdminLastName:  req.LastName,
	})
	if err != nil {
		return nil, err
	}

	// 2. Owner user (full bcrypt; status=active so they can sign in straight away).
	pw, err := hash.Password(req.Password)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "hash password failed", err)
	}
	u, err := s.userSvc.Create(ctx, user.CreateUserInput{
		Email:                 req.Email,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		PasswordHash:          pw,
		Status:                "active",
		PrimaryTenantID:       ptrUUID(t.ID),
		PrimaryOrganizationID: ptrUUID(defaultOrg.ID),
	})
	if err != nil {
		return nil, err
	}

	// 3. Owner membership.
	m, err := s.userSvc.CreateMembership(ctx, user.CreateMembershipInput{
		UserID:           u.ID,
		TenantID:         t.ID,
		OrganizationID:   defaultOrg.ID,
		Status:           "active",
		IsDefault:        true,
		IsOwner:          true,
	})
	if err != nil {
		return nil, err
	}

	// 4. Seed system roles (owner/admin/member) for the new org.
	actor := ptrUUID(u.ID)
	if _, err := s.rbacSvc.SeedSystemRolesForOrg(ctx, t.ID, defaultOrg.ID, actor); err != nil {
		return nil, err
	}

	// 5. Assign Owner role to the new membership.
	if err := s.rbacSvc.AssignSystemRoleByKey(ctx, defaultOrg.ID, m.ID, "owner", actor); err != nil {
		return nil, err
	}

	// 6. Issue tokens — the response shape mirrors /auth/login.
	return s.issueLoginPair(ctx, u, t, m, ip, userAgent)
}

// ── invite ─────────────────────────────────────────────────────────────────

func (s *Service) Invite(ctx context.Context, req InviteRequest) (*Invite, string, error) {
	tid := appctx.TenantID(ctx)
	if tid == uuid.Nil {
		return nil, "", apperr.New(apperr.CodeForbidden, "no tenant context", nil)
	}
	if _, err := s.tenantSvc.GetOrganization(ctx, tid, req.OrganizationID); err != nil {
		return nil, "", err
	}
	tokenStr, err := tokens.New(32)
	if err != nil {
		return nil, "", apperr.New(apperr.CodeInternal, "generate token failed", err)
	}
	// Resolve role keys → uuid[] for the role_ids column (NOT NULL).
	roleIDs := pq.StringArray{}
	for _, k := range req.RoleKeys {
		key := strings.ToLower(strings.TrimSpace(k))
		if key == "" {
			continue
		}
		r, rerr := s.rbacSvc.GetRoleByKey(ctx, req.OrganizationID, key)
		if rerr == nil && r != nil {
			roleIDs = append(roleIDs, r.ID.String())
		}
	}
	inv := &Invite{
		TenantID:       tid,
		OrganizationID: req.OrganizationID,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		JobTitle:       req.JobTitle,
		Department:     req.Department,
		TokenHash:      hash.SHA256(tokenStr),
		RoleIDs:        roleIDs,
		InvitedBy:      ptrUUID(appctx.UserID(ctx)),
		Message:        req.Message,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(defaultInviteTTLHours * time.Hour),
		Metadata:       []byte("{}"),
	}
	if err := s.repo.CreateInvite(ctx, inv); err != nil {
		return nil, "", apperr.New(apperr.CodeInternal, "create invite failed", err)
	}
	s.publishInviteEmail(ctx, inv, tokenStr)
	return inv, tokenStr, nil
}

func (s *Service) AcceptInvite(ctx context.Context, req AcceptInviteRequest, ip, userAgent string) (*LoginResponse, error) {
	inv, err := s.repo.GetInviteByTokenHash(ctx, hash.SHA256(req.Token))
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "invalid invite token", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "invite lookup failed", err)
	}
	if inv.Status != "pending" {
		return nil, apperr.New(apperr.CodeInvalidInput, "invite no longer valid", nil)
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return nil, apperr.New(apperr.CodeTokenExpired, "invite expired", nil)
	}
	pw, err := hash.Password(req.NewPasswordOrInvite(req.Password))
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "hash password failed", err)
	}

	// Reuse existing user record by email if present; else create.
	u, err := s.userSvc.GetByEmail(ctx, inv.Email)
	if err != nil {
		u, err = s.userSvc.Create(ctx, user.CreateUserInput{
			Email:                 inv.Email,
			FirstName:             req.FirstName,
			LastName:              req.LastName,
			PasswordHash:          pw,
			Status:                "active",
			PrimaryTenantID:       ptrUUID(inv.TenantID),
			PrimaryOrganizationID: ptrUUID(inv.OrganizationID),
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Existing user accepting an invite — update password + reactivate.
		if err := s.userSvc.UpdatePasswordHash(ctx, u.ID, pw); err != nil {
			return nil, err
		}
		_ = s.userSvc.Reactivate(ctx, u.ID)
	}

	// Create membership.
	m, err := s.userSvc.CreateMembership(ctx, user.CreateMembershipInput{
		UserID:         u.ID,
		TenantID:       inv.TenantID,
		OrganizationID: inv.OrganizationID,
		Status:         "active",
		IsDefault:      true,
		InvitedBy:      inv.InvitedBy,
		JobTitle:       inv.JobTitle,
		Department:     inv.Department,
	})
	if err != nil {
		return nil, err
	}

	// Assign initial roles. If none specified on the invite, fall back to the
	// 'member' default role.
	if len(inv.RoleIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(inv.RoleIDs))
		for _, s := range inv.RoleIDs {
			if id, err := uuid.Parse(s); err == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			_ = s.rbacSvc.AssignRolesToMembership(ctx, inv.OrganizationID, m.ID, []string{}, ptrUUID(appctx.UserID(ctx)))
		}
	}
	_ = s.rbacSvc.AssignSystemRoleByKey(ctx, inv.OrganizationID, m.ID, "member", ptrUUID(appctx.UserID(ctx)))

	if err := s.repo.MarkInviteAccepted(ctx, inv.ID, u.ID); err != nil {
		s.log.Warn("mark invite accepted failed", zap.Error(err))
	}

	t, err := s.tenantSvc.GetTenant(ctx, inv.TenantID)
	if err != nil {
		return nil, err
	}
	return s.issueLoginPair(ctx, u, t, m, ip, userAgent)
}

// ── password reset ─────────────────────────────────────────────────────────

func (s *Service) ForgotPassword(ctx context.Context, email, ip, userAgent string) error {
	u, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		// Silent success — don't leak existence.
		return nil
	}
	tokenStr, err := tokens.New(32)
	if err != nil {
		return apperr.New(apperr.CodeInternal, "generate token failed", err)
	}
	rec := &PasswordResetToken{
		UserID:    u.ID,
		TokenHash: hash.SHA256(tokenStr),
		ExpiresAt: time.Now().Add(defaultPasswordResetTTLHours * time.Hour),
		UserAgent: userAgent,
	}
	if err := s.repo.CreatePasswordReset(ctx, rec); err != nil {
		return apperr.New(apperr.CodeInternal, "create reset token failed", err)
	}
	s.publishPasswordResetEmail(ctx, u, tokenStr)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	rec, err := s.repo.GetPasswordResetByHash(ctx, hash.SHA256(req.Token))
	if err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "invalid token", nil)
		}
		return apperr.New(apperr.CodeInternal, "reset lookup failed", err)
	}
	if rec.ExpiresAt.Before(time.Now()) {
		return apperr.New(apperr.CodeTokenExpired, "token expired", nil)
	}
	pw, err := hash.Password(req.NewPassword)
	if err != nil {
		return apperr.New(apperr.CodeInternal, "hash failed", err)
	}
	if err := s.directPasswordUpdate(ctx, rec.UserID, pw); err != nil {
		return err
	}
	return s.repo.MarkPasswordResetUsed(ctx, rec.ID)
}

func (s *Service) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	uid := appctx.UserID(ctx)
	u, err := s.userSvc.GetByID(ctx, uid)
	if err != nil {
		return err
	}
	if !hash.ComparePassword(u.PasswordHash, req.CurrentPassword) {
		return apperr.New(apperr.CodeInvalidCredentials, "current password incorrect", nil)
	}
	pw, err := hash.Password(req.NewPassword)
	if err != nil {
		return apperr.New(apperr.CodeInternal, "hash failed", err)
	}
	return s.directPasswordUpdate(ctx, uid, pw)
}

// ── internal helpers ───────────────────────────────────────────────────────

func (s *Service) issueLoginPair(
	ctx context.Context,
	u *user.User,
	t *tenant.Tenant,
	m *user.Membership,
	ip, userAgent string,
) (*LoginResponse, error) {
	claims := jwt.Claims{
		UserID:       u.ID,
		TenantID:     t.ID,
		Email:        u.Email,
		IsSuperAdmin: u.IsSuperAdmin,
	}
	if m != nil {
		claims.OrganizationID = m.OrganizationID
		claims.MembershipID = m.ID
	}
	access, accessExp, _, err := s.signer.Issue(claims)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "issue access token failed", err)
	}

	refreshStr, err := tokens.New(32)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "generate refresh token failed", err)
	}
	familyID := uuid.New()
	refreshRec := &RefreshToken{
		UserID:    u.ID,
		TenantID:  t.ID,
		TokenHash: hash.SHA256(refreshStr),
		FamilyID:  familyID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.refreshTTL),
		IP:        ipToInet(ip),
		UserAgent: userAgent,
		Metadata:  []byte("{}"),
	}
	if m != nil {
		oid := m.OrganizationID
		refreshRec.OrganizationID = &oid
		mid := m.ID
		refreshRec.MembershipID = &mid
	}
	if err := s.repo.CreateRefresh(ctx, refreshRec); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "persist refresh token failed", err)
	}

	resp := &LoginResponse{
		AccessToken:           access,
		RefreshToken:          refreshStr,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshRec.ExpiresAt,
		TokenType:             "Bearer",
		User: UserSummary{
			ID:          u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			AvatarURL:   u.AvatarURL,
		},
		Tenant: DiscoveredTenant{
			ID:           t.ID,
			Name:         t.Name,
			Slug:         t.Slug,
			LogoURL:      t.LogoURL,
			PrimaryColor: t.PrimaryColor,
		},
	}
	if m != nil {
		// Hydrate the active organization summary.
		org, err := s.tenantSvc.GetOrganization(ctx, t.ID, m.OrganizationID)
		if err == nil {
			resp.ActiveOrganization = &OrgSummary{
				ID:        org.ID,
				Name:      org.Name,
				Slug:      org.Slug,
				LogoURL:   org.LogoURL,
				IsDefault: org.IsDefault,
				Status:    org.Status,
			}
		}
	}
	return resp, nil
}

func (s *Service) collectOrgs(ctx context.Context, memberships []user.Membership) ([]OrgSummary, error) {
	out := make([]OrgSummary, 0, len(memberships))
	for _, m := range memberships {
		org, err := s.tenantSvc.GetOrganization(ctx, m.TenantID, m.OrganizationID)
		if err != nil {
			continue
		}
		out = append(out, OrgSummary{
			ID:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			LogoURL:   org.LogoURL,
			IsDefault: m.IsDefault,
			Status:    org.Status,
		})
	}
	return out, nil
}

func (s *Service) directPasswordUpdate(ctx context.Context, userID uuid.UUID, pwHash string) error {
	// Reach into the user repo via the service helper. user.Service.Update only
	// updates profile fields; password updates go through a dedicated path.
	return s.userSvc.UpdatePasswordHash(ctx, userID, pwHash)
}

func (s *Service) lockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	return s.userSvc.LockUser(ctx, userID, until)
}

func (s *Service) publishInviteEmail(ctx context.Context, inv *Invite, tokenStr string) {
	if s.producer == nil {
		return
	}
	_ = s.producer.Publish(ctx, queue.ChannelInviteEmail, map[string]interface{}{
		"inviteId":       inv.ID,
		"email":          inv.Email,
		"token":          tokenStr,
		"organizationId": inv.OrganizationID,
		"tenantId":       inv.TenantID,
	})
}

func (s *Service) publishPasswordResetEmail(ctx context.Context, u *user.User, tokenStr string) {
	if s.producer == nil {
		return
	}
	_ = s.producer.Publish(ctx, queue.ChannelPasswordResetEmail, map[string]interface{}{
		"userId": u.ID,
		"email":  u.Email,
		"token":  tokenStr,
		"baseUrl": s.cfg.Auth.PasswordResetBaseURL,
	})
}

func pickDefaultMembership(ms []user.Membership) user.Membership {
	for _, m := range ms {
		if m.IsDefault {
			return m
		}
	}
	return ms[0]
}

// ipToInet parses a request IP for the Postgres inet column. Returns nil
// when the input is empty or not parseable so GORM writes NULL.
func ipToInet(ip string) *string {
	if ip == "" {
		return nil
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil
	}
	s := parsed.String()
	return &s
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	if u == uuid.Nil {
		return nil
	}
	id := u
	return &id
}

// pickPassword keeps the request-side password readable (allows future hooks
// like enforcing a policy without changing call sites).
func (r AcceptInviteRequest) NewPasswordOrInvite(p string) string { return p }
