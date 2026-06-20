package user

import "time"

type User struct {
	ID           uint64
	Email        string
	PasswordHash string
	Name         string
	RefreshToken *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
