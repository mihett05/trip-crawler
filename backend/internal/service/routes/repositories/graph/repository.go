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
			city.location: geo @index(geo) .
			city.updated_at: datetime .
			station.name: string @index(term) .
			station.transport_type: string @index(term) .
			trip.external_id: string @index(exact) .
			trip.departure_at: datetime @index(hour) .
			trip.arrival_at: datetime .
			trip.price: float @index(float) .
			ticket.type: string .
			ticket.price: float @index(float) .
			ticket.count: int .

			has_station: [uid] @reverse .
			departs: [uid] @reverse .
			has_ticket: [uid] .
			destination: uid .

			type City {
				city.name
				city.location
				city.updated_at
				has_station
			}

			type Station {
				station.name
				station.transport_type
				departs
			}

			type Trip {
				trip.external_id
				trip.departure_at
				trip.arrival_at
				trip.price
				has_ticket
				destination
			}

			type Ticket {
				ticket.type
				ticket.price
				ticket.count
			}
		`,
	}
	if err := r.dg.Client.Alter(ctx, op); err != nil {
		return fmt.Errorf("client.Alter: %w", err)
	}
	return nil
}

func (r *Repository) SaveTrip(ctx context.Context, trip *models.Trip) error {
	query := fmt.Sprintf(`{
		v as var(func: eq(trip.external_id, "%s"))
	}`, trip.ExternalID)

	tripDTO := &TripDTO{
		Uid:         "uid(v)",
		ExternalID:  trip.ExternalID,
		DepartureAt: trip.DepartureAt,
		ArrivalAt:   trip.ArrivalAt,
		Type:        []string{"Trip"},
		Price:       trip.Price,
	}

	if trip.Destination != nil {
		tripDTO.Destination = &StationDTO{Uid: trip.Destination.ID}
	}

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	tripJSON, err := json.Marshal(tripDTO)
	if err != nil {
		return fmt.Errorf("marshal trip (ID: %s): json.Marshal: %w", trip.ExternalID, err)
	}

	mutation := &api.Mutation{
		SetJson: tripJSON,
	}

	request := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mutation},
		CommitNow: true,
	}

	_, err = transaction.Do(ctx, request)
	if err != nil {
		return fmt.Errorf("transaction.Do: %w", err)
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

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	response, err := transaction.QueryWithVars(ctx, query, map[string]string{"$name": cityName})
	if err != nil {
		return nil, fmt.Errorf("transaction.QueryWithVars (City: %s): %w", cityName, err)
	}

	var decode struct {
		City []struct {
			Stations []StationDTO `json:"has_station"`
		} `json:"city"`
	}

	if err := json.Unmarshal(response.Json, &decode); err != nil {
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

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	response, err := transaction.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("transaction.Query: %w", err)
	}

	var decode struct {
		Station []struct {
			Departs []TripDTO `json:"departs"`
		} `json:"station"`
	}

	if err := json.Unmarshal(response.Json, &decode); err != nil {
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

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	response, err := transaction.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("transaction.Query (UID: %s): %w", uid, err)
	}
	return response.Json, nil
}

func (r *Repository) HasConnection(ctx context.Context, fromUID, toUID string) (bool, error) {
	query := fmt.Sprintf(`{
		check(func: uid(%s)) {
			departs @filter(uid_in(destination, %s)) { uid }
		}
	}`, fromUID, toUID)

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	response, err := transaction.Query(ctx, query)
	if err != nil {
		return false, fmt.Errorf("transaction.Query: %w", err)
	}

	var decode struct {
		Check []struct {
			Departs []interface{} `json:"departs"`
		} `json:"check"`
	}

	if err := json.Unmarshal(response.Json, &decode); err != nil {
		return false, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.Check) > 0 && len(decode.Check[0].Departs) > 0 {
		return true, nil
	}

	return false, nil
}

func (r *Repository) HasCityConnection(ctx context.Context, fromCity, toCity string) (bool, error) {
	query := `query all($from: string, $to: string) {
		from_stations as var(func: eq(city.name, $from)) {
			has_station
		}
		to_stations as var(func: eq(city.name, $to)) {
			has_station
		}
		trips(func: uid(from_stations)) {
			departs @filter(uid_in(destination, to_stations)) {
				uid
			}
		}
	}`

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	response, err := transaction.QueryWithVars(ctx, query, map[string]string{
		"$from": fromCity,
		"$to":   toCity,
	})
	if err != nil {
		return false, fmt.Errorf("transaction.QueryWithVars: %w", err)
	}

	var decode struct {
		Trips []struct {
			Departs []interface{} `json:"departs"`
		} `json:"trips"`
	}

	if err := json.Unmarshal(response.Json, &decode); err != nil {
		return false, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.Trips) > 0 && len(decode.Trips[0].Departs) > 0 {
		return true, nil
	}

	return false, nil
}

func (r *Repository) SaveCity(ctx context.Context, city *models.City) error {
	query := fmt.Sprintf(`{
		v as var(func: eq(city.name, "%s"))
	}`, city.Name)

	stationsDTO := make([]StationDTO, 0, len(city.Stations))
	for _, station := range city.Stations {
		stationsDTO = append(stationsDTO, StationDTO{
			Uid:           fmt.Sprintf("_:station_%s", station.Name),
			Name:          station.Name,
			TransportType: station.TransportType,
			Type:          []string{"Station"},
		})
	}

	cityDTO := &CityDTO{
		Uid:  "uid(v)",
		Name: city.Name,
		Type: []string{"City"},
		Location: GeoJSON{
			Type:        "Point",
			Coordinates: []float64{city.Longitude, city.Latitude},
		},
		Stations: stationsDTO,
	}

	transaction := r.dg.Client.NewTxn()
	defer transaction.Discard(ctx)

	cityJSON, err := json.Marshal(cityDTO)
	if err != nil {
		return fmt.Errorf("marshal city (name: %s): %w", city.Name, err)
	}

	mutation := &api.Mutation{
		SetJson: cityJSON,
	}

	request := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{mutation},
		CommitNow: true,
	}

	_, err = transaction.Do(ctx, request)
	if err != nil {
		return fmt.Errorf("transaction.Do (city: %s): %w", city.Name, err)
	}

	return nil
}
