package user

import (
	"errors"
	"testing"
)

func TestRegisterInput_Validate(t *testing.T) {
	tests := []struct {
		name      string
		input     RegisterInput
		wantErr   error
		wantEmail string
	}{
		{
			name:      "valid",
			input:     RegisterInput{Email: "test@example.com", Password: "password123", Name: "Test User"},
			wantEmail: "test@example.com",
		},
		{
			name:      "email normalized to lowercase",
			input:     RegisterInput{Email: "TEST@EXAMPLE.COM", Password: "password123", Name: "Test User"},
			wantEmail: "test@example.com",
		},
		{
			name:    "invalid email",
			input:   RegisterInput{Email: "not-an-email", Password: "password123", Name: "Test User"},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "empty email",
			input:   RegisterInput{Email: "", Password: "password123", Name: "Test User"},
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "weak password",
			input:   RegisterInput{Email: "test@example.com", Password: "short", Name: "Test User"},
			wantErr: ErrWeakPassword,
		},
		{
			name:    "empty name",
			input:   RegisterInput{Email: "test@example.com", Password: "password123", Name: ""},
			wantErr: ErrInvalidName,
		},
		{
			name:    "blank name",
			input:   RegisterInput{Email: "test@example.com", Password: "password123", Name: "   "},
			wantErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && tt.input.Email != tt.wantEmail {
				t.Errorf("expected normalized email %q, got %q", tt.wantEmail, tt.input.Email)
			}
		})
	}
}
