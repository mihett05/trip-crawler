package messages

type Station struct {
	ID            string   `json:"id"` // uid станции, который парсер сам задаёт, чтобы ссылаться на другие станции
	Name          string   `json:"name"`
	CityID        string   `json:"city_id"` // uid города из городов
	TransportType string   `json:"type"`
	Stations      []string `json:"stations"` // Связи с другими станциями, через их
}
