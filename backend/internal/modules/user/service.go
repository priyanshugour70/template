package user

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/pkg/pagination"
)

type Service struct {
	repo  *Repository
	log   *zap.Logger
	cache cache.Cache
}

func NewService(repo *Repository, log *zap.Logger, c cache.Cache) *Service {
	return &Service{repo: repo, log: log, cache: c}
}

// ── users ──────────────────────────────────────────────────────────────────

// Create is used by the auth module (invite → accept flow). Returns an error
// if email already exists.
func (s *Service) Create(ctx context.Context, in CreateUserInput) (*User, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" {
		return nil, apperr.New(apperr.CodeValidation, "email is required", nil)
	}
	if existing, _ := s.repo.GetUserByEmail(ctx, in.Email); existing != nil {
		return nil, apperr.New(apperr.CodeAlreadyExists, "email already in use", nil)
	}
	u := &User{
		Email:                 in.Email,
		PasswordHash:          in.PasswordHash,
		PasswordAlgo:          "bcrypt",
		FirstName:             in.FirstName,
		MiddleName:            in.MiddleName,
		LastName:              in.LastName,
		DisplayName:           firstNonEmpty(in.DisplayName, joinNames(in.FirstName, in.LastName), in.Email),
		Status:                firstNonEmpty(in.Status, "invited"),
		Locale:                firstNonEmpty(in.Locale, "en-IN"),
		Timezone:              firstNonEmpty(in.Timezone, "Asia/Kolkata"),
		Phone:                 in.Phone,
		JobTitle:              in.JobTitle,
		Department:            in.Department,
		EmployeeCode:          in.EmployeeCode,
		PrimaryTenantID:       in.PrimaryTenantID,
		PrimaryOrganizationID: in.PrimaryOrganizationID,
		IsSuperAdmin:          in.IsSuperAdmin,
		Preferences:           []byte("{}"),
		NotificationPreferences: []byte("{}"),
		Metadata:              []byte("{}"),
	}
	if in.Status == "active" {
		now := time.Now()
		u.PasswordChangedAt = &now
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create user failed", err)
	}
	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch user failed", err)
	}
	return u, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch user failed", err)
	}
	return u, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*User, error) {
	patch := map[string]interface{}{}
	setIfNotNil(patch, "first_name", req.FirstName)
	setIfNotNil(patch, "middle_name", req.MiddleName)
	setIfNotNil(patch, "last_name", req.LastName)
	setIfNotNil(patch, "display_name", req.DisplayName)
	setIfNotNil(patch, "username", req.Username)
	setIfNotNil(patch, "avatar_url", req.AvatarURL)
	setIfNotNil(patch, "cover_url", req.CoverURL)
	setIfNotNil(patch, "bio", req.Bio)
	setIfNotNil(patch, "phone", req.Phone)
	setIfNotNil(patch, "alt_email", req.AltEmail)
	if req.DateOfBirth != nil {
		patch["date_of_birth"] = *req.DateOfBirth
	}
	setIfNotNil(patch, "gender", req.Gender)
	setIfNotNil(patch, "job_title", req.JobTitle)
	setIfNotNil(patch, "department", req.Department)
	setIfNotNil(patch, "employee_code", req.EmployeeCode)
	setIfNotNil(patch, "locale", req.Locale)
	setIfNotNil(patch, "timezone", req.Timezone)
	setIfNotNil(patch, "country", req.Country)
	setIfNotNil(patch, "state", req.State)
	setIfNotNil(patch, "city", req.City)
	if req.Address != nil {
		patch["address"] = []byte(req.Address)
	}
	if req.Preferences != nil {
		patch["preferences"] = []byte(req.Preferences)
	}
	if req.NotificationPreferences != nil {
		patch["notification_preferences"] = []byte(req.NotificationPreferences)
	}
	if req.Metadata != nil {
		patch["metadata"] = []byte(req.Metadata)
	}
	if len(patch) == 0 {
		return s.GetByID(ctx, id)
	}
	if err := s.repo.UpdateUser(ctx, id, patch); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "update user failed", err)
	}
	return s.GetByID(ctx, id)
}

