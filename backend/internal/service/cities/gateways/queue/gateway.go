package queue

import (
	"context"

	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

// Контракт +
// Продьюсер задач +-
// Консьюмер задач
// Продьюсер ответов
// Консьюмер ответов

type JetStream interface {
	jetstream.JetStream
}

type Gateway struct {
	js JetStream
}

func New(js JetStream) *Gateway {
	return &Gateway{
		js: js,
	}
}

func (r *Gateway) Request(ctx context.Context, request messages.CitiesRequested) error {
	return natsutils.Publish(ctx, r.js, messages.CitiesSubjectParsed, request)
}
