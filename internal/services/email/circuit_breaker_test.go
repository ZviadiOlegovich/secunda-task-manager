package email

import (
	"context"
	"errors"
	"testing"

	"github.com/sony/gobreaker/v2"
)

type failService struct{ err error }

func (f *failService) SendInvite(_ context.Context, _, _ string) error { return f.err }

func TestCBService_OpensAfterConsecutiveFailures(t *testing.T) {
	svc := NewCBService(&failService{err: errors.New("smtp error")})

	for range cbFailureThreshold {
		_ = svc.SendInvite(context.Background(), "user@example.com", "Alpha")
	}

	err := svc.SendInvite(context.Background(), "user@example.com", "Alpha")
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Errorf("want ErrOpenState, got %v", err)
	}
}

func TestCBService_PassesThroughOnSuccess(t *testing.T) {
	svc := NewCBService(&MockService{})

	if err := svc.SendInvite(context.Background(), "user@example.com", "Alpha"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
