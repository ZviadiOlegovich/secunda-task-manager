package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

func newTestCache(t *testing.T) (*TaskCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewTaskCache(rdb), mr
}

func TestTaskCache_MissAndSet(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	ver, _ := c.GetVersion(ctx, 1)
	got, err := c.GetTaskList(ctx, 1, ver, "filter1")
	if err != nil || got != nil {
		t.Fatalf("want cache miss, got err=%v tasks=%v", err, got)
	}

	tasks := []*task.Task{{ID: 1, Title: "Fix bug"}, {ID: 2, Title: "Review PR"}}
	if err := c.SetTaskListIfVersion(ctx, 1, ver, "filter1", tasks); err != nil {
		t.Fatalf("SetTaskListIfVersion: %v", err)
	}

	got, err = c.GetTaskList(ctx, 1, ver, "filter1")
	if err != nil {
		t.Fatalf("GetTaskList: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("want 2 tasks, got %d", len(got))
	}
}

func TestTaskCache_IncrVersionInvalidates(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	ver, _ := c.GetVersion(ctx, 1)
	_ = c.SetTaskListIfVersion(ctx, 1, ver, "filter1", []*task.Task{{ID: 1, Title: "Task"}})

	if err := c.IncrVersion(ctx, 1); err != nil {
		t.Fatalf("IncrVersion: %v", err)
	}

	newVer, _ := c.GetVersion(ctx, 1)
	got, err := c.GetTaskList(ctx, 1, newVer, "filter1")
	if err != nil {
		t.Fatalf("GetTaskList: %v", err)
	}
	if got != nil {
		t.Error("want cache miss after version increment, got data")
	}
}

func TestTaskCache_StaleWriteDoesNotPollute(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	ver, _ := c.GetVersion(ctx, 1)
	_ = c.IncrVersion(ctx, 1)
	_ = c.SetTaskListIfVersion(ctx, 1, ver, "filter1", []*task.Task{{ID: 1, Title: "Stale"}})

	newVer, _ := c.GetVersion(ctx, 1)
	got, _ := c.GetTaskList(ctx, 1, newVer, "filter1")
	if got != nil {
		t.Error("want cache miss on new version, stale write must not pollute it")
	}
}
