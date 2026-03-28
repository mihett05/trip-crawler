package scheduler

import (
	"testing"
	"time"
)

func defaultConfig() Config {
	return Config{
		TopNCities:     100,
		DaysAhead:      90,
		BucketSizeMin:  5,
		BucketSizeMax:  10,
		BucketPauseMin: 15 * time.Second,
		BucketPauseMax: 30 * time.Second,
	}
}

func neverParsedConn(originCode, destCode string, originPop, destPop int) Connection {
	return Connection{
		OriginCode:       originCode,
		DestinationCode:  destCode,
		OriginPopulation: originPop,
		DestPopulation:   destPop,
		LastParsedAt:     time.Time{},
		LastUsedAt:       time.Now(),
	}
}

// --- GenerateStationTask ---

func TestGenerateStationTask_TopN(t *testing.T) {
	s := New(Config{TopNCities: 42, DaysAhead: 90, BucketSizeMin: 5, BucketSizeMax: 10,
		BucketPauseMin: 15 * time.Second, BucketPauseMax: 30 * time.Second})
	task := s.GenerateStationTask(time.Now())
	if task.TopN != 42 {
		t.Errorf("TopN = %d, want 42", task.TopN)
	}
}

func TestGenerateStationTask_ScheduledAt(t *testing.T) {
	now := time.Now()
	s := New(defaultConfig())
	task := s.GenerateStationTask(now)
	want := truncateToDay(now)
	if !task.ScheduledAt.Equal(want) {
		t.Errorf("ScheduledAt = %v, want %v", task.ScheduledAt, want)
	}
}

// --- GenerateTicketTasks ---

func TestGenerateTicketTasks_EmptyConnectionsReturnsEmpty(t *testing.T) {
	s := New(defaultConfig())
	tasks := s.GenerateTicketTasks(time.Now(), nil)
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for no connections, got %d", len(tasks))
	}
}

func TestGenerateTicketTasks_NeverParsedReturnsAllDays(t *testing.T) {
	s := New(defaultConfig())
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})
	if len(tasks) != defaultConfig().DaysAhead {
		t.Errorf("expected %d tasks, got %d", defaultConfig().DaysAhead, len(tasks))
	}
}

func TestGenerateTicketTasks_RecentlyParsedSkipsDates(t *testing.T) {
	s := New(defaultConfig())
	today := truncateToDay(time.Now())
	conn := Connection{
		OriginCode: "0000001", DestinationCode: "0000002",
		OriginPopulation: 2_000_000, DestPopulation: 2_000_000,
		LastParsedAt: today, LastUsedAt: today,
	}
	tasks := s.GenerateTicketTasks(today, []Connection{conn})
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks for recently parsed connection, got %d", len(tasks))
	}
}

func TestGenerateTicketTasks_DaysAheadRespected(t *testing.T) {
	cfg := defaultConfig()
	cfg.DaysAhead = 10
	s := New(cfg)
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	today := truncateToDay(time.Now())

	tasks := s.GenerateTicketTasks(today, []Connection{conn})

	if len(tasks) != 10 {
		t.Errorf("expected 10 tasks (DaysAhead=10), got %d", len(tasks))
	}
	for _, task := range tasks {
		daysOut := int(truncateToDay(task.DepartureDate).Sub(today).Hours() / 24)
		if daysOut < 1 || daysOut > 10 {
			t.Errorf("DepartureDate is %d days out, expected 1–10", daysOut)
		}
	}
}

func TestGenerateTicketTasks_MultipleConnections(t *testing.T) {
	s := New(defaultConfig())
	conns := []Connection{
		neverParsedConn("1000001", "1000002", 2_000_000, 2_000_000),
		neverParsedConn("2000001", "2000002", 750_000, 750_000),
	}
	tasks := s.GenerateTicketTasks(time.Now(), conns)

	seen := make(map[string]bool)
	for _, task := range tasks {
		seen[task.OriginCode+"|"+task.DestinationCode] = true
	}
	if len(seen) != 2 {
		t.Errorf("expected tasks for 2 distinct connections, got %d", len(seen))
	}
}

// --- Bucketing ---

func TestGenerateTicketTasks_SortedByPriority(t *testing.T) {
	s := New(defaultConfig())
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})

	for i := 1; i < len(tasks); i++ {
		if tasks[i].Priority < tasks[i-1].Priority {
			t.Fatalf("tasks not sorted: tasks[%d].Priority=%d < tasks[%d].Priority=%d",
				i, tasks[i].Priority, i-1, tasks[i-1].Priority)
		}
	}
}

