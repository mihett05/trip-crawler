package worker

import (
	"context"
	"fmt"
)

// Runner is implemented by every processor. Run blocks until ctx is cancelled
// or a fatal error occurs.
type Runner interface {
	Run(ctx context.Context) error
}

// Worker starts all processors concurrently and returns as soon as one of them
// exits (either due to an error or ctx cancellation).
type Worker struct {
	runners []Runner
}

func New(runners ...Runner) *Worker {
	return &Worker{runners: runners}
}

func (w *Worker) Run(ctx context.Context) error {
	errc := make(chan error, len(w.runners))

	for _, r := range w.runners {
		go func(r Runner) { errc <- r.Run(ctx) }(r)
	}

	select {
	case err := <-errc:
		return fmt.Errorf("worker exited: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}
