//go:build integration

package riverjobs_test

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/tests/integration/testdb"
)

type compatibilityArgs struct {
	DocumentID string `json:"document_id"`
}

func (compatibilityArgs) Kind() string { return "scheduled_publish_compatibility_probe" }

func TestRiverInsertTxCompatibilityBoundary(t *testing.T) {
	t.Helper()

	// Open the isolated test DB first; template bootstrap can take minutes on
	// first run. The 30-second operation timeout below applies only to the
	// river-specific work after the DB is ready. The testdb template already
	// bakes the River schema into the default schema (ApplyCuratedBootstrap ->
	// bootstrap.MigrateRiverSchema), so no per-test River migration is needed.
	db, _ := testdb.Open(t)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	bundle, err := riverjobs.NewClientBundle(db, riverjobs.Config{
		Schema:              "",
		SkipUnknownJobCheck: true,
	}, nil)
	if err != nil {
		t.Fatalf("new client bundle: %v", err)
	}

	clientType := reflect.TypeOf(bundle.Client).String()
	if !strings.Contains(clientType, "database/sql.Tx") {
		t.Fatalf("client type = %q; want database/sql transaction client", clientType)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin sql tx: %v", err)
	}
	defer tx.Rollback()

	if _, ok := any(tx).(pgx.Tx); ok {
		t.Fatal("expected current repo sql.Tx not to satisfy pgx.Tx directly")
	}

	scheduledAt := time.Now().UTC().Add(5 * time.Minute).Round(time.Microsecond)
	insertRes, err := bundle.Client.InsertTx(ctx, tx, compatibilityArgs{DocumentID: "doc-compat-1"}, &river.InsertOpts{
		Queue:       "temporal",
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		t.Fatalf("insert tx: %v", err)
	}

	if got := countRiverJobs(ctx, t, db); got != 0 {
		t.Fatalf("visible jobs outside tx before commit = %d, want 0", got)
	}
	if got := countRiverJobs(ctx, t, tx); got != 1 {
		t.Fatalf("visible jobs inside tx before commit = %d, want 1", got)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	if got := countRiverJobs(ctx, t, db); got != 1 {
		t.Fatalf("visible jobs outside tx after commit = %d, want 1", got)
	}
	if insertRes.Job.Queue != "temporal" {
		t.Fatalf("job queue = %q, want temporal", insertRes.Job.Queue)
	}
	if insertRes.Job.Kind != (compatibilityArgs{}).Kind() {
		t.Fatalf("job kind = %q, want %q", insertRes.Job.Kind, (compatibilityArgs{}).Kind())
	}
	if !insertRes.Job.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("job scheduled_at = %v, want %v", insertRes.Job.ScheduledAt, scheduledAt)
	}
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func countRiverJobs(ctx context.Context, t *testing.T, rower queryRower) int {
	t.Helper()

	var count int
	if err := rower.QueryRowContext(ctx, "SELECT COUNT(*) FROM river_job").Scan(&count); err != nil {
		t.Fatalf("count river jobs: %v", err)
	}
	return count
}
