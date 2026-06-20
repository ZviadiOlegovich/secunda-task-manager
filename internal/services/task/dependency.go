package task

import "context"

type Repository interface {
	Create(ctx context.Context, task *Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	Update(ctx context.Context, task *Task) error
	List(ctx context.Context, filter ListFilter) ([]*Task, error)
}

type TeamRepository interface {
	AreMembersOf(ctx context.Context, teamID int64, userIDs []int64) error
}
