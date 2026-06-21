//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

type noopTaskCache struct{}

func (noopTaskCache) GetVersion(_ context.Context, _ int64) (int64, error)                               { return 0, nil }
func (noopTaskCache) GetTaskList(_ context.Context, _, _ int64, _ string) ([]*task.Task, error)           { return nil, nil }
func (noopTaskCache) SetTaskListIfVersion(_ context.Context, _, _ int64, _ string, _ []*task.Task) error { return nil }
func (noopTaskCache) IncrVersion(_ context.Context, _ int64) error                                        { return nil }

type TaskRepoSuite struct {
	baseSuite
	repo task.Repository
}

func TestTaskRepo(t *testing.T) {
	suite.Run(t, new(TaskRepoSuite))
}

func (s *TaskRepoSuite) SetupSuite() {
	s.baseSuite.SetupSuite()
	s.repo = repository.NewTaskRepository(s.db, noopTaskCache{})
}

func (s *TaskRepoSuite) TestCreate() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, userID)

	desc := "some description"
	id, err := s.repo.Create(ctx, &task.Task{
		TeamID:      teamID,
		Title:       "First task",
		Description: &desc,
		Status:      task.StatusTodo,
		Priority:    task.PriorityMedium,
		CreatedBy:   userID,
	})
	s.Require().NoError(err)
	s.Greater(id, int64(0))
}

func (s *TaskRepoSuite) TestGetByID() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, userID)

	desc := "desc"
	id, err := s.repo.Create(ctx, &task.Task{
		TeamID:      teamID,
		Title:       "Get task",
		Description: &desc,
		Status:      task.StatusInProgress,
		Priority:    task.PriorityHigh,
		CreatedBy:   userID,
	})
	s.Require().NoError(err)

	got, err := s.repo.GetByID(ctx, id)
	s.Require().NoError(err)
	s.Equal("Get task", got.Title)
	s.Equal(task.StatusInProgress, got.Status)
	s.Require().NotNil(got.Description)
	s.Equal(desc, *got.Description)
}

func (s *TaskRepoSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(context.Background(), 999999)
	s.True(err == errs.ErrNotFound, "want ErrNotFound, got %v", err)
}

func (s *TaskRepoSuite) TestUpdate() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, userID)

	id, err := s.repo.Create(ctx, &task.Task{
		TeamID:    teamID,
		Title:     "Old title",
		Status:    task.StatusTodo,
		Priority:  task.PriorityLow,
		CreatedBy: userID,
	})
	s.Require().NoError(err)

	got, err := s.repo.GetByID(ctx, id)
	s.Require().NoError(err)

	got.Title = "New title"
	got.Status = task.StatusDone
	got.Priority = task.PriorityHigh

	history := []task.TaskHistoryEntry{
		{TaskID: id, UserID: userID, Field: "title", OldValue: strPtr("Old title"), NewValue: strPtr("New title")},
		{TaskID: id, UserID: userID, Field: "status", OldValue: strPtr("todo"), NewValue: strPtr("done")},
	}
	s.Require().NoError(s.repo.UpdateWithHistory(ctx, got, history))

	updated, err := s.repo.GetByID(ctx, id)
	s.Require().NoError(err)
	s.Equal("New title", updated.Title)
	s.Equal(task.StatusDone, updated.Status)
}

func (s *TaskRepoSuite) TestList() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, userID)
	otherTeamID := seedTeam(s.T(), s.db, userID)

	for i, st := range []task.TaskStatus{task.StatusTodo, task.StatusInProgress, task.StatusDone} {
		_, err := s.repo.Create(ctx, &task.Task{
			TeamID:    teamID,
			Title:     fmt.Sprintf("Task %d", i),
			Status:    st,
			Priority:  task.PriorityMedium,
			CreatedBy: userID,
		})
		s.Require().NoError(err)
	}
	_, err := s.repo.Create(ctx, &task.Task{
		TeamID: otherTeamID, Title: "Other team task",
		Status: task.StatusTodo, Priority: task.PriorityLow, CreatedBy: userID,
	})
	s.Require().NoError(err)

	s.Run("all tasks for team", func() {
		tasks, err := s.repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 1, Limit: 20})
		s.Require().NoError(err)
		s.Len(tasks, 3)
	})

	s.Run("filter by status", func() {
		st := task.StatusTodo
		tasks, err := s.repo.List(ctx, task.ListFilter{TeamID: teamID, Status: &st, Page: 1, Limit: 20})
		s.Require().NoError(err)
		s.Len(tasks, 1)
		s.Equal(task.StatusTodo, tasks[0].Status)
	})

	s.Run("pagination", func() {
		page1, err := s.repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 1, Limit: 2})
		s.Require().NoError(err)
		s.Len(page1, 2)

		page2, err := s.repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 2, Limit: 2})
		s.Require().NoError(err)
		s.Len(page2, 1)
	})
}

func strPtr(s string) *string { return &s }
