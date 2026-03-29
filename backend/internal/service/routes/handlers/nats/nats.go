package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/routes/repositories/graph"
	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	repo *graph.Repository
}

func New(repo *graph.Repository) *Handler {
	return &Handler{repo: repo}
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
