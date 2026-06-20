package stats

import "context"

type Repository interface {
	TeamStats(ctx context.Context) ([]TeamStat, error)
	TopUsers(ctx context.Context) ([]TopUser, error)
}
