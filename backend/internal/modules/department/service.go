package department

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
)

// CacheBuster is the narrow interface this module needs from rbac.Service —
// we inject just the invalidation method to avoid a circular import.
type CacheBuster interface {
	InvalidateMembership(ctx context.Context, membershipID uuid.UUID)
}

type Service struct {
	repo *Repository
	bust CacheBuster
	log  *zap.Logger
}

func NewService(repo *Repository, bust CacheBuster, log *zap.Logger) *Service {
	return &Service{repo: repo, bust: bust, log: log}
}

func (s *Service) Create(ctx context.Context, tenantID, orgID uuid.UUID, in CreateInput) (*Department, error) {
	if in.Slug = strings.TrimSpace(in.Slug); in.Slug == "" {
		return nil, apperr.New(apperr.CodeValidation, "slug is required", nil)
	}
	if in.Name = strings.TrimSpace(in.Name); in.Name == "" {
		return nil, apperr.New(apperr.CodeValidation, "name is required", nil)
	}
	if in.ParentID != nil {
		if _, err := s.repo.Get(ctx, orgID, *in.ParentID); err != nil {
			if IsNotFound(err) {
				return nil, apperr.New(apperr.CodeValidation, "parent department not found in this org", nil)
			}
			return nil, apperr.New(apperr.CodeInternal, "load parent failed", err)
		}
	}
	d := &Department{
		ParentID:      in.ParentID,
		Slug:          in.Slug,
		Name:          in.Name,
		Description:   in.Description,
		CostCenter:    in.CostCenter,
		ManagerUserID: in.ManagerUserID,
		Color:         in.Color,
		Icon:          in.Icon,
		SortOrder:     in.SortOrder,
	}
	d.TenantID = tenantID
	d.OrganizationID = &orgID
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create department failed", err)
	}
	return d, nil
}

func (s *Service) Update(ctx context.Context, orgID, id uuid.UUID, in UpdateInput) (*Department, error) {
	d, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load department failed", err)
	}
	patch := map[string]any{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.Description != nil {
		patch["description"] = *in.Description
	}
	if in.CostCenter != nil {
		patch["cost_center"] = *in.CostCenter
	}
	if in.ManagerUserID != nil {
		patch["manager_user_id"] = *in.ManagerUserID
	}
	if in.Color != nil {
		patch["color"] = *in.Color
	}
	if in.Icon != nil {
		patch["icon"] = *in.Icon
	}
	if in.IsArchived != nil {
		patch["is_archived"] = *in.IsArchived
	}
	if in.SortOrder != nil {
		patch["sort_order"] = *in.SortOrder
	}
	if len(patch) > 0 {
		if err := s.repo.Update(ctx, id, patch); err != nil {
			return nil, apperr.New(apperr.CodeInternal, "update department failed", err)
		}
	}
	d, _ = s.repo.Get(ctx, orgID, id)
	return d, nil
}

// Move reparents — guards against cycles via closure lookup.
func (s *Service) Move(ctx context.Context, orgID, id uuid.UUID, parentID *uuid.UUID) error {
	if parentID != nil {
		if *parentID == id {
			return apperr.New(apperr.CodeValidation, "cannot parent a department to itself", nil)
		}
		// Cycle check: is `id` an ancestor of `parentID`?
		isAncestor, err := s.repo.IsAncestor(ctx, id, *parentID)
		if err != nil {
			return apperr.New(apperr.CodeInternal, "cycle check failed", err)
		}
		if isAncestor {
			return apperr.New(apperr.CodeValidation, "cannot move department under its own descendant", nil)
		}
		if _, err := s.repo.Get(ctx, orgID, *parentID); err != nil {
			return apperr.New(apperr.CodeValidation, "new parent not found in this org", nil)
		}
	}
	if err := s.repo.Reparent(ctx, id, parentID); err != nil {
		return apperr.New(apperr.CodeInternal, "reparent failed", err)
	}
	s.bustForDept(ctx, id)
	return nil
}

func (s *Service) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "department not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load department failed", err)
	}
	s.bustForDept(ctx, id)
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "delete department failed", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Department, int64, error) {
	rows, total, err := s.repo.ListByOrg(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list departments failed", err)
	}
	return rows, total, nil
}

// Tree builds a nested tree from a flat list. Roots are entries with no
// parent or a parent that's not visible in this org.
func (s *Service) Tree(ctx context.Context, orgID uuid.UUID) ([]*Node, error) {
	// Tree is a non-paginated client of the repo — we want every node
	// regardless of caller's `?limit=` because the UI needs the full hierarchy.
	rows, _, err := s.repo.ListByOrg(ctx, orgID, 0, 0)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list departments failed", err)
	}
	byID := make(map[uuid.UUID]*Node, len(rows))
	for i := range rows {
		byID[rows[i].ID] = &Node{Department: rows[i]}
	}
	roots := []*Node{}
	for _, n := range byID {
		if n.ParentID != nil {
			if parent, ok := byID[*n.ParentID]; ok {
				parent.Children = append(parent.Children, n)
				continue
			}
		}
		roots = append(roots, n)
	}
	return roots, nil
}

func (s *Service) AssignRoles(ctx context.Context, orgID, id uuid.UUID, roleIDs []uuid.UUID, by *uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "department not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load department failed", err)
	}
	if err := s.repo.ReplaceRoles(ctx, id, roleIDs, by); err != nil {
		return apperr.New(apperr.CodeInternal, "assign roles failed", err)
	}
	s.bustForDept(ctx, id)
	return nil
}

func (s *Service) ListRoles(ctx context.Context, orgID, id uuid.UUID) ([]uuid.UUID, error) {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "department not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load department failed", err)
	}
	out, err := s.repo.ListRoles(ctx, id)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list department roles failed", err)
	}
	return out, nil
}

// bustForDept invalidates the permission cache for every membership in the
// department's subtree.
func (s *Service) bustForDept(ctx context.Context, deptID uuid.UUID) {
	if s.bust == nil {
		return
	}
	ids, err := s.repo.MembershipIDsAffectedByDept(ctx, deptID)
	if err != nil {
		s.log.Warn("bustForDept: query failed", zap.Error(err))
		return
	}
	for _, mid := range ids {
		s.bust.InvalidateMembership(ctx, mid)
	}
}
