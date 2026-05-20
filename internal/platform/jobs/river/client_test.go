package riverjobs

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"

	"metaldocs/internal/testsupport/pgtest"
)

type compatibilityArgs struct {
	DocumentID string `json:"document_id"`
}

func (compatibilityArgs) Kind() string { return "scheduled_publish_compatibility_probe" }

func TestRiverInsertTxCompatibilityBoundary(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := pgtest.OpenAndMigrate(t)
	schema := fmt.Sprintf("river_task1_%d", time.Now().UTC().UnixNano())

	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
	})

	migrator, err := rivermigrate.New(riverdatabasesql.New(db), &rivermigrate.Config{Schema: schema})
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		t.Fatalf("migrate river schema up: %v", err)
	}

	bundle, err := NewClientBundle(db, Config{
		Schema:              schema,
		SkipUnknownJobCheck: true,
	}, nil)
	if err != nil {
		t.Fatalf("new client bundle: %v", err)
	}

	clientType := reflect.TypeOf(bundle.Client).String()
	if !strings.Contains(clientType, "*sql.Tx") {
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

	if got := countRiverJobs(ctx, t, db, schema); got != 0 {
		t.Fatalf("visible jobs outside tx before commit = %d, want 0", got)
	}
	if got := countRiverJobs(ctx, t, tx, schema); got != 1 {
		t.Fatalf("visible jobs inside tx before commit = %d, want 1", got)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	if got := countRiverJobs(ctx, t, db, schema); got != 1 {
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

func countRiverJobs(ctx context.Context, t *testing.T, rower queryRower, schema string) int {
	t.Helper()

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.river_job", schema)
	if err := rower.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatalf("count river jobs: %v", err)
	}
	return count
}
