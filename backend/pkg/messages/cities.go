package messages

type City struct {
	ID          string      `json:"id"` // uid города, задаётся парсером для ссылки на него
	Name        string      `json:"name"`
	Population  int64       `json:"population"`
	Coordinates Coordinates `json:"coordinates"`
	Stations    []Station   `json:"stations"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CitiesRequested struct {
	TopCities int `json:"top_cities"`
}

type CitiesParsed struct {
	Cities []City `json:"cities"`
}
