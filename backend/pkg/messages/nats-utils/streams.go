package natsutils

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/messages"
	"github.com/nats-io/nats.go/jetstream"
)

type Streams struct {
	Cities jetstream.Stream
	Trips  jetstream.Stream
}

func NewStreams(ctx context.Context, js JetStream) (*Streams, error) {
	citiesStream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     messages.CitiesStream,
		Subjects: []string{messages.CitiesSubjectPrefix + ">"},
	})
	if err != nil {
		return nil, fmt.Errorf("js.CreateOrUpdateStream with name=%s: %w", messages.CitiesStream, err)
	}

	tripsStream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:              messages.TripsStream,
		Subjects:          []string{messages.TripsSubjectPrefix + ">", messages.SchedulesTripsSubjectPrefix + ">"},
		AllowMsgSchedules: true,
	})
	if err != nil {
		return nil, fmt.Errorf("js.CreateOrUpdateStream with name=%s: %w", messages.TripsStream, err)
	}

	return &Streams{
		Cities: citiesStream,
		Trips:  tripsStream,
	}, nil
}
