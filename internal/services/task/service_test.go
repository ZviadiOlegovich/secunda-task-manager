package task

import (
	"context"
	"errors"
	"testing"

	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type mockRepo struct {
	createFn            func(ctx context.Context, task *Task) (int64, error)
	getByIDFn           func(ctx context.Context, id int64) (*Task, error)
	updateWithHistoryFn func(ctx context.Context, task *Task, history []TaskHistoryEntry) error
	listFn              func(ctx context.Context, filter ListFilter) ([]*Task, error)
	listHistoryFn       func(ctx context.Context, taskID int64) ([]HistoryRecord, error)
	createCommentFn     func(ctx context.Context, comment *Comment) (int64, error)
	listCommentsFn      func(ctx context.Context, taskID int64) ([]Comment, error)
}

func (m *mockRepo) Create(ctx context.Context, t *Task) (int64, error) {
	return m.createFn(ctx, t)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*Task, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) UpdateWithHistory(ctx context.Context, t *Task, history []TaskHistoryEntry) error {
	return m.updateWithHistoryFn(ctx, t, history)
}

func (m *mockRepo) List(ctx context.Context, f ListFilter) ([]*Task, error) {
	return m.listFn(ctx, f)
}

func (m *mockRepo) ListHistory(ctx context.Context, taskID int64) ([]HistoryRecord, error) {
	return m.listHistoryFn(ctx, taskID)
}

func (m *mockRepo) CreateComment(ctx context.Context, c *Comment) (int64, error) {
	if m.createCommentFn != nil {
		return m.createCommentFn(ctx, c)
	}
	return 1, nil
}

func (m *mockRepo) ListComments(ctx context.Context, taskID int64) ([]Comment, error) {
	if m.listCommentsFn != nil {
		return m.listCommentsFn(ctx, taskID)
	}
	return nil, nil
}

type mockTeamRepo struct {
	areMembersOfFn func(ctx context.Context, teamID int64, userIDs []int64) error
}

func (m *mockTeamRepo) AreMembersOf(ctx context.Context, teamID int64, userIDs []int64) error {
	return m.areMembersOfFn(ctx, teamID, userIDs)
}

var errDB = errors.New("db error")

var okRepo = &mockRepo{
	createFn: func(_ context.Context, _ *Task) (int64, error) { return 1, nil },
}

var memberTeamRepo = &mockTeamRepo{
	areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return nil },
}

var notMemberTeamRepo = &mockTeamRepo{
	areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return errs.ErrNotFound },
}

func TestService_ListTasks(t *testing.T) {
	status := StatusTodo
	tasks := []*Task{
		{ID: 1, TeamID: 1, Title: "A", Status: StatusTodo, Priority: PriorityMedium},
		{ID: 2, TeamID: 1, Title: "B", Status: StatusDone, Priority: PriorityHigh},
	}

	okListRepo := &mockRepo{
		listFn: func(_ context.Context, _ ListFilter) ([]*Task, error) { return tasks, nil },
	}

	tests := []struct {
		name     string
		input    ListFilter
		repo     Repository
		teamRepo TeamRepository
		wantLen  int
		wantErr  error
	}{
		{
			name:     "success",
			input:    ListFilter{TeamID: 1, RequestedBy: 1, Page: 1, Limit: 20},
			repo:     okListRepo,
			teamRepo: memberTeamRepo,
			wantLen:  2,
		},
		{
			name: "with status filter",
			input: ListFilter{TeamID: 1, RequestedBy: 1, Status: &status, Page: 1, Limit: 20},
			repo: &mockRepo{
				listFn: func(_ context.Context, f ListFilter) ([]*Task, error) {
					if f.Status == nil || *f.Status != StatusTodo {
						return nil, errors.New("unexpected filter")
					}
					return tasks[:1], nil
				},
			},
			teamRepo: memberTeamRepo,
			wantLen:  1,
		},
		{
			name:  "repo error propagated",
			input: ListFilter{TeamID: 1, RequestedBy: 1, Page: 1, Limit: 20},
			repo: &mockRepo{
				listFn: func(_ context.Context, _ ListFilter) ([]*Task, error) { return nil, errDB },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.teamRepo)
			got, err := svc.ListTasks(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && len(got) != tt.wantLen {
				t.Errorf("want %d tasks, got %d", tt.wantLen, len(got))
			}
		})
	}
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
			name: "default priority when empty",
			input: CreateTaskInput{TeamID: 1, Title: "Fix bug", CreatedBy: 1},
			repo: &mockRepo{
				createFn: func(_ context.Context, t *Task) (int64, error) {
					if t.Priority != PriorityMedium {
						return 0, errors.New("expected default priority medium")
					}
					return 1, nil
				},
			},
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
		{
			name:  "membership check repo error",
			input: CreateTaskInput{TeamID: 1, Title: "Fix bug", Priority: PriorityMedium, CreatedBy: 1},
			repo:  okRepo,
			teamRepo: &mockTeamRepo{
				areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return errDB },
			},
			wantErr: errDB,
		},
		{
			name:     "repo create error",
			input:    CreateTaskInput{TeamID: 1, Title: "Fix bug", Priority: PriorityMedium, CreatedBy: 1},
			repo:     &mockRepo{createFn: func(_ context.Context, _ *Task) (int64, error) { return 0, errDB }},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.teamRepo)
			id, err := svc.CreateTask(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && id <= 0 {
				t.Error("expected positive id, got 0")
			}
		})
	}
}

