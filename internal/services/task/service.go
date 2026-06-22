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
	logger := zerolog.Ctx(ctx).With().Int64("user_id", create.CreatedBy).Logger()

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
	logger := zerolog.Ctx(ctx).With().Int64("user_id", update.UpdatedBy).Logger()

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

func (s *Service) GetTaskHistory(ctx context.Context, taskID int64) ([]HistoryRecord, error) {
	logger := zerolog.Ctx(ctx)

	records, err := s.repo.ListHistory(ctx, taskID)
	if err != nil {
		logger.Error().Err(err).Msg("list history")
		return nil, err
	}
	return records, nil
}

func (s *Service) AddComment(ctx context.Context, input CreateCommentInput) (int64, error) {
	logger := zerolog.Ctx(ctx).With().Int64("user_id", input.UserID).Logger()

	if err := input.validate(); err != nil {
		return 0, err
	}

	t, err := s.repo.GetByID(ctx, input.TaskID)
	if err != nil {
		logger.Error().Err(err).Msg("get task for comment")
		return 0, err
	}

	if err := s.teamRepo.AreMembersOf(ctx, t.TeamID, []int64{input.UserID}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return 0, ErrNotMember
		}
		logger.Error().Err(err).Msg("check membership for comment")
		return 0, err
	}

	id, err := s.repo.CreateComment(ctx, &Comment{
		TaskID:  input.TaskID,
		UserID:  input.UserID,
		Content: input.Content,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create comment")
		return 0, err
	}
	return id, nil
}

func (s *Service) ListComments(ctx context.Context, taskID int64) ([]Comment, error) {
	logger := zerolog.Ctx(ctx)

	comments, err := s.repo.ListComments(ctx, taskID)
	if err != nil {
		logger.Error().Err(err).Msg("list comments")
		return nil, err
	}
	return comments, nil
}

func (s *Service) ListTasks(ctx context.Context, filter ListFilter) ([]*Task, error) {
	logger := zerolog.Ctx(ctx).With().Int64("user_id", filter.RequestedBy).Logger()

	filter.normalize()
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
