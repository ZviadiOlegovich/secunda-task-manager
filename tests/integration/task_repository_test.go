package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
)

func setupMySQL(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "root",
			"MYSQL_DATABASE":      "testdb",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("get container port: %v", err)
	}

	dsn := fmt.Sprintf("root:root@tcp(%s:%s)/testdb?parseTime=true&multiStatements=true", host, port.Port())
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	runMigrations(t, db)
	return db
}

func runMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	migrationsPath := "file://../../migrations"

	driver, err := migratemysql.WithInstance(db, &migratemysql.Config{})
	if err != nil {
		t.Fatalf("create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "mysql", driver)
	if err != nil {
		t.Fatalf("create migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
}

func seedUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		fmt.Sprintf("user-%d@test.com", time.Now().UnixNano()), "hash", "Test User",
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedTeam(t *testing.T, db *sql.DB, createdBy int64) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO teams (name, created_by) VALUES (?, ?)`, "Test Team", createdBy,
	)
	if err != nil {
		t.Fatalf("seed team: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestTaskRepository_Create(t *testing.T) {
	db := setupMySQL(t)
	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	userID := seedUser(t, db)
	teamID := seedTeam(t, db, userID)

	desc := "some description"
	tsk := &task.Task{
		TeamID:      teamID,
		Title:       "First task",
		Description: &desc,
		Status:      task.StatusTodo,
		Priority:    task.PriorityMedium,
		CreatedBy:   userID,
	}

	id, err := repo.Create(ctx, tsk)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

func TestTaskRepository_GetByID(t *testing.T) {
	db := setupMySQL(t)
	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	userID := seedUser(t, db)
	teamID := seedTeam(t, db, userID)

	desc := "desc"
	tsk := &task.Task{
		TeamID:      teamID,
		Title:       "Get task",
		Description: &desc,
		Status:      task.StatusInProgress,
		Priority:    task.PriorityHigh,
		CreatedBy:   userID,
	}
	id, err := repo.Create(ctx, tsk)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Title != tsk.Title {
		t.Errorf("title: want %q, got %q", tsk.Title, got.Title)
	}
	if got.Status != tsk.Status {
		t.Errorf("status: want %q, got %q", tsk.Status, got.Status)
	}
	if got.Description == nil || *got.Description != desc {
		t.Errorf("description: want %q, got %v", desc, got.Description)
	}
}

func TestTaskRepository_GetByID_NotFound(t *testing.T) {
	db := setupMySQL(t)
	repo := repository.NewTaskRepository(db)

	_, err := repo.GetByID(context.Background(), 999999)
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestTaskRepository_Update(t *testing.T) {
	db := setupMySQL(t)
	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	userID := seedUser(t, db)
	teamID := seedTeam(t, db, userID)

	id, err := repo.Create(ctx, &task.Task{
		TeamID:    teamID,
		Title:     "Old title",
		Status:    task.StatusTodo,
		Priority:  task.PriorityLow,
		CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, _ := repo.GetByID(ctx, id)
	got.Title = "New title"
	got.Status = task.StatusDone
	got.Priority = task.PriorityHigh

	history := []task.TaskHistoryEntry{
		{TaskID: id, UserID: userID, Field: "title", OldValue: strPtr("Old title"), NewValue: strPtr("New title")},
		{TaskID: id, UserID: userID, Field: "status", OldValue: strPtr("todo"), NewValue: strPtr("done")},
	}
	if err := repo.UpdateWithHistory(ctx, got, history); err != nil {
		t.Fatalf("update: %v", err)
	}

	updated, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("title: want %q, got %q", "New title", updated.Title)
	}
	if updated.Status != task.StatusDone {
		t.Errorf("status: want %q, got %q", task.StatusDone, updated.Status)
	}
}

func TestTaskRepository_List(t *testing.T) {
	db := setupMySQL(t)
	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	userID := seedUser(t, db)
	teamID := seedTeam(t, db, userID)
	otherTeamID := seedTeam(t, db, userID)

	for i, s := range []task.TaskStatus{task.StatusTodo, task.StatusInProgress, task.StatusDone} {
		_, err := repo.Create(ctx, &task.Task{
			TeamID:    teamID,
			Title:     fmt.Sprintf("Task %d", i),
			Status:    s,
			Priority:  task.PriorityMedium,
			CreatedBy: userID,
		})
		if err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	_, err := repo.Create(ctx, &task.Task{
		TeamID: otherTeamID, Title: "Other team task",
		Status: task.StatusTodo, Priority: task.PriorityLow, CreatedBy: userID,
	})
	if err != nil {
		t.Fatalf("create other-team task: %v", err)
	}

	t.Run("all tasks for team", func(t *testing.T) {
		tasks, err := repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 1, Limit: 20})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(tasks) != 3 {
			t.Errorf("want 3 tasks, got %d", len(tasks))
		}
	})

	t.Run("filter by status", func(t *testing.T) {
		s := task.StatusTodo
		tasks, err := repo.List(ctx, task.ListFilter{TeamID: teamID, Status: &s, Page: 1, Limit: 20})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(tasks) != 1 {
			t.Errorf("want 1 task, got %d", len(tasks))
		}
		if tasks[0].Status != task.StatusTodo {
			t.Errorf("want status todo, got %q", tasks[0].Status)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		page1, err := repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 1, Limit: 2})
		if err != nil {
			t.Fatalf("list page 1: %v", err)
		}
		if len(page1) != 2 {
			t.Errorf("want 2 tasks (page 1), got %d", len(page1))
		}

		page2, err := repo.List(ctx, task.ListFilter{TeamID: teamID, Page: 2, Limit: 2})
		if err != nil {
			t.Fatalf("list page 2: %v", err)
		}
		if len(page2) != 1 {
			t.Errorf("want 1 task (page 2), got %d", len(page2))
		}
	})
}

func strPtr(s string) *string { return &s }
