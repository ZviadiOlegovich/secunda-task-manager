package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *taskRepository {
	return &taskRepository{db: db}
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
	return result.LastInsertId()
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

	return tx.Commit()
}

func (r *taskRepository) List(ctx context.Context, filter task.ListFilter) ([]*task.Task, error) {
	wb := newWhereBuilder(4) // 1 обязательный + 3 опциональных
	wb.add("team_id = ?", filter.TeamID)
	if filter.Status != nil     { wb.add("status = ?", *filter.Status) }
	if filter.Priority != nil   { wb.add("priority = ?", *filter.Priority) }
	if filter.AssigneeID != nil { wb.add("assignee_id = ?", *filter.AssigneeID) }

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	q := fmt.Sprintf(
		`SELECT id, team_id, title, description, status, priority, estimate, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		wb.clause(),
	)

	rows, err := r.db.QueryContext(ctx, q, append(wb.args, limit, offset)...)
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
	return tasks, rows.Err()
}

type whereBuilder struct {
	clauses []string
	args    []any
}

func newWhereBuilder(cap int) *whereBuilder {
	return &whereBuilder{
		clauses: make([]string, 0, cap),
		args:    make([]any, 0, cap),
	}
}

func (b *whereBuilder) add(clause string, arg any) {
	b.clauses = append(b.clauses, clause)
	b.args = append(b.args, arg)
}

func (b *whereBuilder) clause() string {
	return strings.Join(b.clauses, " AND ")
}
