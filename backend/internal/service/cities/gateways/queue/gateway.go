package queue

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/core/nats"
	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
)

type Gateway struct {
	nats *nats.Client
}

func New(nats *nats.Client) *Gateway {
	return &Gateway{
		nats: nats,
	}
}

func (r *Gateway) Request(ctx context.Context, request messages.CitiesRequested) error {
	return natsutils.Publish(ctx, r.nats.Connection.JetStream, messages.CitiesSubjectRequested, request)
}
