package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/apierr"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/handler/model"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/middleware"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
)

type TeamService interface {
	CreateTeam(ctx context.Context, userID int64, name string) (int64, error)
	ListTeams(ctx context.Context, userID int64) ([]*team.Team, error)
	InviteUser(ctx context.Context, invite team.InviteUserInput) error
}

type teamHandler struct {
	makeRouter router.MakeRouter
	svc        TeamService
}

func NewTeamHandler(auth fiber.Handler, svc TeamService) *teamHandler {
	h := &teamHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		g := r.Group("/teams", auth)
		g.Post("/", h.create)
		g.Get("/", h.list)
		g.Post("/:id/invite", h.invite)
	}
	return h
}

func (h *teamHandler) Router() router.MakeRouter { return h.makeRouter }

func (h *teamHandler) create(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req model.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	id, err := h.svc.CreateTeam(c.UserContext(), userID, req.Name)
	if err != nil {
		return apierr.Response(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *teamHandler) list(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	teams, err := h.svc.ListTeams(c.UserContext(), userID)
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]model.TeamResponse, len(teams))
	for i, t := range teams {
		resp[i] = model.ToTeamResponse(t)
	}
	return c.JSON(resp)
}

func (h *teamHandler) invite(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	teamID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid team id"})
	}

	var req model.InviteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.InviteUser(c.UserContext(), team.InviteUserInput{
		TeamID:       teamID,
		InviterID:    userID,
		InviteeID:    req.UserID,
		InviteeEmail: req.Email,
		Role:         req.Role,
	}); err != nil {
		return apierr.Response(c, err)
	}

	return c.SendStatus(fiber.StatusOK)
}
