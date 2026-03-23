package rzd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURLSuggest = "https://ticket.rzd.ru/api/v1/suggests"
	baseURLTrains  = "https://ticket.rzd.ru/api/v1/railway-service/prices/train-pricing"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) do(req *http.Request, dest interface{}) error {
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; rzd-rid-client-dto/1.0)")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", "https://rzd.ru")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		return fmt.Errorf("rzd.Client.do request=%s response=%s: %w", req, string(body), err)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// ParseTrains получает список поездов между двумя станциями на указанную дату
func (c *Client) ParseTrains(origin, destination string, departureDate time.Time) (*TrainResponse, error) {
	if len(origin) != 7 || len(destination) != 7 {
		return nil, fmt.Errorf("коды станций должны состоять ровно из 7 символов")
	}

	req, err := http.NewRequest(http.MethodGet, baseURLTrains, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("origin", origin)
	q.Add("destination", destination)
	q.Add("departureDate", departureDate.Format(time.RFC3339)[:10])

	q.Add("serviceProvider", "B2B_RZD")
	q.Add("getByLocalTime", "true")
	q.Add("carGrouping", "DontGroup")
	q.Add("specialPlacesDemand", "StandardPlacesAndForDisabledPersons")
	q.Add("carIssuingType", "Passenger")
	q.Add("getTrainsFromSchedule", "false")
	q.Add("adultPassengersQuantity", "1")
	q.Add("childrenPassengersQuantity", "0")
	req.URL.RawQuery = q.Encode()

	var result TrainResponse
	if err := c.do(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SuggestCity ищет станции/города по префиксу
func (c *Client) SuggestCity(prefix string) (*SuggestResponse, error) {
	req, err := http.NewRequest(http.MethodGet, baseURLSuggest, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("Query", prefix)
	q.Add("TransportType", "bus,avia,rail,aeroexpress,suburban,boat")
	q.Add("GroupResults", "true")
	q.Add("RailwaySortPriority", "true")
	q.Add("SynonymOn", "1")
	q.Add("Language", "ru")
	req.URL.RawQuery = q.Encode()

	var result SuggestResponse
	if err := c.do(req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
