//go:build integration
// +build integration

package testdb

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"
	"time"
)

// resetPolicyVersion identifies the SEMANTICS of resetLeasedDatabase: which
// tables are delete targets, which are excluded, which are snapshotted, and
// how sequences are handled. Bump it by hand whenever any of that changes.
//
// It is embedded in every lease slot's name (see leaseSlotName) precisely so
// a process running a NEW reset policy can never adopt a database that was
// last reset under an OLD policy. Adopting a database whose reset semantics
// you don't know is exactly how a false green happens: the adopting process
// would trust resetLeasedDatabase to have restored clone-equivalence, but the
// database on disk might have been left in a shape only the old policy knew
// how to fully clean up.
const resetPolicyVersion = 1

// leaseSlotCount is the number of deterministic, cross-process-adoptable lease
// slots per (fingerprint, resetPolicyVersion) pair. `go test ./...` defaults
// -p to NumCPU, so up to that many package processes run concurrently and
// each needs at least one lease; this is fixed at 16 to match a 16-core
// default rather than derived at runtime, so the slot-name set (and therefore
// what a future process can adopt) is stable across machines. Fewer slots
// than concurrent processes would make processes block on each other
// indefinitely instead of merely queuing briefly.
const leaseSlotCount = 16

func init() {
	if leasedPoolCap > leaseSlotCount {
		panic("testdb: leasedPoolCap must be <= leaseSlotCount")
	}
}

// leasedPoolCap bounds how many physical leased databases a single package
// test process will ever hold open at once. Grown lazily: the pool starts
// empty and a new leased database is claimed (adopted or cloned) only when
// checkout finds no free lease AND the pool is below cap. This is per-package
// process (the pool is a package-level var), which is exactly the boundary
// go test -p=NumCPU already parallelizes across, so it does not need to be
// cross-process by itself — cross-process coordination is what the slot
// claim (claimSlot) provides.
const leasedPoolCap = 8

