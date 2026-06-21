package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zoshc/secunda-task-manager/internal/cache"
	"github.com/zoshc/secunda-task-manager/internal/config"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/email"
	"github.com/zoshc/secunda-task-manager/internal/services/stats"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
	http "github.com/zoshc/secunda-task-manager/internal/transport/http"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/handler"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/middleware"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
	"github.com/zoshc/secunda-task-manager/pkg/jwt"
	"github.com/zoshc/secunda-task-manager/pkg/logger"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Server.LogForcePlain, logger.ParseLevel(cfg.Server.LogLevel), cfg.Server.ServiceName)

	db, err := repository.NewMySQL(cfg.MySQL)
	if err != nil {
		return fmt.Errorf("connect to mysql: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("close mysql")
		}
	}()

	if err = repository.RunMigrations(db, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	redis, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect to redis: %w", err)
	}
	defer func() {
		if err := redis.Close(); err != nil {
			log.Error().Err(err).Msg("close redis")
		}
	}()

	jwtProvider := jwt.NewProvider(cfg.JWT)
	authMiddleware := middleware.Auth(jwtProvider)

	rateLimiter := middleware.NewRateLimiter(redis)
	userRateLimit := rateLimiter.LimitByUserID(100, time.Minute)

	userRepo := repository.NewUserRepository(db)
	userSvc := user.New(userRepo, jwtProvider)
	authHandler := handler.NewAuthHandler(userSvc)

	emailSvc := email.NewCBService(email.NewMock(), log)

	teamRepo := repository.NewTeamRepository(db)
	teamSvc := team.New(teamRepo, emailSvc)
	teamHandler := handler.NewTeamHandler(authMiddleware, userRateLimit, teamSvc)

	taskCache := cache.NewTaskCache(redis)
	taskRepo := repository.NewTaskRepository(db, taskCache)
	taskSvc := task.New(taskRepo, teamRepo)
	taskHandler := handler.NewTaskHandler(authMiddleware, userRateLimit, taskSvc)

	statsRepo := repository.NewStatsRepository(db)
	statsSvc := stats.New(statsRepo)
	statsHandler := handler.NewStatsHandler(authMiddleware, userRateLimit, statsSvc)

	srv := http.NewServer(*cfg,
		router.NewRouter(authHandler.Router()),
		router.NewRouter(teamHandler.Router()),
		router.NewRouter(taskHandler.Router()),
		router.NewRouter(statsHandler.Router()),
	)
	privateSrv := http.NewPrivateServer(cfg.Server.PrivatePort, []http.Pinger{
		db,
		redis,
	})

	errCh := make(chan error, 2)

	go func() {
		log.Info().Str("addr", fmt.Sprintf(":%d", cfg.Server.Port)).Msg("starting public server")
		errCh <- srv.Run()
	}()

	go func() {
		log.Info().Str("addr", fmt.Sprintf(":%d", cfg.Server.PrivatePort)).Msg("starting private server")
		errCh <- privateSrv.Run()
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case sig := <-signalCh:
		log.Info().Str("signal", sig.String()).Msg("shutting down")
	case err := <-errCh:
		log.Error().Err(err).Msg("server error, shutting down")
	}

	serverCtx, serverCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer serverCancel()

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

	log.Info().Msg("server stopped")
	return nil
}
