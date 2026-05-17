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

            var totalRoutePrice float64
            for _, p := range pathCopy {
                totalRoutePrice += p.Price
            }

            finalPoint := models.RoutePoint{
                Name:           points[idx],
                StartTimestamp: currentTime.Unix(),
                EndTimestamp:   currentTime.Unix(),
                Price:          0.0,
            }
            finalPoint.SetDescription("Finish", totalRoutePrice)

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

        
		waitDuration := trip.DepartureAt.Sub(currentTime)

        fmt.Printf("  -> Поезд %s: Найдено категорий билетов: %d, Время ожидания: %v\n", 
            trip.ExternalID, len(trip.Tickets), waitDuration)

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

		var totalTickets int64
        for _, ticket := range trip.Tickets {
            totalTickets += int64(ticket.Count)
        }

        point := models.RoutePoint{
            Name:            points[idx],
            StartTimestamp:  currentTime.Unix(),
            EndTimestamp:    trip.DepartureAt.Unix(),
            Price:           trip.Price,
            AvailableAmount: totalTickets,
            Latitude:        latitude,
            Longitude:       longitude,
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
        # 1. Находим координаты целевого города назначения (toCity), чтобы отобразить на карте
        target_city(func: eq(city.name, $toCity)) {
            city.name
            city.location
            has_station {
                uid
            }
        }

        # 2. Находим вокзалы города отправления и их рейсы
        origin_stations(func: eq(city.name, $fromCity)) {
            has_station {
                uid
                departs @filter(ge(trip.departure_at, $minDeparture) AND le(trip.departure_at, $maxArrival)) (orderasc: trip.departure_at, first: %d) {
                    uid
                    dgraph.type
                    trip.external_id
                    trip.departure_at
                    trip.arrival_at
                    trip.price            
                    trip.transport_type   
    
                    has_ticket {
                        uid
                        dgraph.type
                        ticket.type
                        ticket.price
                        ticket.count
                    }
                    destination {
                        uid
                        station.name
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
        TargetCity []struct {
            Name     string `json:"city.name"`
            Location *struct {
                Type        string    `json:"type"`
                Coordinates []float64 `json:"coordinates"`
            } `json:"city.location"`
            HasStation []struct {
                Uid string `json:"uid"`
            } `json:"has_station"`
        } `json:"target_city"`

        OriginStations []struct {
            HasStation []struct {
                Uid     string `json:"uid"`
                Departs []struct {
                    Uid           string    `json:"uid"`
                    DgraphType    []string  `json:"dgraph.type"`
                    ExternalID    string    `json:"trip.external_id"`
                    DepartureAt   time.Time `json:"trip.departure_at"`
                    ArrivalAt     time.Time `json:"trip.arrival_at"`
                    Price         float64   `json:"trip.price"`
                    TransportType string    `json:"trip.transport_type"`
                    HasTicket []struct {
                        Uid   string  `json:"uid"`
                        DgraphType []string `json:"dgraph.type"`
                        Type  string  `json:"ticket.type"`
                        Price float64 `json:"ticket.price"`
                        Count int     `json:"ticket.count"`
                    } `json:"has_ticket"`
                    Destination   *struct {
                        Uid  string `json:"uid"`
                        Name string `json:"station.name"`
                    } `json:"destination"`
                } `json:"departs"`
            } `json:"has_station"`
        } `json:"origin_stations"`
    }

    if err := json.Unmarshal(resp.Json, &decode); err != nil {
        return nil, fmt.Errorf("decode dgraph response: %w", err)
    }

    // 1. Извлекаем координаты города назначения и маппим его станции
    var targetCityLocation graph.GeoJSON
    targetStationMap := make(map[string]bool)

    if len(decode.TargetCity) > 0 {
        tc := decode.TargetCity[0]
        if tc.Location != nil {
            targetCityLocation = graph.GeoJSON{
                Type:        tc.Location.Type,
                Coordinates: tc.Location.Coordinates,
            }
        }
        for _, hs := range tc.HasStation {
            targetStationMap[hs.Uid] = true
        }
    }

    // 2. Собираем пулл рейсов
    var result []graph.TripDTO
    for _, s := range decode.OriginStations {
        for _, hs := range s.HasStation {
            for _, d := range hs.Departs {
                if d.Destination == nil || !targetStationMap[d.Destination.Uid] {
                    continue
                }
                // === ВРЕМЕННЫЙ ТЕСТ ДЛЯ ПРОВЕРКИ ТИПОВ В БАЗЕ ===
                fmt.Printf("[DEBUG_DGRAPH_TYPES] Рейс %s: dgraph.type=%v | Найдено сырых билетов в ответе JSON: %d\n", 
                    d.ExternalID, d.DgraphType, len(d.HasTicket))
                // ===============================================

            // Перекладываем билеты из структуры декодера Dgraph в официальные TicketDTO
                var ticketsPool []*graph.TicketDTO
                for _, t := range d.HasTicket {
                    ticketsPool = append(ticketsPool, &graph.TicketDTO{
                        Uid:   t.Uid,
                        Type:  t.Type,
                        Price: t.Price,
                        Count: t.Count,
                    })
                }

                result = append(result, graph.TripDTO{
                    Uid:           d.Uid,
                    ExternalID:    d.ExternalID,
                    Price:         d.Price,
                    DepartureAt:   d.DepartureAt,
                    ArrivalAt:     d.ArrivalAt,
                    TransportType: d.TransportType, 
                    Tickets:       ticketsPool,
                    Destination: &graph.StationDTO{
                        Uid:  d.Destination.Uid,
                        Name: d.Destination.Name,
                        City: &graph.CityDTO{
                            Name:     toCity,
                            Location: targetCityLocation,
                        },
                    },
                })
            }
        }
    }

    return result, nil
}
