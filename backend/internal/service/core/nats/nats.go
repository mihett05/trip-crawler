package nats

import (
	"context"
	"fmt"

	citieshandlers "github.com/mihett05/trip-crawler/internal/service/cities/handlers/nats"
	routeshandlers "github.com/mihett05/trip-crawler/internal/service/routes/handlers/nats"
	"github.com/mihett05/trip-crawler/pkg/application/config"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Client struct {
	Connection *natsutils.Connection
	Streams    *natsutils.Streams
	Logger     *zap.Logger

	consumeContexts []jetstream.ConsumeContext
}

func New(ctx context.Context, config config.Config, logger *zap.Logger) (*Client, error) {
	connection, err := natsutils.NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("natsutils.NewConnection: %w", err)
	}

	streams, err := natsutils.NewStreams(ctx, connection.JetStream)
	if err != nil {
		return nil, fmt.Errorf("natsutils.NewStreams: %w", err)
	}

	return &Client{
		Connection: connection,
		Streams:    streams,
		Logger:     logger,
	}, nil
}

func (c *Client) RunConsumers(ctx context.Context, citiesHandler *citieshandlers.Handler, routesHandler *routeshandlers.Handler) error {
	citiesCtx, err := natsutils.RunConsumer(ctx, c.Streams.Cities, citiesHandler, c.Logger)
	if err != nil {
		return fmt.Errorf("natsutils.RunConsumer: %w", err)
	}

	tripsCtx, err := natsutils.RunConsumer(ctx, c.Streams.Trips, routesHandler, c.Logger)
	if err != nil {
		return fmt.Errorf("natsutils.RunConsumer: %w", err)
	}

	c.consumeContexts = append(c.consumeContexts, citiesCtx, tripsCtx)

	return nil
}

func (c *Client) Shutdown() {
	c.Connection.Connection.Drain()

	for _, ctx := range c.consumeContexts {
		ctx.Stop()
	}
}
