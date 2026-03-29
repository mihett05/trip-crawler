package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/mihett05/trip-crawler/pkg/messages"
)

const rzdDateTimeLayout = "2006-01-02T15:04:05"

func (h *Handler) HandleTripRequested(ctx context.Context, request messages.TripRequested) error {
	fmt.Printf("[trips] received request: %s (%s) -> %s (%s) at %d\n",
		request.DepartStation, request.DepartStationID,
		request.DestinationStation, request.DestinationStationID,
		request.DepartureAtTimestamp,
	)

	departureAt := time.Unix(request.DepartureAtTimestamp, 0)

	originCode := request.DepartStationID
	if originCode == "" {
		originCode = request.DepartStation
	}
	destCode := request.DestinationStationID
	if destCode == "" {
		destCode = request.DestinationStation
	}

	resp, err := h.rzd.ParseTrains(originCode, destCode, departureAt)
	if err != nil {
		return fmt.Errorf("rzd.ParseTrains: %w", err)
	}
	fmt.Printf("[trips] rzd: got %d trains\n", len(resp.Trains))

	for _, train := range resp.Trains {
		fmt.Printf("[trips] processing train %s (%s -> %s, depart=%s)\n",
			train.TrainNumber, train.OriginName, train.DestinationName, train.DepartureDateTime,
		)
		departureTS, err := time.ParseInLocation(rzdDateTimeLayout, train.DepartureDateTime, time.Local)
		if err != nil {
			return fmt.Errorf("parse DepartureDateTime(%s): %w", train.DepartureDateTime, err)
		}

		arrivalTS, err := time.ParseInLocation(rzdDateTimeLayout, train.ArrivalDateTime, time.Local)
		if err != nil {
			return fmt.Errorf("parse ArrivalDateTime(%s): %w", train.ArrivalDateTime, err)
		}

		tickets := make([]messages.Ticket, 0, len(train.CarGroups))
		for _, cg := range train.CarGroups {
			var price float64
			if cg.MinPrice != nil {
				price = *cg.MinPrice
			}
			tickets = append(tickets, messages.Ticket{
				Type:            cg.CarTypeName,
				Price:           price,
				AvailableAmount: int64(cg.TotalPlaceQuantity),
				TotalAmount:     int64(cg.TotalPlaceQuantity),
			})
		}

		trip := messages.Trip{
			ExternalID:           train.TrainNumber,
			DepartStationID:      resp.OriginCode,
			DestinationStationID: resp.DestinationCode,
			DepartureAtTimestamp: departureTS.Unix(),
			ArrivalAtTimestamp:   arrivalTS.Unix(),
			TransportType:        "train",
			Availability:         tickets,
		}

		fmt.Printf("[trips] publishing trip: train=%s depart=%d arrival=%d tickets=%d\n",
			trip.ExternalID, trip.DepartureAtTimestamp, trip.ArrivalAtTimestamp, len(trip.Availability),
		)
		if err := h.gateway.SendTrip(ctx, messages.TripParsed{Trip: trip}); err != nil {
			return fmt.Errorf("gateway.SendTrip: %w", err)
		}
	}

	return nil
}
