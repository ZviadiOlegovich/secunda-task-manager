package http

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type privateServer struct {
	fiber   *fiber.App
	address string
}

func NewPrivateServer(port int) *privateServer {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/livez", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	return &privateServer{
		fiber:   app,
		address: fmt.Sprintf(":%d", port),
	}
}

func (s *privateServer) Run() error {
	return s.fiber.Listen(s.address)
}

func (s *privateServer) Shutdown(ctx context.Context) error {
	return s.fiber.ShutdownWithContext(ctx)
}
