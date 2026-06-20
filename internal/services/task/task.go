package task

import "time"

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

type Task struct {
	ID          int64
	TeamID      int64
	Title       string
	Description *string
	Status      TaskStatus
	Priority    TaskPriority
	AssigneeID  *int64
	CreatedBy   int64
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
