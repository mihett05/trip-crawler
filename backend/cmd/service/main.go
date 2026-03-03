package main

import (
	"context"

	"github.com/mihett05/trip-crawler/internal/mainservice"
)

func main() {
	ctx := context.Background()

	app, err := mainservice.New(ctx)
	if err != nil {
		panic(err)
	}

	app.Run()
}
