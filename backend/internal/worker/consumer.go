package worker

import (
	"context"
	"time"
)

// Message wraps a raw queue message with lifecycle control.
// Ack removes it from the queue; Nak returns it for redelivery after delay.
type Message interface {
	Payload() []byte
	Ack() error
	Nak(delay time.Duration) error
}

// Consumer delivers messages to handler one at a time and blocks until ctx
// is cancelled or a fatal error occurs.
type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, msg Message) error) error
}
