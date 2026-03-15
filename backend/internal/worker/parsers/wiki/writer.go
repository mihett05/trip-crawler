package wiki

import (
	"github.com/mihett05/trip-crawler/internal/worker/parsers/utils"
)

// WriteFullData сохраняет полный массив объектов
func WriteFullData(payload []CityData, filename string) error {
	return utils.SaveJSON(payload, filename)
}

// WriteCitiesList сохраняет и сырой список, и сгруппированный
func WriteCitiesList(citiesData []CityData, rawFilename, groupedFilename string) error {
	var cityNames []string
	for _, city := range citiesData {
		cityNames = append(cityNames, city.Name)
	}

	if err := utils.SaveJSON(cityNames, rawFilename); err != nil {
		return err
	}

	groupedDict := utils.BuildPrefixGroups(cityNames, func(c string) string {
		return c
	}, 3, 20)

	return utils.SaveJSON(groupedDict, groupedFilename)
}
