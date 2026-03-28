package queue

import (
	"context"

	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
)

type Gateway struct {
	js natsutils.JetStream
}

func New(js natsutils.JetStream) *Gateway {
	return &Gateway{
		js: js,
	}
}

func (r *Gateway) Request(ctx context.Context, request messages.CitiesRequested) error {
	return natsutils.Publish(ctx, r.js, messages.CitiesSubjectRequested, request)
}
