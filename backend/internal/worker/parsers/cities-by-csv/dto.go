package city

type Station struct {
	ID   string
	Name string
}

type StationRoute struct {
	Departure Station
	Arrival   Station
}

type GroupedStations map[string][]Station
