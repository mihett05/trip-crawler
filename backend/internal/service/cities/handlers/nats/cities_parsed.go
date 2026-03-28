package nats

import (
	"context"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleCitiesParsed(ctx context.Context, request messages.CitiesParsed) error {
	// TODO: сохранение ответа от парсера
	return nil
}
