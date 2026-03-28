package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) GetSubjects() []string {
	return []string{messages.CitiesSubjectParsed}
}

func (h *Handler) Handle(ctx context.Context, msg jetstream.Msg) error {
	switch msg.Subject() {
	case messages.CitiesSubjectParsed:
		return natsutils.Consume(ctx, msg, h.HandleCitiesParsed)
	}

	return natsutils.ErrHandlerNotFound
}
