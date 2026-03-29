package handlers

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	"github.com/mihett05/trip-crawler/internal/worker/queue/gateways"
	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	rzd     *rzd.Client
	gateway *gateways.Gateway
}

func New(rzdClient *rzd.Client, gateway *gateways.Gateway) *Handler {
	return &Handler{rzd: rzdClient, gateway: gateway}
}

func (h *Handler) GetSubjects() []string {
	return []string{messages.CitiesSubjectRequested, messages.TripsSubjectRequested}
}

func (h *Handler) Handle(ctx context.Context, msg jetstream.Msg) error {
	switch msg.Subject() {
	case messages.CitiesSubjectRequested:
		return natsutils.Consume(ctx, msg, h.HandleCitiesRequested)
	case messages.TripsSubjectRequested:
		return natsutils.Consume(ctx, msg, h.HandleTripRequested)
	}

	return natsutils.ErrHandlerNotFound
}
