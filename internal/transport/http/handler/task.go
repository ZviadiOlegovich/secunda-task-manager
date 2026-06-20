package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/apierr"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/handler/model"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/middleware"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
)

type TaskService interface {
	CreateTask(ctx context.Context, create task.CreateTaskInput) (*task.Task, error)
	ListTasks(ctx context.Context, filter task.ListFilter) ([]*task.Task, error)
	UpdateTask(ctx context.Context, update task.UpdateTaskInput) error
}

type taskHandler struct {
	makeRouter router.MakeRouter
	svc        TaskService
}

func NewTaskHandler(auth fiber.Handler, svc TaskService) *taskHandler {
	h := &taskHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		g := r.Group("/tasks", auth)
		g.Post("/", h.create)
		g.Get("/", h.list)
		g.Put("/:id", h.update)
	}
	return h
}

func (h *taskHandler) Router() router.MakeRouter { return h.makeRouter }

func (h *taskHandler) create(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req model.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	t, err := h.svc.CreateTask(c.UserContext(), task.CreateTaskInput{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Estimate:    req.Estimate,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   userID,
		DueDate:     req.DueDate,
	})
	if err != nil {
		return apierr.Response(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(model.ToTaskResponse(t))
}

func (h *taskHandler) list(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	teamID, err := strconv.ParseInt(c.Query("team_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "team_id is required"})
	}

	filter := task.ListFilter{
		TeamID:      teamID,
		RequestedBy: userID,
		Page:        c.QueryInt("page", 1),
		Limit:       c.QueryInt("limit", 20),
	}

	if s := c.Query("status"); s != "" {
		status := task.TaskStatus(s)
		filter.Status = &status
	}
	if p := c.Query("priority"); p != "" {
		priority := task.TaskPriority(p)
		filter.Priority = &priority
	}
	if a := c.Query("assignee_id"); a != "" {
		id, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid assignee_id"})
		}
		filter.AssigneeID = &id
	}

	tasks, err := h.svc.ListTasks(c.UserContext(), filter)
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]model.TaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = model.ToTaskResponse(t)
	}
	return c.JSON(resp)
}

func (h *taskHandler) update(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	var req model.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.UpdateTask(c.UserContext(), task.UpdateTaskInput{
		TaskID:      taskID,
		TeamID:      req.TeamID,
		UpdatedBy:   userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		Estimate:    req.Estimate,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
	}); err != nil {
		return apierr.Response(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
