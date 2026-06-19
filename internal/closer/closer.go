package closer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type closeFn struct {
	name string
	fn   func(context.Context) error
}

type closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
}

var globalCloser = &closer{}

func Add(name string, fn func(context.Context) error) {
	globalCloser.add(name, fn)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.closeAll(ctx)
}

func (c *closer) add(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

func (c *closer) closeAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		log.Info().Int("count", len(funcs)).Msg("starting graceful shutdown")

		var errs []error

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			start := time.Now()
			log.Info().Str("resource", f.name).Msg("closing")

			if err := f.fn(ctx); err != nil {
				log.Error().Err(err).Str("resource", f.name).Dur("duration", time.Since(start)).Msg("close error")
				errs = append(errs, err)
			} else {
				log.Info().Str("resource", f.name).Dur("duration", time.Since(start)).Msg("closed")
			}
		}

		result = errors.Join(errs...)
	})

	return result
}