func (s *Service) Suspend(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SuspendUser(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "suspend user failed", err)
	}
	return nil
}

func (s *Service) Reactivate(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.ActivateUser(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "reactivate user failed", err)
	}
	return nil
}

func (s *Service) Archive(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.ArchiveUser(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "archive user failed", err)
	}
	return nil
}

func (s *Service) RecordLogin(ctx context.Context, id uuid.UUID, ip, userAgent string) error {
	return s.repo.UpdateUserLogin(ctx, id, ip, userAgent)
}

func (s *Service) RecordFailedLogin(ctx context.Context, id uuid.UUID) error {
	return s.repo.IncrementFailedLogin(ctx, id)
}

func (s *Service) ListInOrg(ctx context.Context, tenantID, orgID uuid.UUID, filter ListFilter, p pagination.Params) ([]User, int64, error) {
	rows, total, err := s.repo.ListUsersInOrg(ctx, tenantID, orgID, filter, p)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list users failed", err)
	}
	return rows, total, nil
}

// ── memberships ────────────────────────────────────────────────────────────

func (s *Service) CreateMembership(ctx context.Context, in CreateMembershipInput) (*Membership, error) {
	m := &Membership{
		UserID:         in.UserID,
		TenantID:       in.TenantID,
		OrganizationID: in.OrganizationID,
		Status:         firstNonEmpty(in.Status, "active"),
		IsDefault:      in.IsDefault,
		IsOwner:        in.IsOwner,
		JobTitle:       in.JobTitle,
		Department:     in.Department,
		EmployeeCode:   in.EmployeeCode,
		InvitedBy:      in.InvitedBy,
		Settings:       []byte("{}"),
		Metadata:       []byte("{}"),
	}
	now := time.Now()
	if m.Status == "active" {
		m.JoinedAt = &now
	} else if m.Status == "invited" {
		m.InvitedAt = &now
	}
	if err := s.repo.CreateMembership(ctx, m); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create membership failed", err)
	}
	return m, nil
}

func (s *Service) GetMembership(ctx context.Context, id uuid.UUID) (*Membership, error) {
	m, err := s.repo.GetMembership(ctx, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "membership not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch membership failed", err)
	}
	return m, nil
}

func (s *Service) ListMembershipsByUser(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
	return s.repo.ListMembershipsByUser(ctx, userID)
}

func (s *Service) ListTenantIDsByEmail(ctx context.Context, email string) ([]uuid.UUID, error) {
	return s.repo.ListTenantIDsByUserEmail(ctx, email)
}

func (s *Service) SuspendMembership(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.SuspendMembership(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "membership not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "suspend membership failed", err)
	}
	return nil
}

// UpdatePasswordHash sets the user's password hash and stamps the change time.
// Called by the auth module from accept-invite, reset-password, change-password.
func (s *Service) UpdatePasswordHash(ctx context.Context, id uuid.UUID, pwHash string) error {
	patch := map[string]interface{}{
		"password_hash":       pwHash,
		"password_changed_at": time.Now(),
		"failed_login_count":  0,
		"locked_until":        nil,
	}
	if err := s.repo.UpdateUser(ctx, id, patch); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "user not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "update password failed", err)
	}
	return nil
}

// LockUser locks the account until the given time (used by auth on too many
// failed logins).
func (s *Service) LockUser(ctx context.Context, id uuid.UUID, until time.Time) error {
	return s.repo.LockUser(ctx, id, until)
}

func (s *Service) ArchiveMembership(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.ArchiveMembership(ctx, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "membership not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "archive membership failed", err)
	}
	return nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func firstNonEmpty(parts ...string) string {
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			return p
		}
	}
	return ""
}

func joinNames(parts ...string) string {
	out := []string{}
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return strings.Join(out, " ")
}

func setIfNotNil(m map[string]interface{}, col string, v *string) {
	if v != nil {
		m[col] = *v
	}
}
