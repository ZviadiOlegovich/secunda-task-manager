package task

import "errors"

var (
	ErrNotMember        = errors.New("user is not a member of this team")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidTitle     = errors.New("task title is required")
	ErrInvalidStatus    = errors.New("invalid task status")
	ErrInvalidPriority  = errors.New("invalid task priority")
)
