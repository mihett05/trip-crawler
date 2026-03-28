package scheduler

import "time"

// StationLoadTask is published once daily to trigger a full city/station refresh.
type StationLoadTask struct {
	RequestedAt time.Time `json:"requested_at"`
	TopN        int       `json:"top_n"`
}

// TicketParseTask is published for each (origin, destination, date) 
// triple that the scoring matrix decides needs a refresh today.
type TicketParseTask struct {
	OriginCode      string    `json:"origin_code"`
	DestinationCode string    `json:"destination_code"`
	DepartureDate   time.Time `json:"departure_date"`
	Priority        int       `json:"priority"` // lower = higher priority
}
