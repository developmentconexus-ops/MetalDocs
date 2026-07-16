//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

// These two tests are the falsifiable guards demanded by the GPT final-round
// review of unit 4.6's leased-DB factory. Each pins a specific laundering hole
// the review identified and each FAILS against the pre-fix code:
//
//   - TestRetirementRequiresOwnershipProof (finding 2): retireStaleTemplates
//     used to overwrite ANY prefix-matching database's COMMENT with a retired
//     marker, ignoring the marker it had just read. That manufactured the very
//     recognised marker GCRetiredDatabases trusts, laundering an unmarked or
//     foreign database this factory never built into a droppable one — one hop
//     upstream of "never drop a database you cannot prove you own".
//   - TestNonVirginBaselineGuardIsNotBypassable (finding 1): the virgin-sequence
//     guard in snapshotLeaseBaseline used to run AFTER the baseline snapshot
//     tables were written. A guard failure left those tables behind;
//     leaseBaselineIsComplete keys off exactly them, so the next adopter saw a
//     "complete" baseline, skipped the guard, and adopted a template whose
//     sequences reset would slam to the wrong nextval.

// TestRetirementRequiresOwnershipProof asserts that retireStaleTemplates retires
// ONLY a database carrying a recognised ready marker whose fingerprint matches
// the database name, and that an unmarked or foreign-marked prefix match
// survives both retirement AND a subsequent GC pass untouched. (GPT finding 2.)
func TestRetirementRequiresOwnershipProof(t *testing.T) {
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

	// keep is the real current template, so this sweep never nukes a template
	// later tests clone from — same discipline as
	// TestTemplateRetirementRetiresStaleAndKeepsCurrent.
	fingerprint, err := schemaFingerprint()
	if err != nil {
		t.Fatalf("schema fingerprint: %v", err)
	}
	keep := templateNamePrefix + fingerprint[:16]

	// Two prefix-matching databases this factory did NOT build:
	//   unmarked  — no COMMENT at all.
	//   foreign   — a well-formed ready marker for a DIFFERENT fingerprint (its
	//               first char is forced to differ from the name suffix, so the
	//               marker fingerprint's first 16 hex cannot match the name).
	unmarked := templateNamePrefix + randomSuffix(t) + randomSuffix(t)
	foreign := templateNamePrefix + randomSuffix(t) + randomSuffix(t)
	foreignSuffix := foreign[len(templateNamePrefix):]
	firstChar := "0"
	if foreignSuffix[0] == '0' {
		firstChar = "1"
	}
	foreignFP := firstChar + foreignSuffix[1:] // differs from foreignSuffix at index 0
	foreignMarker := readyMarker(foreignFP)

	for _, n := range []string{unmarked, foreign} {
		if !dbExists(ctx, t, conn, n) {
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
		for _, n := range []string{unmarked, foreign} {
			_, _ = c.ExecContext(context.Background(),
				fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(n)))
		}
	})
	if _, err := conn.ExecContext(ctx, fmt.Sprintf(
		"COMMENT ON DATABASE %s IS %s", quoteIdent(foreign), quoteLiteral(foreignMarker),
	)); err != nil {
		t.Fatalf("mark foreign db: %v", err)
	}

	if err := retireStaleTemplates(ctx, conn, keep); err != nil {
		t.Fatalf("retireStaleTemplates: %v", err)
	}

	// Neither database's marker may have been rewritten. A retired marker here is
	// the laundering the fix forbids.
	if m := readMarker(ctx, t, conn, unmarked); m.Valid {
		t.Errorf("retirement stamped a marker onto UNMARKED %s (%q) — an unowned database was laundered toward droppable", unmarked, m.String)
	}
	if m := readMarker(ctx, t, conn, foreign); !m.Valid || m.String != foreignMarker {
		t.Errorf("retirement rewrote FOREIGN %s marker to %q, want it left as %q — a database we did not build is not ours to retire", foreign, markerOrNull(m), foreignMarker)
	}

	// And a real GC pass must leave both physically standing: neither carries a
	// recognised retired/lease marker, so both are skipped, never dropped.
	if _, err := GCRetiredDatabases(t); err != nil {
		t.Fatalf("GCRetiredDatabases: %v", err)
	}
	if !dbExists(ctx, t, conn, unmarked) {
		t.Errorf("GC dropped UNMARKED %s — a database with no recognised marker must never be dropped", unmarked)
	}
	if !dbExists(ctx, t, conn, foreign) {
		t.Errorf("GC dropped FOREIGN %s — a foreign-ready-marked database is not ours to drop", foreign)
	}
}

