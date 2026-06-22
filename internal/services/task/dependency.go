package task

import "context"

type Repository interface {
	Create(ctx context.Context, task *Task) (int64, error)
	GetByID(ctx context.Context, id int64) (*Task, error)
	UpdateWithHistory(ctx context.Context, task *Task, history []TaskHistoryEntry) error
	List(ctx context.Context, filter ListFilter) ([]*Task, error)
	ListHistory(ctx context.Context, taskID int64) ([]HistoryRecord, error)
	CreateComment(ctx context.Context, comment *Comment) (int64, error)
	ListComments(ctx context.Context, taskID int64) ([]Comment, error)
}

type TeamRepository interface {
	AreMembersOf(ctx context.Context, teamID int64, userIDs []int64) error
}
