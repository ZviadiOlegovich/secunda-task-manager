package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtpkg "github.com/zoshc/secunda-task-manager/pkg/jwt"
)

type TokenValidator interface {
	ValidateAccess(token string) (*jwtpkg.Claims, error)
}

const userIDKey = "userID"

func Auth(validator TokenValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(header, "Bearer ") {
			return fiber.ErrUnauthorized
		}

		claims, err := validator.ValidateAccess(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return fiber.ErrUnauthorized
		}

		c.Locals(userIDKey, claims.UserID)
		return c.Next()
	}
}

func UserIDFromCtx(c *fiber.Ctx) (int64, bool) {
	id, ok := c.Locals(userIDKey).(int64)
	return id, ok
}
