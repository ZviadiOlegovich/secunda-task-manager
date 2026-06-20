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

func (s *Service) CreateTask(ctx context.Context, create CreateTaskInput) (*Task, error) {
	logger := zerolog.Ctx(ctx)

	create.applyDefaults()
	if err := create.validate(); err != nil {
		return nil, err
	}

	if err := s.teamRepo.AreMembersOf(ctx, create.TeamID, create.participantIDs()); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, ErrNotMember
		}
		logger.Error().Err(err).Msg("check membership")
		return nil, err
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
		return nil, err
	}

	t.ID = id
	return t, nil
}
