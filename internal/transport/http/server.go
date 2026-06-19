package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/zoshc/secunda-task-manager/internal/config"
)

type Router interface {
	Init(r fiber.Router)
}

type server struct {
	fiber   *fiber.App
	address string
}

func NewServer(cfg config.Config, routers ...Router) *server {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		AppName:       cfg.Server.ServiceName,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.App.AllowedOrigins,
	}))

	api := app.Group("/api/v1")
	for _, r := range routers {
		r.Init(api)
	}

	return &server{
		fiber:   app,
		address: fmt.Sprintf(":%d", cfg.Server.Port),
	}
}

func (s *server) Run() error {
	return s.fiber.Listen(s.address)
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.fiber.ShutdownWithContext(ctx)
}
