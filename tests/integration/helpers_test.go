//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/zoshc/secunda-task-manager/internal/repository"
)

var (
	testDB      *sql.DB
	seedCounter atomic.Int64
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "testdb",
			},
			WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		log.Printf("start mysql container: %v", err)
		return 1
	}
	defer container.Terminate(ctx)

	apiCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	host, err := container.Host(apiCtx)
	if err != nil {
		log.Printf("get container host: %v", err)
		return 1
	}
	port, err := container.MappedPort(apiCtx, "3306")
	if err != nil {
		log.Printf("get container port: %v", err)
		return 1
	}

	dsn := fmt.Sprintf("root:root@tcp(%s:%s)/testdb?parseTime=true&multiStatements=true", host, port.Port())

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("open db: %v", err)
		return 1
	}
	defer db.Close()

	if err := db.PingContext(apiCtx); err != nil {
		log.Printf("ping db: %v", err)
		return 1
	}

	if err := repository.RunMigrations(db, "../../migrations"); err != nil {
		log.Printf("run migrations: %v", err)
		return 1
	}

	testDB = db
	return m.Run()
}

type baseSuite struct {
	suite.Suite
	db *sql.DB
}

func (s *baseSuite) SetupSuite() {
	s.db = testDB
}

func seedUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	n := seedCounter.Add(1)
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)`,
		fmt.Sprintf("user-%d@test.com", n), "hash", "Test User",
	)
	require.NoError(t, err)
	id, _ := res.LastInsertId()
	return id
}

func seedTeam(t *testing.T, db *sql.DB, createdBy int64) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO teams (name, created_by) VALUES (?, ?)`, "Test Team", createdBy,
	)
	require.NoError(t, err)
	id, _ := res.LastInsertId()
	return id
}
