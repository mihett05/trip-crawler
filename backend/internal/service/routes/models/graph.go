package models

import ("time"

		"github.com/mihett05/trip-crawler/pkg/messages"
)

type City struct {
	ID        string
	Name      string
	Latitude  float64
	Longitude float64
	Stations  []*Station
}

type Station struct {
	ID                  string
	ExternalID          string
	Name                string
	TransportType       string
	ConnectedStationsID []string
	Departs             []*Trip
}

type Trip struct {
	ID            string
	ExternalID    string
	Price         float64
	DepartureAt   time.Time
	ArrivalAt     time.Time
	TransportType string
	Tickets       []messages.Ticket
	Destination   *Station
}
