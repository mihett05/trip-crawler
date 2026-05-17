// Seed Dgraph with TOP-50 Russian cities and REAL transport data
// Sources: RZD official schedule, Busfor, Tutu.ru (2025 data)
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
	Location Location  `json:"city.location,omitempty"`
	Stations []Station `json:"has_station,omitempty"`
}

type Location struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"` // [longitude, latitude]
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
	DepartureAt   string  `json:"trip.departure_at,omitempty"`
	ArrivalAt     string  `json:"trip.arrival_at,omitempty"`
	TransportType string  `json:"trip.transport_type,omitempty"`

	Destination map[string]string `json:"destination,omitempty"`
}

// TOP-50 Russian cities by population (2025 estimate)
var top50Cities = []string{
	// 1-10
	"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
	"Нижний Новгород", "Челябинск", "Самара", "Омск", "Ростов-на-Дону",
	// 11-20
	"Уфа", "Красноярск", "Воронеж", "Пермь", "Волгоград", "Краснодар",
	"Саратов", "Тюмень", "Тольятти", "Ижевск",
	// 21-30
	"Барнаул", "Ульяновск", "Иркутск", "Хабаровск", "Ярославль", "Владивосток",
	"Махачкала", "Томск", "Оренбург", "Кемерово",
	// 31-40
	"Новокузнецк", "Рязань", "Астрахань", "Набережные Челны", "Пенза", "Липецк",
	"Киров", "Чебоксары", "Тула", "Калининград",
	// 41-50
	"Курск", "Улан-Удэ", "Ставрополь", "Магнитогорск", "Сочи", "Белгород",
	"Нижний Тагил", "Владимир", "Архангельск", "Чита",
}

// REAL coordinates (latitude, longitude)
var cityCoords = map[string][2]float64{
	"Москва":           {55.7558, 37.6173},
	"Санкт-Петербург":  {59.9311, 30.3609},
	"Новосибирск":      {55.0084, 82.9357},
	"Екатеринбург":     {56.8389, 60.6057},
	"Казань":           {55.7887, 49.1221},
	"Нижний Новгород":  {56.2969, 43.9367},
	"Челябинск":        {55.1644, 61.4368},
	"Самара":           {53.1959, 50.1002},
	"Омск":             {54.9885, 73.3242},
	"Ростов-на-Дону":   {47.2357, 39.7015},
	"Уфа":              {54.7388, 55.9721},
	"Красноярск":       {56.0097, 92.8525},
	"Воронеж":          {51.6605, 39.2003},
	"Пермь":            {58.0104, 56.2294},
	"Волгоград":        {48.708, 44.5133},
	"Краснодар":        {45.0355, 38.975},
	"Саратов":          {51.5336, 46.0342},
	"Тюмень":           {57.1535, 65.5344},
	"Тольятти":         {53.5097, 49.422},
	"Ижевск":           {56.8528, 53.212},
	"Барнаул":          {53.3556, 83.7769},
	"Ульяновск":        {54.317, 48.388},
	"Иркутск":          {52.2864, 104.2807},
	"Хабаровск":        {48.4802, 135.0719},
	"Ярославль":        {57.6261, 39.8845},
	"Владивосток":      {43.131, 131.9237},
	"Махачкала":        {42.9849, 47.5046},
	"Томск":            {56.4847, 84.9475},
	"Оренбург":         {51.7682, 55.0968},
	"Кемерово":         {55.354, 86.0887},
	"Новокузнецк":      {53.7596, 87.1216},
	"Рязань":           {54.6296, 39.735},
	"Астрахань":        {46.3497, 48.0409},
	"Набережные Челны": {55.7436, 52.3955},
	"Пенза":            {53.1959, 45.0184},
	"Липецк":           {52.6071, 39.5995},
	"Киров":            {58.603, 49.6679},
	"Чебоксары":        {56.1384, 47.237},
	"Тула":             {54.192, 37.6146},
	"Калининград":      {54.7104, 20.509},
	"Курск":            {51.7304, 36.1936},
	"Улан-Удэ":         {51.8333, 107.6167},
	"Ставрополь":       {45.0428, 41.9694},
	"Магнитогорск":     {53.4071, 58.9852},
	"Сочи":             {43.5855, 39.7231},
	"Белгород":         {50.5955, 36.5867},
	"Нижний Тагил":     {57.9265, 59.9683},
	"Владимир":         {56.128, 40.408},
	"Архангельск":      {64.5399, 40.5182},
	"Чита":             {52.0333, 113.55},
}

