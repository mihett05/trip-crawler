package city

import (
	"bytes"
	"encoding/csv"
	"io"

	"github.com/mihett05/trip-crawler/internal/worker/parsers/cities-by-csv/static"
)

// GetRoutesFromCSV читает CSV и возвращает список маршрутов (откуда -> куда)
func GetRoutesFromCSV() ([]StationRoute, error) {
	bytesReader := bytes.NewReader(static.Data)

	reader := csv.NewReader(bytesReader)
	reader.Comma = ';'

	if _, err := reader.Read(); err != nil {
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
