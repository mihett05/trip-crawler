package ticket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mihett05/trip-crawler/internal/worker"
	"github.com/mihett05/trip-crawler/internal/worker/dto"
	"github.com/mihett05/trip-crawler/internal/worker/parsers/rzd"
	"github.com/mihett05/trip-crawler/internal/worker/throttle"
)

const banNakDelay = time.Minute

// Train is the domain type written to the graph store after a successful parse.
type Train struct {
	ExternalID    string
	OriginCode    string
	DestCode      string
	DepartureAt   time.Time
	ArrivalAt     time.Time
	MinPrice      *float64
	TransportType string
}

// TripRepository persists parsed train data.
type TripRepository interface {
	SaveTrains(ctx context.Context, trains []Train) error
}

// ProxyManager provides HTTP clients backed by proxy IPs and allows banning
// a client whose proxy was detected as blocked by the remote service.
type ProxyManager interface {
	GetClient(ctx context.Context) (*http.Client, error)
	Ban(ctx context.Context, client *http.Client) error
}

// TicketParseProcessor fetches train prices for a single (origin, dest, date)
// triple and persists the result.
type TicketParseProcessor struct {
	throttler    *throttle.Throttler
	proxyManager ProxyManager
	repo         TripRepository
}

func NewTicketParseProcessor(
	throttler *throttle.Throttler,
	proxyManager ProxyManager,
	repo TripRepository,
) *TicketParseProcessor {
	return &TicketParseProcessor{throttler: throttler, proxyManager: proxyManager, repo: repo}
}

func (p *TicketParseProcessor) Process(ctx context.Context, task dto.TicketParseTask) error {
	if err := p.throttler.Wait(ctx); err != nil {
		return err
	}

	httpClient, err := p.proxyManager.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("proxyManager.GetClient: %w", err)
	}

	rzdClient := rzd.NewClientWithHTTP(httpClient)
	resp, err := rzdClient.ParseTrains(task.OriginCode, task.DestinationCode, task.DepartureDate)
	if err != nil {
		var statusErr *rzd.StatusError
		if errors.As(err, &statusErr) && isBanned(statusErr.Code) {
			_ = p.proxyManager.Ban(ctx, httpClient)
			return &worker.NakError{
				Cause: fmt.Errorf("proxy banned (HTTP %d) on %s→%s: %w", statusErr.Code, task.OriginCode, task.DestinationCode, err),
				Delay: banNakDelay,
			}
		}
		return fmt.Errorf("rzd.ParseTrains(%s→%s %s): %w",
			task.OriginCode, task.DestinationCode, task.DepartureDate.Format(time.DateOnly), err)
	}

	trains := convertTrains(task.OriginCode, task.DestinationCode, resp)
	if len(trains) == 0 {
		return nil
	}

	if err := p.repo.SaveTrains(ctx, trains); err != nil {
		return fmt.Errorf("repo.SaveTrains: %w", err)
	}

	return nil
}

func isBanned(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusForbidden
}

func convertTrains(originCode, destCode string, resp *rzd.TrainResponse) []Train {
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
			ExternalID:    t.TrainNumber,
			OriginCode:    originCode,
			DestCode:      destCode,
			DepartureAt:   dep,
			ArrivalAt:     arr,
			MinPrice:      minPrice,
			TransportType: "rail",
		})
	}
	return trains
}
