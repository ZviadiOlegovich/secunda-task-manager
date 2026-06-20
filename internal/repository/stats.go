package repository

import (
	"context"
	"database/sql"

	"github.com/zoshc/secunda-task-manager/internal/services/stats"
)

type statsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *statsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) TeamStats(ctx context.Context) ([]stats.TeamStat, error) {
	const q = `
		SELECT
			t.id,
			t.name,
			COALESCE(m.member_count, 0)   AS member_count,
			COALESCE(d.done_last_week, 0) AS done_last_week
		FROM teams t
		LEFT JOIN (
			SELECT team_id, COUNT(*) AS member_count
			FROM team_members
			GROUP BY team_id
		) m ON m.team_id = t.id
		LEFT JOIN (
			SELECT team_id, COUNT(*) AS done_last_week
			FROM tasks
			WHERE status = 'done'
			  AND updated_at >= NOW() - INTERVAL 7 DAY
			GROUP BY team_id
		) d ON d.team_id = t.id
		ORDER BY t.id`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []stats.TeamStat
	for rows.Next() {
		var s stats.TeamStat
		if err := rows.Scan(&s.TeamID, &s.TeamName, &s.MemberCount, &s.DoneLastWeek); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *statsRepository) TopUsers(ctx context.Context) ([]stats.TopUser, error) {
	const q = `
		WITH ranked AS (
			SELECT
				t.id              AS team_id,
				t.name            AS team_name,
				u.id              AS user_id,
				u.name            AS user_name,
				COUNT(ta.id)      AS task_count,
				-- RANK даёт одинаковый ранг при ничьей, поэтому результат может содержать >3 строк на команду.
				-- Если нужно ровно 3 — заменить на ROW_NUMBER() (но тогда ничья разрешается произвольно).
				RANK() OVER (PARTITION BY t.id ORDER BY COUNT(ta.id) DESC) AS rn
			FROM teams t
			JOIN team_members tm ON tm.team_id = t.id
			JOIN users u         ON u.id = tm.user_id
			LEFT JOIN tasks ta   ON ta.team_id = t.id
			                    AND ta.created_by = u.id
			                    AND ta.created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')
			GROUP BY t.id, t.name, u.id, u.name
		)
		SELECT team_id, team_name, user_id, user_name, task_count, rn
		FROM ranked
		WHERE rn <= 3
		ORDER BY team_id, rn`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []stats.TopUser
	for rows.Next() {
		var u stats.TopUser
		if err := rows.Scan(&u.TeamID, &u.TeamName, &u.UserID, &u.UserName, &u.TaskCount, &u.Rank); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
