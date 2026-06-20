package user

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByRefreshToken(ctx context.Context, token string) (*User, error)
	UpdateRefreshToken(ctx context.Context, userID int64, token string) error
}

type TokenProvider interface {
	GenerateAccess(userID int64) (string, error)
	GenerateRefresh(userID int64) (string, error)
	ValidateRefresh(token string) (int64, error)
}
