package email

import (
	"context"

	"github.com/rs/zerolog"
)

type MockService struct{}

func NewMock() *MockService {
	return &MockService{}
}

func (m *MockService) SendInvite(ctx context.Context, to, teamName string) error {
	zerolog.Ctx(ctx).Info().
		Str("to", to).
		Str("team", teamName).
		Msg("mock: send invite email")
	return nil
}
