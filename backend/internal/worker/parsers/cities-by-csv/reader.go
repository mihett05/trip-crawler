package city

import (
	"encoding/csv"
	"io"
	"os"
)

// GetRoutesFromCSV читает CSV и возвращает список маршрутов (откуда -> куда)
func GetRoutesFromCSV(filePath string) ([]StationRoute, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	if _, err = reader.Read(); err != nil {
		return nil, err
	}

	var routes []StationRoute

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		route := StationRoute{
			Departure: Station{
				ID:   record[0],
				Name: record[1],
			},
			Arrival: Station{
				ID:   record[2],
				Name: record[3],
			},
		}

		routes = append(routes, route)
	}

	return routes, nil
}
