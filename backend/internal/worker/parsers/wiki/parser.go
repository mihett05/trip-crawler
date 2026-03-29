package wiki

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// ParseTables находит таблицы со списком городов и извлекает из них данные
func ParseTables(doc *goquery.Document) []CityData {
	html, err := doc.Html()
	if err == nil {
		os.WriteFile("wiki_page.html", []byte(html), 0644)
		fmt.Printf("[wiki] save wiki page\n")
	}

	var citiesData []CityData

	keyHeaders := []string{
		"№",
		"город",
		"субъектрф",
		"населениеперепись2021[29]",
		"населениеперепись2010[30]",
	}

	fmt.Printf("[wiki] scanning tables in document\n")
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		var headers []string
		table.Find("th").Each(func(i int, th *goquery.Selection) {
			headers = append(headers, cleanHeaderText(th.Text()))
		})

		if !isValidTable(keyHeaders, headers) {
			return
		}

		fmt.Printf("[wiki] found valid cities table (index=%d)\n", i)
		table.Find("tr").Each(func(i int, tr *goquery.Selection) {
			if i == 0 {
				return
			}

			tds := tr.Find("td")
			if tds.Length() != 5 {
				return
			}

			citiesData = append(citiesData, CityData{
				ID:         strings.TrimSpace(tds.Eq(0).Text()),
				Name:       strings.TrimSpace(tds.Eq(1).Text()),
				Region:     strings.TrimSpace(tds.Eq(2).Text()),
				Population: strings.TrimSpace(tds.Eq(3).Text()),
			})
		})
	})

	fmt.Printf("[wiki] parsed %d cities total\n", len(citiesData))

	sort.Slice(citiesData, func(i, j int) bool {
		return citiesData[i].Name < citiesData[j].Name
	})

	return citiesData
}

// cleanHeaderText удаляет все пробелы и переводит строку в нижний регистр
func cleanHeaderText(text string) string {
	var builder strings.Builder
	for _, r := range text {
		if !unicode.IsSpace(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}

// isValidTable проверяет, содержат ли заголовки таблицы нужные нам ключи
func isValidTable(keys, headers []string) bool {
	if len(headers) < len(keys) {
		return false
	}
	for i, key := range keys {
		if !strings.Contains(headers[i], key) {
			return false
		}
	}
	return true
}
