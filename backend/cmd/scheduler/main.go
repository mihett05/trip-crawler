package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihett05/trip-crawler/internal/service"
	"github.com/mihett05/trip-crawler/internal/service/scheduler"
	"github.com/mihett05/trip-crawler/pkg/messages"
)

func main() {
	envFile := flag.String("envFile", "", "env file for load (e.g., .env.service.local)")
	mode := flag.String("mode", "citites", "scheduler mode (cities or trips)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := service.New(ctx, *envFile)
	if err != nil {
		panic(err)
	}

	switch *mode {
	case "cities":
		// TODO: вынести в сервис
		task := app.Scheduler.GenerateCititesTask(time.Now())
		app.CitiesQueue.Request(ctx, messages.CitiesRequested{
			TopCities: task.TopN,
		})
	case "trips":
		// TODO: сделать загрузку из репозитория
		tasks := app.Scheduler.GenerateTicketTasks(time.Now(), []scheduler.Connection{})
		for _, task := range tasks {
			app.RoutesQueue.ScheduleTrip(ctx, messages.TripRequested{
				DepartStation:        task.OriginCode,
				DepartStationID:      "", // из репозитория
				DepartureAtTimestamp: task.DepartureDate.Unix(),
				DestinationStation:   task.DestinationCode,
				DestinationStationID: "", // из репозитория
			}, task.ScheduledAt)
		}
	default:
		panic("unknown mode: " + *mode)
	}

	app.Shutdown()
}
