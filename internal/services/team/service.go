package team

import (
	"context"
	"errors"
	"strings"

	"github.com/rs/zerolog"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

type InviteUserInput struct {
	TeamID    int64
	InviterID int64
	InviteeID int64
	Role      Role
}

func (s *Service) CreateTeam(ctx context.Context, userID int64, name string) (*Team, error) {
	logger := zerolog.Ctx(ctx)

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	t := &Team{
		Name:      name,
		CreatedBy: userID,
	}

	id, err := s.repo.CreateWithOwner(ctx, t, userID)
	if err != nil {
		logger.Error().Err(err).Msg("create team with owner")
		return nil, err
	}

	t.ID = id
	return t, nil
}

func (s *Service) InviteUser(ctx context.Context, invite InviteUserInput) error {
	logger := zerolog.Ctx(ctx)

	inviter, err := s.repo.GetMember(ctx, invite.TeamID, invite.InviterID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return ErrPermissionDenied
		}
		logger.Error().Err(err).Msg("get inviter member")
		return err
	}

	if !canInvite(inviter.Role, invite.Role) {
		return ErrPermissionDenied
	}

	if err := s.repo.AddMember(ctx, &TeamMember{
		TeamID: invite.TeamID,
		UserID: invite.InviteeID,
		Role:   invite.Role,
	}); err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			return err
		}
		logger.Error().Err(err).Msg("add member")
		return err
	}

	return nil
}

func (s *Service) ListTeams(ctx context.Context, userID int64) ([]*Team, error) {
	teams, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("list teams")
		return nil, err
	}
	return teams, nil
}

func canInvite(inviterRole, targetRole Role) bool {
	switch inviterRole {
	case RoleOwner:
		return true
	case RoleAdmin:
		return targetRole == RoleMember
	default:
		return false
	}
}
