package models

import "time"

type City struct {
	ID       string
	Name     string
	Stations []*Station
}

type Station struct {
	ID            string
	Name          string
	TransportType string
	Departs       []*Trip
}

type Trip struct {
	ID            string
	ExternalID    string
	Price         float64
	DepartureAt   time.Time
	ArrivalAt     time.Time
	TransportType string
	Destination   *Station
}