// leaseCloneGateKey is the single cluster-wide advisory-lock key that
// serializes the checkpoint-forcing CREATE DATABASE clone across ALL processes
// and goroutines (see withLeaseCloneGate). Namespaced + sha256'd, distinct by
// construction from the template BUILD key (advisoryKey, db.go — raw fingerprint
// bytes) and the per-slot claim key (leaseSlotLockKey — "…-lease-slot:" prefix),
// so the three advisory-lock spaces cannot collide by chance.
func leaseCloneGateKey() int64 {
	sum := sha256.Sum256([]byte("metaldocs-testdb-lease-clone-gate"))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// withLeaseCloneGate runs fn while holding leaseCloneGateKey on a dedicated
// connection, so at most one CREATE DATABASE clone runs anywhere on the cluster
// at a time.
//
// This SUPERSEDES the old process-local leaseCreateMu (a sync.Mutex).
// `go test ./...` runs each package as its own OS process, so a process-local
// mutex serialized clones only WITHIN a package: on a COLD cluster (no slots
// yet) up to leaseSlotCount package processes each win a DISTINCT slot's
// advisory lock and issue a concurrent clone — reproducing the exact 8-wide
// clone storm the mutex was meant to prevent (measured 60s CREATE DATABASE
// timeouts on this virtualized-storage host). The per-slot lock (claimSlot)
// prevents two processes cloning the SAME slot; it does nothing about different
// slots cloned at once. A Postgres advisory lock is held per SESSION, so one
// dedicated conn here serializes clones across BOTH goroutines (t.Parallel) and
// processes — the coverage leaseCreateMu could never provide. Adoption (the
// warm path, no clone) never takes this gate, so steady-state contention is
// nil; the gate only bites the once-per-cluster cold clone burst.
func withLeaseCloneGate(ctx context.Context, adminDB *sql.DB, fn func(conn *sql.Conn) error) error {
	conn, err := adminDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("pin conn for clone gate: %w", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", leaseCloneGateKey()); err != nil {
		return fmt.Errorf("acquire clone gate: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", leaseCloneGateKey())
	}()
	return fn(conn)
}

// globalLeasePool is the package-process-wide pool of leased databases
// checked out by Open. It replaces the old one-CREATE-DATABASE/DROP-DATABASE-
// per-test pattern, which forced a cluster-wide checkpoint (DROP DATABASE)
// once per test, 16 packages wide under go test's default -p=NumCPU.
var globalLeasePool = &leasePool{}

// leaseClaimTimeout bounds how long checkout will wait for a cross-process
// slot to free up before failing loudly. It is a hang-guard against a
// genuinely saturated cluster (all leaseSlotCount slots held by other live
// processes), not a normal-path budget — claiming an already-adopted slot is
// a handful of catalog queries, not a clone.
const leaseClaimTimeout = 60 * time.Second

type leasePool struct {
	mu   sync.Mutex
	cond *sync.Cond
	free []string
	// held tracks every slot this process has successfully claimed (via
	// claimSlot's per-slot advisory lock) across the process lifetime,
	// keyed by slot name. The dedicated *sql.Conn holding each lock is kept
	// open here — closing it would release the lock — so it lives exactly as
	// long as this process does. Process death (including a crash) closes
	// the connection and releases the lock automatically; that IS the
	// release mechanism for cross-process adoption, deliberately not
	// replicated via a signal handler or atexit.
	held map[string]*sql.Conn
	// adminDB is the shared connection used to issue the per-slot
	// pg_try_advisory_lock probes and adopt-or-create DDL. Kept open for the
	// process lifetime alongside held.
	adminDB *sql.DB
}

func (p *leasePool) init() {
	if p.cond == nil {
		p.cond = sync.NewCond(&p.mu)
	}
	if p.held == nil {
		p.held = make(map[string]*sql.Conn)
	}
}

// checkout returns the name of a leased database exclusively owned by the
// caller until release is called. If a previously-returned lease is free, it
// is reused (still needs a reset by the caller — checkout does not reset). If
// none are free and this process holds fewer than leasedPoolCap slots, a new
// slot is claimed cross-process (claimSlot) and adopted-or-created. Otherwise
// checkout blocks until another test in THIS process releases a lease —
// cross-process contention is handled inside claimSlot itself.
func (p *leasePool) checkout(t *testing.T, baseDSN, templateDBName, fingerprint16 string) string {
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
		if len(p.held) < leasedPoolCap {
			p.mu.Unlock()
			name, err := p.claimSlot(baseDSN, templateDBName, fingerprint16)
			if err != nil {
				t.Fatalf("claim leased test database slot: %v", err)
			}
			return name
		}
		p.cond.Wait()
	}
}

// release returns a leased database to the free list. It is NEVER dropped and
// its cross-process claim (the held advisory lock) is NEVER released here —
// that is the entire point of both the lease pool and slot adoption. The next
// checkout resets it (resetLeasedDatabase) before handing it back out.
func (p *leasePool) release(name string) {
	p.mu.Lock()
	p.init()
	p.free = append(p.free, name)
	p.mu.Unlock()
	p.cond.Signal()
}

