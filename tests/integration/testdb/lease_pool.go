//go:build integration
// +build integration

package testdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"
)

// leasedPoolCap bounds how many physical leased databases a single package
// test process will ever hold open at once. Grown lazily: the pool starts
// empty and a new leased database is CREATE DATABASE ... TEMPLATE'd only when
// checkout finds no free lease AND the pool is below cap. This is per-package
// process (the pool is a package-level var), which is exactly the boundary
// go test -p=NumCPU already parallelizes across, so it does not need to be
// cross-process.
const leasedPoolCap = 8

// globalLeasePool is the package-process-wide pool of leased databases
// checked out by Open. It replaces the old one-CREATE-DATABASE/DROP-DATABASE-
// per-test pattern, which forced a cluster-wide checkpoint (DROP DATABASE)
// once per test, 16 packages wide under go test's default -p=NumCPU.
var globalLeasePool = &leasePool{}

type leasePool struct {
	mu    sync.Mutex
	cond  *sync.Cond
	free  []string
	total int
}

func (p *leasePool) init() {
	if p.cond == nil {
		p.cond = sync.NewCond(&p.mu)
	}
}

// checkout returns the name of a leased database exclusively owned by the
// caller until release is called. If a previously-returned lease is free, it
// is reused (still needs a reset by the caller — checkout does not reset). If
// none are free and the pool has not reached leasedPoolCap, a new leased
// database is physically cloned from the template. Otherwise checkout blocks
// until another test releases a lease.
func (p *leasePool) checkout(t *testing.T, baseDSN string) string {
	t.Helper()
	p.mu.Lock()
	p.init()
	for {
		if n := len(p.free); n > 0 {
			name := p.free[n-1]
			p.free = p.free[:n-1]
			p.mu.Unlock()
			return name
		}
		if p.total < leasedPoolCap {
			p.total++
			p.mu.Unlock()
			name, err := createLeasedDatabase(baseDSN)
			if err != nil {
				// Give the reserved slot back so a failed create does not
				// permanently shrink the pool's effective capacity.
				p.mu.Lock()
				p.total--
				p.mu.Unlock()
				t.Fatalf("create leased test database: %v", err)
			}
			return name
		}
		p.cond.Wait()
	}
}

// release returns a leased database to the free list. It is NEVER dropped —
// that is the entire point of the lease pool. The next checkout resets it
// (resetLeasedDatabase) before handing it back out.
//
// KNOWN GAP (reported, not silently worked around): "never dropped" is correct
// within a process, but nothing reclaims a lease at process EXIT. Lease names
// are random (randomHex) and the free list is in-memory, so a subsequent
// `go test` process cannot adopt an orphan and clones a fresh one instead.
// A full-suite run therefore leaks roughly one lease per package process and
// still pays one clone per package. Fixing both at once wants deterministic
// per-package lease names (reuse across runs, bounded set) rather than a
// sweeper; that is a design change beyond this slice's authorized scope.
func (p *leasePool) release(name string) {
	p.mu.Lock()
	p.init()
	p.free = append(p.free, name)
	p.mu.Unlock()
	p.cond.Signal()
}

func createLeasedDatabase(baseDSN string) (string, error) {
	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		return "", fmt.Errorf("open admin db: %w", err)
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	suffix, err := randomHex(5)
	if err != nil {
		return "", fmt.Errorf("generate lease suffix: %w", err)
	}
	name := "metaldocs_test_lease_" + suffix
	if _, err := adminDB.ExecContext(ctx,
		fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", quoteIdent(name), quoteIdent(templateDBName)),
	); err != nil {
		return "", fmt.Errorf("create leased database %s: %w", name, err)
	}

	// Snapshot the baseline NOW, while the clone is by definition still
	// byte-identical to the template. This is the only moment at which
	// "which rows are baseline?" is knowable without trusting a hardcoded
	// list — after the first test runs, baseline and test-created rows are
	// indistinguishable by inspection.
	if err := snapshotLeaseBaseline(ctx, baseDSN, name); err != nil {
		return "", fmt.Errorf("snapshot lease baseline on %s: %w", name, err)
	}
	return name, nil
}

