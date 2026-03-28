package scheduler

import (
	"context"
	"fmt"
	"time"
)

// Publisher is the only messaging contract the scheduler needs.
// The concrete implementation (NATS, in-memory for tests) is injected externally.
type Publisher interface {
	PublishStationLoad(ctx context.Context, task StationLoadTask) error
	PublishTicketParse(ctx context.Context, task TicketParseTask) error
}

type Scheduler struct {
	cfg  Config
	repo Repository
	pub  Publisher
}

func New(cfg Config, repo Repository, pub Publisher) *Scheduler {
	return &Scheduler{cfg: cfg, repo: repo, pub: pub}
}

// Run blocks until ctx is cancelled, firing two daily jobs:
//   - jobLoadStations  at StationLoadHour:StationLoadMinute
//   - jobEnqueueTickets at TicketEnqueueHour:TicketEnqueueMinute
func (s *Scheduler) Run(ctx context.Context) error {
	stationErr := make(chan error, 1)
	ticketErr := make(chan error, 1)

	go func() {
		stationErr <- s.runDaily(ctx, s.cfg.StationLoadHour, s.cfg.StationLoadMinute, s.jobLoadStations)
	}()
	go func() {
		ticketErr <- s.runDaily(ctx, s.cfg.TicketEnqueueHour, s.cfg.TicketEnqueueMinute, s.jobEnqueueTickets)
	}()

	select {
	case err := <-stationErr:
		return err
	case err := <-ticketErr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// runDaily waits until the next wall-clock trigger (hour:minute) and
// calls job, then repeats every 24 hours. A job error is logged and
// skipped so the scheduler keeps running.
func (s *Scheduler) runDaily(ctx context.Context, hour, minute int, job func(context.Context) error) error {
	for {
		next := nextTrigger(hour, minute)
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			return ctx.Err()
		}

		if err := job(ctx); err != nil {
			// non-fatal: surface the error but keep the loop alive
			_ = fmt.Errorf("scheduler job failed: %w", err)
		}
	}
}

// nextTrigger returns the next wall-clock time at which hour:minute occurs.
// If that time has already passed today it returns tomorrow's occurrence.
func nextTrigger(hour, minute int) time.Time {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// jobLoadStations publishes a single StationLoadTask for the worker to refresh
// the city/station/connection graph.
func (s *Scheduler) jobLoadStations(ctx context.Context) error {
	task := StationLoadTask{
		RequestedAt: time.Now(),
		TopN:        s.cfg.TopNCities,
	}
	return s.pub.PublishStationLoad(ctx, task)
}

// jobEnqueueTickets queries all known connections and, for each (connection, date)
// pair that the scoring matrix decides needs a refresh, publishes a TicketParseTask.
func (s *Scheduler) jobEnqueueTickets(ctx context.Context) error {
	today := truncateToDay(time.Now())

	connections, err := s.repo.GetTopConnections(ctx, s.cfg.TopNCities)
	if err != nil {
		return fmt.Errorf("repo.GetTopConnections: %w", err)
	}

	throttle := time.NewTicker(time.Second / time.Duration(s.cfg.PublishRatePerSec))
	defer throttle.Stop()

	for _, conn := range connections {
		anyPublished := false

		for daysAhead := 1; daysAhead <= s.cfg.DaysAhead; daysAhead++ {
			date := today.AddDate(0, 0, daysAhead)
			if !shouldParse(conn, today, date) {
				continue
			}

			task := TicketParseTask{
				OriginCode:      conn.OriginCode,
				DestinationCode: conn.DestinationCode,
				DepartureDate:   date,
				Priority:        computePriority(conn, date),
			}

			select {
			case <-throttle.C:
			case <-ctx.Done():
				return ctx.Err()
			}

			if err := s.pub.PublishTicketParse(ctx, task); err != nil {
				return fmt.Errorf("pub.PublishTicketParse (%s->%s %s): %w",
					conn.OriginCode, conn.DestinationCode, date.Format(time.DateOnly), err)
			}
			anyPublished = true
		}

		if anyPublished {
			if err := s.repo.MarkParsed(ctx, conn.OriginCode, conn.DestinationCode, today); err != nil {
				return fmt.Errorf("repo.MarkParsed (%s->%s): %w",
					conn.OriginCode, conn.DestinationCode, err)
			}
		}
	}

	return nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
