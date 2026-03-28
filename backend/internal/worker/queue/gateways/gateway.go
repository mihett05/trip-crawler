package gateways

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

func (g *Gateway) SendCities(ctx context.Context, cities messages.CitiesParsed) error {
	return natsutils.Publish(ctx, g.js, messages.CitiesSubjectParsed, cities)
}

func (g *Gateway) SendTrip(ctx context.Context, trip messages.TripParsed) error {
	return natsutils.Publish(ctx, g.js, messages.TripsSubjectParsed, trip)
}
