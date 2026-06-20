package apierr

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
)

type errResponse struct {
	Error string `json:"error"`
}

func Response(c *fiber.Ctx, err error) error {
	code := statusCode(err)
	msg := err.Error()
	if code == http.StatusInternalServerError {
		msg = "internal server error"
	}
	return c.Status(code).JSON(errResponse{Error: msg})
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, user.ErrEmailTaken):
		return http.StatusConflict
	case errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword),
		errors.Is(err, user.ErrInvalidName):
		return http.StatusBadRequest
	case errors.Is(err, user.ErrInvalidCredentials):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
