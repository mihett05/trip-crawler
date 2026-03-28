package natsutils

import (
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/application/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Connection struct {
	Connection *nats.Conn
	JetStream  JetStream
}

func NewConnection(config config.Config) (*Connection, error) {
	nc, err := nats.Connect(config.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream.New: %w", err)
	}

	return &Connection{
		Connection: nc,
		JetStream:  js,
	}, nil
}
