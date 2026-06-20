package user

import (
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	createFn     func(ctx context.Context, u *User) error
	getByEmailFn func(ctx context.Context, email string) (*User, error)
}

func (m *mockRepo) Create(ctx context.Context, u *User) error {
	return m.createFn(ctx, u)
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	return m.getByEmailFn(ctx, email)
}

func TestService_Register(t *testing.T) {
	validInput := RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	tests := []struct {
		name    string
		input   RegisterInput
		repo    Repository
		wantErr error
	}{
		{
			name:  "success",
			input: validInput,
			repo:  &mockRepo{createFn: func(_ context.Context, _ *User) error { return nil }},
		},
		{
			name:    "email taken",
			input:   validInput,
			repo:    &mockRepo{createFn: func(_ context.Context, _ *User) error { return ErrEmailTaken }},
			wantErr: ErrEmailTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo)
			err := svc.Register(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
