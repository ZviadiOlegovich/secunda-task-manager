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
			COUNT(DISTINCT tm.user_id)                                                                               AS member_count,
			COUNT(DISTINCT CASE WHEN ta.status = 'done' AND ta.updated_at >= NOW() - INTERVAL 7 DAY THEN ta.id END) AS done_last_week
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks        ta ON ta.team_id = t.id
		GROUP BY t.id, t.name
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
