package model

import (
	"time"

	"github.com/zoshc/secunda-task-manager/internal/services/team"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	UserID int64     `json:"user_id"`
	Email  string    `json:"email"`
	Role   team.Role `json:"role"`
}

type TeamResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func ToTeamResponse(t *team.Team) TeamResponse {
	return TeamResponse{
		ID:        t.ID,
		Name:      t.Name,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
	}
}
