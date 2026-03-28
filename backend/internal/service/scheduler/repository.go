package scheduler

import "time"

// Connection represents a known direct route between two stations
// along with the metadata needed for scoring decisions.
// The caller is responsible for fetching connections and tracking LastParsedAt.
type Connection struct {
	OriginCode       string
	DestinationCode  string
	OriginPopulation int
	DestPopulation   int
	LastParsedAt     time.Time // when tasks were last generated for this connection
	LastUsedAt       time.Time // when the route builder last used this connection
}
