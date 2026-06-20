package user

import "errors"

var (
	ErrEmailTaken        = errors.New("email already taken")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrWeakPassword      = errors.New("password must be at least 8 characters")
	ErrInvalidName       = errors.New("name is required")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
