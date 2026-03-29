package graph

import "time"

type CityDTO struct {
	Uid      string       `json:"uid,omitempty"`
	Name     string       `json:"city.name,omitempty"`
	Location GeoJSON      `json:"city.location,omitempty"`
	Stations []StationDTO `json:"has_station,omitempty"`
	Type     []string     `json:"dgraph.type,omitempty"`
}

type GeoJSON struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type TicketDTO struct {
	Uid   string  `json:"uid,omitempty"`
	Type  string  `json:"ticket.type,omitempty"`
	Price float64 `json:"ticket.price,omitempty"`
	Count int     `json:"ticket.count,omitempty"`
}

type StationDTO struct {
	Uid           string     `json:"uid,omitempty"`
	Type          []string   `json:"dgraph.type,omitempty"`
	Name          string     `json:"station.name,omitempty"`
	TransportType string     `json:"station.transport_type,omitempty"`
	Departs       []*TripDTO `json:"departs,omitempty"`
	City          *CityDTO   `json:"city_info,omitempty"`
}

type TripDTO struct {
	Uid           string       `json:"uid,omitempty"`
	Type          []string     `json:"dgraph.type,omitempty"`
	ExternalID    string       `json:"trip.external_id,omitempty"`
	Price         float64      `json:"trip.price,omitempty"`
	DepartureAt   time.Time    `json:"trip.departure_at,omitempty"`
	ArrivalAt     time.Time    `json:"trip.arrival_at,omitempty"`
	TransportType string       `json:"trip.transport_type,omitempty"`
	Tickets       []*TicketDTO `json:"has_ticket,omitempty"`
	Destination   *StationDTO  `json:"destination,omitempty"`
}