// claimSlot finds a deterministic lease slot (see leaseSlotName) not already
// claimed by another process, proven by this process winning that slot's
// per-slot pg_try_advisory_lock, then ensures the slot's database exists
// (adopt-or-create) before returning its name. The winning connection is kept
// open in p.held for the rest of the process lifetime — see the held field's
// doc comment for why.
func (p *leasePool) claimSlot(baseDSN, templateDBName, fingerprint16 string) (string, error) {
	p.mu.Lock()
	if p.adminDB == nil {
		db, err := openDBWithDatabase(baseDSN, "postgres")
		if err != nil {
			p.mu.Unlock()
			return "", fmt.Errorf("open admin db for slot claim: %w", err)
		}
		p.adminDB = db
	}
	adminDB := p.adminDB
	p.mu.Unlock()

	deadline := time.Now().Add(leaseClaimTimeout)
	for {
		for nn := 0; nn < leaseSlotCount; nn++ {
			slotName := leaseSlotName(fingerprint16, nn)

			p.mu.Lock()
			_, alreadyHeld := p.held[slotName]
			p.mu.Unlock()
			if alreadyHeld {
				continue
			}

			claimCtx, claimCancel := context.WithTimeout(context.Background(), 15*time.Second)
			conn, err := adminDB.Conn(claimCtx)
			if err != nil {
				claimCancel()
				return "", fmt.Errorf("pin conn to probe lease slot %s: %w", slotName, err)
			}

			var locked bool
			err = conn.QueryRowContext(claimCtx, "SELECT pg_try_advisory_lock($1)", leaseSlotLockKey(slotName)).Scan(&locked)
			claimCancel()
			if err != nil {
				_ = conn.Close()
				return "", fmt.Errorf("probe advisory lock for lease slot %s: %w", slotName, err)
			}
			if !locked {
				// Another process holds this slot right now. Try the next one.
				_ = conn.Close()
				continue
			}

			// Won the slot. Record it as held BEFORE any fallible work below,
			// so a failure here does not leave this process's bookkeeping
			// unaware that it is (correctly, at the Postgres level) still
			// holding the lock — a stray retry-the-same-slot loop would
			// otherwise never make progress.
			p.mu.Lock()
			p.held[slotName] = conn
			p.mu.Unlock()

			// Concurrent CREATE DATABASE is the measured failure mode, not a
			// theoretical one: with t.Parallel and -parallel=8, eight goroutines
			// reached this line at once, issued eight concurrent clones, and every
			// one blew the 60s timeout below (60.02s, repeatedly) on this
			// virtualized-storage host. Cloning is I/O the host serializes anyway —
			// doing it 8-wide converts a ~0.9s operation into a timeout for all 8.
			//
			// The serialization now lives INSIDE ensureLeaseSlotDatabase, around
			// only the clone DDL, as the cluster-wide advisory clone gate
			// (withLeaseCloneGate) — deliberately NOT a process-local mutex here,
			// because `go test ./...` runs packages as separate processes and a
			// mutex would leave cross-process cold-start clones unserialized (see
			// withLeaseCloneGate's doc). Adoption — the steady state, where the
			// first run cloned N slots once and every later process pays ~0.02s to
			// ADOPT them — never touches the gate.
			ensureCtx, ensureCancel := context.WithTimeout(context.Background(), 60*time.Second)
			ensureErr := ensureLeaseSlotDatabase(ensureCtx, baseDSN, templateDBName, fingerprint16, slotName)
			ensureCancel()
			if ensureErr != nil {
				return "", fmt.Errorf("prepare lease slot database %s: %w", slotName, ensureErr)
			}

			return slotName, nil
		}

		if time.Now().After(deadline) {
			return "", fmt.Errorf("all %d lease slots are held by other test processes", leaseSlotCount)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// leaseSlotName is the deterministic slot name two processes running the same
// code (same fingerprint, same resetPolicyVersion) will compute identically —
// which is the entire point: it lets a fresh process ADOPT an existing slot
// database instead of cloning a new one.
func leaseSlotName(fingerprint16 string, nn int) string {
	return fmt.Sprintf("metaldocs_test_lease_%s_%d_%02d", fingerprint16, resetPolicyVersion, nn)
}

// leaseSlotLockKey derives a per-slot advisory lock key from the slot's full
// name. It is namespaced ("metaldocs-testdb-lease-slot:" prefix) and hashed
// with sha256 — deliberately distinct in both prefix and algorithm from
// advisoryKey (db.go), which derives the template BUILD lock's key by
// reading raw fingerprint bytes — so the two lock spaces cannot collide by
// construction, not by chance.
func leaseSlotLockKey(slotName string) int64 {
	sum := sha256.Sum256([]byte("metaldocs-testdb-lease-slot:" + slotName))
	return int64(binary.BigEndian.Uint64(sum[:8]))
}

// leaseMarker is the COMMENT ON DATABASE value written on every lease slot
// database, proving ownership (fingerprint + reset policy) for both adoption
// and GCRetiredDatabases. Format: "metaldocs-testdb-lease:<fp16>:<policy>".
func leaseMarker(fingerprint16 string) string {
	return fmt.Sprintf("metaldocs-testdb-lease:%s:%d", fingerprint16, resetPolicyVersion)
}

// ensureLeaseSlotDatabase implements adopt-or-create for a claimed lease
// slot:
//
//   - If the slot's database does not exist, this is the (only) clone: CREATE
//     DATABASE ... TEMPLATE, take the baseline snapshot while the clone is
//     still byte-identical to the template, and mark ownership.
//   - If it already exists, ADOPT it: no drop, no recreate. The per-checkout
//     reset (resetLeasedDatabase) is what makes an adopted slot clean; this
//     function's only adoption-path responsibility is making sure the
//     baseline snapshot schema that reset depends on is actually present —
//     if a prior run's clone step never completed it, this heals it exactly
//     as the create path would.
func ensureLeaseSlotDatabase(ctx context.Context, baseDSN, templateDBName, fingerprint16, slotName string) error {
	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		return fmt.Errorf("open admin db: %w", err)
	}
	defer adminDB.Close()

	exists, err := databaseExists(ctx, adminDB, slotName)
	if err != nil {
		return fmt.Errorf("check lease slot database existence: %w", err)
	}

	needRebuild := false
	if exists {
		// Adoption is only legitimate if the slot carries a baseline this code
		// can trust. If it does, adopt it: no drop, no recreate — the
		// per-checkout reset is what makes it clean, and skipping the clone is
		// the entire point of deterministic slot names.
		complete, err := leaseBaselineIsComplete(ctx, baseDSN, slotName)
		if err != nil {
			return fmt.Errorf("check baseline snapshot on adopted slot %s: %w", slotName, err)
		}
		if !complete {
			// The slot exists but its baseline is missing or half-written: a
			// prior owner died between CREATE DATABASE and the snapshot, or the
			// slot predates the snapshot convention.
			//
			// It CANNOT be healed by snapshotting it now. This database may
			// already have been used, and snapshotting it would enshrine
			// whatever rows a previous test left behind as the "pristine"
			// baseline that every future reset restores to — a permanent,
			// silent false green. Nothing here can distinguish a pristine clone
			// from a dirty one, and guessing is exactly the assumption this
			// factory must not make.
			//
			// So rebuild it from the template (DROP + CREATE below, both under
			// the clone gate). This is the only DROP DATABASE on the lease path
			// and it is deliberately bounded: it takes a crash to reach, it runs
			// under this slot's exclusive advisory-lock claim (no live process
			// can be using it) and NOT under the global template-build lock
			// (dropping under that lock is what wedged the cluster). Failing
			// loudly instead would be safe but would brick this slot for every
			// future run until a human intervened.
			needRebuild = true
		}
	}

	if !exists || needRebuild {
		// The DROP (rebuild only) and CREATE are the checkpoint-forcing clone
		// I/O — serialize them across every process and goroutine via the
		// cluster-wide clone gate, on the gate's own pinned connection.
		//
		// STRATEGY = WAL_LOG pinned for the same reason as db.go's isolated-clone
		// site: FILE_COPY forces a checkpoint per clone, which on this storage
		// stack blocks on IPC:CheckpointDone for an unbounded time (measured
		// 36.4s worst-case vs WAL_LOG 2.6s worst-case against the 12 MB
		// template). Lease slots are created once and reused, so this fires far
		// less often than the isolated site — but a forced-checkpoint stall here
		// convoys every test waiting on the pool, so the strategy still matters.
		if err := withLeaseCloneGate(ctx, adminDB, func(conn *sql.Conn) error {
			if needRebuild {
				if _, err := conn.ExecContext(ctx,
					fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(slotName)),
				); err != nil {
					return fmt.Errorf("rebuild lease slot %s (drop incomplete slot): %w", slotName, err)
				}
			}
			if _, err := conn.ExecContext(ctx,
				fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s STRATEGY = WAL_LOG", quoteIdent(slotName), quoteIdent(templateDBName)),
			); err != nil {
				return fmt.Errorf("create leased database %s: %w", slotName, err)
			}
			return nil
		}); err != nil {
			return err
		}
		// Snapshot the baseline NOW, while the clone is by definition still
		// byte-identical to the template — see snapshotLeaseBaseline's own
		// doc comment for why full row content, not just presence, matters.
		// This is the ONLY path that snapshots, which is precisely what lets
		// snapshotLeaseBaseline assume a pristine database and fail loudly if
		// that assumption is ever violated.
		if err := snapshotLeaseBaseline(ctx, baseDSN, slotName); err != nil {
			return fmt.Errorf("snapshot lease baseline on %s: %w", slotName, err)
		}
	}

	// The database MUST be ownership-marked so GCRetiredDatabases can prove
	// it owns it before ever touching it. Reasserted unconditionally (not
	// only on create) so an adopted slot's marker always reflects the
	// fingerprint/resetPolicyVersion of the code that most recently verified
	// it — which is what GC's staleness comparison depends on being honest.
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf(
		"COMMENT ON DATABASE %s IS %s", quoteIdent(slotName), quoteLiteral(leaseMarker(fingerprint16)),
	)); err != nil {
		return fmt.Errorf("mark leased database %s: %w", slotName, err)
	}
	return nil
}

