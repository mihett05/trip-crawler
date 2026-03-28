package worker

import (
	"context"
	"time"
)

// Topics used by the worker processors.
const (
	TopicStationsLoad   = "STATIONS.load"
	TopicStationsResult = "STATIONS.result"
	TopicTicketsParse   = "TICKETS.parse"
	TopicTicketsResult  = "TICKETS.result"
)

// Gateway abstracts the message broker for both directions of communication.
//
//   - Subscribe delivers incoming tasks one at a time to handler. It blocks
//     until ctx is cancelled or a fatal broker error occurs. The broker decides
//     Ack/Nak based on the handler's return value:
//     nil      → Ack (task done)
//     *NakError → Nak with NakError.Delay (custom retry interval, e.g. proxy ban)
//     other    → Nak with a default delay (transient error, retry shortly)
//
//   - Publish sends a result payload to topic.
type Gateway interface {
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, payload []byte) error) error
	Publish(ctx context.Context, topic string, data []byte) error
}

// NakError signals that a message should be requeued after Delay instead of
// the broker's default retry interval.
type NakError struct {
	Cause error
	Delay time.Duration
}

func (e *NakError) Error() string { return e.Cause.Error() }
func (e *NakError) Unwrap() error { return e.Cause }
