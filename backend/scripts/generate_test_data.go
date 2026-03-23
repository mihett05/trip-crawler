package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// City represents a city in the database
type City struct {
	Uid  string   `json:"uid,omitempty"`
	Type []string `json:"dgraph.type,omitempty"`
	Name string   `json:"city.name,omitempty"`
}

// Station represents a station in the database
type Station struct {
	Uid           string   `json:"uid,omitempty"`
	Type          []string `json:"dgraph.type,omitempty"`
	Name          string   `json:"station.name,omitempty"`
	TransportType string   `json:"station.transport_type,omitempty"`
	City          *City    `json:"has_station,omitempty"` // Reverse relationship
}

// Trip represents a trip in the database
type Trip struct {
	Uid           string    `json:"uid,omitempty"`
	Type          []string  `json:"dgraph.type,omitempty"`
	ExternalID    string    `json:"trip.external_id,omitempty"`
	Price         float64   `json:"trip.price,omitempty"`
	DepartureAt   time.Time `json:"trip.departure_at,omitempty"`
	ArrivalAt     time.Time `json:"trip.arrival_at,omitempty"`
	TransportType string    `json:"trip.transport_type,omitempty"`
	Source        *Station  `json:"departs,omitempty"`     // Reverse relationship from station
	Destination   *Station  `json:"destination,omitempty"` // Forward relationship to destination
}

