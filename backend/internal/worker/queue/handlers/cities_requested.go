package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	parserutils "github.com/mihett05/trip-crawler/internal/worker/parsers/utils"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/wiki"
	"github.com/mihett05/trip-crawler/pkg/messages"
)

var nonDigits = regexp.MustCompile(`\D`)

func (h *Handler) HandleCitiesRequested(ctx context.Context, request messages.CitiesRequested) error {
	fmt.Printf("[cities] received request: top_cities=%d\n", request.TopCities)

	cityData, err := fetchTopCities(request.TopCities)
	if err != nil {
		return err
	}

	num := 0
	grouped := parserutils.BuildPrefixGroups(cityData, func(c wiki.CityData) string { return c.Name }, 3, 20)
	fmt.Printf("[cities] grouped into %d prefix buckets\n", len(grouped))

	for prefix, group := range grouped {
		cities, err := h.buildCitiesFromGroup(prefix, group)
		if err != nil {
			return err
		}

		fmt.Printf("[cities] sending group №%d prefix=%q count=%d\n", num, prefix, len(cities))
		if err := h.gateway.SendCities(ctx, messages.CitiesParsed{Cities: cities}); err != nil {
			return fmt.Errorf("gateway.SendCities(prefix=%s): %w", prefix, err)
		}
		num++

		sleep := time.Duration(15+rand.Intn(15)) * 10 * time.Millisecond
		fmt.Printf("[cities] sleeping for %s\n", sleep)
		time.Sleep(sleep)
	}

	return nil
}

func fetchTopCities(topN int) ([]wiki.CityData, error) {
	doc, err := wiki.LoadWikiPage()
	if err != nil {
		return nil, fmt.Errorf("wiki.LoadWikiPage: %w", err)
	}

	cities := wiki.ParseTables(doc)
	fmt.Printf("[cities] wiki: parsed %d cities\n", len(cities))

	sort.Slice(cities, func(i, j int) bool {
		return parsePopulation(cities[i].Population) > parsePopulation(cities[j].Population)
	})

	if topN > 0 && topN < len(cities) {
		cities = cities[:topN]
	}
	fmt.Printf("[cities] processing %d cities after top-N filter\n", len(cities))

	return cities, nil
}

func (h *Handler) buildCitiesFromGroup(prefix string, group []wiki.CityData) ([]messages.City, error) {
	fmt.Printf("[cities] rzd suggest: querying prefix=%q (%d cities)\n", prefix, len(group))
	suggest, err := h.rzd.SuggestCity(prefix)
	if err != nil {
		return nil, fmt.Errorf("rzd.SuggestCity(%s): %w", prefix, err)
	}
	fmt.Printf("[cities] rzd suggest: prefix=%q got %d city nodes, %d train nodes\n", prefix, len(suggest.City), len(suggest.Train))

	cityIDByName := make(map[string]string, len(suggest.City))
	for _, node := range suggest.City {
		cityIDByName[strings.ToLower(node.Name)] = node.CityID
	}

	trainsByCityID := mapSuggestToStations(suggest.Train)

	cities := make([]messages.City, 0, len(group))
	for _, cd := range group {
		cityID := cd.ID
		if id, ok := cityIDByName[strings.ToLower(cd.Name)]; ok {
			cityID = id
		}
		cities = append(cities, messages.City{
			ID:         cityID,
			Name:       cd.Name,
			Population: parsePopulation(cd.Population),
			Stations:   trainsByCityID[cityID],
		})
	}
	return cities, nil
}

func mapSuggestToStations(nodes []rzd.Node) map[string][]messages.Station {
	trainsByCityID := make(map[string][]messages.Station)
	for _, node := range nodes {
		trainsByCityID[node.CityID] = append(trainsByCityID[node.CityID], messages.Station{
			ID:            node.ExpressCode,
			Name:          node.Name,
			CityID:        node.CityID,
			TransportType: node.TransportType,
		})
	}
	return trainsByCityID
}

func parsePopulation(s string) int64 {
	digits := nonDigits.ReplaceAllString(s, "")
	n, _ := strconv.ParseInt(digits, 10, 64)
	return n
}