func databaseExists(ctx context.Context, adminDB *sql.DB, name string) (bool, error) {
	var exists bool
	err := adminDB.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists)
	return exists, err
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

	// Guard the reset's load-bearing assumption BEFORE writing any baseline
	// state — at the ONE moment it is verifiable, on the pristine clone, before
	// any test has touched it. resetLeasedDatabase restores every ADVANCED
	// sequence (last_value IS NOT NULL) to setval(start_value, false), which
	// reproduces clone semantics ONLY if every template sequence is virgin
	// (last_value IS NULL) at clone time — i.e. start_value IS the nextval a
	// fresh clone hands out. That is true today (verified: all sequences virgin
	// in the template), but it is an assumption, not an invariant the schema
	// enforces. If a future migration seeds a sequence, its clone-fresh nextval
	// would be start_value+N while reset would slam it back to start_value — a
	// silent wrong-nextval that surfaces only when some test asserts a specific
	// revision_num or audit_sequence, reading as a product bug. Fail LOUD here so
	// a human decides birth-capture.
	//
	// The probe runs FIRST, before CREATE SCHEMA and the snapshot tables, so a
	// guard failure leaves NO baseline artifacts behind. leaseBaselineIsComplete
	// keys off exactly those tables, so a failed guard means the next adopter
	// sees an incomplete baseline, re-runs this function, and hits the same guard
	// again — the check is non-bypassable across runs, never laundered by debris
	// a later run mistakes for a finished baseline. (GPT final-round finding 1.)
	var nonVirgin []string
	seqRows, err := db.QueryContext(ctx,
		`SELECT schemaname||'.'||sequencename
		   FROM pg_sequences
		  WHERE schemaname IN ('public', 'metaldocs')
		    AND last_value IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("probe template sequence virginity on %s: %w", dbName, err)
	}
	for seqRows.Next() {
		var name string
		if err := seqRows.Scan(&name); err != nil {
			seqRows.Close()
			return fmt.Errorf("scan non-virgin sequence on %s: %w", dbName, err)
		}
		nonVirgin = append(nonVirgin, name)
	}
	seqRows.Close()
	if err := seqRows.Err(); err != nil {
		return fmt.Errorf("iterate template sequences on %s: %w", dbName, err)
	}
	if len(nonVirgin) > 0 {
		return fmt.Errorf("lease baseline on %s: %d template sequence(s) are NON-VIRGIN at clone time (%v) — resetLeasedDatabase restores advanced sequences to setval(start_value), which would hand tests the wrong nextval for these; capture each sequence's birth nextval or exclude it before this can be trusted", dbName, len(nonVirgin), nonVirgin)
	}

	if _, err := db.ExecContext(ctx,
		"CREATE SCHEMA IF NOT EXISTS "+quoteIdent(leaseBaselineSchema)); err != nil {
		return fmt.Errorf("create %s schema: %w", leaseBaselineSchema, err)
	}

	for _, st := range baselineSnapshotTables {
		// Plain CREATE TABLE, NOT "IF NOT EXISTS". This function may only ever
		// run against a database in pristine template-fresh state, and its
		// callers guarantee that. If a snapshot table already exists, that
		// guarantee is broken and the correct outcome is a loud 42P07, not a
		// silent skip: "IF NOT EXISTS" would skip the SELECT entirely and leave
		// a stale snapshot standing in as the baseline, which every later reset
		// would restore to. That is a false green by construction.
		if _, err := db.ExecContext(ctx, fmt.Sprintf(
			"CREATE TABLE %s AS SELECT * FROM %s", st.snapshot(), st.qualified,
		)); err != nil {
			return fmt.Errorf("snapshot baseline rows for %s: %w", st.qualified, err)
		}
	}

	return nil
}

// leaseBaselineIsComplete reports whether dbName carries a baseline snapshot
// this code can trust: the schema AND every table in baselineSnapshotTables.
//
// Presence of the schema alone is not enough. A half-written baseline (the
// clone step died between two CREATE TABLEs) would pass a schema-only check,
// and the missing tables would only surface later as a restore error — or, if
// the set of snapshot tables changed, not surface at all.
func leaseBaselineIsComplete(ctx context.Context, baseDSN, dbName string) (bool, error) {
	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		return false, err
	}
	defer db.Close()

	for _, st := range baselineSnapshotTables {
		var exists bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM pg_tables
			    WHERE schemaname = $1 AND tablename = $2)`,
			leaseBaselineSchema, st.snapName,
		).Scan(&exists); err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

