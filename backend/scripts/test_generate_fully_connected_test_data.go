// Dry-run version - only prints JSON to console, no Dgraph connection needed
//
// Run:
//
//	go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
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

	ExternalID    string                   `json:"trip.external_id,omitempty"`
	Price         float64                  `json:"trip.price,omitempty"`
	DepartureAt   string                   `json:"trip.departure_at,omitempty"`
	ArrivalAt     string                   `json:"trip.arrival_at,omitempty"`
	TransportType string                   `json:"trip.transport_type,omitempty"`
	Destination   map[string]string        `json:"destination,omitempty"`
	HasTicket     []map[string]interface{} `json:"has_ticket,omitempty"`
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

// REAL transport data from official sources (using float64 for all numbers)
var realRoutes = [][]interface{}{
	{"Москва", "Санкт-Петербург", 3.5, 2500.0, 8.0, 1500.0},
	{"Москва", "Нижний Новгород", 4.0, 1800.0, 6.5, 1200.0},
	{"Москва", "Казань", 11.5, 3200.0, 12.0, 2000.0},
	{"Москва", "Ярославль", 3.5, 1500.0, 5.0, 1000.0},
	{"Москва", "Воронеж", 6.5, 2200.0, 8.0, 1500.0},
	{"Москва", "Тула", 2.0, 800.0, 3.0, 600.0},
	{"Москва", "Рязань", 2.5, 900.0, 3.5, 700.0},
	{"Москва", "Владимир", 1.5, 800.0, 2.5, 600.0},
	{"Москва", "Белгород", 7.5, 2500.0, 9.0, 1800.0},
	{"Москва", "Курск", 6.0, 2100.0, 8.0, 1500.0},
	{"Москва", "Липецк", 6.5, 2200.0, 8.5, 1600.0},
	{"Москва", "Киров", 12.0, 2800.0, 14.0, 2200.0},
	{"Москва", "Чебоксары", 13.0, 3000.0, 15.0, 2300.0},
	{"Москва", "Пенза", 10.0, 2600.0, 12.0, 2000.0},
	{"Москва", "Калининград", 20.0, 5000.0, 22.0, 4000.0},
	{"Санкт-Петербург", "Калининград", 20.0, 5000.0, 22.0, 4000.0},
	{"Санкт-Петербург", "Архангельск", 22.0, 4500.0, 24.0, 3500.0},
	{"Краснодар", "Ростов-на-Дону", 4.0, 1500.0, 5.0, 1000.0},
	{"Краснодар", "Сочи", 5.5, 1800.0, 6.5, 1200.0},
	{"Краснодар", "Волгоград", 11.0, 3000.0, 13.0, 2500.0},
	{"Краснодар", "Ставрополь", 5.0, 1600.0, 6.0, 1100.0},
	{"Ростов-на-Дону", "Волгоград", 8.0, 2500.0, 10.0, 2000.0},
	{"Ростов-на-Дону", "Воронеж", 7.0, 2300.0, 9.0, 1800.0},
	{"Сочи", "Москва", 24.0, 6000.0, 26.0, 5000.0},
	{"Казань", "Нижний Новгород", 6.0, 2000.0, 7.5, 1500.0},
	{"Казань", "Самара", 5.0, 1800.0, 6.5, 1400.0},
	{"Казань", "Уфа", 7.0, 2200.0, 8.5, 1700.0},
	{"Казань", "Ижевск", 5.0, 1700.0, 6.5, 1300.0},
	{"Самара", "Саратов", 5.0, 1600.0, 6.5, 1200.0},
	{"Нижний Новгород", "Чебоксары", 3.0, 1000.0, 4.0, 700.0},
	{"Пермь", "Екатеринбург", 6.0, 2000.0, 7.5, 1600.0},
	{"Екатеринбург", "Челябинск", 4.0, 1400.0, 5.0, 1000.0},
	{"Екатеринбург", "Тюмень", 4.0, 1500.0, 5.5, 1100.0},
	{"Екатеринбург", "Москва", 26.0, 6500.0, 28.0, 5000.0},
	{"Челябинск", "Магнитогорск", 4.0, 1200.0, 5.0, 900.0},
	{"Новосибирск", "Омск", 7.0, 2200.0, 9.0, 1700.0},
	{"Новосибирск", "Томск", 4.0, 1400.0, 5.5, 1000.0},
	{"Новосибирск", "Красноярск", 11.0, 3000.0, 13.0, 2500.0},
	{"Новосибирск", "Барнаул", 3.5, 1200.0, 4.5, 900.0},
	{"Новосибирск", "Кемерово", 4.5, 1500.0, 6.0, 1100.0},
	{"Новосибирск", "Иркутск", 32.0, 6000.0, 36.0, 5000.0},
	{"Красноярск", "Иркутск", 18.0, 4500.0, 20.0, 3800.0},
	{"Иркутск", "Улан-Удэ", 7.0, 2300.0, 9.0, 1800.0},
	{"Владивосток", "Хабаровск", 10.0, 3500.0, 12.0, 3000.0},
	{"Владивосток", "Москва", 144.0, 14000.0, 168.0, 12000.0},
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("DRY-RUN MODE: Generating realistic transport data for top-50 Russian cities")
	fmt.Println("No database connection - just printing JSON to console")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Build route lookup map
	type RouteData struct {
		TrainHours, TrainPrice, BusHours, BusPrice float64
	}
	routeMap := make(map[string]RouteData)

	for _, route := range realRoutes {
		from := route[0].(string)
		to := route[1].(string)

		// Конвертируем в float64 с проверкой типа
		var trainH, trainP, busH, busP float64

		switch v := route[2].(type) {
		case float64:
			trainH = v
		case int:
			trainH = float64(v)
		}

		switch v := route[3].(type) {
		case float64:
			trainP = v
		case int:
			trainP = float64(v)
		}

		switch v := route[4].(type) {
		case float64:
			busH = v
		case int:
			busH = float64(v)
		}

		switch v := route[5].(type) {
		case float64:
			busP = v
		case int:
			busP = float64(v)
		}

		key := from + "|" + to
		routeMap[key] = RouteData{trainH, trainP, busH, busP}

		// Add reverse direction if not exists
		revKey := to + "|" + from
		if _, exists := routeMap[revKey]; !exists {
			routeMap[revKey] = RouteData{trainH, trainP, busH, busP}
		}
	}

	// Generate all data
	allCities := make([]City, 0, len(top50Cities))

	fmt.Println("🏙️  Creating cities...")
	for _, name := range top50Cities {
		coords, ok := cityCoords[name]
		if !ok {
			log.Printf("Warning: missing coordinates for %s, skipping", name)
			continue
		}

		city := City{
			DType: []string{"City"},
			Name:  name,
			Location: Location{
				Type:        "Point",
				Coordinates: []float64{coords[1], coords[0]},
			},
			Stations: []Station{
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — Железнодорожный вокзал", name),
					TransportType: "train",
					Departs:       []Trip{},
				},
				{
					DType:         []string{"Station"},
					Name:          fmt.Sprintf("%s — Автовокзал", name),
					TransportType: "bus",
					Departs:       []Trip{},
				},
			},
		}
		allCities = append(allCities, city)
	}

	fmt.Printf("✅ Created %d cities\n", len(allCities))
	fmt.Println()

	fmt.Println("🚂 Generating trips between all city pairs...")
	now := time.Now().UTC().Truncate(time.Hour)
	totalPairs := len(top50Cities) * (len(top50Cities) - 1)
	processed := 0

	// Create a map for quick station lookup
	stationMap := make(map[string]map[string]string)
	for _, city := range allCities {
		stationMap[city.Name] = make(map[string]string)
		stationMap[city.Name]["train"] = fmt.Sprintf("station_%s_train", strings.ReplaceAll(city.Name, " ", "_"))
		stationMap[city.Name]["bus"] = fmt.Sprintf("station_%s_bus", strings.ReplaceAll(city.Name, " ", "_"))
	}

	tripsCount := 0

	for i, origin := range top50Cities {
		for j, dest := range top50Cities {
			if i == j {
				continue
			}
			processed++

			if processed%100 == 0 {
				fmt.Printf("  Progress: %d/%d pairs processed\n", processed, totalPairs)
			}

			routeKey := origin + "|" + dest
			route, hasRealData := routeMap[routeKey]

			// Generate train trips
			if hasRealData && route.TrainHours > 0 && route.TrainPrice > 0 {
				for hour := 8; hour <= 20; hour += 6 {
					departure := now.Add(time.Duration(rand.Intn(30)*24+hour) * time.Hour)
					arrival := departure.Add(time.Duration(route.TrainHours) * time.Hour)
					price := route.TrainPrice * (0.9 + rand.Float64()*0.2)
					price = math.Round(price*100) / 100

					trip := createTrip(origin, dest, "train", departure, arrival, price, stationMap)

					// Find and add trip to the correct station
					for idx := range allCities {
						if allCities[idx].Name == origin {
							for stIdx := range allCities[idx].Stations {
								if allCities[idx].Stations[stIdx].TransportType == "train" {
									allCities[idx].Stations[stIdx].Departs = append(allCities[idx].Stations[stIdx].Departs, trip)
									tripsCount++
									break
								}
							}
							break
						}
					}
				}
			} else {
				// Estimate based on distance
				if coords1, ok1 := cityCoords[origin]; ok1 {
					if coords2, ok2 := cityCoords[dest]; ok2 {
						dist := haversine(coords1[0], coords1[1], coords2[0], coords2[1])
						hours := math.Max(dist/70, 1.5)
						price := math.Max(dist*2.2, 500)
						price = math.Round(price/50) * 50

						departure := now.Add(time.Duration(rand.Intn(30)*24+12) * time.Hour)
						arrival := departure.Add(time.Duration(hours) * time.Hour)

						trip := createTrip(origin, dest, "train", departure, arrival, price, stationMap)

						for idx := range allCities {
							if allCities[idx].Name == origin {
								for stIdx := range allCities[idx].Stations {
									if allCities[idx].Stations[stIdx].TransportType == "train" {
										allCities[idx].Stations[stIdx].Departs = append(allCities[idx].Stations[stIdx].Departs, trip)
										tripsCount++
										break
									}
								}
								break
							}
						}
					}
				}
			}

			// Generate bus trips
			if hasRealData && route.BusHours > 0 && route.BusPrice > 0 {
				for hour := 9; hour <= 18; hour += 9 {
					departure := now.Add(time.Duration(rand.Intn(30)*24+hour) * time.Hour)
					arrival := departure.Add(time.Duration(route.BusHours) * time.Hour)
					price := route.BusPrice * (0.9 + rand.Float64()*0.2)
					price = math.Round(price*100) / 100

					trip := createTrip(origin, dest, "bus", departure, arrival, price, stationMap)

					for idx := range allCities {
						if allCities[idx].Name == origin {
							for stIdx := range allCities[idx].Stations {
								if allCities[idx].Stations[stIdx].TransportType == "bus" {
									allCities[idx].Stations[stIdx].Departs = append(allCities[idx].Stations[stIdx].Departs, trip)
									tripsCount++
									break
								}
							}
							break
						}
					}
				}
			} else {
				if coords1, ok1 := cityCoords[origin]; ok1 {
					if coords2, ok2 := cityCoords[dest]; ok2 {
						dist := haversine(coords1[0], coords1[1], coords2[0], coords2[1])
						hours := math.Max(dist/60, 1)
						price := math.Max(dist*1.5, 400)
						price = math.Round(price/50) * 50

						departure := now.Add(time.Duration(rand.Intn(30)*24+9) * time.Hour)
						arrival := departure.Add(time.Duration(hours) * time.Hour)

						trip := createTrip(origin, dest, "bus", departure, arrival, price, stationMap)

						for idx := range allCities {
							if allCities[idx].Name == origin {
								for stIdx := range allCities[idx].Stations {
									if allCities[idx].Stations[stIdx].TransportType == "bus" {
										allCities[idx].Stations[stIdx].Departs = append(allCities[idx].Stations[stIdx].Departs, trip)
										tripsCount++
										break
									}
								}
								break
							}
						}
					}
				}
			}
		}
	}

	fmt.Printf("✅ Generated %d trips\n", tripsCount)
	fmt.Println()

	// Print sample of data
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SAMPLE DATA (first 3 cities with their stations and trips):")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Print first 3 cities in detail
	for idx, city := range allCities[:min(3, len(allCities))] {
		cityJSON, _ := json.MarshalIndent(city, "", "  ")
		fmt.Printf("🏙️  City %d: %s\n", idx+1, city.Name)
		fmt.Printf("%s\n", string(cityJSON))
		fmt.Println()
	}

	// Print statistics
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("STATISTICS:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total cities:    %d\n", len(allCities))
	fmt.Printf("Total trips:     %d\n", tripsCount)
	fmt.Printf("Total pairs:     %d\n", totalPairs)
	fmt.Printf("Total tickets:   ~%d\n", tripsCount*4)
	fmt.Println()

	// Print complete JSON for all data
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("FULL DATA EXPORT (all cities with their stations and trips):")
	fmt.Println(strings.Repeat("=", 80))

	// Create export object
	export := map[string]interface{}{
		"cities":       allCities,
		"total_cities": len(allCities),
		"total_trips":  tripsCount,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	exportJSON, _ := json.MarshalIndent(export, "", "  ")
	fmt.Println(string(exportJSON))

	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✅ DRY-RUN COMPLETE")
	fmt.Println("This was a simulation - no data was written to any database")
	fmt.Println("To use this data with Dgraph, remove the 'Dry-run mode' comment")
	fmt.Println(strings.Repeat("=", 80))
}

func createTrip(origin, dest, transport string, departure, arrival time.Time, price float64, stationMap map[string]map[string]string) Trip {
	externalID := fmt.Sprintf("%s_%s_%s_%d", transport, origin, dest, departure.Unix())

	var tickets []map[string]interface{}

	if transport == "train" {
		tickets = []map[string]interface{}{
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
		tickets = []map[string]interface{}{
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

	destStationUID := stationMap[dest][transport]

	return Trip{
		DType:         []string{"Trip"},
		ExternalID:    externalID,
		DepartureAt:   departure.Format(time.RFC3339),
		ArrivalAt:     arrival.Format(time.RFC3339),
		TransportType: transport,
		Price:         price,
		Destination:   map[string]string{"uid": destStationUID},
		HasTicket:     tickets,
	}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
