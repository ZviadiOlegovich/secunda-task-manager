package model

import (
	"time"

	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

type CreateTaskRequest struct {
	TeamID      int64              `json:"team_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Priority    task.TaskPriority  `json:"priority"`
	Estimate    *task.TaskEstimate `json:"estimate"`
	AssigneeID  *int64             `json:"assignee_id"`
	DueDate     *time.Time         `json:"due_date"`
}

type UpdateTaskRequest struct {
	TeamID      int64              `json:"team_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Status      task.TaskStatus    `json:"status"`
	Priority    task.TaskPriority  `json:"priority"`
	Estimate    *task.TaskEstimate `json:"estimate"`
	AssigneeID  *int64             `json:"assignee_id"`
	DueDate     *time.Time         `json:"due_date"`
}

type TaskResponse struct {
	ID          int64              `json:"id"`
	TeamID      int64              `json:"team_id"`
	Title       string             `json:"title"`
	Description *string            `json:"description"`
	Status      task.TaskStatus    `json:"status"`
	Priority    task.TaskPriority  `json:"priority"`
	Estimate    *task.TaskEstimate `json:"estimate"`
	AssigneeID  *int64             `json:"assignee_id"`
	CreatedBy   int64              `json:"created_by"`
	DueDate     *time.Time         `json:"due_date"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func ToTaskResponse(t *task.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Estimate:    t.Estimate,
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
