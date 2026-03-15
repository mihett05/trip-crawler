package city

import "github.com/mihett05/trip-crawler/internal/worker/parsers/utils"

const (
	MinPrefixLen = 3
	MaxGroupSize = 20
)

// BuildPrefixGroups использует общую утилиту
func BuildPrefixGroups(routes []StationRoute) map[string][]Station {
	var uniqueStations []Station

	return utils.BuildPrefixGroups(uniqueStations, func(s Station) string {
		return s.Name
	}, MinPrefixLen, MaxGroupSize)
}
