package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/your-org/your-service/internal/cache"
	apperr "github.com/your-org/your-service/internal/pkg/errors"
	"github.com/your-org/your-service/internal/queue"
)

const (
	permCacheTTL     = 15 * time.Minute
	permCachePrefix  = "perm:"
)

type Service struct {
	repo     *Repository
	log      *zap.Logger
	cache    cache.Cache
	producer queue.Producer
}

func NewService(repo *Repository, log *zap.Logger, c cache.Cache, p queue.Producer) *Service {
	return &Service{repo: repo, log: log, cache: c, producer: p}
}

// ── permissions ────────────────────────────────────────────────────────────

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list permissions failed", err)
	}
	return rows, nil
}

// ResolveForUser returns the union of permission keys the user has in the org.
// Reads cache first; on miss, hits DB and primes cache.
func (s *Service) ResolveForUser(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	if userID == uuid.Nil || orgID == uuid.Nil {
		return nil, nil
	}
	key := permCacheKey(userID, orgID)
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil && raw != "" {
			var out []string
			if err := json.Unmarshal([]byte(raw), &out); err == nil {
				return out, nil
			}
		}
	}
	keys, err := s.repo.ResolvePermissionsForUserOrg(ctx, userID, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "resolve permissions failed", err)
	}
	if s.cache != nil {
		if b, err := json.Marshal(keys); err == nil {
			_ = s.cache.Set(ctx, key, string(b), permCacheTTL)
		}
	}
	return keys, nil
}

