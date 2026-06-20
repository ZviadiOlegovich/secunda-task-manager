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
	const q = `INSERT INTO tasks (team_id, title, description, status, priority, assignee_id, created_by, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, q,
		t.TeamID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.CreatedBy, t.DueDate,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *taskRepository) GetByID(ctx context.Context, id int64) (*task.Task, error) {
	const q = `SELECT id, team_id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = ?`
	t := &task.Task{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority,
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

func (r *taskRepository) Update(ctx context.Context, t *task.Task) error {
	const q = `UPDATE tasks SET title = ?, description = ?, status = ?, priority = ?, assignee_id = ?, due_date = ?
		WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q,
		t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.DueDate, t.ID,
	)
	return err
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
		`SELECT id, team_id, title, description, status, priority, assignee_id, created_by, due_date, created_at, updated_at
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
			&t.ID, &t.TeamID, &t.Title, &t.Description, &t.Status, &t.Priority,
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
