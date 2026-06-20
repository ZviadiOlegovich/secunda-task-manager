package stats

import (
	"context"

	"github.com/rs/zerolog"
)

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TeamStats(ctx context.Context) ([]TeamStat, error) {
	stats, err := s.repo.TeamStats(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("team stats")
		return nil, err
	}
	return stats, nil
}

func (s *Service) TopUsers(ctx context.Context) ([]TopUser, error) {
	users, err := s.repo.TopUsers(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("top users")
		return nil, err
	}
	return users, nil
}
