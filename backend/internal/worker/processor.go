package worker

import (
	"context"
	"time"
)

// Processor handles one task of type T.
//   - Return nil  → caller Acks the message (success).
//   - Return *NakError → caller Naks with the error's delay.
//   - Return any other error → caller Naks with the default delay.
type Processor[T any] interface {
	Process(ctx context.Context, task T) error
}

// NakError signals that a message should be requeued after Delay instead of
// the default retry interval. Wrap it around the original cause so callers
// can still inspect the underlying error with errors.Is / errors.As.
type NakError struct {
	Cause error
	Delay time.Duration
}

func (e *NakError) Error() string  { return e.Cause.Error() }
func (e *NakError) Unwrap() error  { return e.Cause }
