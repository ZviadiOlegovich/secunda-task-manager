package user

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByRefreshToken(ctx context.Context, token string) (*User, error)
	UpdateRefreshToken(ctx context.Context, userID uint64, token string) error
}

type TokenProvider interface {
	GenerateAccess(userID uint64) (string, error)
	GenerateRefresh(userID uint64) (string, error)
	ValidateRefresh(token string) (uint64, error)
}
