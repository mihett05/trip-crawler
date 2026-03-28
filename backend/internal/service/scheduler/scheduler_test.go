package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// --- mocks ---

type mockPublisher struct {
	mu             sync.Mutex
	stationCalls   []StationLoadTask
	ticketCalls    []TicketParseTask
	stationErr     error
	ticketErr      error
}

func (m *mockPublisher) PublishStationLoad(_ context.Context, task StationLoadTask) error {
	if m.stationErr != nil {
		return m.stationErr
	}
	m.mu.Lock()
	m.stationCalls = append(m.stationCalls, task)
	m.mu.Unlock()
	return nil
}

func (m *mockPublisher) PublishTicketParse(_ context.Context, task TicketParseTask) error {
	if m.ticketErr != nil {
		return m.ticketErr
	}
	m.mu.Lock()
	m.ticketCalls = append(m.ticketCalls, task)
	m.mu.Unlock()
	return nil
}

type mockRepository struct {
	connections []Connection
	marked      []struct{ origin, dest string }
	getErr      error
	markErr     error
}

func (m *mockRepository) GetTopConnections(_ context.Context, _ int) ([]Connection, error) {
	return m.connections, m.getErr
}

func (m *mockRepository) MarkParsed(_ context.Context, originCode, destCode string, _ time.Time) error {
	if m.markErr != nil {
		return m.markErr
	}
	m.marked = append(m.marked, struct{ origin, dest string }{originCode, destCode})
	return nil
}

// --- helpers ---

func defaultConfig() Config {
	return Config{
		TopNCities:          100,
		DaysAhead:           90,
		PublishRatePerSec:   10_000, // high rate so tests aren't slow
		StationLoadHour:     2,
		StationLoadMinute:   0,
		TicketEnqueueHour:   3,
		TicketEnqueueMinute: 0,
	}
}

func neverParsedConn(originPop, destPop int) Connection {
	return Connection{
		OriginCode:       "0000001",
		DestinationCode:  "0000002",
		OriginPopulation: originPop,
		DestPopulation:   destPop,
		LastParsedAt:     time.Time{},
		LastUsedAt:       time.Now(),
	}
}

// --- nextTrigger ---

func TestNextTrigger_FutureToday(t *testing.T) {
	now := time.Now()
	// Pick a time far enough in the future (23:59) to avoid flakiness
	next := nextTrigger(23, 59)
	if !next.After(now) {
		t.Error("nextTrigger should return a future time")
	}
}

func TestNextTrigger_PastTimeTodayReturnsTomorrow(t *testing.T) {
	// Hour 0, minute 0 has definitely already passed today.
	next := nextTrigger(0, 0)
	tomorrow := truncateToDay(time.Now()).Add(24 * time.Hour)
	if !next.Equal(tomorrow) {
		t.Errorf("nextTrigger(0,0) should return tomorrow 00:00, got %v", next)
	}
}

// --- jobLoadStations ---

func TestJobLoadStations_PublishesTask(t *testing.T) {
	pub := &mockPublisher{}
	s := New(defaultConfig(), &mockRepository{}, pub)

	if err := s.jobLoadStations(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pub.stationCalls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(pub.stationCalls))
	}
	if pub.stationCalls[0].TopN != defaultConfig().TopNCities {
		t.Errorf("TopN = %d, want %d", pub.stationCalls[0].TopN, defaultConfig().TopNCities)
	}
}

func TestJobLoadStations_PropagatesPublisherError(t *testing.T) {
	pub := &mockPublisher{stationErr: errors.New("nats down")}
	s := New(defaultConfig(), &mockRepository{}, pub)

	err := s.jobLoadStations(context.Background())
	if err == nil {
		t.Error("expected error from publisher, got nil")
	}
}

// --- jobEnqueueTickets ---

