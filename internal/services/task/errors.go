package task

import "errors"

var (
	ErrNotMember       = errors.New("user is not a member of this team")
	ErrInvalidTitle    = errors.New("task title is required")
	ErrInvalidStatus   = errors.New("invalid task status")
	ErrInvalidPriority = errors.New("invalid task priority")
	ErrInvalidEstimate = errors.New("invalid task estimate")
	ErrEmptyComment    = errors.New("comment content is required")
)