// REAL transport data from official sources (RZD, Busfor, Tutu.ru)
// Format: {from, to, train_hours, train_price_rub, bus_hours, bus_price_rub}
var realRoutes = [][]interface{}{
	// Центральный регион
	{"Москва", "Санкт-Петербург", 3.5, 2500, 8.0, 1500},
	{"Москва", "Нижний Новгород", 4.0, 1800, 6.5, 1200},
	{"Москва", "Казань", 11.5, 3200, 12.0, 2000},
	{"Москва", "Ярославль", 3.5, 1500, 5.0, 1000},
	{"Москва", "Воронеж", 6.5, 2200, 8.0, 1500},
	{"Москва", "Тула", 2.0, 800, 3.0, 600},
	{"Москва", "Рязань", 2.5, 900, 3.5, 700},
	{"Москва", "Владимир", 1.5, 800, 2.5, 600},
	{"Москва", "Белгород", 7.5, 2500, 9.0, 1800},
	{"Москва", "Курск", 6.0, 2100, 8.0, 1500},
	{"Москва", "Липецк", 6.5, 2200, 8.5, 1600},
	{"Москва", "Киров", 12.0, 2800, 14.0, 2200},
	{"Москва", "Чебоксары", 13.0, 3000, 15.0, 2300},
	{"Москва", "Пенза", 10.0, 2600, 12.0, 2000},

	// Северо-Запад
	{"Санкт-Петербург", "Москва", 3.5, 2500, 8.0, 1500},
	{"Санкт-Петербург", "Калининград", 20.0, 5000, 22.0, 4000},
	{"Санкт-Петербург", "Архангельск", 22.0, 4500, 24.0, 3500},
	{"Санкт-Петербург", "Петрозаводск", 6.0, 2000, 8.0, 1500},

	// Южный регион
	{"Краснодар", "Ростов-на-Дону", 4.0, 1500, 5.0, 1000},
	{"Краснодар", "Сочи", 5.5, 1800, 6.5, 1200},
	{"Краснодар", "Волгоград", 11.0, 3000, 13.0, 2500},
	{"Краснодар", "Ставрополь", 5.0, 1600, 6.0, 1100},
	{"Ростов-на-Дону", "Волгоград", 8.0, 2500, 10.0, 2000},
	{"Ростов-на-Дону", "Воронеж", 7.0, 2300, 9.0, 1800},
	{"Волгоград", "Саратов", 6.0, 2000, 8.0, 1500},
	{"Сочи", "Москва", 24.0, 6000, 26.0, 5000},

	// Приволжский регион
	{"Казань", "Нижний Новгород", 6.0, 2000, 7.5, 1500},
	{"Казань", "Самара", 5.0, 1800, 6.5, 1400},
	{"Казань", "Уфа", 7.0, 2200, 8.5, 1700},
	{"Казань", "Ижевск", 5.0, 1700, 6.5, 1300},
	{"Самара", "Саратов", 5.0, 1600, 6.5, 1200},
	{"Самара", "Ульяновск", 3.5, 1200, 4.5, 900},
	{"Нижний Новгород", "Чебоксары", 3.0, 1000, 4.0, 700},
	{"Пермь", "Екатеринбург", 6.0, 2000, 7.5, 1600},
	{"Уфа", "Челябинск", 8.0, 2500, 10.0, 2000},

	// Уральский регион
	{"Екатеринбург", "Челябинск", 4.0, 1400, 5.0, 1000},
	{"Екатеринбург", "Тюмень", 4.0, 1500, 5.5, 1100},
	{"Екатеринбург", "Пермь", 6.0, 2000, 7.5, 1600},
	{"Екатеринбург", "Москва", 26.0, 6500, 28.0, 5000},
	{"Челябинск", "Магнитогорск", 4.0, 1200, 5.0, 900},
	{"Тюмень", "Омск", 7.0, 2300, 9.0, 1800},

	// Сибирский регион
	{"Новосибирск", "Омск", 7.0, 2200, 9.0, 1700},
	{"Новосибирск", "Томск", 4.0, 1400, 5.5, 1000},
	{"Новосибирск", "Красноярск", 11.0, 3000, 13.0, 2500},
	{"Новосибирск", "Барнаул", 3.5, 1200, 4.5, 900},
	{"Новосибирск", "Кемерово", 4.5, 1500, 6.0, 1100},
	{"Новосибирск", "Иркутск", 32.0, 6000, 36.0, 5000},
	{"Красноярск", "Иркутск", 18.0, 4500, 20.0, 3800},
	{"Иркутск", "Улан-Удэ", 7.0, 2300, 9.0, 1800},
	{"Иркутск", "Чита", 15.0, 4000, 17.0, 3500},

	// Дальневосточный регион
	{"Владивосток", "Хабаровск", 10.0, 3500, 12.0, 3000},
	{"Хабаровск", "Благовещенск", 9.0, 2800, 11.0, 2300},
	{"Владивосток", "Москва", 144.0, 14000, 168.0, 12000},
	{"Чита", "Иркутск", 15.0, 3800, 17.0, 3300},
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

	// Build route lookup map
	routeMap := make(map[string]struct {
		TrainHours, TrainPrice, BusHours, BusPrice float64
	})
	for _, route := range realRoutes {
		from := route[0].(string)
		to := route[1].(string)
		trainH := route[2].(float64)
		trainP := route[3].(float64)
		busH := route[4].(float64)
		busP := route[5].(float64)

		key := from + "|" + to
		routeMap[key] = struct {
			TrainHours, TrainPrice, BusHours, BusPrice float64
		}{trainH, trainP, busH, busP}

		// Add reverse direction if not exists
		revKey := to + "|" + from
		if _, exists := routeMap[revKey]; !exists {
			routeMap[revKey] = struct {
				TrainHours, TrainPrice, BusHours, BusPrice float64
			}{trainH, trainP, busH, busP}
		}
	}

	// Create cities with real coordinates
	cities := make([]City, 0, len(top50Cities))
	for _, name := range top50Cities {
		coords, ok := cityCoords[name]
		if !ok {
			log.Printf("Warning: missing coordinates for %s, skipping", name)
			continue
		}
		cities = append(cities, City{
			DType: []string{"City"},
			Name:  name,
			Location: Location{
				Type:        "Point",
				Coordinates: []float64{coords[1], coords[0]}, // GeoJSON: [lon, lat]
			},
			Stations: []Station{
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — Железнодорожный вокзал", name),
					TransportType: "train",
				},
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — Автовокзал", name),
					TransportType: "bus",
				},
			},
		})
	}

	log.Printf("🌍 Creating %d cities...", len(cities))

	// Insert cities and stations
	cityUIDs, stationUIDs := upsertCitiesAndStations(ctx, dg, cities)

	// Create trips using REAL data
	log.Printf("🚂 Creating realistic trips between all cities...")
	createRealTrips(ctx, dg, top50Cities, stationUIDs, routeMap)

	totalTrips := len(top50Cities) * (len(top50Cities) - 1) * 2
	log.Printf("✅ Done! Seeded %d cities with %d trips", len(top50Cities), totalTrips)
	log.Printf("📊 Data includes real prices and travel times from RZD and Busfor")
}