// snapshotLeaseBaseline copies the FULL ROW CONTENT of the semi-static tables
// (see baselineSnapshotTables) into a private schema inside the leased
// database itself, immediately after the clone — while it is still identical
// to the template.
//
// Full content, not primary keys: a PK snapshot would restore row PRESENCE
// but not row CONTENT, so a test that UPDATEs the baseline tenant row would
// leave it mutated and the row-count assert would still pass (unchanged
// count) — a silent leak of the same class as the one the exclusion set
// caused. Restoring content is also why this is a restore and not a
// fail-loud: under clone semantics a test mutating the baseline row is LEGAL
// (it is that test's private database), so the reset's job is to make the
// lease clone-equivalent again, not to forbid the mutation.
//
// CREATE TABLE ... AS SELECT * is column-agnostic, so this needs no schema
// knowledge and self-heals when columns are added or dropped. Verified: none
// of these tables has an identity or generated column, so the restoring
// INSERT ... SELECT * needs no OVERRIDING SYSTEM VALUE.
//
// That omission is deliberate and must stay. If one of these tables later gains
// a GENERATED ALWAYS AS IDENTITY column, the restoring INSERT will fail loudly
// with 428C9 — which is the outcome we want. Pre-emptively writing OVERRIDING
// SYSTEM VALUE here would silence that signal and force the snapshot's stored
// identity values back into a column the schema declares the database owns,
// which is a correctness question that deserves a human, not a default.
func snapshotLeaseBaseline(ctx context.Context, baseDSN, dbName string) error {
	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		return fmt.Errorf("open leased database for baseline snapshot: %w", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx,
		"CREATE SCHEMA IF NOT EXISTS "+quoteIdent(leaseBaselineSchema)); err != nil {
		return fmt.Errorf("create %s schema: %w", leaseBaselineSchema, err)
	}

	for _, st := range baselineSnapshotTables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf(
			"CREATE TABLE %s AS SELECT * FROM %s", st.snapshot(), st.qualified,
		)); err != nil {
			return fmt.Errorf("snapshot baseline rows for %s: %w", st.qualified, err)
		}
	}
	return nil
}

// randomHex generates n random bytes hex-encoded, without requiring a
// *testing.T (unlike randomSuffix), since createLeasedDatabase runs outside
// any single test's lifecycle — it is shared pool infrastructure.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// leaseBaselineSchema holds the per-lease baseline snapshot. It lives OUTSIDE
// schemas 'public' and 'metaldocs' precisely so the reset's table scan (which
// filters to those two) cannot see — and therefore cannot delete — it.
const leaseBaselineSchema = "metaldocs_testkit"

// Exactly 6 tables in the template carry seed rows (verified by scanning every
// table in the template for rows): role_capabilities 112, schema_migrations
// 101, river_migration 6, tenants 1, templates_template 1,
// templates_template_version 1. EVERY other table in public + metaldocs is
// empty in the template — including the counter tables. So deleting the other
// 59 to empty is exactly what a fresh clone gives, and no per-table seed
// logic is needed for anything outside these 6.
//
// resetExclusions are the 3 STATIC ones: seeded by the curated bootstrap and
// never written by any test, so there is no reason to churn them.
//
// Contrast baselineSnapshotTables below — the distinction is not cosmetic.
// Excluding a table that tests DO write lets a leased database drift from
// clone-equivalence with every reuse: the audit package's invariant assert
// caught this at 2 tenant rows, and a lease left unguarded reached 51.
var resetExclusions = map[string]struct{}{
	"metaldocs.role_capabilities": {},
	"public.schema_migrations":    {},
	"public.river_migration":      {},
}

// snapshotTable is a SEMI-static table: the curated bootstrap seeds baseline
// rows into it, AND test helpers (testdb.NewTenant, testdb/fixtures.go) write
// to it at runtime. Neither blanket-exclude (drifts) nor plain blanket-delete
// (loses the seed) is correct. The reset deletes it like any other table and
// then restores the CREATE-time snapshot verbatim.
type snapshotTable struct {
	qualified string
	snapName  string
}

func (s snapshotTable) snapshot() string {
	return quoteIdent(leaseBaselineSchema) + "." + quoteIdent(s.snapName)
}

var baselineSnapshotTables = []snapshotTable{
	{qualified: "metaldocs.tenants", snapName: "snap_metaldocs_tenants"},
	{qualified: "public.templates_template", snapName: "snap_public_templates_template"},
	{qualified: "public.templates_template_version", snapName: "snap_public_templates_template_version"},
}