func TestJobEnqueueTickets_NeverParsedPublishesAllEligibleDates(t *testing.T) {
	pub := &mockPublisher{}
	repo := &mockRepository{
		connections: []Connection{neverParsedConn(2_000_000, 2_000_000)},
	}
	cfg := defaultConfig()
	s := New(cfg, repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A↔A never parsed: should publish for all 90 days.
	if len(pub.ticketCalls) != cfg.DaysAhead {
		t.Errorf("expected %d tasks, got %d", cfg.DaysAhead, len(pub.ticketCalls))
	}
}

func TestJobEnqueueTickets_RecentlyParsedSkipsDates(t *testing.T) {
	today := truncateToDay(time.Now())
	pub := &mockPublisher{}
	// A↔A mid interval = 2; last parsed today → near dates (interval=1, elapsed=0) skipped.
	conn := Connection{
		OriginCode:       "0000001",
		DestinationCode:  "0000002",
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     today,
		LastUsedAt:       today,
	}
	repo := &mockRepository{connections: []Connection{conn}}
	s := New(defaultConfig(), repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No interval elapsed → no tasks at all.
	if len(pub.ticketCalls) != 0 {
		t.Errorf("expected 0 tasks for recently parsed connection, got %d", len(pub.ticketCalls))
	}
}

func TestJobEnqueueTickets_MarkParsedCalledOncePerConnection(t *testing.T) {
	pub := &mockPublisher{}
	repo := &mockRepository{
		connections: []Connection{
			neverParsedConn(2_000_000, 2_000_000),
			neverParsedConn(750_000, 750_000),
		},
	}
	s := New(defaultConfig(), repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.marked) != 2 {
		t.Errorf("MarkParsed should be called once per connection, got %d calls", len(repo.marked))
	}
}

func TestJobEnqueueTickets_MarkParsedNotCalledWhenNothingPublished(t *testing.T) {
	today := truncateToDay(time.Now())
	pub := &mockPublisher{}
	// Parsed today → nothing to publish.
	conn := Connection{
		OriginCode:       "0000001",
		DestinationCode:  "0000002",
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     today,
		LastUsedAt:       today,
	}
	repo := &mockRepository{connections: []Connection{conn}}
	s := New(defaultConfig(), repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.marked) != 0 {
		t.Errorf("MarkParsed should not be called when nothing was published, got %d calls", len(repo.marked))
	}
}

func TestJobEnqueueTickets_TaskFieldsAreCorrect(t *testing.T) {
	pub := &mockPublisher{}
	conn := neverParsedConn(2_000_000, 2_000_000)
	repo := &mockRepository{connections: []Connection{conn}}
	s := New(defaultConfig(), repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	today := truncateToDay(time.Now())
	for _, task := range pub.ticketCalls {
		if task.OriginCode != conn.OriginCode {
			t.Errorf("OriginCode = %q, want %q", task.OriginCode, conn.OriginCode)
		}
		if task.DestinationCode != conn.DestinationCode {
			t.Errorf("DestinationCode = %q, want %q", task.DestinationCode, conn.DestinationCode)
		}
		dep := truncateToDay(task.DepartureDate)
		if !dep.After(today) {
			t.Errorf("DepartureDate %v should be strictly after today %v", dep, today)
		}
	}
}

func TestJobEnqueueTickets_PriorityOrdering(t *testing.T) {
	pub := &mockPublisher{}
	conn := neverParsedConn(2_000_000, 2_000_000)
	repo := &mockRepository{connections: []Connection{conn}}
	s := New(defaultConfig(), repo, pub)

	if err := s.jobEnqueueTickets(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Near dates should have lower priority numbers than far dates.
	if len(pub.ticketCalls) < 2 {
		t.Skip("not enough tasks to compare")
	}
	first := pub.ticketCalls[0]
	last := pub.ticketCalls[len(pub.ticketCalls)-1]
	if first.Priority >= last.Priority {
		t.Errorf("first task priority (%d) should be lower than last (%d)", first.Priority, last.Priority)
	}
}

func TestJobEnqueueTickets_PropagatesRepoError(t *testing.T) {
	pub := &mockPublisher{}
	repo := &mockRepository{getErr: errors.New("dgraph unreachable")}
	s := New(defaultConfig(), repo, pub)

	err := s.jobEnqueueTickets(context.Background())
	if err == nil {
		t.Error("expected error from repository, got nil")
	}
}

func TestJobEnqueueTickets_PropagatesPublishError(t *testing.T) {
	pub := &mockPublisher{ticketErr: errors.New("nats down")}
	repo := &mockRepository{connections: []Connection{neverParsedConn(2_000_000, 2_000_000)}}
	s := New(defaultConfig(), repo, pub)

	err := s.jobEnqueueTickets(context.Background())
	if err == nil {
		t.Error("expected error from publisher, got nil")
	}
}

func TestJobEnqueueTickets_RespectsContextCancellation(t *testing.T) {
	pub := &mockPublisher{}
	// Many connections so the loop takes a while without a high publish rate
	connections := make([]Connection, 50)
	for i := range connections {
		connections[i] = neverParsedConn(2_000_000, 2_000_000)
	}
	repo := &mockRepository{connections: connections}

	cfg := defaultConfig()
	cfg.PublishRatePerSec = 1 // very slow → context cancels before finishing
	s := New(cfg, repo, pub)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.jobEnqueueTickets(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

// --- Run ---

func TestRun_ReturnsOnContextCancel(t *testing.T) {
	s := New(defaultConfig(), &mockRepository{}, &mockPublisher{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Error("Run did not return after context cancellation")
	}
}
