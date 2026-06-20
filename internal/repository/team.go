package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
)

type teamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *teamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateWithOwner(ctx context.Context, t *team.Team, ownerID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `INSERT INTO teams (name, created_by) VALUES (?, ?)`, t.Name, t.CreatedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`,
		id, ownerID, team.RoleOwner,
	)
	if err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

func (r *teamRepository) AddMember(ctx context.Context, m *team.TeamMember) error {
	const q = `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, m.TeamID, m.UserID, m.Role)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return team.ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (r *teamRepository) GetByID(ctx context.Context, id int64) (*team.Team, error) {
	const q = `SELECT id, name, created_by, created_at FROM teams WHERE id = ?`
	t := &team.Team{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *teamRepository) GetByUserID(ctx context.Context, userID int64) ([]*team.Team, error) {
	const q = `
		SELECT t.id, t.name, t.created_by, t.created_at
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = ?`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*team.Team
	for rows.Next() {
		t := &team.Team{}
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *teamRepository) GetMember(ctx context.Context, teamID, userID int64) (*team.TeamMember, error) {
	const q = `SELECT team_id, user_id, role, created_at FROM team_members WHERE team_id = ? AND user_id = ?`
	m := &team.TeamMember{}
	err := r.db.QueryRowContext(ctx, q, teamID, userID).Scan(&m.TeamID, &m.UserID, &m.Role, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return m, nil
}

func (r *teamRepository) AreMembersOf(ctx context.Context, teamID int64, userIDs []int64) error {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",")
	q := fmt.Sprintf(`SELECT COUNT(*) FROM team_members WHERE team_id = ? AND user_id IN (%s)`, placeholders)

	args := make([]any, 0, len(userIDs)+1)
	args = append(args, teamID)
	for _, id := range userIDs {
		args = append(args, id)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return err
	}
	if count < len(userIDs) {
		return errs.ErrNotFound
	}
	return nil
}
