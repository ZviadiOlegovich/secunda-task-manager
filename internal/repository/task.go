package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

type taskCache interface {
	GetVersion(ctx context.Context, teamID int64) (int64, error)
	GetTaskList(ctx context.Context, teamID, ver int64, filterKey string) ([]*task.Task, error)
	SetTaskListIfVersion(ctx context.Context, teamID, ver int64, filterKey string, tasks []*task.Task) error
	IncrVersion(ctx context.Context, teamID int64) error
}

type taskRepository struct {
	db    *sql.DB
	cache taskCache
}

func NewTaskRepository(db *sql.DB, cache taskCache) *taskRepository {
	return &taskRepository{db: db, cache: cache}
}

func (r *taskRepository) Create(ctx context.Context, t *task.Task) (int64, error) {
	const q = `INSERT INTO tasks (team_id, title, description, status, priority, estimate, assignee_id, created_by, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, q,
		t.TeamID, t.Title, t.Description, t.Status, t.Priority, t.Estimate, t.AssigneeID, t.CreatedBy, t.DueDate,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := r.cache.IncrVersion(ctx, t.TeamID); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("cache incr version")
	}
	return id, nil
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (*task.Task, error) {
	const q = `SELECT id, team_id, title, description, status, priority, estimate, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = ?`
	t := &task.Task{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Estimate,
		&t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *taskRepository) UpdateWithHistory(ctx context.Context, t *task.Task, history []task.TaskHistoryEntry) error {
	if len(history) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE tasks SET title = ?, description = ?, status = ?, priority = ?, estimate = ?, assignee_id = ?, due_date = ? WHERE id = ? AND team_id = ?`,
		t.Title, t.Description, t.Status, t.Priority, t.Estimate, t.AssigneeID, t.DueDate, t.ID, t.TeamID,
	)
	if err != nil {
		return err
	}

	var b strings.Builder
	b.Grow(12 * len(history))
	args := make([]any, 0, len(history)*5)
	for i, h := range history {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString("(?,?,?,?,?)")
		args = append(args, h.TaskID, h.UserID, h.Field, h.OldValue, h.NewValue)
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO task_history (task_id, user_id, field, old_value, new_value) VALUES `+b.String(),
		args...,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	if err := r.cache.IncrVersion(ctx, t.TeamID); err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("cache incr version")
	}
	return nil
}

func (r *taskRepository) List(ctx context.Context, filter task.ListFilter) ([]*task.Task, error) {
	logger := zerolog.Ctx(ctx)
	key := filter.FilterKey()

	ver, cacheErr := r.cache.GetVersion(ctx, filter.TeamID)
	if cacheErr != nil {
		logger.Warn().Err(cacheErr).Msg("cache get version")
	} else {
		if cached, err := r.cache.GetTaskList(ctx, filter.TeamID, ver, key); err != nil {
			logger.Warn().Err(err).Msg("cache get task list")
		} else if cached != nil {
			return cached, nil
		}
	}

	wb := newWhereBuilder(4)
	wb.add("team_id = ?", filter.TeamID)
	if filter.Status != nil {
		wb.add("status = ?", *filter.Status)
	}
	if filter.Priority != nil {
		wb.add("priority = ?", *filter.Priority)
	}
	if filter.AssigneeID != nil {
		wb.add("assignee_id = ?", *filter.AssigneeID)
	}

	offset := (filter.Page - 1) * filter.Limit

	q := fmt.Sprintf(
		`SELECT id, team_id, title, description, status, priority, estimate, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		wb.clause(),
	)

	rows, err := r.db.QueryContext(ctx, q, append(wb.args, filter.Limit, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*task.Task
	for rows.Next() {
		t := &task.Task{}
		if err := rows.Scan(
			&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Estimate,
			&t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if cacheErr == nil {
		if werr := r.cache.SetTaskListIfVersion(ctx, filter.TeamID, ver, key, tasks); werr != nil {
			logger.Warn().Err(werr).Msg("cache set task list")
		}
	}

	return tasks, nil
}

func (r *taskRepository) ListHistory(ctx context.Context, taskID int64) ([]task.HistoryRecord, error) {
	const q = `SELECT h.id, h.task_id, h.user_id, u.name, h.field, h.old_value, h.new_value, h.created_at
		FROM task_history h
		JOIN users u ON u.id = h.user_id
		WHERE h.task_id = ? ORDER BY h.created_at ASC`

	rows, err := r.db.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []task.HistoryRecord
	for rows.Next() {
		var h task.HistoryRecord
		if err := rows.Scan(&h.ID, &h.TaskID, &h.UserID, &h.UserName, &h.Field, &h.OldValue, &h.NewValue, &h.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, h)
	}
	return records, rows.Err()
}
