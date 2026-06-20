package team

import "context"

type Repository interface {
	Create(ctx context.Context, team *Team) (int64, error)
	AddMember(ctx context.Context, member *TeamMember) error
	GetByID(ctx context.Context, id int64) (*Team, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Team, error)
	GetMember(ctx context.Context, teamID, userID int64) (*TeamMember, error)
}