// leaseBaselineSchema holds the per-lease baseline snapshot. It lives OUTSIDE
// schemas 'public' and 'metaldocs' precisely so the reset's table scan (which
// filters to those two) cannot see — and therefore cannot delete — it.
const leaseBaselineSchema = "metaldocs_testkit"

// The template's seed-row tables (verified by scanning every table in the
// template for rows): role_capabilities 112, roles 8 (A8.1, migration 0318 —
// hand-maintained catalog, see the migration's own header comment), schema_migrations
// 101, river_migration 6, tenants 1, templates_template 1,
// templates_template_version 1. EVERY other table in public + metaldocs is
// empty in the template — including the counter tables. So deleting the rest
// to empty is exactly what a fresh clone gives, and no per-table seed logic
// is needed for anything outside this set.
//
// resetExclusions are the STATIC ones: seeded by the curated bootstrap and
// never written by any test, so there is no reason to churn them.
//
// Contrast baselineSnapshotTables below — the distinction is not cosmetic.
// Excluding a table that tests DO write lets a leased database drift from
// clone-equivalence with every reuse: the audit package's invariant assert
// caught this at 2 tenant rows, and a lease left unguarded reached 51.
//
// metaldocs.roles (added here, not to baselineSnapshotTables): it is a fixed,
// hand-maintained 8-row role catalog (migration 0318; A8.3 repoints the seed
// at a Go-registry-driven generator, but until then this is it), and no test
// or application code path writes to it (verified: the only reference outside
// this migration is migration_0318_test.go's own read-only row-count assert).
// Omitting it here was a real defect, not a hypothetical one: without this
// entry, deleteTargets (below) treats metaldocs.roles as an ordinary
// ephemeral table — ok for the FIRST lease of a template clone (pristine,
// carries all 8 rows), but resetLeasedDatabase deletes it on every REUSE and
// nothing restores it (it is not in baselineSnapshotTables either), so a
// reused lease's metaldocs.roles is permanently empty and any ordinary test's
// attempt to insert a metaldocs.capability_bindings row fails FK validation
// against role_code — not because the test did anything wrong, but because
// the leased database no longer represents the curated bootstrap state it is
// supposed to model.
var resetExclusions = map[string]struct{}{
	"metaldocs.role_capabilities": {},
	"metaldocs.roles":             {},
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

	// Guard the restore: every semi-static table must carry its baseline row
	// back after the DELETE+snapshot-restore, or the lease has silently
	// under-seeded (an empty snapshot restores 0 rows without erroring). Drive
	// the assertion from baselineSnapshotTables itself so the guard can NEVER
	// again drift from the restore set — the earlier hand-listed form checked
	// only 2 of the 3 restored tables (templates_template_version was
	// unguarded), the same one-directional-guard hole recorded in §11.22. All
	// three tables carry exactly 1 baseline row in the template (lease_pool.go
	// classification comment: tenants 1, templates_template 1,
	// templates_template_version 1).
	for _, st := range baselineSnapshotTables {
		if err := assertReferenceRowCount(ctx, db, dbName, st.qualified, 1); err != nil {
			return err
		}
	}
	return nil
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
