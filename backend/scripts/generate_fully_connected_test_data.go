// Seed Dgraph with top ~15 Russian cities (names in Russian) and connect them
// with stations + trips between every pair of cities.
//
// Requirements:
//   - Dgraph Alpha running at localhost:9080
//   - Schema already applied (as in the prompt)
//   - go get github.com/dgraph-io/dgo/v210
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"google.golang.org/grpc"
)

type City struct {
	UID   string   `json:"uid,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`

	Name     string    `json:"city.name,omitempty"`
	Stations []Station `json:"has_station,omitempty"`
}

type Station struct {
	UID   string   `json:"uid,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`

	Name          string `json:"station.name,omitempty"`
	TransportType string `json:"station.transport_type,omitempty"`
	Departs       []Trip `json:"departs,omitempty"`
}

type Trip struct {
	UID   string   `json:"uid,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`

	ExternalID    string  `json:"trip.external_id,omitempty"`
	Price         float64 `json:"trip.price,omitempty"`
	DepartureAt   string  `json:"trip.departure_at,omitempty"` // RFC3339
	ArrivalAt     string  `json:"trip.arrival_at,omitempty"`   // RFC3339
	TransportType string  `json:"trip.transport_type,omitempty"`

	Destination map[string]string `json:"destination,omitempty"` // {"uid": "..."}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()

	conn, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	// Top cities by population (common list). Names in Russian.
	cityNames := []string{
		"Москва",
		"Санкт-Петербург",
		"Новосибирск",
		"Екатеринбург",
		"Казань",
		"Нижний Новгород",
		"Челябинск",
		"Самара",
		"Омск",
		"Ростов-на-Дону",
		"Уфа",
		"Красноярск",
		"Воронеж",
		"Пермь",
		"Волгоград",
	}

	// Create cities + 2 stations per city (ЖД вокзал + Автовокзал).
	// Then create trips between every pair of cities for both transport types.
	// Trips are attached to the origin station via "departs", and point to destination city via "destination".
	cities := make([]City, 0, len(cityNames))
	for _, n := range cityNames {
		cities = append(cities, City{
			DType: []string{"City"},
			Name:  n,
			Stations: []Station{
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — ЖД вокзал", n),
					TransportType: "train",
				},
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — Автовокзал", n),
					TransportType: "bus",
				},
			},
		})
	}

	// First mutation: create cities+stations, capture assigned UIDs.
	cityUIDs, stationUIDsByCityAndType := upsertCitiesAndStations(ctx, dg, cities)

	// Second mutation: create trips and attach to stations.
	createTripsBetweenAllCities(ctx, dg, cityNames, cityUIDs, stationUIDsByCityAndType)

	log.Println("Done.")
}

func upsertCitiesAndStations(ctx context.Context, dg *dgo.Dgraph, cities []City) (map[string]string, map[string]map[string]string) {
	// We’ll create everything in one mutation using blank nodes.
	// Then map city name -> uid, and city name -> (transport_type -> station uid).
	type cityOut struct {
		UID string `json:"uid"`
	}
	type stationOut struct {
		UID           string `json:"uid"`
		TransportType string `json:"station.transport_type"`
	}

	// Prepare payload with deterministic blank node names.
	payload := make([]map[string]any, 0, len(cities))
	for i, c := range cities {
		cBlank := fmt.Sprintf("_:city_%d", i)
		obj := map[string]any{
			"uid":         cBlank,
			"dgraph.type": c.DType,
			"city.name":   c.Name,
			"city.location": map[string]any{
				"type": "Point",
				"coordinates": []float64{
					30.0 + rand.Float64()*10.0, // Longitude
					50.0 + rand.Float64()*10.0, // Latitude
				},
			},
		}

		sts := make([]map[string]any, 0, len(c.Stations))
		for j, s := range c.Stations {
			sBlank := fmt.Sprintf("_:station_%d_%d", i, j)
			sts = append(sts, map[string]any{
				"uid":                    sBlank,
				"dgraph.type":            s.DType,
				"station.name":           s.Name,
				"station.transport_type": s.TransportType,
			})
		}
		obj["has_station"] = sts
		payload = append(payload, obj)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("marshal cities payload: %v", err)
	}

	txn := dg.NewTxn()
	defer txn.Discard(ctx)

	mu := &api.Mutation{SetJson: b, CommitNow: true}
	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		log.Fatalf("mutate cities/stations: %v", err)
	}

	// Query back to build name->uid maps (more robust than relying on blank node mapping only).
	// (We still could use resp.Uids, but querying ensures we get station UIDs by transport type.)
	q := `
{
  cities(func: has(city.name)) {
    uid
    city.name
    has_station {
      uid
      station.transport_type
    }
  }
}`
	txn2 := dg.NewTxn()
	defer txn2.Discard(ctx)

	qr, err := txn2.Query(ctx, q)
	if err != nil {
		log.Fatalf("query back cities: %v", err)
	}

	var out struct {
		Cities []struct {
			UID      string       `json:"uid"`
			Name     string       `json:"city.name"`
			Stations []stationOut `json:"has_station"`
		} `json:"cities"`
	}
	if err := json.Unmarshal(qr.Json, &out); err != nil {
		log.Fatalf("unmarshal query: %v", err)
	}

	cityUIDs := map[string]string{}
	stationUIDsByCityAndType := map[string]map[string]string{}
	for _, c := range out.Cities {
		cityUIDs[c.Name] = c.UID
		if _, ok := stationUIDsByCityAndType[c.Name]; !ok {
			stationUIDsByCityAndType[c.Name] = map[string]string{}
		}
		for _, s := range c.Stations {
			stationUIDsByCityAndType[c.Name][s.TransportType] = s.UID
		}
	}

	_ = resp // resp used implicitly; keep for debugging if needed.
	return cityUIDs, stationUIDsByCityAndType
}

