package team

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type Service struct {
	repo     Repository
	emailSvc EmailService
}

func New(repo Repository, emailSvc EmailService) *Service {
	return &Service{repo: repo, emailSvc: emailSvc}
}

type InviteUserInput struct {
	TeamID       int64
	InviterID    int64
	InviteeID    int64
	InviteeEmail string
	Role         Role
}

func (s *Service) CreateTeam(ctx context.Context, userID int64, name string) (int64, error) {
	logger := zerolog.Ctx(ctx)

	name = strings.TrimSpace(name)
	if name == "" {
		return 0, ErrInvalidName
	}

	id, err := s.repo.CreateWithOwner(ctx, &Team{Name: name, CreatedBy: userID}, userID)
	if err != nil {
		logger.Error().Err(err).Msg("create team with owner")
		return 0, err
	}
	return id, nil
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

	t, err := s.repo.GetByID(ctx, invite.TeamID)
	if err != nil {
		logger.Error().Err(err).Msg("get team")
		return err
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

	// Требования не описывают поведение email-сервиса и гарантии доставки.
	// Fire-and-forget: если письмо не отправится — приглашение всё равно создано.
	// Для гарантированной доставки нужен transactional outbox + брокер (например, Kafka).
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.emailSvc.SendInvite(ctx, invite.InviteeEmail, t.Name); err != nil {
			logger.Error().Err(err).Str("to", invite.InviteeEmail).Msg("send invite email")
		}
	}()

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
