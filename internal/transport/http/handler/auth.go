package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/apierr"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/handler/model"
	"github.com/zoshc/secunda-task-manager/internal/transport/http/router"
)

type AuthService interface {
	Register(ctx context.Context, reg user.RegisterInput) error
	Login(ctx context.Context, creds user.LoginInput) (*user.Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*user.Tokens, error)
}

type authHandler struct {
	makeRouter router.MakeRouter
	svc        AuthService
}

func NewAuthHandler(svc AuthService) *authHandler {
	h := &authHandler{svc: svc}
	h.makeRouter = func(r fiber.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
	}
	return h
}

func (h *authHandler) Router() router.MakeRouter { return h.makeRouter }

func (h *authHandler) register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.svc.Register(c.UserContext(), user.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}); err != nil {
		return apierr.Response(c, err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

func (h *authHandler) login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tokens, err := h.svc.Login(c.UserContext(), user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return apierr.Response(c, err)
	}

	return c.JSON(model.TokensResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	})
}

func (h *authHandler) refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "refresh_token is required"})
	}

	tokens, err := h.svc.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return apierr.Response(c, err)
	}

	return c.JSON(model.TokensResponse{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
	})
}
