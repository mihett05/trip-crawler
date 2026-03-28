package natsutils

import (
	"context"
	"encoding/json"
	"fmt"
)

func Publish[Request any](ctx context.Context, js JetStream, subject string, request Request) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	_, err = js.Publish(ctx, subject, body)
	if err != nil {
		return fmt.Errorf("js.Publish: %w", err)
	}

	return nil
}
