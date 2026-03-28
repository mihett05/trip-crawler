package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	"github.com/mihett05/trip-crawler/internal/service/routes/models"
	"github.com/mihett05/trip-crawler/internal/service/routes/repositories/graph"
)

const (
	minStayDuration = 24 * time.Hour
	maxStayDuration = 168 * time.Hour
	idealStayHours  = 48.0
	pricePenalty    = 2000.0
	waitBonusWeight = 0.5
	maxTimeScore    = 500.0
)

type DgraphItineraryBuilder struct {
	dg *dgraph.Client
}

func NewDgraphItineraryBuilder(dg *dgraph.Client) *DgraphItineraryBuilder {
	return &DgraphItineraryBuilder{dg: dg}
}

func (b *DgraphItineraryBuilder) Build(ctx context.Context, points []string, startAt int64, minDays, maxDays int) ([]models.RoutePoint, error) {
	startDate := time.Unix(startAt, 0).UTC()

	if minDays <= 0 {
		return nil, fmt.Errorf("validate minDays: got %d, want positive", minDays)
	}

	if maxDays < minDays {
		return nil, fmt.Errorf("validate duration interval: maxDays (%d) < minDays (%d)", maxDays, minDays)
	}

	minFinish := startDate.Add(time.Duration(minDays) * 24 * time.Hour)
	targetFinish := startDate.Add(time.Duration(maxDays) * 24 * time.Hour)
	searchDeadline := targetFinish.Add(maxStayDuration)

	tripsPool, err := b.fetchFullPool(ctx, points, startDate, searchDeadline)
	if err != nil {
		return nil, fmt.Errorf("build itinerary pool: %w", err)
	}

	var allPaths [][]models.RoutePoint
	b.generatePaths(points, 0, startDate, nil, &allPaths, minFinish, targetFinish, tripsPool)

	if len(allPaths) == 0 {
		return nil, fmt.Errorf("route not found for interval %d-%d days", minDays, maxDays)
	}

	sort.Slice(allPaths, func(i, j int) bool {
		return b.score(allPaths[i], targetFinish) > b.score(allPaths[j], targetFinish)
	})

	return allPaths[0], nil
}

func (b *DgraphItineraryBuilder) generatePaths(
	points []string,
	idx int,
	currentTime time.Time,
	currentPath []models.RoutePoint,
	results *[][]models.RoutePoint,
	minFinish time.Time,
	maxFinish time.Time,
	pool map[string][]graph.TripDTO,
) {
	if len(*results) >= 500 {
		return
	}

	if idx == len(points)-1 {
		if currentTime.After(minFinish) && currentTime.Before(maxFinish.Add(minStayDuration)) {
			pathCopy := make([]models.RoutePoint, len(currentPath))
			copy(pathCopy, currentPath)

			finalPoint := models.RoutePoint{
				Name:           points[idx],
				StartTimestamp: currentTime.Unix(),
				EndTimestamp:   currentTime.Unix(),
			}
			finalPoint.SetDescription("Finish", 0)

			*results = append(*results, append(pathCopy, finalPoint))
		}
		return
	}

	segmentKey := fmt.Sprintf("%s->%s", points[idx], points[idx+1])
	candidates := pool[segmentKey]

	minWait := minStayDuration
	if idx == 0 {
		minWait = 0
	}

	for _, trip := range candidates {
		if len(trip.Tickets) == 0 {
			continue
		}

		waitDuration := trip.DepartureAt.Sub(currentTime)

		if waitDuration < minWait || waitDuration > maxStayDuration {
			continue
		}
		if trip.ArrivalAt.After(maxFinish.Add(minStayDuration)) {
			continue
		}

		var latitude, longitude *float64
		if trip.Destination != nil && trip.Destination.City != nil && len(trip.Destination.City.Location.Coordinates) >= 2 {
			longitudeVal := trip.Destination.City.Location.Coordinates[0]
			latitudeVal := trip.Destination.City.Location.Coordinates[1]
			latitude, longitude = &longitudeVal, &latitudeVal
		}

		point := models.RoutePoint{
			Name:           points[idx],
			StartTimestamp: currentTime.Unix(),
			EndTimestamp:   trip.DepartureAt.Unix(),
			Price:          trip.Price,
			Latitude:       latitude,
			Longitude:      longitude,
		}
		point.SetDescription(trip.ExternalID, trip.Price)

		if idx == 0 {
			point.StartTimestamp = trip.DepartureAt.Unix()
		}

		b.generatePaths(points, idx+1, trip.ArrivalAt, append(currentPath, point), results, minFinish, maxFinish, pool)
	}
}

func (b *DgraphItineraryBuilder) score(path []models.RoutePoint, targetFinish time.Time) float64 {
	var totalPrice float64
	var totalWaitBonus float64

	for i, point := range path {
		totalPrice += point.Price
		// бонус за близость к идеальным 48ч (первый и последний город не учитываем)
		if i > 0 && i < len(path)-1 {
			stay := time.Duration(point.EndTimestamp-point.StartTimestamp) * time.Second
			totalWaitBonus += 100.0 / (1.0 + math.Abs(stay.Hours()-idealStayHours))
		}
	}

	actualFinish := time.Unix(path[len(path)-1].EndTimestamp, 0)

	// штраф за отклонение от целевой даты
	timeDiffHours := targetFinish.Sub(actualFinish).Hours()
	timeScore := maxTimeScore / (1.0 + math.Pow(timeDiffHours/24.0, 2))

	return (totalWaitBonus * waitBonusWeight) + timeScore - (totalPrice / pricePenalty)
}

