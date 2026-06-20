package team

import "errors"

var (
	ErrAlreadyMember  = errors.New("user is already a member of this team")
	ErrPermissionDenied = errors.New("permission denied")
)
