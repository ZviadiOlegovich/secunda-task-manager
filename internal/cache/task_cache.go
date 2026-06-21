package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

const (
	taskCacheTTL = 5 * time.Minute
	taskLockTTL  = 10 * time.Second
)

type TaskCache struct {
	rdb *redis.Client
}

func NewTaskCache(rdb *redis.Client) *TaskCache {
	return &TaskCache{rdb: rdb}
}

func (c *TaskCache) versionKey(teamID int64) string {
	return fmt.Sprintf("tasks:team:%d:version", teamID)
}

func (c *TaskCache) listKey(teamID, version int64, filterKey string) string {
	return fmt.Sprintf("tasks:team:%d:v%d:%s", teamID, version, filterKey)
}

func (c *TaskCache) lockKey(teamID int64, filterKey string) string {
	return fmt.Sprintf("tasks:lock:%d:%s", teamID, filterKey)
}

func (c *TaskCache) version(ctx context.Context, teamID int64) (int64, error) {
	v, err := c.rdb.Get(ctx, c.versionKey(teamID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}

// GetTaskList возвращает список задач из кеша. Возвращает nil, nil при cache miss.
func (c *TaskCache) GetTaskList(ctx context.Context, teamID int64, filterKey string) ([]*task.Task, error) {
	v, err := c.version(ctx, teamID)
	if err != nil {
		return nil, err
	}

	data, err := c.rdb.Get(ctx, c.listKey(teamID, v, filterKey)).Bytes()
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

// SetTaskList сохраняет список задач в кеш с защитой от cache stampede через Redis lock.
func (c *TaskCache) SetTaskList(ctx context.Context, teamID int64, filterKey string, tasks []*task.Task) error {
	lockKey := c.lockKey(teamID, filterKey)

	// Пробуем захватить lock — SET NX
	acquired, err := c.rdb.SetNX(ctx, lockKey, 1, taskLockTTL).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return nil // другой воркер уже пишет в кеш
	}
	defer c.rdb.Del(ctx, lockKey)

	v, err := c.version(ctx, teamID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, c.listKey(teamID, v, filterKey), data, taskCacheTTL).Err()
}

// IncrVersion инвалидирует весь кеш задач команды увеличением версии.
func (c *TaskCache) IncrVersion(ctx context.Context, teamID int64) error {
	return c.rdb.Incr(ctx, c.versionKey(teamID)).Err()
}