func (b *DgraphItineraryBuilder) fetchFullPool(ctx context.Context, points []string, startTime, endTime time.Time) (map[string][]graph.TripDTO, error) {
	resultMap := make(map[string][]graph.TripDTO)
	type segmentResult struct {
		segmentKey string
		trips      []graph.TripDTO
		err        error
	}
	numSegments := len(points) - 1
	resultsChan := make(chan segmentResult, numSegments)

	for i := 0; i < numSegments; i++ {
		go func(fromCity, toCity string) {
			trips, err := b.fetchTrips(ctx, fromCity, toCity, startTime, endTime, 100)
			resultsChan <- segmentResult{
				segmentKey: fmt.Sprintf("%s->%s", fromCity, toCity),
				trips:      trips,
				err:        err,
			}
		}(points[i], points[i+1])
	}

	for i := 0; i < numSegments; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultsChan:
			if result.err != nil {
				return nil, fmt.Errorf("fetch segment %s: %w", result.segmentKey, result.err)
			}
			resultMap[result.segmentKey] = result.trips
		}
	}
	return resultMap, nil
}

func (b *DgraphItineraryBuilder) fetchTrips(ctx context.Context, fromCity, toCity string, minDeparture, maxArrival time.Time, limit int) ([]graph.TripDTO, error) {
	query := fmt.Sprintf(`query all($fromCity: string, $toCity: string, $minDeparture: string, $maxArrival: string) {
		stations(func: eq(city.name, $fromCity)) {
			has_station {
				departs @filter(ge(trip.departure_at, $minDeparture) AND le(trip.departure_at, $maxArrival)) (orderasc: trip.departure_at, first: %d) {
					uid
					trip.external_id
					trip.departure_at
					trip.arrival_at
					destination {
						uid
						station.name
						city_info: ~has_station { 
							city.name
							city.location
						}
					}
					has_ticket (orderasc: ticket.price, first: 1) {
						ticket.type
						ticket.price
					}
				}
			}
		}
	}`, limit)

	vars := map[string]string{
		"$fromCity":     fromCity,
		"$toCity":       toCity,
		"$minDeparture": minDeparture.UTC().Format(time.RFC3339),
		"$maxArrival":   maxArrival.UTC().Format(time.RFC3339),
	}

	txn := b.dg.Client.NewTxn()
	defer txn.Discard(ctx)

	resp, err := txn.QueryWithVars(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("dgraph query (from %s to %s): %w", fromCity, toCity, err)
	}

	var decode struct {
		Stations []struct {
			HasStation []struct {
				Departs []struct {
					Uid         string    `json:"uid"`
					ExternalID  string    `json:"trip.external_id"`
					DepartureAt time.Time `json:"trip.departure_at"`
					ArrivalAt   time.Time `json:"trip.arrival_at"`
					Destination *struct {
						Uid      string `json:"uid"`
						Name     string `json:"station.name"`
						CityInfo []struct {
							Name     string `json:"city.name"`
							Location *struct {
								Type        string    `json:"type"`
								Coordinates []float64 `json:"coordinates"`
							} `json:"city.location"`
						} `json:"city_info"`
					} `json:"destination"`
					Tickets []struct {
						Type  string  `json:"ticket.type"`
						Price float64 `json:"ticket.price"`
					} `json:"has_ticket"`
				} `json:"departs"`
			} `json:"has_station"`
		} `json:"stations"`
	}

	if err := json.Unmarshal(resp.Json, &decode); err != nil {
		return nil, fmt.Errorf("decode dgraph response: %w", err)
	}

	var result []graph.TripDTO
	for _, s := range decode.Stations {
		for _, hs := range s.HasStation {
			for _, d := range hs.Departs {
				if d.Destination == nil || len(d.Destination.CityInfo) == 0 {
					continue
				}

				if d.Destination.CityInfo[0].Name != toCity {
					continue
				}

				price := 0.0
				var tickets []*graph.TicketDTO
				if len(d.Tickets) > 0 {
					price = d.Tickets[0].Price
					tickets = []*graph.TicketDTO{{
						Type:  d.Tickets[0].Type,
						Price: d.Tickets[0].Price,
					}}
				}

				var cityLocation graph.GeoJSON
				if d.Destination.CityInfo[0].Location != nil {
					cityLocation = graph.GeoJSON{
						Type:        d.Destination.CityInfo[0].Location.Type,
						Coordinates: d.Destination.CityInfo[0].Location.Coordinates,
					}
				}

				result = append(result, graph.TripDTO{
					Uid:         d.Uid,
					ExternalID:  d.ExternalID,
					Price:       price,
					DepartureAt: d.DepartureAt,
					ArrivalAt:   d.ArrivalAt,
					Destination: &graph.StationDTO{
						Uid:  d.Destination.Uid,
						Name: d.Destination.Name,
						City: &graph.CityDTO{
							Name:     d.Destination.CityInfo[0].Name,
							Location: cityLocation,
						},
					},
					Tickets: tickets,
				})
			}
		}
	}

	return result, nil
}
