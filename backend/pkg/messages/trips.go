package messages

type Trip struct {
	ExternalID           string   `json:"external_id"`
	DepartStationID      string   `json:"depart_station_id"`
	DestinationStationID string   `json:"destination_station_id"`
	DepartureAtTimestamp int64    `json:"departure_at_timestamp"`
	ArrivalAtTimestamp   int64    `json:"arrival_at_timestamp"`
	TransportType        string   `json:"transport_type"`
	Availability         []Ticket `json:"availability"`
}

type Ticket struct {
	Type            string  `json:"type"` // тип билетов: купе, плацкарт, бизнес-класс
	Price           float64 `json:"price"`
	AvailableAmount int64   `json:"available_amount"` // количество доступных билетов данного типа
	TotalAmount     int64   `json:"total_amount"`     // количество всего билетов данного типа
}

type TripRequested struct {
	DepartStation        string `json:"depart"`
	DepartStationID      string `json:"depart_station_id"`
	DepartureAtTimestamp int64
	DestinationStation   string `json:"destination"`
	DestinationStationID string `json:"destination_station_id"`
}

type TripParsed struct {
	Trip Trip `json:"trip"`
}
