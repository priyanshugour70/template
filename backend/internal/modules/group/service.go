package group

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	apperr "github.com/your-org/your-service/internal/pkg/errors"
)

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

func (s *Service) Create(ctx context.Context, tenantID, orgID uuid.UUID, in CreateInput) (*Group, error) {
	if in.Slug = strings.TrimSpace(in.Slug); in.Slug == "" {
		return nil, apperr.New(apperr.CodeValidation, "slug is required", nil)
	}
	if in.Name = strings.TrimSpace(in.Name); in.Name == "" {
		return nil, apperr.New(apperr.CodeValidation, "name is required", nil)
	}
	kind := strings.TrimSpace(in.Kind)
	if kind == "" {
		kind = "custom"
	}
	g := &Group{
		Slug: in.Slug, Name: in.Name, Description: in.Description, Kind: kind,
		Color: in.Color, Icon: in.Icon,
	}
	g.TenantID = tenantID
	g.OrganizationID = &orgID
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, apperr.New(apperr.CodeInternal, "create group failed", err)
	}
	return g, nil
}

func (s *Service) Update(ctx context.Context, orgID, id uuid.UUID, in UpdateInput) (*Group, error) {
	g, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	patch := map[string]any{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.Description != nil {
		patch["description"] = *in.Description
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
	if len(patch) > 0 {
		if err := s.repo.Update(ctx, id, patch); err != nil {
			return nil, apperr.New(apperr.CodeInternal, "update group failed", err)
		}
	}
	g, _ = s.repo.Get(ctx, orgID, id)
	return g, nil
}

func (s *Service) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	s.bustForGroup(ctx, id)
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperr.New(apperr.CodeInternal, "delete group failed", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Group, int64, error) {
	rows, total, err := s.repo.List(ctx, orgID, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list groups failed", err)
	}
	return rows, total, nil
}

func (s *Service) AddMember(ctx context.Context, orgID, id uuid.UUID, in AddMemberInput, by *uuid.UUID) error {
	if (in.UserID == nil) == (in.GroupID == nil) {
		return apperr.New(apperr.CodeValidation, "exactly one of userId or groupId must be set", nil)
	}
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	var err error
	if in.UserID != nil {
		err = s.repo.AddUser(ctx, id, *in.UserID, by)
	} else {
		// Verify sub-group is in the same org.
		if _, gerr := s.repo.Get(ctx, orgID, *in.GroupID); gerr != nil {
			return apperr.New(apperr.CodeValidation, "sub-group not found in this org", nil)
		}
		err = s.repo.AddSubGroup(ctx, id, *in.GroupID, by)
	}
	if err != nil {
		return apperr.New(apperr.CodeInternal, "add member failed", err)
	}
	s.bustForGroup(ctx, id)
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, orgID, groupID, memberID uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, groupID); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	if err := s.repo.RemoveMember(ctx, memberID); err != nil {
		return apperr.New(apperr.CodeInternal, "remove member failed", err)
	}
	s.bustForGroup(ctx, groupID)
	return nil
}

func (s *Service) ListMembers(ctx context.Context, orgID, id uuid.UUID, limit, offset int) ([]Member, int64, error) {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return nil, 0, apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return nil, 0, apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	out, total, err := s.repo.ListMembers(ctx, id, limit, offset)
	if err != nil {
		return nil, 0, apperr.New(apperr.CodeInternal, "list members failed", err)
	}
	return out, total, nil
}

func (s *Service) AssignRoles(ctx context.Context, orgID, id uuid.UUID, roleIDs []uuid.UUID, by *uuid.UUID) error {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	if err := s.repo.ReplaceRoles(ctx, id, roleIDs, by); err != nil {
		return apperr.New(apperr.CodeInternal, "assign roles failed", err)
	}
	s.bustForGroup(ctx, id)
	return nil
}

func (s *Service) ListRoles(ctx context.Context, orgID, id uuid.UUID) ([]uuid.UUID, error) {
	if _, err := s.repo.Get(ctx, orgID, id); err != nil {
		if IsNotFound(err) {
			return nil, apperr.New(apperr.CodeNotFound, "group not found", nil)
		}
		return nil, apperr.New(apperr.CodeInternal, "load group failed", err)
	}
	out, err := s.repo.ListRoles(ctx, id)
	if err != nil {
		return nil, apperr.New(apperr.CodeInternal, "list group roles failed", err)
	}
	return out, nil
}

func (s *Service) bustForGroup(ctx context.Context, groupID uuid.UUID) {
	if s.bust == nil {
		return
	}
	ids, err := s.repo.MembershipIDsAffectedByGroup(ctx, groupID)
	if err != nil {
		s.log.Warn("bustForGroup: query failed", zap.Error(err))
		return
	}
	for _, mid := range ids {
		s.bust.InvalidateMembership(ctx, mid)
	}
}
