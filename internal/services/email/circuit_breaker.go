package email

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/sony/gobreaker/v2"
)

const (
	cbMaxRequests      uint32        = 1
	cbFailureThreshold uint32        = 5
	cbInterval         time.Duration = 60 * time.Second
)

type CBService struct {
	delegate Service
	cb       *gobreaker.CircuitBreaker[struct{}]
}

func NewCBService(delegate Service, log zerolog.Logger) *CBService {
	st := gobreaker.Settings{
		Name:        "email",
		MaxRequests: cbMaxRequests,
		Interval:    cbInterval,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cbFailureThreshold
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Warn().
				Str("circuit_breaker", name).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("email circuit breaker state changed")
		},
	}
	return &CBService{
		delegate: delegate,
		cb:       gobreaker.NewCircuitBreaker[struct{}](st),
	}
}

func (s *CBService) SendInvite(ctx context.Context, to, teamName string) error {
	_, err := s.cb.Execute(func() (struct{}, error) {
		return struct{}{}, s.delegate.SendInvite(ctx, to, teamName)
	})
	return err
}
