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

// TestTemplateRetirementRetiresStaleAndKeepsCurrent proves the content-addressed
// template sweep neutralises orphan templates from previous schema fingerprints
// while preserving the current one. This is the regression guard for the M3
// close-out leak: under the old per-pid scheme the template was never retired at
// process exit, so a single dev machine reached 144 orphan templates after ~4h.
//
// This test used to assert the orphan was physically DROPPED. That assertion is
// deliberately gone, and the distinction is the entire point of the change it
// guards:
//
//   - Retirement is now LOGICAL. The sweep rewrites the orphan's COMMENT marker
//     so templateIsReady can never mistake it for a live template. It does not
//     drop it. Dropping here forced a checkpoint while the sweep held the global
//     template-build advisory lock, which wedged the whole cluster for >5 minutes
//     on IPC:CheckpointStart and blocked every other test process.
//   - Physical space reclamation moved to GCRetiredDatabases, which is idle-only,
//     never automatic, and never holds that lock.
//
// So the correctness invariant (a stale template can never be cloned from) is
// asserted here, and is what this test owns. The disk-space invariant is NOT
// asserted here and is no longer automatic: it is GC's job, on purpose. A test
// that called GC to prove boundedness would itself have to drop every retired
// database on the cluster — serialized, one forced checkpoint each — which is
// precisely the hot-path stall this design exists to prevent.
//
// It drives retireStaleTemplates directly (not via Open) so the assertion is
// deterministic and independent of the per-process sync.Once that gates the
// real build.
func TestTemplateRetirementRetiresStaleAndKeepsCurrent(t *testing.T) {
	baseDSN := DSN(t)
	admin, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer admin.Close()

	// 60s was exceeded under full-suite parallel load (CREATE DATABASE + checkpoint
	// contention); this test's ops are serial and cheap when idle, the budget is
	// generous headroom, not expected runtime.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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
	keep := templateNamePrefix + fingerprint[:16]
	orphan := templateNamePrefix + randomSuffix(t) + randomSuffix(t) // 20 hex, distinct from keep

	// Ensure keep exists (empty is fine — a later Open rebuilds it via the
	// readiness marker check) and create the orphan.
	for _, n := range []string{keep, orphan} {
		if !dbExists(ctx, t, conn, n) {
			if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(n))); err != nil {
				t.Fatalf("create %s: %v", n, err)
			}
		}
	}
	// Give the orphan a ready marker for a fingerprint that is not the current
	// one. Without this the test would prove nothing: an unmarked database is
	// already not-ready, so the retirement assertion below would pass whether or
	// not retireStaleTemplates did anything at all.
	orphanFP16 := orphan[len(templateNamePrefix):]
	if _, err := conn.ExecContext(ctx, fmt.Sprintf(
		"COMMENT ON DATABASE %s IS %s", quoteIdent(orphan), quoteLiteral(readyMarker(orphanFP16)),
	)); err != nil {
		t.Fatalf("mark orphan ready: %v", err)
	}
	ready, err := templateIsReady(ctx, conn, orphan, orphanFP16)
	if err != nil {
		t.Fatalf("read orphan readiness before sweep: %v", err)
	}
	if !ready {
		t.Fatalf("precondition failed: orphan %s does not read as ready before the sweep — the retirement assertion below would be vacuous", orphan)
	}

	t.Cleanup(func() {
		c, err := admin.Conn(context.Background())
		if err != nil {
			return
		}
		defer c.Close()
		// Drop only the orphan this test created; keep is the shared real
		// template. This is the test cleaning up a database it provably owns,
		// not the sweep's job — see the doc comment.
		_, _ = c.ExecContext(context.Background(),
			fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(orphan)))
	})

	if err := retireStaleTemplates(ctx, conn, keep); err != nil {
		t.Fatalf("retireStaleTemplates: %v", err)
	}

	if !dbExists(ctx, t, conn, keep) {
		t.Errorf("retirement dropped the current template %s — it must be kept", keep)
	}
	stillReady, err := templateIsReady(ctx, conn, orphan, orphanFP16)
	if err != nil {
		t.Fatalf("read orphan readiness after retirement: %v", err)
	}
	if stillReady {
		t.Errorf("orphan template %s still reads as ready after retirement — a stale template must never be clonable", orphan)
	}
	if !dbExists(ctx, t, conn, orphan) {
		t.Errorf("retirement physically dropped orphan %s — retirement must be logical only; "+
			"dropping under the global template lock is what wedged the cluster", orphan)
	}
}

// dbExists reports whether name exists in pg_database. A query failure (e.g.
// an expired context under load) must fail LOUD — returning false here once
// misattributed an infra timeout as "GC dropped an unowned database".
func dbExists(ctx context.Context, t *testing.T, conn *sql.Conn, name string) bool {
	var n int
	if err := conn.QueryRowContext(ctx,
		"SELECT count(*) FROM pg_database WHERE datname = $1", name).Scan(&n); err != nil {
		t.Fatalf("dbExists(%s): %v", name, err)
	}
	return n > 0
}
