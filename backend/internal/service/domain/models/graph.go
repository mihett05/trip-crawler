package models

import "time"

type City struct {
	Uid      string     `json:"uid,omitempty"`
	Type     []string   `json:"dgraph.type,omitempty"`
	Name     string     `json:"city.name,omitempty"`
	Stations []*Station `json:"has_station,omitempty"`
}

type Station struct {
	Uid           string   `json:"uid,omitempty"`
	Type          []string `json:"dgraph.type,omitempty"`
	Name          string   `json:"station.name,omitempty"`
	TransportType string   `json:"station.transport_type,omitempty"`
	Departs       []*Trip  `json:"departs,omitempty"`
}

type Trip struct {
	Uid           string    `json:"uid,omitempty"`
	Type          []string  `json:"dgraph.type,omitempty"`
	ExternalID    string    `json:"trip.external_id,omitempty"`
	Price         float64   `json:"trip.price,omitempty"`
	DepartureAt   time.Time `json:"trip.departure_at,omitempty"`
	ArrivalAt     time.Time `json:"trip.arrival_at,omitempty"`
	TransportType string    `json:"trip.transport_type,omitempty"`
	Destination   *Station  `json:"destination,omitempty"`
}
