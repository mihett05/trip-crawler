package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sort"

	citycsv "github.com/mihett05/trip-crawler/internal/worker/parsers/cities-by-csv"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/wiki"
	"github.com/mihett05/trip-crawler/pkg/messages"
)

// nonDigits нужно для очистки строки населения от пробелов и ссылок типа [29]
var nonDigits = regexp.MustCompile(`\D`)

func (h *Handler) HandleCitiesRequested(ctx context.Context, request messages.CitiesRequested) error {
	fmt.Printf("[cities] starting hybrid parsing: Top-%d cities + CSV routes\n", request.TopCities)

	// 1. Получаем ТОП городов из Википедии
	topWikiCities, err := fetchTopCities(request.TopCities)
	if err != nil {
		return fmt.Errorf("fetchTopCities: %w", err)
	}

	fmt.Println("[cities] Target cities from Wiki:")
	for i, c := range topWikiCities {
		fmt.Printf("  %d. %s (pop: %s)\n", i+1, c.Name, c.Population)
	}

	// 2. Читаем все маршруты из CSV
	allRoutes, err := citycsv.GetRoutesFromCSV()
	if err != nil {
		return fmt.Errorf("citycsv.GetRoutesFromCSV: %w", err)
	}

	// 3. Создаем карту городов
	cityMap := make(map[string]*messages.City)
	for _, wc := range topWikiCities {
		cityMap[strings.ToLower(wc.Name)] = &messages.City{
			ID:         wc.ID,
			Name:       wc.Name,
			Population: parsePopulation(wc.Population),
			Stations:   []messages.Station{},
		}
	}

	// 4. Фильтруем станции из CSV
	uniqueStations := make(map[string]bool)
	for _, route := range allRoutes {
		processStation := func(csvID, csvName string) {
			csvNameLower := strings.ToLower(csvName)
			for cityNameLower, cityObj := range cityMap {
				if strings.Contains(csvNameLower, cityNameLower) {
					stationKey := cityObj.ID + "_" + csvID
					if !uniqueStations[stationKey] {
						cityObj.Stations = append(cityObj.Stations, messages.Station{
							ID:   csvID,
							Name: csvName,
						})
						uniqueStations[stationKey] = true
					}
					break 
				}
			}
		}
		processStation(route.Departure.ID, route.Departure.Name)
		processStation(route.Arrival.ID, route.Arrival.Name)
	}

	// 5. Собираем финальный список и выводим статистику
	var finalCities []messages.City
	fmt.Println("[cities] Matching results:")
	for _, city := range cityMap {
		if len(city.Stations) > 0 {
			finalCities = append(finalCities, *city)
			fmt.Printf("  [MATCH] %s: found %d stations\n", city.Name, len(city.Stations))
		} else {
			fmt.Printf("  [SKIP ] %s: no stations found in CSV\n", city.Name)
		}
	}

	fmt.Printf("[cities] Total: %d cities with stations ready to send\n", len(finalCities))

	// 6. Отправка в NATS
	batchSize := 20
	for i := 0; i < len(finalCities); i += batchSize {
		end := i + batchSize
		if end > len(finalCities) {
			end = len(finalCities)
		}

		fmt.Printf("[cities] sending batch: %d - %d\n", i, end)
		err := h.gateway.SendCities(ctx, messages.CitiesParsed{
			Cities: finalCities[i:end],
		})
		if err != nil {
			return fmt.Errorf("gateway.SendCities: %w", err)
		}
	}

	fmt.Println("[cities] hybrid parsing finished successfully")
	return nil
}

func fetchTopCities(topN int) ([]wiki.CityData, error) {
	doc, err := wiki.LoadWikiPage()
	if err != nil {
		return nil, fmt.Errorf("wiki.LoadWikiPage: %w", err)
	}

	// 1. Получаем все города (они придут по алфавиту)
	cities := wiki.ParseTables(doc)
	fmt.Printf("[cities] wiki: parsed %d total cities\n", len(cities))

	// 2. Сортируем по населению (ОТ БОЛЬШЕГО К МЕНЬШЕМУ)
	sort.Slice(cities, func(i, j int) bool {
		// Используем parsePopulation для сравнения чисел, а не строк
		return parsePopulation(cities[i].Population) > parsePopulation(cities[j].Population)
	})

	// 3. Теперь отрезаем ТОП-N
	if topN > 0 && topN < len(cities) {
		cities = cities[:topN]
	}

	return cities, nil
}

func parsePopulation(s string) int64 {
	digits := nonDigits.ReplaceAllString(s, "")
	n, _ := strconv.ParseInt(digits, 10, 64)
	return n
}