package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

const taskCacheTTL = 5 * time.Minute

type TaskCache struct {
	rdb *Client
}

func NewTaskCache(rdb *Client) *TaskCache {
	return &TaskCache{rdb: rdb}
}

func (c *TaskCache) versionKey(teamID int64) string {
	return fmt.Sprintf("tasks:team:%d:version", teamID)
}

func (c *TaskCache) listKey(teamID, version int64, filterKey string) string {
	return fmt.Sprintf("tasks:team:%d:v%d:%s", teamID, version, filterKey)
}

func (c *TaskCache) GetVersion(ctx context.Context, teamID int64) (int64, error) {
	ver, err := c.rdb.Get(ctx, c.versionKey(teamID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return ver, err
}

func (c *TaskCache) GetTaskList(ctx context.Context, teamID, ver int64, filterKey string) ([]*task.Task, error) {
	data, err := c.rdb.Get(ctx, c.listKey(teamID, ver, filterKey)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *TaskCache) SetTaskListIfVersion(ctx context.Context, teamID, ver int64, filterKey string, tasks []*task.Task) error {
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, c.listKey(teamID, ver, filterKey), data, taskCacheTTL).Err()
}

func (c *TaskCache) IncrVersion(ctx context.Context, teamID int64) error {
	if err := c.rdb.Incr(ctx, c.versionKey(teamID)).Err(); err != nil {
		return c.rdb.Set(ctx, c.versionKey(teamID), 0, 0).Err()
	}
	return nil
}
