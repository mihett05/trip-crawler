package graph

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"github.com/mihett05/trip-crawler/internal/service/core/dgraph"
	"github.com/mihett05/trip-crawler/internal/service/domain/models"
)

type Repository struct {
	dg *dgraph.Client
}

func NewRepository(dg *dgraph.Client) *Repository {
	return &Repository{dg: dg}
}

func (r *Repository) SaveTrip(ctx context.Context, trip *models.Trip) error {
	trip.Type = []string{"Trip"}

	txn := r.dg.Client.NewTxn()
	defer txn.Discard(ctx)

	pb, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}

	_, err = txn.Mutate(ctx, mu)
	return err
}
