package task

import (
	"strings"
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

func (p TaskPriority) isValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}

type TaskEstimate string

const (
	EstimateXS TaskEstimate = "XS"
	EstimateS  TaskEstimate = "S"
	EstimateM  TaskEstimate = "M"
	EstimateL  TaskEstimate = "L"
	EstimateXL TaskEstimate = "XL"
)

func (e TaskEstimate) isValid() bool {
	switch e {
	case EstimateXS, EstimateS, EstimateM, EstimateL, EstimateXL:
		return true
	}
	return false
}

type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description *string
	Status      TaskStatus
	Priority    TaskPriority
	Estimate    *TaskEstimate
	AssigneeID  *int64
	CreatedBy   int64
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateTaskInput struct {
	TeamID      int64
	Title       string
	Description *string
	Priority    TaskPriority
	Estimate    *TaskEstimate
	AssigneeID  *int64
	CreatedBy   int64
	DueDate     *time.Time
}

func (i *CreateTaskInput) applyDefaults() {
	i.Title = strings.TrimSpace(i.Title)
	if i.Priority == "" {
		i.Priority = PriorityMedium
	}
}

func (i *CreateTaskInput) validate() error {
	if i.Title == "" {
		return ErrInvalidTitle
	}
	if !i.Priority.isValid() {
		return ErrInvalidPriority
	}
	if i.Estimate != nil && !i.Estimate.isValid() {
		return ErrInvalidEstimate
	}
	return nil
}

func (i *CreateTaskInput) participantIDs() []int64 {
	ids := []int64{i.CreatedBy}
	if i.AssigneeID != nil && *i.AssigneeID != i.CreatedBy {
		ids = append(ids, *i.AssigneeID)
	}
	return ids
}

type ListFilter struct {
	TeamID      int64
	RequestedBy int64
	Status      *TaskStatus
	Priority    *TaskPriority
	AssigneeID  *int64
	Page        int
	Limit       int
}
