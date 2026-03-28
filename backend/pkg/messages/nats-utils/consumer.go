package natsutils

import (
	context "context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

var ErrHandlerNotFound = errors.New("no appropriate handler found")

type Consumer interface {
	GetSubjects() []string
	Handle(ctx context.Context, msg jetstream.Msg) error
}

func RunConsumer(ctx context.Context, stream jetstream.Stream, handler Consumer, logger *zap.Logger) (jetstream.ConsumeContext, error) {
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		FilterSubjects: handler.GetSubjects(),
		MaxDeliver:     5,
	})
	if err != nil {
		return nil, fmt.Errorf("stream.CreateOrUpdateConsumer: %w", err)
	}

	return consumer.Consume(func(msg jetstream.Msg) {
		msgCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		if err := handler.Handle(msgCtx, msg); err != nil {
			logger.Error(
				"messages.natsutils.RunConsumer: failed to handle consumed message",
				zap.String("subject", msg.Subject()),
				zap.Error(err),
			)
			msg.Nak()
			return
		}

		msg.Ack()
	})
}

func Consume[Request any](ctx context.Context, msg jetstream.Msg, handler func(context.Context, Request) error) error {
	var request Request
	if err := json.Unmarshal(msg.Data(), &request); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return handler(ctx, request)
}