// TestNonVirginBaselineGuardIsNotBypassable asserts that the virgin-sequence
// guard in snapshotLeaseBaseline fails on a non-virgin template AND leaves no
// baseline artifacts behind, so a second consecutive claim re-runs the guard and
// fails identically. The middle assertion — leaseBaselineIsComplete is false
// after the first failure — is the falsifiable core: under the pre-fix ordering
// the snapshot tables were written before the guard ran, so that call returned
// true and the second adopter would have skipped the guard entirely. (GPT
// finding 1.)
func TestNonVirginBaselineGuardIsNotBypassable(t *testing.T) {
	baseDSN := DSN(t)
	admin, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer admin.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}

	// A non-template name so cluster sweeps (retire/GC) ignore it entirely.
	nv := "metaldocs_test_nonvirgin_" + randomSuffix(t) + randomSuffix(t)
	if _, err := admin.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(nv))); err != nil {
		t.Fatalf("create %s: %v", nv, err)
	}
	t.Cleanup(func() {
		_, _ = admin.ExecContext(context.Background(),
			fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(nv)))
	})

	// Make one public sequence NON-VIRGIN: nextval advances last_value from NULL,
	// which is exactly the state the guard exists to reject.
	seedDB, err := openDBWithDatabase(baseDSN, nv)
	if err != nil {
		t.Fatalf("open %s: %v", nv, err)
	}
	if _, err := seedDB.ExecContext(ctx, "CREATE SEQUENCE public.nonvirgin_probe"); err != nil {
		seedDB.Close()
		t.Fatalf("create sequence: %v", err)
	}
	if _, err := seedDB.ExecContext(ctx, "SELECT nextval('public.nonvirgin_probe')"); err != nil {
		seedDB.Close()
		t.Fatalf("advance sequence: %v", err)
	}
	seedDB.Close()

	// First claim: guard must fire, loudly.
	err1 := snapshotLeaseBaseline(ctx, baseDSN, nv)
	if err1 == nil {
		t.Fatalf("first snapshotLeaseBaseline on non-virgin template returned nil — the virgin guard did not fire")
	}
	if !strings.Contains(err1.Error(), "NON-VIRGIN") {
		t.Fatalf("first failure was not the virgin guard: %v", err1)
	}

	// The load-bearing assertion: a failed guard leaves NO usable baseline. If
	// this reads complete, the guard ran after the writes and the next adopter
	// would skip it — the exact bypass finding 1 forbids.
	complete, err := leaseBaselineIsComplete(ctx, baseDSN, nv)
	if err != nil {
		t.Fatalf("leaseBaselineIsComplete: %v", err)
	}
	if complete {
		t.Fatalf("baseline reads COMPLETE after a failed virgin guard — guard-failure debris was left behind and the guard is bypassable across runs")
	}

	// Second consecutive claim: must fail identically. No debris from the first
	// call may have laundered the template into an adoptable state.
	err2 := snapshotLeaseBaseline(ctx, baseDSN, nv)
	if err2 == nil {
		t.Fatalf("second snapshotLeaseBaseline returned nil — first call laundered the non-virgin template into adoptable")
	}
	if !strings.Contains(err2.Error(), "NON-VIRGIN") {
		t.Fatalf("second failure was not the virgin guard: %v", err2)
	}
}

func readMarker(ctx context.Context, t *testing.T, conn *sql.Conn, name string) sql.NullString {
	t.Helper()
	var m sql.NullString
	if err := conn.QueryRowContext(ctx,
		`SELECT shobj_description(oid, 'pg_database') FROM pg_database WHERE datname = $1`,
		name).Scan(&m); err != nil {
		if err == sql.ErrNoRows {
			return sql.NullString{}
		}
		t.Fatalf("read marker for %s: %v", name, err)
	}
	return m
}

func markerOrNull(m sql.NullString) string {
	if !m.Valid {
		return "<null>"
	}
	return m.String
}