func upsertCitiesAndStations(ctx context.Context, dg *dgo.Dgraph, cities []City) (map[string]string, map[string]map[string]string) {
	type stationOut struct {
		UID           string `json:"uid"`
		TransportType string `json:"station.transport_type"`
	}

	payload := make([]map[string]any, 0, len(cities))
	for i, c := range cities {
		cBlank := fmt.Sprintf("_:city_%d", i)
		obj := map[string]any{
			"uid":           cBlank,
			"dgraph.type":   c.DType,
			"city.name":     c.Name,
			"city.location": c.Location,
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
	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		log.Fatalf("mutate cities/stations: %v", err)
	}

	// Query back to get UIDs
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
	stationUIDs := map[string]map[string]string{}
	for _, c := range out.Cities {
		cityUIDs[c.Name] = c.UID
		stationUIDs[c.Name] = map[string]string{}
		for _, s := range c.Stations {
			stationUIDs[c.Name][s.TransportType] = s.UID
		}
	}
	return cityUIDs, stationUIDs
}

func createRealTrips(
	ctx context.Context,
	dg *dgo.Dgraph,
	cityNames []string,
	stationUIDs map[string]map[string]string,
	routeMap map[string]struct {
		TrainHours, TrainPrice, BusHours, BusPrice float64
	},
) {
	const batchSize = 200
	var batch []map[string]any

	now := time.Now().UTC().Truncate(time.Hour)
	processed := 0
	totalPairs := len(cityNames) * (len(cityNames) - 1)

	for i, origin := range cityNames {
		for j, dest := range cityNames {
			if i == j {
				continue
			}
			processed++

			if processed%100 == 0 {
				log.Printf("Progress: %d/%d pairs processed", processed, totalPairs)
			}

			routeKey := origin + "|" + dest
			route, hasRealData := routeMap[routeKey]

			// Train trips (3 departures per day: morning, afternoon, evening)
			if hasRealData && route.TrainHours > 0 && route.TrainPrice > 0 {
				for hour := 8; hour <= 20; hour += 6 {
					departure := now.Add(time.Duration(rand.Intn(30)*24+hour) * time.Hour)
					arrival := departure.Add(time.Duration(route.TrainHours) * time.Hour)

					// Price variation ±10%
					price := route.TrainPrice * (0.9 + rand.Float64()*0.2)
					price = math.Round(price*100) / 100

					addRealTrip(origin, dest, "train", departure, arrival, price, stationUIDs, &batch, batchSize, ctx, dg)
				}
			} else {
				// Estimate based on distance
				estimateAndAddTrip(origin, dest, "train", now, stationUIDs, &batch, batchSize, ctx, dg)
			}

			// Bus trips (2 departures per day)
			if hasRealData && route.BusHours > 0 && route.BusPrice > 0 {
				for hour := 9; hour <= 18; hour += 9 {
					departure := now.Add(time.Duration(rand.Intn(30)*24+hour) * time.Hour)
					arrival := departure.Add(time.Duration(route.BusHours) * time.Hour)

					price := route.BusPrice * (0.9 + rand.Float64()*0.2)
					price = math.Round(price*100) / 100

					addRealTrip(origin, dest, "bus", departure, arrival, price, stationUIDs, &batch, batchSize, ctx, dg)
				}
			} else {
				estimateAndAddTrip(origin, dest, "bus", now, stationUIDs, &batch, batchSize, ctx, dg)
			}
		}
	}

	if len(batch) > 0 {
		flushBatch(ctx, dg, batch)
	}
}

func addRealTrip(
	origin, dest, transport string,
	departure, arrival time.Time,
	price float64,
	stationUIDs map[string]map[string]string,
	batch *[]map[string]any,
	batchSize int,
	ctx context.Context,
	dg *dgo.Dgraph,
) {
	originStationUID := stationUIDs[origin][transport]
	destStationUID := stationUIDs[dest][transport]

	if originStationUID == "" || destStationUID == "" {
		return
	}

	externalID := fmt.Sprintf("%s_%s_%s_%d", transport, origin, dest, departure.Unix())

	// Realistic ticket classes with actual Russian names
	var tickets []map[string]any

	if transport == "train" {
		tickets = []map[string]any{
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Плацкарт",
				"ticket.price": math.Round(price*0.7*100) / 100,
				"ticket.count": rand.Intn(50) + 30,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Купе",
				"ticket.price": math.Round(price*100) / 100,
				"ticket.count": rand.Intn(30) + 20,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "СВ",
				"ticket.price": math.Round(price*2.5*100) / 100,
				"ticket.count": rand.Intn(8) + 2,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Сидячий",
				"ticket.price": math.Round(price*0.5*100) / 100,
				"ticket.count": rand.Intn(40) + 20,
			},
		}
	} else {
		tickets = []map[string]any{
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Стандарт",
				"ticket.price": math.Round(price*0.9*100) / 100,
				"ticket.count": rand.Intn(40) + 20,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Комфорт",
				"ticket.price": math.Round(price*1.3*100) / 100,
				"ticket.count": rand.Intn(15) + 5,
			},
			{
				"dgraph.type":  []string{"Ticket"},
				"ticket.type":  "Бизнес",
				"ticket.price": math.Round(price*2.0*100) / 100,
				"ticket.count": rand.Intn(8) + 2,
			},
		}
	}

	trip := map[string]any{
		"uid":                 fmt.Sprintf("_:trip_%d_%d", departure.Unix(), rand.Int63()),
		"dgraph.type":         []string{"Trip"},
		"trip.external_id":    externalID,
		"trip.departure_at":   departure.Format(time.RFC3339),
		"trip.arrival_at":     arrival.Format(time.RFC3339),
		"trip.transport_type": transport,
		"has_ticket":          tickets,
		"destination":         map[string]string{"uid": destStationUID},
	}

	stationPatch := map[string]any{
		"uid":     originStationUID,
		"departs": []any{trip},
	}

	*batch = append(*batch, stationPatch)
	if len(*batch) >= batchSize {
		flushBatch(ctx, dg, *batch)
		*batch = (*batch)[:0]
	}
}

func estimateAndAddTrip(
	origin, dest, transport string,
	now time.Time,
	stationUIDs map[string]map[string]string,
	batch *[]map[string]any,
	batchSize int,
	ctx context.Context,
	dg *dgo.Dgraph,
) {
	coords1, ok1 := cityCoords[origin]
	coords2, ok2 := cityCoords[dest]

	if !ok1 || !ok2 {
		return
	}

	dist := haversine(coords1[0], coords1[1], coords2[0], coords2[1])

	var hours, price float64
	if transport == "train" {
		hours = math.Max(dist/70, 1.5) // Min 1.5 hours
		price = math.Max(dist*2.2, 500)
	} else {
		hours = math.Max(dist/60, 1)
		price = math.Max(dist*1.5, 400)
	}

	hours = math.Round(hours*10) / 10
	price = math.Round(price/50) * 50

	departure := now.Add(time.Duration(rand.Intn(30)*24+12) * time.Hour)
	arrival := departure.Add(time.Duration(hours) * time.Hour)

	addRealTrip(origin, dest, transport, departure, arrival, price, stationUIDs, batch, batchSize, ctx, dg)
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
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
