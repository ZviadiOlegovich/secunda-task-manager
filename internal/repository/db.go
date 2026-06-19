package repository

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/zoshc/secunda-task-manager/internal/config"
)

func NewMySQL(cfg config.MySQLConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sqlx.DB, path string) error {
	driver, err := mysqlmigrate.WithInstance(db.DB, &mysqlmigrate.Config{})
	if err != nil {
		return fmt.Errorf("create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
