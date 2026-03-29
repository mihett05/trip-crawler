package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
		rawTasks, err := LoadConnections("connections.json")
		if err != nil {
			panic(err)
		}
		fmt.Println("[scheduler] loaded", len(rawTasks), "rawTasks")
		tasks := app.Scheduler.GenerateTicketTasks(time.Now(), rawTasks)
		fmt.Println("[scheduler] loaded", len(tasks), "tasks")
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

func LoadConnections(filePath string) ([]scheduler.Connection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var connections []scheduler.Connection
	if err := json.NewDecoder(f).Decode(&connections); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return connections, nil
}