func TestService_GetTaskHistory(t *testing.T) {
	records := []HistoryRecord{
		{ID: 1, TaskID: 10, UserID: 1, Field: "title", OldValue: nil, NewValue: func(s string) *string { return &s }("new")},
	}

	tests := []struct {
		name    string
		taskID  int64
		repo    Repository
		wantLen int
		wantErr error
	}{
		{
			name:   "success",
			taskID: 10,
			repo: &mockRepo{
				listHistoryFn: func(_ context.Context, _ int64) ([]HistoryRecord, error) { return records, nil },
			},
			wantLen: 1,
		},
		{
			name:   "repo error propagated",
			taskID: 10,
			repo: &mockRepo{
				listHistoryFn: func(_ context.Context, _ int64) ([]HistoryRecord, error) { return nil, errDB },
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, memberTeamRepo)
			got, err := svc.GetTaskHistory(context.Background(), tt.taskID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && len(got) != tt.wantLen {
				t.Errorf("want %d records, got %d", tt.wantLen, len(got))
			}
		})
	}
}

func TestService_AddComment(t *testing.T) {
	taskWithTeam := &Task{ID: 10, TeamID: 5, Title: "task", Status: StatusTodo, Priority: PriorityMedium, CreatedBy: 1}

	okGetByID := func(_ context.Context, _ int64) (*Task, error) { return taskWithTeam, nil }

	tests := []struct {
		name     string
		input    CreateCommentInput
		repo     Repository
		teamRepo TeamRepository
		wantID   bool
		wantErr  error
	}{
		{
			name:  "success",
			input: CreateCommentInput{TaskID: 10, UserID: 1, Content: "looks good"},
			repo: &mockRepo{
				getByIDFn: okGetByID,
				createCommentFn: func(_ context.Context, c *Comment) (int64, error) {
					return 42, nil
				},
			},
			teamRepo: memberTeamRepo,
			wantID:   true,
		},
		{
			name:     "empty content",
			input:    CreateCommentInput{TaskID: 10, UserID: 1, Content: ""},
			repo:     &mockRepo{},
			teamRepo: memberTeamRepo,
			wantErr:  ErrEmptyComment,
		},
		{
			name:     "whitespace-only content",
			input:    CreateCommentInput{TaskID: 10, UserID: 1, Content: "   "},
			repo:     &mockRepo{},
			teamRepo: memberTeamRepo,
			wantErr:  ErrEmptyComment,
		},
		{
			name:  "task not found",
			input: CreateCommentInput{TaskID: 99, UserID: 1, Content: "hi"},
			repo: &mockRepo{
				getByIDFn: func(_ context.Context, _ int64) (*Task, error) { return nil, errs.ErrNotFound },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errs.ErrNotFound,
		},
		{
			name:  "get task repo error",
			input: CreateCommentInput{TaskID: 10, UserID: 1, Content: "hi"},
			repo: &mockRepo{
				getByIDFn: func(_ context.Context, _ int64) (*Task, error) { return nil, errDB },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
		{
			name:  "user not a member of task team",
			input: CreateCommentInput{TaskID: 10, UserID: 7, Content: "hi"},
			repo: &mockRepo{
				getByIDFn: okGetByID,
			},
			teamRepo: notMemberTeamRepo,
			wantErr:  ErrNotMember,
		},
		{
			name:  "membership check verifies task team_id",
			input: CreateCommentInput{TaskID: 10, UserID: 1, Content: "hi"},
			repo: &mockRepo{
				getByIDFn: okGetByID,
			},
			teamRepo: &mockTeamRepo{
				areMembersOfFn: func(_ context.Context, teamID int64, _ []int64) error {
					if teamID != taskWithTeam.TeamID {
						return errors.New("wrong team_id passed")
					}
					return nil
				},
			},
			wantID: true,
		},
		{
			name:  "membership check repo error",
			input: CreateCommentInput{TaskID: 10, UserID: 1, Content: "hi"},
			repo: &mockRepo{
				getByIDFn: okGetByID,
			},
			teamRepo: &mockTeamRepo{
				areMembersOfFn: func(_ context.Context, _ int64, _ []int64) error { return errDB },
			},
			wantErr: errDB,
		},
		{
			name:  "create comment repo error",
			input: CreateCommentInput{TaskID: 10, UserID: 1, Content: "hi"},
			repo: &mockRepo{
				getByIDFn:       okGetByID,
				createCommentFn: func(_ context.Context, _ *Comment) (int64, error) { return 0, errDB },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.teamRepo)
			id, err := svc.AddComment(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want err %v, got %v", tt.wantErr, err)
			}
			if tt.wantID && id <= 0 {
				t.Error("expected positive id, got 0")
			}
		})
	}
}


func taskForUpdate(createdBy int64) *Task {
	return &Task{ID: 10, TeamID: 1, Title: "old", Status: StatusTodo, Priority: PriorityMedium, CreatedBy: createdBy}
}

func updateRepo(task *Task) *mockRepo {
	return &mockRepo{
		getByIDFn:           func(_ context.Context, _ int64) (*Task, error) { return task, nil },
		updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error { return nil },
	}
}

func TestService_UpdateTask(t *testing.T) {
	assigneeID := int64(5)
	otherAssignee := int64(99)

	validInput := func(updatedBy int64) UpdateTaskInput {
		return UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: updatedBy, Title: "new title", Status: StatusTodo, Priority: PriorityMedium}
	}

	tests := []struct {
		name     string
		input    UpdateTaskInput
		repo     Repository
		teamRepo TeamRepository
		wantErr  error
	}{
		{
			name:     "success",
			input:    validInput(99),
			repo:     updateRepo(taskForUpdate(1)),
			teamRepo: memberTeamRepo,
		},
		{
			name:     "user not in team",
			input:    validInput(99),
			repo:     updateRepo(taskForUpdate(1)),
			teamRepo: notMemberTeamRepo,
			wantErr:  ErrNotMember,
		},
		{
			name: "task belongs to different team",
			input: UpdateTaskInput{TaskID: 10, TeamID: 2, UpdatedBy: 1, Title: "new", Status: StatusTodo, Priority: PriorityMedium},
			repo: &mockRepo{
				getByIDFn:           func(_ context.Context, _ int64) (*Task, error) { return taskForUpdate(1), nil },
				updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error { return nil },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errs.ErrNotFound,
		},
		{
			name:  "no-op update skips DB write",
			input: UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "old", Status: StatusTodo, Priority: PriorityMedium},
			repo: &mockRepo{
				getByIDFn: func(_ context.Context, _ int64) (*Task, error) { return taskForUpdate(1), nil },
				updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error {
					t.Error("UpdateWithHistory should not be called for no-op")
					return nil
				},
			},
			teamRepo: memberTeamRepo,
		},
		{
			name:     "status update",
			input:    UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "old", Status: StatusInProgress, Priority: PriorityMedium},
			repo:     updateRepo(taskForUpdate(1)),
			teamRepo: memberTeamRepo,
		},
		{
			name: "assignee changed to non-member returns ErrNotMember",
			input: UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "old", Status: StatusTodo, Priority: PriorityMedium, AssigneeID: &otherAssignee},
			repo: &mockRepo{
				getByIDFn:           func(_ context.Context, _ int64) (*Task, error) { return taskForUpdate(1), nil },
				updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error { return nil },
			},
			teamRepo: notMemberTeamRepo,
			wantErr:  ErrNotMember,
		},
		{
			name: "assignee unchanged — no extra membership check",
			input: UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "new title", Status: StatusTodo, Priority: PriorityMedium, AssigneeID: &assigneeID},
			repo: &mockRepo{
				getByIDFn: func(_ context.Context, _ int64) (*Task, error) {
					t := taskForUpdate(1)
					t.AssigneeID = &assigneeID
					return t, nil
				},
				updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error { return nil },
			},
			teamRepo: memberTeamRepo,
		},
		{
			name:     "blank title",
			input:    UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "  ", Status: StatusTodo, Priority: PriorityMedium},
			repo:     updateRepo(taskForUpdate(1)),
			teamRepo: memberTeamRepo,
			wantErr:  ErrInvalidTitle,
		},
		{
			name:     "invalid status",
			input:    UpdateTaskInput{TaskID: 10, TeamID: 1, UpdatedBy: 1, Title: "ok", Status: TaskStatus("unknown"), Priority: PriorityMedium},
			repo:     updateRepo(taskForUpdate(1)),
			teamRepo: memberTeamRepo,
			wantErr:  ErrInvalidStatus,
		},
		{
			name:  "get by id repo error",
			input: validInput(1),
			repo: &mockRepo{
				getByIDFn: func(_ context.Context, _ int64) (*Task, error) { return nil, errDB },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
		{
			name:  "update with history repo error",
			input: validInput(1),
			repo: &mockRepo{
				getByIDFn:           func(_ context.Context, _ int64) (*Task, error) { return taskForUpdate(1), nil },
				updateWithHistoryFn: func(_ context.Context, _ *Task, _ []TaskHistoryEntry) error { return errDB },
			},
			teamRepo: memberTeamRepo,
			wantErr:  errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, tt.teamRepo)
			err := svc.UpdateTask(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}
