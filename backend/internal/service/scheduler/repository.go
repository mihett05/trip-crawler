package scheduler

import (
	"context"
	"time"
)

// Connection represents a known direct route between two stations
// along with the metadata needed for scoring decisions.
type Connection struct {
	OriginCode       string
	DestinationCode  string
	OriginPopulation int
	DestPopulation   int
	LastParsedAt     time.Time // when we last enqueued parse tasks for this connection
	LastUsedAt       time.Time // when the route builder last used this connection
}

// Repository is the only Dgraph contract the scheduler needs.
type Repository interface {
	// GetTopConnections returns connections between the top-limit cities by population.
	GetTopConnections(ctx context.Context, limit int) ([]Connection, error)
	// MarkParsed records today as the last time we enqueued tasks for this connection.
	MarkParsed(ctx context.Context, originCode, destCode string, at time.Time) error
}
