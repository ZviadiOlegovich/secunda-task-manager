package user

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   Repository
	tokens TokenProvider
}

func New(repo Repository, tokens TokenProvider) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, reg RegisterInput) error {
	logger := zerolog.Ctx(ctx)

	if err := reg.validate(); err != nil {
		logger.Warn().Err(err).Msg("invalid register input")
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("generate password hash")
		return err
	}

	if err = s.repo.Create(ctx, &User{
		Email:        reg.Email,
		PasswordHash: string(hash),
		Name:         reg.Name,
	}); err != nil {
		if errors.Is(err, ErrEmailTaken) {
			return err
		}
		logger.Error().Err(err).Msg("create user")
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, creds LoginInput) (*Tokens, error) {
	logger := zerolog.Ctx(ctx)

	if err := creds.validate(); err != nil {
		logger.Warn().Err(err).Msg("invalid login input")
		return nil, err
	}

	u, err := s.repo.GetByEmail(ctx, creds.Email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		logger.Error().Err(err).Msg("get user by email")
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(creds.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		logger.Error().Err(err).Msg("issue tokens")
		return nil, err
	}

	return tokens, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	logger := zerolog.Ctx(ctx)

	if _, err := s.tokens.ValidateRefresh(refreshToken); err != nil {
		return nil, ErrInvalidCredentials
	}

	u, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		logger.Error().Err(err).Msg("get user by refresh token")
		return nil, err
	}

	tokens, err := s.issueTokens(ctx, u.ID)
	if err != nil {
		logger.Error().Err(err).Msg("issue tokens")
		return nil, err
	}

	return tokens, nil
}

func (s *Service) issueTokens(ctx context.Context, userID int64) (*Tokens, error) {
	accessToken, err := s.tokens.GenerateAccess(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokens.GenerateRefresh(userID)
	if err != nil {
		return nil, err
	}

	if err = s.repo.UpdateRefreshToken(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	return &Tokens{Access: accessToken, Refresh: refreshToken}, nil
}
