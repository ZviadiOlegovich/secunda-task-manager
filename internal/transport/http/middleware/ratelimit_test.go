package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRateLimiter(t *testing.T) *RateLimiter {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr:        mr.Addr(),
		DialTimeout: 100 * time.Millisecond,
		MaxRetries:  0,
	})
	return NewRateLimiter(rdb)
}

func TestRateLimiter_LimitByUserID(t *testing.T) {
	rl := newTestRateLimiter(t)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(userIDKey, int64(42))
		return c.Next()
	})
	app.Get("/", rl.LimitByUserID(2, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for i := 0; i < 2; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d: want 200, got %d", i+1, resp.StatusCode)
		}
	}

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("3rd request: %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("want 429 after limit, got %d", resp.StatusCode)
	}
}

func TestRateLimiter_NoUserID_PassThrough(t *testing.T) {
	rl := newTestRateLimiter(t)

	app := fiber.New()
	app.Get("/", rl.LimitByUserID(1, time.Minute), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for i := 0; i < 5; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("request %d without userID: want 200, got %d", i+1, resp.StatusCode)
		}
	}
}

