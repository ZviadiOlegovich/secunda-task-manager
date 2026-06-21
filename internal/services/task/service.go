package task

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type Service struct {
	repo     Repository
	teamRepo TeamRepository
}

func New(repo Repository, teamRepo TeamRepository) *Service {
	return &Service{repo: repo, teamRepo: teamRepo}
}

func (s *Service) CreateTask(ctx context.Context, create CreateTaskInput) (int64, error) {
	logger := zerolog.Ctx(ctx)

	create.applyDefaults()
	if err := create.validate(); err != nil {
		return 0, err
	}

	if err := s.teamRepo.AreMembersOf(ctx, create.TeamID, create.participantIDs()); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return 0, ErrNotMember
		}
		logger.Error().Err(err).Msg("check membership")
		return 0, err
	}

	t := &Task{
		TeamID:      create.TeamID,
		Title:       create.Title,
		Description: create.Description,
		Status:      StatusTodo,
		Priority:    create.Priority,
		Estimate:    create.Estimate,
		AssigneeID:  create.AssigneeID,
		CreatedBy:   create.CreatedBy,
		DueDate:     create.DueDate,
	}

	id, err := s.repo.Create(ctx, t)
	if err != nil {
		logger.Error().Err(err).Msg("create task")
		return 0, err
	}
	return id, nil
}

func (s *Service) UpdateTask(ctx context.Context, update UpdateTaskInput) error {
	logger := zerolog.Ctx(ctx)

	update.applyDefaults()
	if err := update.validate(); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(ctx, update.TaskID)
	if err != nil {
		return err
	}
	if existing.TeamID != update.TeamID {
		return errs.ErrNotFound
	}

	if err := s.teamRepo.AreMembersOf(ctx, update.TeamID, update.participantIDs(existing.AssigneeID)); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return ErrNotMember
		}
		logger.Error().Err(err).Msg("check membership")
		return err
	}

	updated, history := applyUpdate(existing, update)
	if len(history) == 0 {
		return nil
	}

	if err := s.repo.UpdateWithHistory(ctx, updated, history); err != nil {
		logger.Error().Err(err).Msg("update task")
		return err
	}
	return nil
}

func (s *Service) GetTaskHistory(ctx context.Context, taskID, teamID int64) ([]HistoryRecord, error) {
	logger := zerolog.Ctx(ctx)

	t, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t.TeamID != teamID {
		return nil, errs.ErrNotFound
	}

	records, err := s.repo.ListHistory(ctx, taskID)
	if err != nil {
		logger.Error().Err(err).Msg("list history")
		return nil, err
	}
	return records, nil
}

func (s *Service) ListTasks(ctx context.Context, filter ListFilter) ([]*Task, error) {
	logger := zerolog.Ctx(ctx)

	if err := filter.validate(); err != nil {
		return nil, err
	}

	tasks, err := s.repo.List(ctx, filter)
	if err != nil {
		logger.Error().Err(err).Msg("list tasks")
		return nil, err
	}
	return tasks, nil
}
