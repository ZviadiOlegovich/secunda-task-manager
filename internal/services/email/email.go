package email

import "context"

type Service interface {
	SendInvite(ctx context.Context, to, teamName string) error
}
