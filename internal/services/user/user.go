package user

import (
	"net/mail"
	"strings"
	"time"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
	RefreshToken *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Tokens struct {
	Access  string
	Refresh string
}

type LoginInput struct {
	Email    string
	Password string
}

func (i *LoginInput) validate() error {
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))
	if _, err := mail.ParseAddress(i.Email); err != nil {
		return ErrInvalidEmail
	}
	if i.Password == "" {
		return ErrInvalidCredentials
	}
	return nil
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

func (i *RegisterInput) validate() error {
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))

	if _, err := mail.ParseAddress(i.Email); err != nil {
		return ErrInvalidEmail
	}
	// TODO: add stronger password validation (complexity, unicode length)
	if len(i.Password) < 8 {
		return ErrWeakPassword
	}
	if strings.TrimSpace(i.Name) == "" {
		return ErrInvalidName
	}
	return nil
}
