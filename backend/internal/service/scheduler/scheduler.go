package scheduler

import (
	"math/rand/v2"
	"sort"
	"time"
)

// Scheduler is a pure task generator. It has no I/O dependencies — it only
// applies the scoring matrix to the connections provided by the caller and
// returns the tasks that should be executed on the given day.
//
// Typical usage (called once a day by the owning service):
//
//	tasks := s.GenerateTicketTasks(time.Now(), connections)
//	repo.SaveTasks(ctx, tasks)
type Scheduler struct {
	cfg Config
}

func New(cfg Config) *Scheduler {
	return &Scheduler{cfg: cfg}
}

// GenerateStationTask returns the daily task that triggers a full city/station
// graph refresh. Always returns exactly one task scheduled for the start of today.
func (s *Scheduler) GenerateStationTask(today time.Time) StationLoadTask {
	return StationLoadTask{
		ScheduledAt: truncateToDay(today),
		TopN:        s.cfg.TopNCities,
	}
}

// GenerateTicketTasks returns all ticket-parse tasks for today.
//
// Steps:
//  1. Collect every (connection, date) pair the scoring matrix selects.
//  2. Sort by Priority ascending so the most important tasks run first.
//  3. Split into buckets of random size [BucketSizeMin, BucketSizeMax].
//     All tasks in the same bucket share a ScheduledAt timestamp.
//  4. Between consecutive buckets add a random pause [BucketPauseMin, BucketPauseMax].
//
// The result is a steady trickle of small bursts spread across the day,
// keeping the request rate well below RZD's ban threshold.
func (s *Scheduler) GenerateTicketTasks(today time.Time, connections []Connection) []TicketParseTask {
	today = truncateToDay(today)

	var tasks []TicketParseTask
	for _, conn := range connections {
		for daysAhead := 1; daysAhead <= s.cfg.DaysAhead; daysAhead++ {
			date := today.AddDate(0, 0, daysAhead)
			if !shouldParse(conn, today, date) {
				continue
			}
			tasks = append(tasks, TicketParseTask{
				OriginCode:      conn.OriginCode,
				DestinationCode: conn.DestinationCode,
				DepartureDate:   date,
				Priority:        computePriority(conn, date),
			})
		}
	}

	if len(tasks) == 0 {
		return tasks
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Priority < tasks[j].Priority
	})

	s.assignScheduledAt(tasks, today)

	return tasks
}

// assignScheduledAt walks the sorted task slice, groups tasks into random-sized
// buckets, and stamps each bucket with the same ScheduledAt.
func (s *Scheduler) assignScheduledAt(tasks []TicketParseTask, start time.Time) {
	cursor := start
	for i := 0; i < len(tasks); {
		bucketSize := randBetween(s.cfg.BucketSizeMin, s.cfg.BucketSizeMax)
		end := min(i+bucketSize, len(tasks))

		for j := i; j < end; j++ {
			tasks[j].ScheduledAt = cursor
		}

		i = end
		cursor = cursor.Add(randDuration(s.cfg.BucketPauseMin, s.cfg.BucketPauseMax))
	}
}

func randBetween(lo, hi int) int {
	if lo >= hi {
		return lo
	}
	return lo + rand.IntN(hi-lo+1)
}

func randDuration(lo, hi time.Duration) time.Duration {
	if lo >= hi {
		return lo
	}
	return lo + time.Duration(rand.Int64N(int64(hi-lo+1)))
}
