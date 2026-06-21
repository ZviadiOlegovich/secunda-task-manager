package email

import (
	"context"

	"github.com/sony/gobreaker/v2"
)

const (
	cbMaxRequests        uint32 = 1
	cbFailureThreshold   uint32 = 5
)

type CBService struct {
	delegate Service
	cb       *gobreaker.CircuitBreaker[struct{}]
}

func NewCBService(delegate Service) *CBService {
	st := gobreaker.Settings{
		Name:        "email",
		MaxRequests: cbMaxRequests,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cbFailureThreshold
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
