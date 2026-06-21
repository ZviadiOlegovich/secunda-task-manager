package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zoshc/secunda-task-manager/internal/config"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/middleware"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
)

type server struct {
	fiber   *fiber.App
	address string
}

func NewServer(cfg config.Config, routers ...router.Router) *server {
	s := &server{
		address: fmt.Sprintf(":%d", cfg.Server.Port),
		fiber: fiber.New(fiber.Config{
			CaseSensitive: true,
			AppName:       cfg.Server.ServiceName,
			ReadTimeout:   30 * time.Second,
			WriteTimeout:  30 * time.Second,
		}),
	}

	s.fiber.Use(cors.New(cors.Config{
		AllowOrigins: cfg.App.AllowedOrigins,
	}))
	s.fiber.Use(middleware.Metrics())

	api := s.fiber.Group("/api/v1")
	for _, r := range routers {
		r.Init(api)
	}

	return s
}

func (s *server) Run() error {
	return s.fiber.Listen(s.address)
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.fiber.ShutdownWithContext(ctx)
}