// HasPermission reports whether the user has the named permission in the org.
func (s *Service) HasPermission(ctx context.Context, userID, orgID uuid.UUID, perm string) (bool, error) {
	if perm == "" {
		return true, nil
	}
	keys, err := s.ResolveForUser(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	for _, k := range keys {
		if k == perm {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) HasAny(ctx context.Context, userID, orgID uuid.UUID, perms ...string) (bool, error) {
	keys, err := s.ResolveForUser(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	have := make(map[string]bool, len(keys))
	for _, k := range keys {
		have[k] = true
	}
	for _, p := range perms {
		if have[p] {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) HasAll(ctx context.Context, userID, orgID uuid.UUID, perms ...string) (bool, error) {
	keys, err := s.ResolveForUser(ctx, userID, orgID)
	if err != nil {
		return false, err
	}
	have := make(map[string]bool, len(keys))
	for _, k := range keys {
		have[k] = true
	}
	for _, p := range perms {
		if !have[p] {
			return false, nil
		}
	}
	return true, nil
}

// ── roles ──────────────────────────────────────────────────────────────────

func (s *Service) CreateRole(ctx context.Context, tenantID, orgID uuid.UUID, req CreateRoleRequest, actor *uuid.UUID) (*Role, error) {
	if existing, _ := s.repo.GetRoleByKey(ctx, orgID, strings.ToLower(req.Key)); existing != nil {
		return nil, apperr.New(apperr.CodeAlreadyExists, "role key already taken", nil)
	}
	r := &Role{
		TenantID:       ptrUUID(tenantID),
		OrganizationID: ptrUUID(orgID),
		Key:            strings.ToLower(req.Key),
		Name:           req.Name,
		Description:    req.Description,
		Priority:       req.Priority,
		Color:          req.Color,
		Icon:           req.Icon,
		IsAssignable:   true,
		IsSystem:       false,
		Metadata:       []byte("{}"),
	}
	if err := s.repo.CreateRole(ctx, r); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create role failed", err)
	}
	if len(req.PermissionKeys) > 0 {
		if err := s.setRolePermissionsByKey(ctx, r.ID, req.PermissionKeys, actor); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (s *Service) GetRole(ctx context.Context, orgID, id uuid.UUID) (*Role, error) {
	r, err := s.repo.GetRoleByID(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "role not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "fetch role failed", err)
	}
	return r, nil
}

func (s *Service) ListRoles(ctx context.Context, orgID uuid.UUID) ([]Role, error) {
	rows, err := s.repo.ListRoles(ctx, orgID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list roles failed", err)
	}
	return rows, nil
}

func (s *Service) UpdateRole(ctx context.Context, orgID, id uuid.UUID, req UpdateRoleRequest, actor *uuid.UUID) (*Role, error) {
	patch := map[string]interface{}{}
	if req.Name != nil {
		patch["name"] = *req.Name
	}
	if req.Description != nil {
		patch["description"] = *req.Description
	}
	if req.Priority != nil {
		patch["priority"] = *req.Priority
	}
	if req.Color != nil {
		patch["color"] = *req.Color
	}
	if req.Icon != nil {
		patch["icon"] = *req.Icon
	}
	if req.IsAssignable != nil {
		patch["is_assignable"] = *req.IsAssignable
	}
	if len(patch) > 0 {
		if err := s.repo.UpdateRole(ctx, orgID, id, patch); err != nil {
			if IsNotFound(err) {
				return nil, apperr.New(apperr.CodeNotFound, "role not found", nil)
			}
			return nil, apperr.New(apperr.CodeInternal, "update role failed", err)
		}
	}
	if req.PermissionKeys != nil {
		if err := s.setRolePermissionsByKey(ctx, id, req.PermissionKeys, actor); err != nil {
			return nil, err
		}
		s.invalidateRoleCache(ctx, id)
	}
	return s.GetRole(ctx, orgID, id)
}

func (s *Service) ArchiveRole(ctx context.Context, orgID, id uuid.UUID) error {
	if err := s.repo.ArchiveRole(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "role not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "archive role failed", err)
	}
	s.invalidateRoleCache(ctx, id)
	return nil
}

func (s *Service) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	rows, err := s.repo.ListRolePermissions(ctx, roleID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list role permissions failed", err)
	}
	return rows, nil
}

func (s *Service) setRolePermissionsByKey(ctx context.Context, roleID uuid.UUID, keys []string, actor *uuid.UUID) error {
	perms, err := s.repo.GetPermissionsByKeys(ctx, keys)
	if err != nil {
		return apperr.New(apperr.CodeInternal, "lookup permissions failed", err)
	}
	if len(perms) != len(keys) {
		known := make(map[string]bool, len(perms))
		for _, p := range perms {
			known[p.Key] = true
		}
		for _, k := range keys {
			if !known[k] {
				return apperr.New(apperr.CodeValidation, fmt.Sprintf("unknown permission key: %s", k), nil)
			}
		}
	}
	ids := make([]uuid.UUID, 0, len(perms))
	for _, p := range perms {
		ids = append(ids, p.ID)
	}
	if err := s.repo.SetRolePermissions(ctx, roleID, ids, actor); err != nil {
		return apperr.New(apperr.CodeInternal, "set role permissions failed", err)
	}
	return nil
}

// ── membership roles ───────────────────────────────────────────────────────

func (s *Service) AssignRolesToMembership(ctx context.Context, orgID, membershipID uuid.UUID, roleKeys []string, actor *uuid.UUID) error {
	if len(roleKeys) == 0 {
		return apperr.New(apperr.CodeValidation, "no roles to assign", nil)
	}
	ids := []uuid.UUID{}
	for _, k := range roleKeys {
		r, err := s.repo.GetRoleByKey(ctx, orgID, strings.ToLower(k))
		if err != nil {
			return apperr.New(apperr.CodeNotFound, fmt.Sprintf("role %q not found", k), nil)
		}
		ids = append(ids, r.ID)
	}
	if err := s.repo.AssignRolesToMembership(ctx, membershipID, ids, actor); err != nil {
		return apperr.New(apperr.CodeInternal, "assign roles failed", err)
	}
	s.invalidateMembershipCache(ctx, membershipID)
	return nil
}

func (s *Service) RemoveRoleFromMembership(ctx context.Context, membershipID, roleID uuid.UUID) error {
	if err := s.repo.RemoveRoleFromMembership(ctx, membershipID, roleID); err != nil {
		return apperr.New(apperr.CodeInternal, "remove role failed", err)
	}
	s.invalidateMembershipCache(ctx, membershipID)
	return nil
}

func (s *Service) ListMembershipRoles(ctx context.Context, membershipID uuid.UUID) ([]Role, error) {
	rows, err := s.repo.ListMembershipRoles(ctx, membershipID)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list membership roles failed", err)
	}
	return rows, nil
}

// ── system role seeding (called by tenant onboarding) ──────────────────────

// SeedSystemRolesForOrg creates Owner / Admin / Member roles for a newly
// minted organization and seeds their permission catalogs from current
// permissions table. Returns the Owner role so the caller can assign it to
// the first user.
func (s *Service) SeedSystemRolesForOrg(ctx context.Context, tenantID, orgID uuid.UUID, actor *uuid.UUID) (ownerRole *Role, err error) {
	perms, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list permissions failed", err)
	}
	allKeys := make([]string, 0, len(perms))
	keyToID := make(map[string]uuid.UUID, len(perms))
	for _, p := range perms {
		allKeys = append(allKeys, p.Key)
		keyToID[p.Key] = p.ID
	}

	var owner *Role
	for _, seed := range DefaultSystemRoles() {
		role := &Role{
			TenantID:       ptrUUID(tenantID),
			OrganizationID: ptrUUID(orgID),
			Key:            seed.Key,
			Name:           seed.Name,
			Description:    seed.Description,
			IsSystem:       true,
			IsDefault:      seed.IsDefault,
			IsAssignable:   true,
			Priority:       seed.Priority,
			Metadata:       []byte("{}"),
		}
		if err := s.repo.CreateRole(ctx, role); err != nil {
			return nil, apperr.New(apperr.CodeInternal, "create system role failed", err)
		}
		chosen := seed.PermissionKey(allKeys)
		ids := make([]uuid.UUID, 0, len(chosen))
		for _, k := range chosen {
			if id, ok := keyToID[k]; ok {
				ids = append(ids, id)
			}
		}
		if err := s.repo.SetRolePermissions(ctx, role.ID, ids, actor); err != nil {
			return nil, apperr.New(apperr.CodeInternal, "seed role permissions failed", err)
		}
		if seed.Key == "owner" {
			owner = role
		}
	}
	return owner, nil
}

// AssignSystemRoleByKey is a helper for the auth/onboarding flow.
func (s *Service) AssignSystemRoleByKey(ctx context.Context, orgID, membershipID uuid.UUID, roleKey string, actor *uuid.UUID) error {
	return s.AssignRolesToMembership(ctx, orgID, membershipID, []string{roleKey}, actor)
}

// ── cache invalidation ─────────────────────────────────────────────────────

func (s *Service) invalidateRoleCache(ctx context.Context, roleID uuid.UUID) {
	if s.producer == nil {
		return
	}
	_ = s.producer.Publish(ctx, queue.ChannelPermissionInvalidate, map[string]interface{}{
		"roleId": roleID.String(),
	})
}

func (s *Service) invalidateMembershipCache(ctx context.Context, membershipID uuid.UUID) {
	if s.producer == nil {
		return
	}
	_ = s.producer.Publish(ctx, queue.ChannelPermissionInvalidate, map[string]interface{}{
		"membershipId": membershipID.String(),
	})
}

// InvalidateUserOrgCache is called by the consumer worker (and on direct
// changes) to drop the cached permission set for a single (user, org).
func (s *Service) InvalidateUserOrgCache(ctx context.Context, userID, orgID uuid.UUID) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Delete(ctx, permCacheKey(userID, orgID))
}

// InvalidateMembership publishes a cache-bust for the given membership.
// Used by department/group services after role-binding changes — implements
// the CacheBuster interface those modules accept.
func (s *Service) InvalidateMembership(ctx context.Context, membershipID uuid.UUID) {
	s.invalidateMembershipCache(ctx, membershipID)
}

func permCacheKey(userID, orgID uuid.UUID) string {
	return permCachePrefix + userID.String() + ":" + orgID.String()
}

func ptrUUID(u uuid.UUID) *uuid.UUID {
	if u == uuid.Nil {
		return nil
	}
	id := u
	return &id
}