func createTripsBetweenAllCities(
	ctx context.Context,
	dg *dgo.Dgraph,
	cityNames []string,
	_ map[string]string,
	stationUIDsByCityAndType map[string]map[string]string,
) {
	const batchSize = 200
	var batch []map[string]any
	now := time.Now().UTC().Truncate(time.Hour)

	addTrip := func(originCity, destCity, transport string, dep time.Time, arr time.Time, price float64) {
		originStationUID := stationUIDsByCityAndType[originCity][transport]
		// ИСПРАВЛЕНИЕ: Берем UID СТАНЦИИ в целевом городе, а не самого города
		destStationUID := stationUIDsByCityAndType[destCity][transport]

		externalID := fmt.Sprintf("%s_%s_%s_%d", transport, originCity, destCity, dep.Unix())

		// Генерируем билеты (Tickets)
		tickets := []map[string]any{
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Economy",
				"ticket.price": price,
				"ticket.count": rand.Intn(50) + 1,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Business",
				"ticket.price": price * 1.5,
				"ticket.count": rand.Intn(10) + 1,
			},
		}

		trip := map[string]any{
			"uid":                 fmt.Sprintf("_:trip_%d", rand.Int63()),
			"dgraph.type":         []string{"Trip"},
			"trip.external_id":    externalID,
			"trip.departure_at":   dep.Format(time.RFC3339),
			"trip.arrival_at":     arr.Format(time.RFC3339),
			"trip.transport_type": transport,
			"has_ticket":          tickets,
			// ИСПРАВЛЕНИЕ: destination теперь указывает на станцию
			"destination": map[string]string{"uid": destStationUID},
		}

		// ИСПРАВЛЕНИЕ: Привязываем рейс к СТАНЦИИ через departs
		stationPatch := map[string]any{
			"uid":     originStationUID,
			"departs": []any{trip},
		}

		batch = append(batch, stationPatch)
		if len(batch) >= batchSize {
			flushBatch(ctx, dg, batch)
			batch = batch[:0]
		}
	}

	for i := 0; i < len(cityNames); i++ {
		for j := 0; j < len(cityNames); j++ {
			if i == j {
				continue
			}
			origin := cityNames[i]
			dest := cityNames[j]

			// Train: 4..30 hours
			dep1 := now.Add(time.Duration(rand.Intn(7*24)) * time.Hour).Add(time.Duration(rand.Intn(24)) * time.Hour)
			durTrainH := 4 + rand.Intn(27)
			arr1 := dep1.Add(time.Duration(durTrainH) * time.Hour)
			priceTrain := round2(1500 + rand.Float64()*8500)
			addTrip(origin, dest, "train", dep1, arr1, priceTrain)

			// Bus: 6..40 hours
			dep2 := now.Add(time.Duration(rand.Intn(7*24)) * time.Hour).Add(time.Duration(rand.Intn(24)) * time.Hour)
			durBusH := 6 + rand.Intn(35)
			arr2 := dep2.Add(time.Duration(durBusH) * time.Hour)
			priceBus := round2(800 + rand.Float64()*5200)
			addTrip(origin, dest, "bus", dep2, arr2, priceBus)
		}
	}

	if len(batch) > 0 {
		flushBatch(ctx, dg, batch)
	}
}

func flushBatch(ctx context.Context, dg *dgo.Dgraph, batch []map[string]any) {
	b, err := json.Marshal(batch)
	if err != nil {
		log.Fatalf("marshal batch: %v", err)
	}
	txn := dg.NewTxn()
	defer txn.Discard(ctx)

	_, err = txn.Mutate(ctx, &api.Mutation{
		SetJson:   b,
		CommitNow: true,
	})
	if err != nil {
		log.Fatalf("mutate batch: %v", err)
	}
}

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}
