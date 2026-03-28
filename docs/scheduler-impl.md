# Scheduler — Implementation Spec

## Package Structure

```
backend/internal/service/scheduler/
├── config.go      — Config struct (env vars)
├── dto.go         — NATS message DTOs (no parsers import)
├── scoring.go     — Tier classification + frequency matrix
├── repository.go  — DgraphRepository interface (scheduler-owned)
└── scheduler.go   — Scheduler struct + daily jobs
```

---

## Contracts

### DTOs (`dto.go`)

The scheduler defines its own message types. It **must not** import `internal/worker/parsers/`.

```go
package scheduler

import "time"

// StationLoadTask is published once daily to trigger a full city/station refresh.
type StationLoadTask struct {
    RequestedAt time.Time `json:"requested_at"`
    TopN        int       `json:"top_n"`
}

// TicketParseTask is published for each (route, date) pair the scoring
// matrix decides needs a refresh today.
type TicketParseTask struct {
    OriginCode      string    `json:"origin_code"`       // RZD 7-char station code
    DestinationCode string    `json:"destination_code"`
    DepartureDate   time.Time `json:"departure_date"`
    Priority        int       `json:"priority"`          // lower = higher priority
}
```

### Repository Interface (`repository.go`)

The scheduler reads connection metadata from Dgraph through its own interface.

```go
package scheduler

import (
    "context"
    "time"
)

type Connection struct {
    OriginCode      string
    DestinationCode string
    OriginPopulation int
    DestPopulation   int
    LastParsedAt    time.Time
    LastUsedAt      time.Time // updated by route builder on each routing call
}

type Repository interface {
    // GetTopConnections returns connections between the top-N cities by population.
    GetTopConnections(ctx context.Context, limit int) ([]Connection, error)
    // MarkParsed records the timestamp a connection was last enqueued for parsing.
    MarkParsed(ctx context.Context, originCode, destCode string, at time.Time) error
}
```

---

## Scoring (`scoring.go`)

### City Tier

```go
type Tier int

const (
    TierA Tier = iota // > 1 000 000
    TierB             // 500 000 – 1 000 000
    TierC             // 100 000 – 500 000
    TierD             // < 100 000
)

func tierFromPopulation(pop int) Tier {
    switch {
    case pop > 1_000_000:
        return TierA
    case pop > 500_000:
        return TierB
    case pop > 100_000:
        return TierC
    default:
        return TierD
    }
}
```

### Frequency Matrix

`parseInterval(origin, dest Tier, daysUntilDeparture int) int` returns how many days
must have passed since last parse before the task should be enqueued again.

| Route | 1–14 d | 15–45 d | 46–90 d |
|-------|--------|---------|---------|
| A↔A   | 1      | 2       | 4       |
| A↔B   | 1      | 3       | 7       |
| A↔C   | 2      | 5       | 14      |
| A↔D   | 3      | 7       | 21      |
| B↔B   | 1      | 4       | 10      |
| B↔C   | 2      | 7       | 14      |
| B↔D   | 3      | 10      | 21      |
| C↔C   | 3      | 10      | 21      |
| C↔D   | 5      | 14      | 30      |
| D↔D   | 7      | 21      | 60      |

Dates beyond 90 days are skipped entirely.

### Pruning (sleep mode)

A connection not used in any route result for **30 days** (`today - LastUsedAt > 30d`)
is considered sleeping. Its parse interval is overridden to **21 days** regardless of tiers.

```go
const sleepThresholdDays = 30
const sleepIntervalDays  = 21

func shouldParse(conn Connection, today, departureDate time.Time) bool {
    daysUntil := int(departureDate.Sub(today).Hours() / 24)
    if daysUntil < 1 || daysUntil > 90 {
        return false
    }

    oTier := tierFromPopulation(conn.OriginPopulation)
    dTier := tierFromPopulation(conn.DestPopulation)

    interval := parseInterval(oTier, dTier, daysUntil)
    if today.Sub(conn.LastUsedAt).Hours()/24 > sleepThresholdDays {
        interval = sleepIntervalDays
    }

    daysSinceLastParse := today.Sub(conn.LastParsedAt).Hours() / 24
    return daysSinceLastParse >= float64(interval)
}
```

---

## Scheduler (`scheduler.go`)

### Config

```go
type Config struct {
    TopNCities        int `env:"SCHEDULER_TOP_N"        envDefault:"100"`
    DaysAhead         int `env:"SCHEDULER_DAYS_AHEAD"   envDefault:"90"`
    PublishRatePerSec int `env:"SCHEDULER_PUBLISH_RATE" envDefault:"500"`
}
```

### Struct

```go
type Scheduler struct {
    cfg    Config
    repo   Repository
    // publisher is an interface for sending StationLoadTask / TicketParseTask;
    // the concrete implementation (NATS, in-memory for tests) is injected.
    pub    Publisher
}

// Publisher is the only external contract the scheduler needs.
type Publisher interface {
    PublishStationLoad(ctx context.Context, task StationLoadTask) error
    PublishTicketParse(ctx context.Context, task TicketParseTask) error
}

func New(cfg Config, repo Repository, pub Publisher) *Scheduler
```

### Daily Jobs

```
02:00  →  jobLoadStations
03:00  →  jobEnqueueTickets
```

#### `jobLoadStations`

Publishes a single `StationLoadTask{RequestedAt: now, TopN: cfg.TopNCities}`.

#### `jobEnqueueTickets`

```
1. connections ← repo.GetTopConnections(ctx, cfg.TopNCities)
2. today ← time.Now().Truncate(24h)
3. for each conn in connections:
     for daysAhead in 1..cfg.DaysAhead:
       date ← today + daysAhead
       if shouldParse(conn, today, date):
         priority ← computePriority(conn, date)   // lower daysAhead + higher tiers = lower value
         pub.PublishTicketParse(ctx, TicketParseTask{
           OriginCode:      conn.OriginCode,
           DestinationCode: conn.DestinationCode,
           DepartureDate:   date,
           Priority:        priority,
         })
         repo.MarkParsed(ctx, conn.OriginCode, conn.DestinationCode, today)
         // rate-limit: sleep 1s every cfg.PublishRatePerSec publishes
```

`computePriority` example — lower number means higher queue priority:

```go
func computePriority(conn Connection, departureDate time.Time) int {
    tierScore := int(tierFromPopulation(conn.OriginPopulation)) +
                 int(tierFromPopulation(conn.DestPopulation))   // 0 (A+A) .. 6 (D+D)
    daysUntil := int(departureDate.Sub(time.Now()).Hours() / 24)
    return tierScore*100 + daysUntil
}
```
