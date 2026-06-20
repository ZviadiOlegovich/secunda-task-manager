package user

import (
	"context"
	"errors"
	"testing"

	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"golang.org/x/crypto/bcrypt"
)

type mockRepo struct {
	createFn              func(ctx context.Context, u *User) error
	getByEmailFn          func(ctx context.Context, email string) (*User, error)
	getByRefreshTokenFn   func(ctx context.Context, token string) (*User, error)
	updateRefreshTokenFn  func(ctx context.Context, userID int64, token string) error
}

func (m *mockRepo) Create(ctx context.Context, u *User) error {
	return m.createFn(ctx, u)
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockRepo) GetByRefreshToken(ctx context.Context, token string) (*User, error) {
	return m.getByRefreshTokenFn(ctx, token)
}

func (m *mockRepo) UpdateRefreshToken(ctx context.Context, userID int64, token string) error {
	return m.updateRefreshTokenFn(ctx, userID, token)
}

type mockTokens struct {
	generateAccessFn   func(userID int64) (string, error)
	generateRefreshFn  func(userID int64) (string, error)
	validateRefreshFn  func(token string) (int64, error)
}

func (m *mockTokens) GenerateAccess(userID int64) (string, error) {
	return m.generateAccessFn(userID)
}

func (m *mockTokens) GenerateRefresh(userID int64) (string, error) {
	return m.generateRefreshFn(userID)
}

func (m *mockTokens) ValidateRefresh(token string) (int64, error) {
	return m.validateRefreshFn(token)
}

var (
	okTokens = &mockTokens{
		generateAccessFn:  func(_ int64) (string, error) { return "access-token", nil },
		generateRefreshFn: func(_ int64) (string, error) { return "refresh-token", nil },
		validateRefreshFn: func(_ string) (int64, error) { return 1, nil },
	}
)

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
			svc := New(tt.repo, okTokens)
			err := svc.Register(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	existingUser := &User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hash),
	}

	okRepo := &mockRepo{
		getByEmailFn:         func(_ context.Context, _ string) (*User, error) { return existingUser, nil },
		updateRefreshTokenFn: func(_ context.Context, _ int64, _ string) error { return nil },
	}

	tests := []struct {
		name    string
		input   LoginInput
		repo    Repository
		tokens  TokenProvider
		wantErr error
	}{
		{
			name:   "success",
			input:  LoginInput{Email: "test@example.com", Password: "password123"},
			repo:   okRepo,
			tokens: okTokens,
		},
		{
			name:    "invalid email",
			input:   LoginInput{Email: "not-an-email", Password: "password123"},
			repo:    okRepo,
			tokens:  okTokens,
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "user not found",
			input:   LoginInput{Email: "missing@example.com", Password: "password123"},
			repo:    &mockRepo{getByEmailFn: func(_ context.Context, _ string) (*User, error) { return nil, errs.ErrNotFound }},
			tokens:  okTokens,
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "wrong password",
			input:   LoginInput{Email: "test@example.com", Password: "wrongpass"},
			repo:    okRepo,
			tokens:  okTokens,
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.tokens)
			tokens, err := svc.Login(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tokens == nil {
				t.Error("expected tokens, got nil")
			}
		})
	}
}
