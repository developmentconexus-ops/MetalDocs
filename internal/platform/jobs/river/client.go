package riverjobs

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
)

// Config holds the River client configuration: queues, schema, retention
// policy, and registered periodic jobs.
type Config struct {
	Queues              map[string]river.QueueConfig
	Schema              string
	SkipUnknownJobCheck bool
	PeriodicJobs        []*river.PeriodicJob

	// CompletedJobRetentionPeriod, CancelledJobRetentionPeriod, and
	// DiscardedJobRetentionPeriod are forwarded verbatim to the identically
	// named fields on river.Config. The zero value (time.Duration(0)) is NOT
	// special-cased here: River's own Config treats an unset (zero)
	// retention period as "use my documented default" (24h / 24h / 7 days
	// respectively) — only the special value -1 disables deletion entirely.
	// Callers that care about the retention policy (contract-bound values
	// per M5 F5.4) must set these explicitly; callers that don't set them
	// simply inherit River's own defaults, which is the same behavior as
	// before these fields existed.
	CompletedJobRetentionPeriod time.Duration
	CancelledJobRetentionPeriod time.Duration
	DiscardedJobRetentionPeriod time.Duration
}

// ClientBundle bundles a River client with its underlying database driver.
type ClientBundle struct {
	// Driver is intentionally public for wiring; do not use outside bootstrap.
	Driver *riverdatabasesql.Driver
	Client *river.Client[*sql.Tx]
}

// NewClientBundle constructs a River client and driver bundle against db,
// registering workers (or an empty river.Workers when nil).
func NewClientBundle(db *sql.DB, cfg Config, workers *river.Workers) (*ClientBundle, error) {
	if db == nil {
		return nil, fmt.Errorf("river jobs: sql db is nil")
	}
	if workers == nil {
		workers = river.NewWorkers()
	}

	driver := riverdatabasesql.New(db)
	client, err := river.NewClient(driver, &river.Config{
		Queues:                      cfg.Queues,
		Schema:                      cfg.Schema,
		SkipUnknownJobCheck:         cfg.SkipUnknownJobCheck,
		Workers:                     workers,
		PeriodicJobs:                cfg.PeriodicJobs,
		CompletedJobRetentionPeriod: cfg.CompletedJobRetentionPeriod,
		CancelledJobRetentionPeriod: cfg.CancelledJobRetentionPeriod,
		DiscardedJobRetentionPeriod: cfg.DiscardedJobRetentionPeriod,
	})
	if err != nil {
		return nil, fmt.Errorf("river jobs: new client: %w", err)
	}

	return &ClientBundle{
		Driver: driver,
		Client: client,
	}, nil
}
