//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// TestTemplateSweepBoundsOrphans proves the content-addressed template sweep
// removes orphan templates from previous schema fingerprints while preserving
// the current one, so metaldocs_test_template_* databases cannot accumulate
// without bound. This is the regression guard for the M3 close-out leak: under
// the old per-pid scheme the template was never dropped at process exit, so a
// single dev machine reached 144 orphan templates after ~4h.
//
// It drives sweepStaleTemplates directly (not via Open) so the assertion is
// deterministic and independent of the per-process sync.Once that gates the
// real build.
func TestTemplateSweepBoundsOrphans(t *testing.T) {
	baseDSN := DSN(t)
	admin, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer admin.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}

	conn, err := admin.Conn(ctx)
	if err != nil {
		t.Fatalf("pin admin conn: %v", err)
	}
	defer conn.Close()

	// keep is the real current template name, so the sweep here reuses it
	// rather than nuking a template later tests clone from. orphan stands in for
	// a template left over from a previous schema fingerprint.
	fingerprint, err := schemaFingerprint()
	if err != nil {
		t.Fatalf("schema fingerprint: %v", err)
	}
	keep := fmt.Sprintf("metaldocs_test_template_%s", fingerprint[:16])
	orphan := "metaldocs_test_template_" + randomSuffix(t) + randomSuffix(t) // 20 hex, distinct from keep

	// Ensure keep exists (empty is fine — a later Open rebuilds it via the
	// readiness marker check) and create the orphan.
	for _, n := range []string{keep, orphan} {
		if !dbExists(ctx, conn, n) {
			if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(n))); err != nil {
				t.Fatalf("create %s: %v", n, err)
			}
		}
	}
	t.Cleanup(func() {
		c, err := admin.Conn(context.Background())
		if err != nil {
			return
		}
		defer c.Close()
		// Drop only the orphan; keep is the shared real template.
		_, _ = c.ExecContext(context.Background(),
			fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(orphan)))
	})

	if err := sweepStaleTemplates(ctx, conn, keep); err != nil {
		t.Fatalf("sweepStaleTemplates: %v", err)
	}

	if !dbExists(ctx, conn, keep) {
		t.Errorf("sweep dropped the current template %s — it must be kept", keep)
	}
	if dbExists(ctx, conn, orphan) {
		t.Errorf("sweep left orphan template %s — it must be dropped", orphan)
	}
}

func dbExists(ctx context.Context, conn *sql.Conn, name string) bool {
	var n int
	if err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM pg_database WHERE datname = $1", name).Scan(&n); err != nil {
		return false
	}
	return n > 0
}
