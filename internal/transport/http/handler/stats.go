package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/services/stats"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/apierr"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
)

type StatsService interface {
	TeamStats(ctx context.Context) ([]stats.TeamStat, error)
}

type TeamStatResponse struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	MemberCount  int    `json:"member_count"`
	DoneLastWeek int    `json:"done_last_week"`
}

type statsHandler struct {
	makeRouter router.MakeRouter
	svc        StatsService
}

func NewStatsHandler(auth fiber.Handler, svc StatsService) *statsHandler {
	h := &statsHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		g := r.Group("/stats", auth)
		g.Get("/teams", h.teams)
	}
	return h
}

func (h *statsHandler) Router() router.MakeRouter { return h.makeRouter }

func (h *statsHandler) teams(c *fiber.Ctx) error {
	result, err := h.svc.TeamStats(c.UserContext())
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]TeamStatResponse, len(result))
	for i, s := range result {
		resp[i] = TeamStatResponse{
			TeamID:       s.TeamID,
			TeamName:     s.TeamName,
			MemberCount:  s.MemberCount,
			DoneLastWeek: s.DoneLastWeek,
		}
	}
	return c.JSON(resp)
}