// resetLeasedDatabase restores a leased database to clone-equivalence using
// DELETE, not TRUNCATE. TRUNCATE's cost is per-RELATION and fixed (one new
// relfilenode + one synchronous DataFileImmediateSync fsync each), which on a
// virtualized-disk host (Docker Desktop / WSL2, ~0.9-1.7s per fsync) cost
// ~99-169s for this schema's 62 tables regardless of row count. DELETE reuses
// the existing relfilenode, issues no per-relation fsync, and costs
// proportional to ROWS — and a test leaves a handful. Measured: ~99,000ms
// (TRUNCATE) vs ~80ms (DELETE). Dead tuples are left for autovacuum; leases
// are short-lived and a VACUUM step would reintroduce the I/O this avoids.
//
// Three-way table classification (see resetExclusions / baselineSnapshotTables):
// static tables are skipped, semi-static tables are emptied then restored from
// their CREATE-time snapshot, everything else is emptied. Sequences are
// restarted. Together that reproduces clone state exactly.
//
// The table and sequence lists are derived at runtime from pg_tables /
// pg_sequences (self-healing against schema drift), never hardcoded.
//
// Returns an error rather than calling t.Fatalf directly: the caller (Open)
// must release the lease back to the pool BEFORE failing the test, because
// t.Fatalf triggers runtime.Goexit and any code after it — including the
// pool release — would never run, permanently leaking the lease slot.
func resetLeasedDatabase(ctx context.Context, baseDSN, dbName string) error {
	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		return fmt.Errorf("open leased database %s for reset: %w", dbName, err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping leased database %s for reset: %w", dbName, err)
	}

	targets, err := deleteTargets(ctx, db)
	if err != nil {
		return fmt.Errorf("compute delete targets for %s: %w", dbName, err)
	}
	sequences, err := leaseSequences(ctx, db)
	if err != nil {
		return fmt.Errorf("compute sequences for %s: %w", dbName, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset tx on %s: %w", dbName, err)
	}
	defer func() { _ = tx.Rollback() }()

	// session_replication_role = replica suspends ALL non-replica triggers for
	// this tx: FK triggers (so the DELETEs need no topological ordering — this
	// is what buys back the ordering-freedom TRUNCATE ... CASCADE gave us) AND
	// the DB tripwire triggers (intended: a reset is not a governed write).
	// SET LOCAL is mandatory — it reverts at commit so the setting can never
	// leak onto a pooled connection a test later uses, where suspended
	// tripwires would silently void the invariant those tests exist to prove.
	if _, err := tx.ExecContext(ctx, "SET LOCAL session_replication_role = replica"); err != nil {
		return fmt.Errorf("suspend triggers for reset on %s: %w", dbName, err)
	}

	for _, target := range targets {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+target); err != nil {
			return fmt.Errorf("delete from %s on %s: %w", target, dbName, err)
		}
	}

	// Semi-static tables were emptied by the loop above (they are ordinary
	// delete targets). Restore the CREATE-time snapshot verbatim, which puts
	// back both the rows AND their exact content.
	for _, st := range baselineSnapshotTables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(
			"INSERT INTO %s SELECT * FROM %s", st.qualified, st.snapshot(),
		)); err != nil {
			return fmt.Errorf("restore baseline rows into %s on %s: %w", st.qualified, dbName, err)
		}
	}

	// DELETE does not restart sequences the way TRUNCATE ... RESTART IDENTITY
	// does, so under lease reuse they advance monotonically and diverge from
	// clone semantics. In the template all sequences are virgin
	// (last_value IS NULL), so a clone hands every test nextval = start_value;
	// setval(seq, start_value, false) reproduces exactly that. Two carry
	// user-visible semantics — document_revisions_revision_num_seq (revision
	// numbering) and audit_events_audit_sequence_seq (the audit chain) — so
	// this is not hypothetical.
	//
	// setval, NOT `ALTER SEQUENCE ... RESTART`, and this is load-bearing:
	// RESTART rewrites the sequence's relfilenode, which forces PostgreSQL to
	// commit synchronously even though this cluster runs synchronous_commit=off.
	// That drags one WAL fsync into every reset, and on the Docker/WSL2
	// virtualized disk a single fsync costs 150-660ms — measured 169-258ms per
	// reset with RESTART vs 13-19ms with setval, i.e. RESTART alone was the
	// entire reset budget. setval is an ordinary row update: no relfilenode
	// rewrite, no forced sync, and nextval is identical (verified). The only
	// divergence is the catalog's last_value display (1 vs NULL), which nothing
	// reads; tests observe nextval, and that is exact.
	for _, seq := range sequences {
		if _, err := tx.ExecContext(ctx,
			"SELECT setval($1::regclass, $2, false)", seq.qualified, seq.startValue,
		); err != nil {
			return fmt.Errorf("reset sequence %s on %s: %w", seq.qualified, dbName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset tx on %s: %w", dbName, err)
	}

	if err := assertReferenceRowCount(ctx, db, dbName, "metaldocs.tenants", 1); err != nil {
		return err
	}
	return assertReferenceRowCount(ctx, db, dbName, "public.templates_template", 1)
}

// deleteTargets queries pg_tables for every base table in schemas public and
// metaldocs, minus resetExclusions. The baselineSnapshotTables ARE included —
// they are emptied like everything else and then restored from their
// snapshots.
//
// The `schemaname IN ('public', 'metaldocs')` restriction is LOAD-BEARING,
// not incidental: it is the only reason leaseBaselineSchema
// ("metaldocs_testkit") survives the very reset that depends on it. Widening
// this scan to all schemas would silently delete the baseline snapshots and
// turn every subsequent reset into an under-seeded false green. Do not
// "fix" this filter.
func deleteTargets(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT schemaname, tablename
		   FROM pg_tables
		  WHERE schemaname IN ('public', 'metaldocs')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []string
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return nil, err
		}
		if _, excluded := resetExclusions[schema+"."+table]; excluded {
			continue
		}
		targets = append(targets, quoteIdent(schema)+"."+quoteIdent(table))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

// advancedSequence is a sequence that a test moved off its virgin state, plus
// the start_value nextval must hand back to the next test.
type advancedSequence struct {
	qualified  string
	startValue int64
}

// leaseSequences returns the sequences in public/metaldocs that a test actually
// advanced, derived at runtime from pg_sequences for the same self-healing
// reason as the table scan.
//
// The `last_value IS NOT NULL` predicate is a correctness-preserving filter, not
// an optimization guess: in the template every sequence is virgin
// (last_value IS NULL), and a sequence nobody called is already in exactly the
// state a fresh clone would hand over. Skipping it keeps its catalog state
// byte-identical to a clone; resetting it could only move it away from that.
// The predicate is safe under reuse too: after the first reset a sequence reads
// back last_value = start_value with is_called = false, so it re-enters this
// list and is reset again. The filter can only ever skip a genuinely virgin
// sequence.
//
// start_value is read per sequence and never assumed to be 1. It is scanned
// into a NOT NULL Go int64 with no COALESCE default on purpose: pg_sequences
// declares start_value NOT NULL, so a NULL here would mean the catalog is not
// what this code believes it is, and the Scan fails loudly. Defaulting to 1
// would paper over that with a guess — and a wrong start_value is invisible
// until some test asserts on a specific revision_num or audit_sequence, at
// which point it reads as a product bug rather than a factory bug.
func leaseSequences(ctx context.Context, db *sql.DB) ([]advancedSequence, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT schemaname, sequencename, start_value
		   FROM pg_sequences
		  WHERE schemaname IN ('public', 'metaldocs')
		    AND last_value IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seqs []advancedSequence
	for rows.Next() {
		var schema, name string
		var start int64
		if err := rows.Scan(&schema, &name, &start); err != nil {
			return nil, err
		}
		seqs = append(seqs, advancedSequence{
			qualified:  quoteIdent(schema) + "." + quoteIdent(name),
			startValue: start,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return seqs, nil
}

func assertReferenceRowCount(ctx context.Context, db *sql.DB, dbName, qualifiedTable string, want int) error {
	var got int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM "+qualifiedTable).Scan(&got); err != nil {
		return fmt.Errorf("count %s in leased database %s: %w", qualifiedTable, dbName, err)
	}
	if got != want {
		return fmt.Errorf("leased database %s: post-reset invariant violated: %s has %d rows, want %d — reset left the database under-seeded (false green)", dbName, qualifiedTable, got, want)
	}
	return nil
}
