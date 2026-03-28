// Package dto holds the worker-side message types.
// These mirror the scheduler's DTOs by JSON shape but are defined independently
// so the worker never imports the scheduler package.
package dto

import "time"

// StationLoadTask is consumed once daily to trigger a city/station graph refresh.
type StationLoadTask struct {
	RequestedAt time.Time `json:"requested_at"`
	TopN        int       `json:"top_n"`
}

// TicketParseTask is consumed to fetch train prices for a single route + date.
type TicketParseTask struct {
	OriginCode      string    `json:"origin_code"`
	DestinationCode string    `json:"destination_code"`
	DepartureDate   time.Time `json:"departure_date"`
	Priority        int       `json:"priority"`
}
