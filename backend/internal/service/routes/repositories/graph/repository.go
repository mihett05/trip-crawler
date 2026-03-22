package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	"github.com/mihett05/trip-crawler/internal/service/routes/models"
)

type Repository struct {
	dg *dgraph.Client
}

func NewRepository(dg *dgraph.Client) *Repository {
	return &Repository{dg: dg}
}

func (r *Repository) InitSchema(ctx context.Context) error {
	op := &api.Operation{
		Schema: `
			city.name: string @index(term) .
			station.name: string @index(term) .
			station.transport_type: string @index(exact) .
			trip.external_id: string @index(exact) .
			trip.price: float @index(float) .
			trip.departure_at: datetime @index(hour) .
			trip.arrival_at: datetime .
			trip.transport_type: string @index(exact) .

			has_station: [uid] @reverse .
			departs: [uid] @reverse .
			destination: uid .

			type City {
				city.name
				has_station
			}

			type Station {
				station.name
				station.transport_type
				departs
			}

			type Trip {
				trip.external_id
				trip.price
				trip.departure_at
				trip.arrival_at
				trip.transport_type
				destination
			}
		`,
	}
	if err := r.dg.Client.Alter(ctx, op); err != nil {
		return fmt.Errorf("client.Alter: %w", err)
	}
	return nil
}

func (r *Repository) SaveTrip(ctx context.Context, trip *models.Trip) error {
	dto := &TripDTO{
		Uid:           trip.ID,
		ExternalID:    trip.ExternalID,
		Price:         trip.Price,
		DepartureAt:   trip.DepartureAt,
		ArrivalAt:     trip.ArrivalAt,
		TransportType: trip.TransportType,
		Type:          []string{"Trip"},
	}

	if trip.Destination != nil {
		dto.Destination = &StationDTO{Uid: trip.Destination.ID}
	}

	txn := r.dg.Client.NewTxn()
	defer txn.Discard(ctx)

	tripJSON, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal trip (ID: %s): json.Marshal: %w", trip.ExternalID, err)
	}

	mutation := &api.Mutation{
		SetJson:   tripJSON,
		CommitNow: true,
	}

	_, err = txn.Mutate(ctx, mutation)
	if err != nil {
		return fmt.Errorf("txn.Mutate: %w", err)
	}

	return nil
}

func (r *Repository) GetCityStations(ctx context.Context, cityName string) ([]models.Station, error) {
	query := `query all($name: string) {
		city(func: eq(city.name, $name)) {
			has_station {
				uid
				station.name
				station.transport_type
			}
		}
	}`

	resp, err := r.dg.Client.NewTxn().QueryWithVars(ctx, query, map[string]string{"$name": cityName})
	if err != nil {
		return nil, fmt.Errorf("txn.QueryWithVars (City: %s): %w", cityName, err)
	}

	var decode struct {
		City []struct {
			Stations []StationDTO `json:"has_station"`
		} `json:"city"`
	}

	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.City) == 0 {
		return nil, nil
	}

	stations := make([]models.Station, 0, len(decode.City[0].Stations))
	for _, s := range decode.City[0].Stations {
		stations = append(stations, models.Station{
			ID:            s.Uid,
			Name:          s.Name,
			TransportType: s.TransportType,
		})
	}

	return stations, nil
}

func (r *Repository) GetStationDepartures(ctx context.Context, stationID string) ([]models.Trip, error) {
	query := fmt.Sprintf(`{
		station(func: uid(%s)) {
			departs {
				uid
				trip.external_id
				trip.price
				trip.departure_at
				trip.arrival_at
				destination { uid station.name }
			}
		}
	}`, stationID)

	resp, err := r.dg.Client.NewTxn().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("txn.Query: %w", err)
	}

	var decode struct {
		Station []struct {
			Departs []TripDTO `json:"departs"`
		} `json:"station"`
	}

	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.Station) == 0 {
		return nil, nil
	}

	trips := make([]models.Trip, 0, len(decode.Station[0].Departs))
	for _, t := range decode.Station[0].Departs {
		trips = append(trips, models.Trip{
			ID:          t.Uid,
			ExternalID:  t.ExternalID,
			Price:       t.Price,
			DepartureAt: t.DepartureAt,
			ArrivalAt:   t.ArrivalAt,
		})
	}

	return trips, nil
}

func (r *Repository) GetNodeByID(ctx context.Context, uid string) ([]byte, error) {
	query := fmt.Sprintf(`{ node(func: uid(%s)) { uid dgraph.type expand(_all_) } }`, uid)
	resp, err := r.dg.Client.NewTxn().Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("txn.Query (UID: %s): %w", uid, err)
	}
	return resp.Json, nil
}

func (r *Repository) HasConnection(ctx context.Context, fromUID, toUID string) (bool, error) {
	query := fmt.Sprintf(`{
		check(func: uid(%s)) {
			departs @filter(uid_in(destination, %s)) { uid }
		}
	}`, fromUID, toUID)

	resp, err := r.dg.Client.NewTxn().Query(ctx, query)
	if err != nil {
		return false, fmt.Errorf("txn.Query: %w", err)
	}

	var decode struct {
		Check []struct {
			Departs []interface{} `json:"departs"`
		} `json:"check"`
	}

	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return false, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.Check) > 0 && len(decode.Check[0].Departs) > 0 {
		return true, nil
	}

	return false, nil
}
