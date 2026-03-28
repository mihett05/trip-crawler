package station

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mihett05/trip-crawler/internal/worker/dto"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/utils"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/wiki"
	"github.com/mihett05/trip-crawler/internal/worker/throttle"
)

const (
	prefixLen    = 3
	maxGroupSize = 20
)

// City and Station are the domain types used to persist data.
// They are defined here so the repository contract does not depend on any
// parser-internal DTOs.

type Station struct {
	ExternalID    string // RZD expressCode (7-char)
	Name          string
	TransportType string
}

type City struct {
	Name       string
	Population int
	Stations   []Station
}

// StationRepository persists city+station data to the graph store.
type StationRepository interface {
	UpsertCity(ctx context.Context, city City) error
}

// StationLoadProcessor refreshes the city/station graph on demand.
// It scrapes Wikipedia for the top-N Russian cities by population, then
// calls the RZD suggest API to resolve their train station codes.
type StationLoadProcessor struct {
	throttler *throttle.Throttler
	rzdClient *rzd.Client
	repo      StationRepository
}

func NewStationLoadProcessor(
	throttler *throttle.Throttler,
	rzdClient *rzd.Client,
	repo StationRepository,
) *StationLoadProcessor {
	return &StationLoadProcessor{throttler: throttler, rzdClient: rzdClient, repo: repo}
}

func (p *StationLoadProcessor) Process(ctx context.Context, task dto.StationLoadTask) error {
	doc, err := wiki.LoadWikiPage()
	if err != nil {
		return fmt.Errorf("wiki.LoadWikiPage: %w", err)
	}

	cities := wiki.ParseTables(doc)
	cities = topNByPopulation(cities, task.TopN)

	// Group cities by name prefix so that one SuggestCity call covers the whole
	// group instead of making a separate request for each city.
	groups := utils.BuildPrefixGroups(cities, func(c wiki.CityData) string {
		return c.Name
	}, prefixLen, maxGroupSize)

	for prefix, group := range groups {
		if err := p.throttler.Wait(ctx); err != nil {
			return err
		}

		resp, err := p.rzdClient.SuggestCity(prefix)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Non-fatal: skip this prefix group and continue.
			continue
		}

		// Index RZD city NodeID by normalised city name.
		// City nodes in the response carry the NodeID; Train nodes carry CityID
		// that references the parent City node.
		rzdIDByName := make(map[string]string, len(resp.City))
		for _, node := range resp.City {
			rzdIDByName[strings.ToLower(node.Name)] = node.NodeID
		}

		// Index rail stations by their parent RZD city NodeID.
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
			if err := p.repo.UpsertCity(ctx, City{
				Name:       c.Name,
				Population: parsePopulation(c.Population),
				Stations:   stationsByRZDID[rzdID], // empty slice is fine
			}); err != nil {
				return fmt.Errorf("repo.UpsertCity(%s): %w", c.Name, err)
			}
		}
	}

	return nil
}

// topNByPopulation sorts cities by parsed population descending and returns
// the top n. It does not modify the original slice.
func topNByPopulation(cities []wiki.CityData, n int) []wiki.CityData {
	type scored struct {
		city wiki.CityData
		pop  int
	}

	scored_ := make([]scored, 0, len(cities))
	for _, c := range cities {
		scored_ = append(scored_, scored{c, parsePopulation(c.Population)})
	}

	sort.Slice(scored_, func(i, j int) bool {
		return scored_[i].pop > scored_[j].pop
	})

	if n > len(scored_) {
		n = len(scored_)
	}

	result := make([]wiki.CityData, n)
	for i := range result {
		result[i] = scored_[i].city
	}
	return result
}

// parsePopulation converts a Russian-formatted population string (e.g. "1 234 567")
// to an integer. Returns 0 if parsing fails.
func parsePopulation(s string) int {
	// Strip everything that isn't a digit.
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
