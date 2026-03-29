package wiki

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const wikiURL = "https://ru.wikipedia.org/wiki/Города_России"

// LoadWikiPage скачивает страницу Википедии и возвращает объект для парсинга
func LoadWikiPage() (*goquery.Document, error) {
	fmt.Printf("[wiki] fetching %s\n", wikiURL)

	req, err := http.NewRequest(http.MethodGet, wikiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; rzd-rid-client-dto/1.0)")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", "https://rzd.ru")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP EXP: %d", resp.StatusCode)
	}

	fmt.Printf("[wiki] page fetched, status=%d\n", resp.StatusCode)
	return goquery.NewDocumentFromReader(resp.Body)
}
