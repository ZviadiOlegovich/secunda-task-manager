package user

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	logger := zerolog.Ctx(ctx)

	if err := input.validate(); err != nil {
		logger.Warn().Err(err).Msg("invalid register input")
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error().Err(err).Msg("generate password hash")
		return err
	}

	if err = s.repo.Create(ctx, &User{
		Email:        input.Email,
		PasswordHash: string(hash),
		Name:         input.Name,
	}); err != nil {
		if errors.Is(err, ErrEmailTaken) {
			return err
		}
		logger.Error().Err(err).Msg("create user")
		return err
	}

	return nil
}
