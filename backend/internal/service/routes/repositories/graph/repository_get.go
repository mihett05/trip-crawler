package graph

import (
	"context"
	"encoding/json"
	"fmt"
)

func (r *Repository) GetAllCities(ctx context.Context) ([]string, error) {
	query := `{
		cities(func: type(City)) {
		    uid
		    city.name
		}
	}`

	resp, err := r.dg.Client.NewTxn().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("txn.Query: %w", err)
	}

	var decode []CityDTO
	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	cities := make([]string, 0, len(decode))
	for _, city := range decode {
		cities = append(cities, city.Name)
	}

	return cities, nil
}
