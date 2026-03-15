package dgraph

import (
	"fmt"

	"github.com/dgraph-io/dgo/v250"

	"github.com/mihett05/trip-crawler/pkg/application/config"
)

type Client struct {
	Client *dgo.Dgraph
}

func New(cfg config.Config) (*Client, error) {
	client, err := dgo.Open(cfg.DGraphConnection)
	if err != nil {
		return nil, fmt.Errorf("dgo.Open: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}
