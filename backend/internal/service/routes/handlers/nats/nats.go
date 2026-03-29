package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/routes"
	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	service *routes.Service
}

func New(svc *routes.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetSubjects() []string {
	return []string{messages.TripsSubjectParsed}
}

func (h *Handler) Handle(ctx context.Context, msg jetstream.Msg) error {
	switch msg.Subject() {
	case messages.TripsSubjectParsed:
		return natsutils.Consume(ctx, msg, h.HandleTripParsed)
	}

	return natsutils.ErrHandlerNotFound
}