func TestGenerateTicketTasks_ScheduledAtNonDecreasing(t *testing.T) {
	s := New(defaultConfig())
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})

	for i := 1; i < len(tasks); i++ {
		if tasks[i].ScheduledAt.Before(tasks[i-1].ScheduledAt) {
			t.Fatalf("ScheduledAt decreased at index %d: %v < %v",
				i, tasks[i].ScheduledAt, tasks[i-1].ScheduledAt)
		}
	}
}

func TestGenerateTicketTasks_ScheduledAtNotBeforeToday(t *testing.T) {
	today := truncateToDay(time.Now())
	s := New(defaultConfig())
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(today, []Connection{conn})

	for _, task := range tasks {
		if task.ScheduledAt.Before(today) {
			t.Fatalf("ScheduledAt %v is before today %v", task.ScheduledAt, today)
		}
	}
}

func TestGenerateTicketTasks_MultipleDistinctScheduledAt(t *testing.T) {
	// With 90 tasks and bucket size max 10, we must have at least 9 buckets → 9+ distinct times.
	s := New(defaultConfig())
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})

	distinct := make(map[time.Time]struct{})
	for _, task := range tasks {
		distinct[task.ScheduledAt] = struct{}{}
	}
	minBuckets := defaultConfig().DaysAhead / defaultConfig().BucketSizeMax
	if len(distinct) < minBuckets {
		t.Errorf("expected at least %d distinct ScheduledAt values, got %d", minBuckets, len(distinct))
	}
}

func TestGenerateTicketTasks_BucketSizeWithinBounds(t *testing.T) {
	cfg := defaultConfig()
	s := New(cfg)
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})

	// Group consecutive tasks by ScheduledAt and check each group size.
	i := 0
	for i < len(tasks) {
		j := i + 1
		for j < len(tasks) && tasks[j].ScheduledAt.Equal(tasks[i].ScheduledAt) {
			j++
		}
		bucketSize := j - i
		// Last bucket may be smaller than BucketSizeMin.
		if j < len(tasks) && bucketSize < cfg.BucketSizeMin {
			t.Errorf("bucket at index %d has size %d, want >= %d", i, bucketSize, cfg.BucketSizeMin)
		}
		if bucketSize > cfg.BucketSizeMax {
			t.Errorf("bucket at index %d has size %d, want <= %d", i, bucketSize, cfg.BucketSizeMax)
		}
		i = j
	}
}

func TestGenerateTicketTasks_PauseBetweenBucketsWithinBounds(t *testing.T) {
	cfg := defaultConfig()
	s := New(cfg)
	conn := neverParsedConn("0000001", "0000002", 2_000_000, 2_000_000)
	tasks := s.GenerateTicketTasks(time.Now(), []Connection{conn})

	// Collect distinct bucket timestamps in order.
	var bucketTimes []time.Time
	for i, task := range tasks {
		if i == 0 || !task.ScheduledAt.Equal(tasks[i-1].ScheduledAt) {
			bucketTimes = append(bucketTimes, task.ScheduledAt)
		}
	}

	for i := 1; i < len(bucketTimes); i++ {
		pause := bucketTimes[i].Sub(bucketTimes[i-1])
		if pause < cfg.BucketPauseMin || pause > cfg.BucketPauseMax {
			t.Errorf("pause between bucket %d and %d is %v, want [%v, %v]",
				i-1, i, pause, cfg.BucketPauseMin, cfg.BucketPauseMax)
		}
	}
}

func TestGenerateTicketTasks_SleepModeReducesTasks(t *testing.T) {
	s := New(defaultConfig())
	today := truncateToDay(time.Now())
	parsedAt := today.AddDate(0, 0, -5)

	active := Connection{
		OriginCode: "0000001", DestinationCode: "0000002",
		OriginPopulation: 2_000_000, DestPopulation: 2_000_000,
		LastParsedAt: parsedAt, LastUsedAt: today,
	}
	sleeping := Connection{
		OriginCode: "0000001", DestinationCode: "0000002",
		OriginPopulation: 2_000_000, DestPopulation: 2_000_000,
		LastParsedAt: parsedAt,
		LastUsedAt:   today.AddDate(0, 0, -31),
	}

	activeTasks := s.GenerateTicketTasks(today, []Connection{active})
	sleepingTasks := s.GenerateTicketTasks(today, []Connection{sleeping})

	if len(sleepingTasks) >= len(activeTasks) {
		t.Errorf("sleeping connection (%d tasks) should produce fewer tasks than active (%d)",
			len(sleepingTasks), len(activeTasks))
	}
}
