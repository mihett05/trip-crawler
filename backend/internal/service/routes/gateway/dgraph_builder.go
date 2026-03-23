package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	"github.com/mihett05/trip-crawler/internal/service/routes/models"
	"github.com/mihett05/trip-crawler/internal/service/routes/repositories/graph"
)

type DgraphItineraryBuilder struct {
	dg *dgraph.Client
}

func NewDgraphItineraryBuilder(dg *dgraph.Client) *DgraphItineraryBuilder {
	return &DgraphItineraryBuilder{dg: dg}
}

func (b *DgraphItineraryBuilder) Build(ctx context.Context, points []string, startAt int64) ([]models.RoutePoint, error) {
	var finalRoute []models.RoutePoint
	currentTimestamp := startAt

	for i := 0; i < len(points)-1; i++ {
		fromCity := points[i]
		toCity := points[i+1]

		timeStr := time.Unix(currentTimestamp, 0).Format(time.RFC3339)

		trip, err := b.findBestTrip(ctx, fromCity, toCity, timeStr)
		if err != nil {
			return nil, fmt.Errorf("find segment %s->%s: %w", fromCity, toCity, err)
		}

		point := models.RoutePoint{
			Name:           fromCity,
			StartTimestamp: trip.DepartureAt.Unix(),
			EndTimestamp:   trip.ArrivalAt.Unix(),
		}

		point.SetDescription(trip.ExternalID, trip.Price)

		finalRoute = append(finalRoute, point)

		currentTimestamp = trip.ArrivalAt.Add(48 * time.Hour).Unix()
	}

	finalRoute = append(finalRoute, models.RoutePoint{Name: points[len(points)-1]})
	return finalRoute, nil
}

func (b *DgraphItineraryBuilder) findBestTrip(ctx context.Context, from, to string, afterTime string) (*graph.TripDTO, error) {
	query := `query all($from: string, $to: string, $time: string) {
		A as var(func: eq(city.name, $from)) {
			has_station { A_stations as uid }
		}
		B as var(func: eq(city.name, $to)) {
			has_station { B_stations as uid }
		}

		trips(func: uid(A_stations)) {
			departs @filter(ge(trip.departure_at, $time)) @orderasc(trip.price) {
				uid
				trip.external_id
				trip.price
				trip.departure_at
				trip.arrival_at
				destination @filter(uid(B_stations)) {
					station.name
				}
			}
		}
	}`

	vars := map[string]string{
		"$from": from,
		"$to":   to,
		"$time": afterTime,
	}

	resp, err := b.dg.Client.NewTxn().QueryWithVars(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("txn.QueryWithVars: %w", err)
	}

	var decode struct {
		Trips []struct {
			Departs []graph.TripDTO `json:"departs"`
		} `json:"trips"`
	}

	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if len(decode.Trips) == 0 || len(decode.Trips[0].Departs) == 0 {
		return nil, fmt.Errorf("no trips found from %s to %s after %s", from, to, afterTime)
	}

	return &decode.Trips[0].Departs[0], nil
}
