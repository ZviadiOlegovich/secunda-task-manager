package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

func (r *RateLimiter) LimitByUserID(maxReqs int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := UserIDFromCtx(c)
		if !ok {
			return c.Next()
		}
		key := fmt.Sprintf("rl:user:%d", userID)
		exceeded, err := r.check(c.Context(), key, maxReqs, window)
		if err != nil {
			return c.Next()
		}
		if exceeded {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		return c.Next()
	}
}

func (r *RateLimiter) check(ctx context.Context, key string, maxReqs int, window time.Duration) (bool, error) {
	count, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := r.rdb.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}
	return count > int64(maxReqs), nil
}
