package router

import "github.com/gofiber/fiber/v2"

type MakeRouter func(fiber.Router)

type Router interface {
	Init(fiber.Router)
}

type router struct {
	makeRouter MakeRouter
}

func NewRouter(fn MakeRouter) Router {
	return &router{makeRouter: fn}
}

func (r *router) Init(api fiber.Router) {
	r.makeRouter(api)
}
