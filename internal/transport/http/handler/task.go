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
	CreateTask(ctx context.Context, create task.CreateTaskInput) (int64, error)
	ListTasks(ctx context.Context, filter task.ListFilter) ([]*task.Task, error)
	UpdateTask(ctx context.Context, update task.UpdateTaskInput) error
	GetTaskHistory(ctx context.Context, taskID int64) ([]task.HistoryRecord, error)
	AddComment(ctx context.Context, input task.CreateCommentInput) (int64, error)
	ListComments(ctx context.Context, taskID int64) ([]task.Comment, error)
}

type taskHandler struct {
	makeRouter router.MakeRouter
	svc        TaskService
}

func NewTaskHandler(auth, userRateLimit fiber.Handler, svc TaskService) *taskHandler {
	h := &taskHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		g := r.Group("/tasks", auth, userRateLimit)
		g.Post("/", h.create)
		g.Get("/", h.list)
		g.Put("/:id", h.update)
		g.Get("/:id/history", h.history)
		g.Post("/:id/comments", h.addComment)
		g.Get("/:id/comments", h.listComments)
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

	id, err := h.svc.CreateTask(c.UserContext(), task.CreateTaskInput{
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

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
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

func (h *taskHandler) history(c *fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	records, err := h.svc.GetTaskHistory(c.UserContext(), taskID)
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]model.HistoryRecordResponse, len(records))
	for i, r := range records {
		resp[i] = model.ToHistoryRecordResponse(r)
	}
	return c.JSON(resp)
}

func (h *taskHandler) addComment(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	var req model.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	id, err := h.svc.AddComment(c.UserContext(), task.CreateCommentInput{
		TaskID:  taskID,
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		return apierr.Response(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *taskHandler) listComments(c *fiber.Ctx) error {
	taskID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid task id"})
	}

	comments, err := h.svc.ListComments(c.UserContext(), taskID)
	if err != nil {
		return apierr.Response(c, err)
	}

	resp := make([]model.CommentResponse, len(comments))
	for i, cm := range comments {
		resp[i] = model.ToCommentResponse(cm)
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
