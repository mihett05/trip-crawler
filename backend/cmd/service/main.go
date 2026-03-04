package main

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/service"
)

func main() {
	ctx := context.Background()

	app, err := service.New(ctx)
	if err != nil {
		panic(err)
	}

	app.Run()
}
