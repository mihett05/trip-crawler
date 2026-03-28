package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mihett05/trip-crawler/internal/worker/dto"
)

const defaultNakDelay = 30 * time.Second

// Worker pairs two consumers (station-load, ticket-parse) with their
// respective processors and handles the Ack/Nak lifecycle.
type Worker struct {
	stationConsumer Consumer
	ticketConsumer  Consumer
	stationProc     Processor[dto.StationLoadTask]
	ticketProc      Processor[dto.TicketParseTask]
}

func New(
	stationConsumer Consumer,
	ticketConsumer Consumer,
	stationProc Processor[dto.StationLoadTask],
	ticketProc Processor[dto.TicketParseTask],
) *Worker {
	return &Worker{
		stationConsumer: stationConsumer,
		ticketConsumer:  ticketConsumer,
		stationProc:     stationProc,
		ticketProc:      ticketProc,
	}
}

// Run starts both consume loops concurrently and blocks until ctx is cancelled
// or one of them returns a fatal error.
func (w *Worker) Run(ctx context.Context) error {
	stationErr := make(chan error, 1)
	ticketErr := make(chan error, 1)

	go func() {
		stationErr <- consume(ctx, w.stationConsumer, w.stationProc)
	}()
	go func() {
		ticketErr <- consume(ctx, w.ticketConsumer, w.ticketProc)
	}()

	select {
	case err := <-stationErr:
		return fmt.Errorf("station consumer: %w", err)
	case err := <-ticketErr:
		return fmt.Errorf("ticket consumer: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// consume is the generic dispatch loop: it deserialises each message into T,
// calls proc.Process, then Acks on success or Naks with an appropriate delay
// on failure. Malformed messages are Acked immediately (dropped) to avoid
// infinite redelivery.
func consume[T any](ctx context.Context, consumer Consumer, proc Processor[T]) error {
	return consumer.Consume(ctx, func(ctx context.Context, msg Message) error {
		var task T
		if err := json.Unmarshal(msg.Payload(), &task); err != nil {
			// Malformed payload — nothing a retry can fix; drop it.
			_ = msg.Ack()
			return nil
		}

		if err := proc.Process(ctx, task); err != nil {
			delay := defaultNakDelay
			var nakErr *NakError
			if errors.As(err, &nakErr) {
				delay = nakErr.Delay
			}
			return msg.Nak(delay)
		}

		return msg.Ack()
	})
}
