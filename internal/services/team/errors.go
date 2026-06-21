package team

import "errors"

var (
	ErrAlreadyMember    = errors.New("user is already a member of this team")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidName      = errors.New("team name is required")
	ErrInviteeNotFound  = errors.New("invitee user not found")
)
