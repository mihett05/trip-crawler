package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mihett05/trip-crawler/internal/worker"
	"github.com/mihett05/trip-crawler/internal/worker/dto"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/utils"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/wiki"
)

const (
	prefixLen    = 3
	maxGroupSize = 20
)

// Station and City are the result domain types published to TopicStationsResult.

type Station struct {
	ExternalID    string `json:"external_id"` // RZD 7-char expressCode
	Name          string `json:"name"`
	TransportType string `json:"transport_type"`
}

type City struct {
	Name       string    `json:"name"`
	Population int       `json:"population"`
	Stations   []Station `json:"stations"`
}

// StationLoadResult is published to TopicStationsResult after a successful run.
type StationLoadResult struct {
	Cities []City `json:"cities"`
}

// StationLoadProcessor subscribes to TopicStationsLoad, scrapes Wikipedia and
// RZD for city/station data, and publishes a StationLoadResult.
type StationLoadProcessor struct {
	rzdClient *rzd.Client
	gateway   worker.Gateway
}

func NewStationLoadProcessor(rzdClient *rzd.Client, gateway worker.Gateway) *StationLoadProcessor {
	return &StationLoadProcessor{rzdClient: rzdClient, gateway: gateway}
}

// Run implements worker.Runner. It blocks until ctx is cancelled.
func (p *StationLoadProcessor) Run(ctx context.Context) error {
	return p.gateway.Subscribe(ctx, worker.TopicStationsLoad,
		func(ctx context.Context, payload []byte) error {
			var task dto.StationLoadTask
			if err := json.Unmarshal(payload, &task); err != nil {
				return nil // malformed payload — drop, no retry
			}
			result, err := p.process(ctx, task)
			if err != nil {
				return err
			}
			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("marshal StationLoadResult: %w", err)
			}
			return p.gateway.Publish(ctx, worker.TopicStationsResult, data)
		},
	)
}

func (p *StationLoadProcessor) process(ctx context.Context, task dto.StationLoadTask) (*StationLoadResult, error) {
	doc, err := wiki.LoadWikiPage()
	if err != nil {
		return nil, fmt.Errorf("wiki.LoadWikiPage: %w", err)
	}

	wikiCities := wiki.ParseTables(doc)
	wikiCities = topNByPopulation(wikiCities, task.TopN)

	groups := utils.BuildPrefixGroups(wikiCities, func(c wiki.CityData) string {
		return c.Name
	}, prefixLen, maxGroupSize)

	var cities []City

	for prefix, group := range groups {
		resp, err := p.rzdClient.SuggestCity(prefix)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			continue // non-fatal: skip this prefix group
		}

		rzdIDByName := make(map[string]string, len(resp.City))
		for _, node := range resp.City {
			rzdIDByName[strings.ToLower(node.Name)] = node.NodeID
		}

		stationsByRZDID := make(map[string][]Station)
		for _, node := range resp.Train {
			if node.TransportType != "rail" {
				continue
			}
			stationsByRZDID[node.CityID] = append(stationsByRZDID[node.CityID], Station{
				ExternalID:    node.ExpressCode,
				Name:          node.Name,
				TransportType: node.TransportType,
			})
		}

		for _, c := range group {
			rzdID := rzdIDByName[strings.ToLower(c.Name)]
			cities = append(cities, City{
				Name:       c.Name,
				Population: parsePopulation(c.Population),
				Stations:   stationsByRZDID[rzdID],
			})
		}
	}

	return &StationLoadResult{Cities: cities}, nil
}

func topNByPopulation(cities []wiki.CityData, n int) []wiki.CityData {
	type scored struct {
		city wiki.CityData
		pop  int
	}
	s := make([]scored, 0, len(cities))
	for _, c := range cities {
		s = append(s, scored{c, parsePopulation(c.Population)})
	}
	sort.Slice(s, func(i, j int) bool { return s[i].pop > s[j].pop })
	if n > len(s) {
		n = len(s)
	}
	result := make([]wiki.CityData, n)
	for i := range result {
		result[i] = s[i].city
	}
	return result
}

func parsePopulation(s string) int {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
	if digits == "" {
		return 0
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return n
}
