package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	const q = `INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q, u.Email, u.PasswordHash, u.Name)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return user.ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, name, refresh_token, created_at, updated_at
		FROM users
		WHERE email = ?`
	u := &user.User{}
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name,
		&u.RefreshToken, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) GetByRefreshToken(ctx context.Context, token string) (*user.User, error) {
	const q = `
		SELECT id, email, password_hash, name, refresh_token, created_at, updated_at
		FROM users
		WHERE refresh_token = ?`
	u := &user.User{}
	err := r.db.QueryRowContext(ctx, q, token).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name,
		&u.RefreshToken, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepository) UpdateRefreshToken(ctx context.Context, userID uint64, token string) error {
	const q = `UPDATE users SET refresh_token = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, q, token, userID)
	return err
}
