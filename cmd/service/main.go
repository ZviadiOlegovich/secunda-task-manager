package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zoshc/secunda-task-manager/internal/cache"
	"github.com/zoshc/secunda-task-manager/internal/closer"
	"github.com/zoshc/secunda-task-manager/internal/config"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/email"
	"github.com/zoshc/secunda-task-manager/internal/services/stats"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/middleware"
	"github.com/zoshc/secunda-task-manager/internal/transport/http"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/handler"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
	"github.com/zoshc/secunda-task-manager/pkg/jwt"
	"github.com/zoshc/secunda-task-manager/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Server.LogForcePlain, logger.ParseLevel(cfg.Server.LogLevel), cfg.Server.ServiceName)

	db, err := repository.NewMySQL(cfg.MySQL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to mysql")
	}
	closer.Add("mysql", func(_ context.Context) error {
		return db.Close()
	})

	if err = repository.RunMigrations(db, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	rdb, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("connect to redis")
	}
	closer.Add("redis", func(_ context.Context) error {
		return rdb.Close()
	})

	jwtProvider := jwt.NewProvider(cfg.JWT)
	authMiddleware := middleware.Auth(jwtProvider)

	userRepo := repository.NewUserRepository(db)
	userSvc := user.New(userRepo, jwtProvider)
	authHandler := handler.NewAuthHandler(userSvc)

	emailSvc := email.NewCBService(email.NewMock())

	teamRepo := repository.NewTeamRepository(db)
	teamSvc := team.New(teamRepo, emailSvc)
	teamHandler := handler.NewTeamHandler(authMiddleware, teamSvc)

	taskRepo := repository.NewTaskRepository(db)
	taskSvc := task.New(taskRepo, teamRepo)
	taskHandler := handler.NewTaskHandler(authMiddleware, taskSvc)

	statsRepo := repository.NewStatsRepository(db)
	statsSvc := stats.New(statsRepo)
	statsHandler := handler.NewStatsHandler(authMiddleware, statsSvc)

	srv := http.NewServer(*cfg,
		router.NewRouter(authHandler.Router()),
		router.NewRouter(teamHandler.Router()),
		router.NewRouter(taskHandler.Router()),
		router.NewRouter(statsHandler.Router()),
	)
	privateSrv := http.NewPrivateServer(cfg.Server.PrivatePort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Str("addr", fmt.Sprintf(":%d", cfg.Server.Port)).Msg("starting public server")
		if err := srv.Run(); err != nil && ctx.Err() == nil {
			log.Error().Err(err).Msg("public server error")
		}
	}()

	go func() {
		log.Info().Str("addr", fmt.Sprintf(":%d", cfg.Server.PrivatePort)).Msg("starting private server")
		if err := privateSrv.Run(); err != nil && ctx.Err() == nil {
			log.Error().Err(err).Msg("private server error")
		}
	}()

	<-ctx.Done()
	stop()

	log.Info().Msg("shutting down")

	serverCtx, serverCancel := context.WithTimeout(context.Background(), 15*time.Second)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := srv.Shutdown(serverCtx); err != nil {
			log.Error().Err(err).Msg("public server shutdown error")
		}
	}()
	go func() {
		defer wg.Done()
		if err := privateSrv.Shutdown(serverCtx); err != nil {
			log.Error().Err(err).Msg("private server shutdown error")
		}
	}()
	wg.Wait()
	serverCancel()

	resourceCtx, resourceCancel := context.WithTimeout(context.Background(), 10*time.Second)

	if err = closer.CloseAll(resourceCtx); err != nil {
		log.Error().Err(err).Msg("shutdown errors")
	}
	resourceCancel()

	log.Info().Msg("server stopped")
}
