package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service/cities/services/cities"
	"github.com/mihett05/trip-crawler/pkg/messages"
	natsutils "github.com/mihett05/trip-crawler/pkg/messages/nats-utils"
	"github.com/nats-io/nats.go/jetstream"
)

type Handler struct {
	service *cities.Service
}

func New(svc *cities.Service) *Handler {
	return &Handler{service: svc}
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
