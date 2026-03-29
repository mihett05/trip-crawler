package nats

import (
	"context"
	"fmt"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

func (h *Handler) HandleCitiesParsed(ctx context.Context, request messages.CitiesParsed) error {
	// TODO: сохранение ответа от парсера
	fmt.Println("cities parsed", request)
	return nil
}
