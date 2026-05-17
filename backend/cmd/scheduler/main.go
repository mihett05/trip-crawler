package main

import (
	"context"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/mihett05/trip-crawler/internal/service"
	"github.com/mihett05/trip-crawler/internal/service/scheduler"
	"github.com/mihett05/trip-crawler/pkg/messages"

	citycsv "github.com/mihett05/trip-crawler/internal/worker/parsers/cities-by-csv"
)

func main() {
	envFile := flag.String("envFile", "", "env file for load")
	mode := flag.String("mode", "cities", "scheduler mode (cities or trips)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app, err := service.New(ctx, *envFile)
	if err != nil {
		panic(err)
	}

	switch *mode {
	case "cities":
		//task := app.Scheduler.GenerateCititesTask(time.Now())
		app.CitiesQueue.Request(ctx, messages.CitiesRequested{
			TopCities: 15,
		})
	case "trips":
		routes, err := citycsv.GetRoutesFromCSV()
		if err != nil {
			panic(err)
		}

		var rawTasks []scheduler.Connection
		for _, r := range routes {
			rawTasks = append(rawTasks, scheduler.Connection{
				OriginCode:       r.Departure.ID,
				DestinationCode:  r.Arrival.ID,
				OriginPopulation: 600000,
				DestPopulation:   600000,
				LastParsedAt:     time.Time{},
				LastUsedAt:       time.Now(),
			})
		}

		tasks := app.Scheduler.GenerateTicketTasks(time.Now(), rawTasks)

		// Ограничение на 10 000 задач
		if len(tasks) > 100000 {
			tasks = tasks[:100000]
		}

		fmt.Printf("[scheduler] Sending %d tasks to NATS...\n", len(tasks))

		for _, task := range tasks {
			err = app.RoutesQueue.ScheduleTrip(ctx, messages.TripRequested{
				DepartStationID:      task.OriginCode,
				DestinationStationID: task.DestinationCode,
				DepartureAtTimestamp: task.DepartureDate.Unix(),
			}, time.Now().Add(-5*time.Hour))
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("[scheduler] Done.")
	default:
		panic("unknown mode: " + *mode)
	}

	app.Shutdown()
}