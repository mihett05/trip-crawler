package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/mihett05/trip-crawler/internal/worker"
)

func main() {
	envFile := flag.String("envFile", "", "env file for load")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := worker.New(ctx, *envFile)
	if err != nil {
		panic(err)
	}

	app.Start(ctx)

	<-ctx.Done()
	app.Shutdown()
}
