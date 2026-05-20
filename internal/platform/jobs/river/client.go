package riverjobs

import (
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
)

type Config struct {
	Queues              map[string]river.QueueConfig
	Schema              string
	SkipUnknownJobCheck bool
}

type ClientBundle struct {
	Driver *riverdatabasesql.Driver
	Client *river.Client[*sql.Tx]
}

func NewClientBundle(db *sql.DB, cfg Config, workers *river.Workers) (*ClientBundle, error) {
	if db == nil {
		return nil, fmt.Errorf("river jobs: sql db is nil")
	}
	if workers == nil {
		workers = river.NewWorkers()
	}

	driver := riverdatabasesql.New(db)
	client, err := river.NewClient(driver, &river.Config{
		Queues:              cfg.Queues,
		Schema:              cfg.Schema,
		SkipUnknownJobCheck: cfg.SkipUnknownJobCheck,
		Workers:             workers,
	})
	if err != nil {
		return nil, fmt.Errorf("river jobs: new client: %w", err)
	}

	return &ClientBundle{
		Driver: driver,
		Client: client,
	}, nil
}
