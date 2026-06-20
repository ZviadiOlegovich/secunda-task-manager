package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	jwtpkg "github.com/zoshc/secunda-task-manager/pkg/jwt"
)

type mockValidator struct {
	fn func(token string) (*jwtpkg.Claims, error)
}

func (m *mockValidator) ValidateAccess(token string) (*jwtpkg.Claims, error) {
	return m.fn(token)
}

func newTestApp(validator TokenValidator) *fiber.App {
	app := fiber.New()
	app.Use(Auth(validator))
	app.Get("/", func(c *fiber.Ctx) error {
		id, ok := UserIDFromCtx(c)
		if !ok {
			return fiber.ErrInternalServerError
		}
		return c.JSON(fiber.Map{"user_id": id})
	})
	return app
}

func TestAuthMiddleware(t *testing.T) {
	okValidator := &mockValidator{
		fn: func(_ string) (*jwtpkg.Claims, error) {
			return &jwtpkg.Claims{UserID: 42}, nil
		},
	}
	errValidator := &mockValidator{
		fn: func(_ string) (*jwtpkg.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}

	tests := []struct {
		name       string
		header     string
		validator  TokenValidator
		wantStatus int
	}{
		{
			name:       "valid token",
			header:     "Bearer valid-token",
			validator:  okValidator,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing header",
			header:     "",
			validator:  okValidator,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no Bearer prefix",
			header:     "Token valid-token",
			validator:  okValidator,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			header:     "Bearer bad-token",
			validator:  errValidator,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set(fiber.HeaderAuthorization, tt.header)
			}

			resp, err := newTestApp(tt.validator).Test(req)
			if err != nil {
				t.Fatalf("test request: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}
