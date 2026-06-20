package team

import "time"

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

type TeamMember struct {
	TeamID    int64
	UserID    int64
	Role      Role
	CreatedAt time.Time
}
