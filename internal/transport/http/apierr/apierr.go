package apierr

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
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
	case errors.Is(err, errs.ErrNotFound),
		errors.Is(err, team.ErrInviteeNotFound):
		return http.StatusNotFound
	case errors.Is(err, user.ErrEmailTaken),
		errors.Is(err, team.ErrAlreadyMember):
		return http.StatusConflict
	case errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword),
		errors.Is(err, user.ErrInvalidName),
		errors.Is(err, team.ErrInvalidName):
		return http.StatusBadRequest
	case errors.Is(err, user.ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, team.ErrPermissionDenied),
		errors.Is(err, task.ErrNotMember):
		return http.StatusForbidden
	case errors.Is(err, task.ErrInvalidTitle),
		errors.Is(err, task.ErrInvalidStatus),
		errors.Is(err, task.ErrInvalidPriority),
		errors.Is(err, task.ErrInvalidEstimate):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
