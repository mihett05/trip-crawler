package processors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mihett05/trip-crawler/internal/worker"
	"github.com/mihett05/trip-crawler/internal/worker/dto"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
)

const banNakDelay = time.Minute

// Train and TicketParseResult are published to TopicTicketsResult.

type Train struct {
	ExternalID  string    `json:"external_id"`
	DepartureAt time.Time `json:"departure_at"`
	ArrivalAt   time.Time `json:"arrival_at"`
	MinPrice    *float64  `json:"min_price,omitempty"`
}

type TicketParseResult struct {
	OriginCode      string    `json:"origin_code"`
	DestinationCode string    `json:"destination_code"`
	DepartureDate   time.Time `json:"departure_date"`
	Trains          []Train   `json:"trains"`
}

// ProxyManager provides HTTP clients backed by proxy IPs and can ban a client
// whose proxy was detected as blocked by the remote service.
type ProxyManager interface {
	GetClient(ctx context.Context) (*http.Client, error)
	Ban(ctx context.Context, client *http.Client) error
}

// TicketParseProcessor subscribes to TopicTicketsParse, fetches train prices
// from RZD, and publishes a TicketParseResult.
type TicketParseProcessor struct {
	proxyManager ProxyManager
	gateway      worker.Gateway
}

func NewTicketParseProcessor(proxyManager ProxyManager, gateway worker.Gateway) *TicketParseProcessor {
	return &TicketParseProcessor{proxyManager: proxyManager, gateway: gateway}
}

// Run implements worker.Runner. It blocks until ctx is cancelled.
func (p *TicketParseProcessor) Run(ctx context.Context) error {
	return p.gateway.Subscribe(ctx, worker.TopicTicketsParse,
		func(ctx context.Context, payload []byte) error {
			var task dto.TicketParseTask
			if err := json.Unmarshal(payload, &task); err != nil {
				return nil // malformed payload — drop, no retry
			}
			result, err := p.process(ctx, task)
			if err != nil {
				return err // NakError for ban, regular error for transient failures
			}
			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("marshal TicketParseResult: %w", err)
			}
			return p.gateway.Publish(ctx, worker.TopicTicketsResult, data)
		},
	)
}

func (p *TicketParseProcessor) process(ctx context.Context, task dto.TicketParseTask) (*TicketParseResult, error) {
	httpClient, err := p.proxyManager.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("proxyManager.GetClient: %w", err)
	}

	resp, err := rzd.NewClientWithHTTP(httpClient).ParseTrains(
		task.OriginCode, task.DestinationCode, task.DepartureDate,
	)
	if err != nil {
		var statusErr *rzd.StatusError
		if errors.As(err, &statusErr) && isBanned(statusErr.Code) {
			_ = p.proxyManager.Ban(ctx, httpClient)
			return nil, &worker.NakError{
				Cause: fmt.Errorf("proxy banned (HTTP %d) on %s→%s: %w",
					statusErr.Code, task.OriginCode, task.DestinationCode, err),
				Delay: banNakDelay,
			}
		}
		return nil, fmt.Errorf("rzd.ParseTrains(%s→%s %s): %w",
			task.OriginCode, task.DestinationCode, task.DepartureDate.Format(time.DateOnly), err)
	}

	return &TicketParseResult{
		OriginCode:      task.OriginCode,
		DestinationCode: task.DestinationCode,
		DepartureDate:   task.DepartureDate,
		Trains:          convertTrains(resp),
	}, nil
}

func isBanned(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusForbidden
}

func convertTrains(resp *rzd.TrainResponse) []Train {
	trains := make([]Train, 0, len(resp.Trains))
	for _, t := range resp.Trains {
		dep, err := time.Parse("2006-01-02T15:04:05", t.DepartureDateTime)
		if err != nil {
			continue
		}
		arr, err := time.Parse("2006-01-02T15:04:05", t.ArrivalDateTime)
		if err != nil {
			continue
		}
		var minPrice *float64
		for _, cg := range t.CarGroups {
			if cg.MinPrice != nil && (minPrice == nil || *cg.MinPrice < *minPrice) {
				v := *cg.MinPrice
				minPrice = &v
			}
		}
		trains = append(trains, Train{
			ExternalID:  t.TrainNumber,
			DepartureAt: dep,
			ArrivalAt:   arr,
			MinPrice:    minPrice,
		})
	}
	return trains
}
