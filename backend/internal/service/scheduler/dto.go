package scheduler

import "time"

// StationLoadTask triggers a full city/station graph refresh.
type StationLoadTask struct {
	ScheduledAt time.Time `json:"scheduled_at"` // when the task should be dispatched
	TopN        int       `json:"top_n"`
}

// TicketParseTask requests train prices for a single (origin, destination, date) triple.
type TicketParseTask struct {
	ScheduledAt     time.Time `json:"scheduled_at"` // when the task should be dispatched
	OriginCode      string    `json:"origin_code"`
	DestinationCode string    `json:"destination_code"`
	DepartureDate   time.Time `json:"departure_date"`
	Priority        int       `json:"priority"` // lower = higher priority
}
