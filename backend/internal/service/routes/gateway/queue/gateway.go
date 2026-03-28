package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go"
)

type Gateway struct {
	js natsutils.JetStream
}

func New(js natsutils.JetStream) *Gateway {
	return &Gateway{
		js: js,
	}
}

func (r *Gateway) ScheduleTrip(ctx context.Context, request messages.TripRequested, scheduleAt time.Time) error {
	msg := nats.NewMsg(messages.SchedulesTripsSubjectParsed)
	msg.Header.Add("Nats-Schedule", fmt.Sprintf("@at %s", scheduleAt.Format(time.RFC3339)))
	msg.Header.Add("Nats-Schedule-Target", messages.TripsSubjectRequested)

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	msg.Data = body

	_, err = r.js.PublishMsg(ctx, msg)
	if err != nil {
		return fmt.Errorf("js.PublishMsg: %w", err)
	}

	return nil
}