func main() {
	// Connect to Dgraph
	conn, err := grpc.Dial("localhost:9080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Dgraph: %v", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	ctx := context.Background()

	// Initialize schema
	err = initSchema(ctx, dc)
	if err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Generate Russian cities
	russianCities := []string{
		"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань",
		"Нижний Новгород", "Челябинск", "Самара", "Омск", "Ростов-на-Дону",
		"Уфа", "Красноярск", "Воронеж", "Пермь", "Волгоград",
	}

	// Create cities
	cityUIDs := make(map[string]string)
	for _, cityName := range russianCities {
		uid, err := createCity(ctx, dc, cityName)
		if err != nil {
			log.Printf("Failed to create city %s: %v", cityName, err)
			continue
		}
		cityUIDs[cityName] = uid
		log.Printf("Created city: %s with UID: %s", cityName, uid)
	}

	// Create stations for each city
	stationUIDs := make(map[string][]string)
	for cityName, cityUID := range cityUIDs {
		stationTypes := []string{"train", "bus", "plane"}

		// Create 2-3 stations per city with different transport types
		numStations := rand.Intn(2) + 2 // 2 to 3 stations
		var cityStations []string

		for j := 0; j < numStations; j++ {
			transportType := stationTypes[j%len(stationTypes)]
			stationName := fmt.Sprintf("%s %s Станция", cityName, transportType)

			uid, err := createStation(ctx, dc, stationName, transportType, cityUID)
			if err != nil {
				log.Printf("Failed to create station %s: %v", stationName, err)
				continue
			}
			cityStations = append(cityStations, uid)
			log.Printf("Created station: %s with UID: %s for city: %s", stationName, uid, cityName)
		}
		stationUIDs[cityName] = cityStations
	}

	// Generate trips between cities for the next 30 days
	log.Println("Generating trips...")

	// For each day in the next 30 days
	for dayOffset := 0; dayOffset < 30; dayOffset++ {
		startOfDay := time.Now().AddDate(0, 0, dayOffset).Truncate(24 * time.Hour)

		// Generate some trips for today
		numTrips := rand.Intn(10) + 5 // 5 to 15 trips per day
		for i := 0; i < numTrips; i++ {
			sourceCityIndex := rand.Intn(len(russianCities))
			destCityIndex := rand.Intn(len(russianCities))

			// Make sure source and destination are different
			if destCityIndex == sourceCityIndex {
				destCityIndex = (destCityIndex + 1) % len(russianCities)
			}

			sourceCity := russianCities[sourceCityIndex]
			destCity := russianCities[destCityIndex]

			// Pick random stations from source and destination cities
			if len(stationUIDs[sourceCity]) == 0 || len(stationUIDs[destCity]) == 0 {
				continue
			}

			sourceStationUID := stationUIDs[sourceCity][rand.Intn(len(stationUIDs[sourceCity]))]
			destStationUID := stationUIDs[destCity][rand.Intn(len(stationUIDs[destCity]))]

			// Generate random departure time during the day
			hour := rand.Intn(24)
			minute := rand.Intn(60)
			departureTime := time.Date(
				startOfDay.Year(), startOfDay.Month(), startOfDay.Day(),
				hour, minute, 0, 0, time.UTC,
			)

			// Arrival time is departure time + duration (min 1 hour, max 24 hours)
			durationHours := rand.Float64()*23 + 1
			arrivalTime := departureTime.Add(time.Duration(durationHours) * time.Hour)

			// Random price between 500 and 15000
			price := rand.Float64()*14500 + 500

			// Transport type based on source station
			// Note: We would need to get the actual transport type from the station
			// For now, we'll derive it from the station UID
			transportType := "train" // Default, would need to query actual station data

			// Create readable external ID using transliterated city names
			externalID := fmt.Sprintf("TRIP_%s_TO_%s_%d_%d",
				transliterateAndSanitize(sourceCity), transliterateAndSanitize(destCity), dayOffset, i)

			err := createTrip(ctx, dc, externalID, price, departureTime, arrivalTime,
				transportType, sourceStationUID, destStationUID)
			if err != nil {
				log.Printf("Failed to create trip %s: %v", externalID, err)
				continue
			}

			log.Printf("Created trip: %s -> %s on %s", sourceCity, destCity, departureTime.Format("2006-01-02 15:04"))
		}
	}

	log.Println("Test data generation completed!")
}

func initSchema(ctx context.Context, dc *dgo.Dgraph) error {
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

	return dc.Alter(ctx, op)
}

func createCity(ctx context.Context, dc *dgo.Dgraph, name string) (string, error) {
	city := City{
		Type: []string{"City"},
		Name: name,
	}

	jsonData, err := json.Marshal(city)
	if err != nil {
		return "", err
	}

	txn := dc.NewTxn()
	defer txn.Discard(ctx)

	mutation := &api.Mutation{
		SetJson:   jsonData,
		CommitNow: true,
	}

	response, err := txn.Mutate(ctx, mutation)
	if err != nil {
		return "", err
	}

	if len(response.Uids) == 0 {
		return "", fmt.Errorf("no UIDs returned for city creation")
	}

	// Find the first UID
	for _, uid := range response.Uids {
		return uid, nil
	}

	return "", fmt.Errorf("no UID found in response")
}

func createStation(ctx context.Context, dc *dgo.Dgraph, name, transportType, cityUID string) (string, error) {
	station := Station{
		Type:          []string{"Station"},
		Name:          name,
		TransportType: transportType,
		City:          &City{Uid: cityUID}, // This creates the reverse link
	}

	jsonData, err := json.Marshal(station)
	if err != nil {
		return "", err
	}

	txn := dc.NewTxn()
	defer txn.Discard(ctx)

	mutation := &api.Mutation{
		SetJson:   jsonData,
		CommitNow: true,
	}

	response, err := txn.Mutate(ctx, mutation)
	if err != nil {
		return "", err
	}

	if len(response.Uids) == 0 {
		return "", fmt.Errorf("no UIDs returned for station creation")
	}

	// Find the first UID
	for _, uid := range response.Uids {
		return uid, nil
	}

	return "", fmt.Errorf("no UID found in response")
}

func createTrip(ctx context.Context, dc *dgo.Dgraph, externalID string, price float64, departureAt, arrivalAt time.Time,
	transportType, sourceStationUID, destStationUID string) error {

	trip := Trip{
		Type:          []string{"Trip"},
		ExternalID:    externalID,
		Price:         price,
		DepartureAt:   departureAt,
		ArrivalAt:     arrivalAt,
		TransportType: transportType,
		Source:        &Station{Uid: sourceStationUID}, // This creates the reverse link from station to trip
		Destination:   &Station{Uid: destStationUID},
	}

	jsonData, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	txn := dc.NewTxn()
	defer txn.Discard(ctx)

	mutation := &api.Mutation{
		SetJson:   jsonData,
		CommitNow: true,
	}

	_, err = txn.Mutate(ctx, mutation)
	return err
}

func transliterateAndSanitize(name string) string {
	// Cyrillic to Latin transliteration map
	translitMap := map[rune]string{
		'А': "A", 'а': "a", 'Б': "B", 'б': "b", 'В': "V", 'в': "v",
		'Г': "G", 'г': "g", 'Д': "D", 'д': "d", 'Е': "E", 'е': "e",
		'Ё': "YO", 'ё': "yo", 'Ж': "ZH", 'ж': "zh", 'З': "Z", 'з': "z",
		'И': "I", 'и': "i", 'Й': "J", 'й': "j", 'К': "K", 'к': "k",
		'Л': "L", 'л': "l", 'М': "M", 'м': "m", 'Н': "N", 'н': "n",
		'О': "O", 'о': "o", 'П': "P", 'п': "p", 'Р': "R", 'р': "r",
		'С': "S", 'с': "s", 'Т': "T", 'т': "t", 'У': "U", 'у': "u",
		'Ф': "F", 'ф': "f", 'Х': "KH", 'х': "kh", 'Ц': "TS", 'ц': "ts",
		'Ч': "CH", 'ч': "ch", 'Ш': "SH", 'ш': "sh", 'Щ': "SCH", 'щ': "sch",
		'Ъ': "", 'ъ': "", 'Ы': "Y", 'ы': "y", 'Ь': "", 'ь': "",
		'Э': "E", 'э': "e", 'Ю': "YU", 'ю': "yu", 'Я': "YA", 'я': "ya",
		' ': "_", '-': "_", '.': "", ',': "",
	}

	var result string
	for _, char := range name {
		if latin, exists := translitMap[char]; exists {
			result += latin
		} else if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else {
			result += "_"
		}
	}
	return result
}

func sanitizeName(name string) string {
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}
