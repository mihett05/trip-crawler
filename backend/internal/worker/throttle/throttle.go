// Package throttle provides a burst-then-pause rate limiter for outbound HTTP
// calls. Instead of a steady token-bucket rate, it allows BurstSize requests
// to pass through immediately, then sleeps for a random duration in
// [SleepMin, SleepMax] before the next burst. This mimics human browsing
// behaviour and keeps average RPS well below the burst peak.
package throttle

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type Config struct {
	BurstSize int           `env:"THROTTLE_BURST_SIZE" envDefault:"5"`
	SleepMin  time.Duration `env:"THROTTLE_SLEEP_MIN"  envDefault:"5s"`
	SleepMax  time.Duration `env:"THROTTLE_SLEEP_MAX"  envDefault:"15s"`
}

// Throttler is safe for concurrent use. All goroutines share the same counter,
// so the burst limit is global across workers — not per-goroutine.
type Throttler struct {
	cfg     Config
	mu      sync.Mutex
	counter int
}

func New(cfg Config) *Throttler {
	return &Throttler{cfg: cfg}
}

// Wait must be called before every outbound request. It returns immediately
// while the current burst has capacity; once the burst is exhausted it sleeps
// for a random interval in [SleepMin, SleepMax] and resets the counter.
// Returns ctx.Err() if the context is cancelled during the sleep.
func (t *Throttler) Wait(ctx context.Context) error {
	t.mu.Lock()
	t.counter++
	burst := t.counter >= t.cfg.BurstSize
	if burst {
		t.counter = 0
	}
	t.mu.Unlock()

	if !burst {
		return nil
	}

	sleep := t.randomSleep()
	select {
	case <-time.After(sleep):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Throttler) randomSleep() time.Duration {
	diff := t.cfg.SleepMax - t.cfg.SleepMin
	if diff <= 0 {
		return t.cfg.SleepMin
	}
	return t.cfg.SleepMin + time.Duration(rand.Int64N(int64(diff)))
}
