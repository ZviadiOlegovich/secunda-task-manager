package team

import "context"

type Repository interface {
	CreateWithOwner(ctx context.Context, team *Team, ownerID int64) (int64, error)
	AddMember(ctx context.Context, member *TeamMember) error
	GetByID(ctx context.Context, id int64) (*Team, error)
	GetByUserID(ctx context.Context, userID int64) ([]*Team, error)
	GetMember(ctx context.Context, teamID, userID int64) (*TeamMember, error)
}

type EmailService interface {
	SendInvite(ctx context.Context, to, teamName string) error
}
