package task

import "context"

type ListFilter struct {
	TeamID     int64
	Status     *TaskStatus
	Priority   *TaskPriority
	AssigneeID *int64
	Page       int
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, task *Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	Update(ctx context.Context, task *Task) error
	List(ctx context.Context, filter ListFilter) ([]*Task, error)
}
