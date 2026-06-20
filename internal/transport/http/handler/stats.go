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
	TopUsers(ctx context.Context) ([]stats.TopUser, error)
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

type TopUserResponse struct {
	TeamID    int64  `json:"team_id"`
	TeamName  string `json:"team_name"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	TaskCount int    `json:"task_count"`
	Rank      int    `json:"rank"`
}

func NewStatsHandler(auth fiber.Handler, svc StatsService) *statsHandler {
	h := &statsHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		g := r.Group("/stats", auth)
		g.Get("/teams", h.teams)
		g.Get("/top-users", h.topUsers)
	}
	return h
}

func (h *statsHandler) Router() router.MakeRouter { return h.makeRouter }

func (h *statsHandler) topUsers(c *fiber.Ctx) error {
	result, err := h.svc.TopUsers(c.UserContext())
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]TopUserResponse, len(result))
	for i, u := range result {
		resp[i] = TopUserResponse{
			TeamID:    u.TeamID,
			TeamName:  u.TeamName,
			UserID:    u.UserID,
			UserName:  u.UserName,
			TaskCount: u.TaskCount,
			Rank:      u.Rank,
		}
	}
	return c.JSON(resp)
}

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
