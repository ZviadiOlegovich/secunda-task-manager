package task

import (
	"context"
	"errors"
	"testing"

	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type mockRepo struct {
	createFn  func(ctx context.Context, task *Task) (int64, error)
	getByIDFn func(ctx context.Context, id int64) (*Task, error)
	updateFn  func(ctx context.Context, task *Task) error
	listFn    func(ctx context.Context, filter ListFilter) ([]*Task, error)
}

func (m *mockRepo) Create(ctx context.Context, t *Task) (int64, error) {
	return m.createFn(ctx, t)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*Task, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) Update(ctx context.Context, t *Task) error {
	return m.updateFn(ctx, t)
}

func (m *mockRepo) List(ctx context.Context, f ListFilter) ([]*Task, error) {
	return m.listFn(ctx, f)
}

type mockTeamRepo struct {
	areMembersOfFn func(ctx context.Context, teamID int64, userIDs []int64) error
}

func (m *mockTeamRepo) AreMembersOf(ctx context.Context, teamID int64, userIDs []int64) error {
	return m.areMembersOfFn(ctx, teamID, userIDs)
}

var okRepo = &mockRepo{
	createFn: func(_ context.Context, _ *Task) (int64, error) { return 1, nil },
}

var memberTeamRepo = &mockTeamRepo{
	areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return nil },
}

var notMemberTeamRepo = &mockTeamRepo{
	areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return errs.ErrNotFound },
}

func TestService_CreateTask(t *testing.T) {
	assigneeID := int64(2)

	tests := []struct {
		name     string
		input    CreateTaskInput
		repo     Repository
		teamRepo TeamRepository
		wantErr  error
	}{
		{
			name:     "success",
			input:    CreateTaskInput{TeamID: 1, Title: "Fix bug", Priority: PriorityMedium, CreatedBy: 1},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
		},
		{
			name: "success with assignee",
			input: CreateTaskInput{
				TeamID: 1, Title: "Fix bug", Priority: PriorityHigh,
				CreatedBy: 1, AssigneeID: &assigneeID,
			},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
		},
		{
			name:     "default priority when empty",
			input:    CreateTaskInput{TeamID: 1, Title: "Fix bug", CreatedBy: 1},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
		},
		{
			name:     "empty title",
			input:    CreateTaskInput{TeamID: 1, Priority: PriorityMedium, CreatedBy: 1},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
			wantErr:  ErrInvalidTitle,
		},
		{
			name:     "blank title",
			input:    CreateTaskInput{TeamID: 1, Title: "   ", Priority: PriorityMedium, CreatedBy: 1},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
			wantErr:  ErrInvalidTitle,
		},
		{
			name:     "invalid priority",
			input:    CreateTaskInput{TeamID: 1, Title: "Fix bug", Priority: "urgent", CreatedBy: 1},
			repo:     okRepo,
			teamRepo: memberTeamRepo,
			wantErr:  ErrInvalidPriority,
		},
		{
			name:     "creator not in team",
			input:    CreateTaskInput{TeamID: 1, Title: "Fix bug", Priority: PriorityMedium, CreatedBy: 1},
			repo:     okRepo,
			teamRepo: notMemberTeamRepo,
			wantErr:  ErrNotMember,
		},
		{
			name: "assignee not in team",
			input: CreateTaskInput{
				TeamID: 1, Title: "Fix bug", Priority: PriorityMedium,
				CreatedBy: 1, AssigneeID: &assigneeID,
			},
			repo:     okRepo,
			teamRepo: notMemberTeamRepo,
			wantErr:  ErrNotMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.teamRepo)
			task, err := svc.CreateTask(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && task == nil {
				t.Error("expected task, got nil")
			}
		})
	}
}
