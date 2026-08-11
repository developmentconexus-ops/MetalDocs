//go:build integration
// +build integration

package testdb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"metaldocs/internal/platform/bootstrap"
)

// testNamespace is a fixed UUID v5 namespace for deterministic fixture IDs.
var testNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

// schemaFingerprintSalt is bumped by hand whenever the curated template's
// contents can change WITHOUT a change to the hashed SQL files below — e.g.
// when bootstrap.MigrateRiverSchema (Go, not a *.sql file) starts emitting a
// different schema. Bumping it changes the fingerprint, so every dev's cached
// template rebuilds on the next run instead of silently serving stale schema.
const schemaFingerprintSalt = "v1"

var (
	templateDBOnce sync.Once
	// templateDBName is content-addressed and resolved during
	// ensureTemplateDatabase (empty until then). It embeds a fingerprint of
	// every curated bootstrap input, so exactly one template per schema version
	// exists at a time and the template is reused across processes and runs.
	templateDBName string
	// templateFingerprint16 is the same 16-char fingerprint embedded in
	// templateDBName, kept separately so the lease pool (lease_pool.go) can
	// compute deterministic, cross-process-adoptable slot names without
	// re-deriving or re-hashing the fingerprint itself.
	templateFingerprint16 string
	templateDBErr         error
)

// DeterministicID returns a deterministic UUID v5 based on test name + suffix.
func DeterministicID(t *testing.T, suffix string) string {
	t.Helper()
	return uuid.NewSHA1(testNamespace, []byte(t.Name()+":"+suffix)).String()
}

// AnyUUID is a fixed, syntactically valid UUID for tests that need a
// path-parameter VALUE that never has to exist — tier-1 authorization runs
// before any handler looks the id up, so a real row is never required.
const AnyUUID = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"

// New returns a *sql.DB connected to a leased test database, discarding the
// database name Open also returns. A thin convenience wrapper for callers
// (Task 17a's conformance suite) that never need Qualified()'s name.
func New(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := Open(t)
	return db
}

// DSN returns the test database connection string from env, or skips.
func DSN(t *testing.T) string {
	t.Helper()
	if v := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("DATABASE_URL")); v != "" {
		return v
	}
	t.Skip("DATABASE_URL/METALDOCS_DATABASE_URL not set")
	return ""
}

// Open returns a *sql.DB connected to a reset-safe leased database checked
// out of the package-process lease pool (see leasePool below). This is the
// default ~130-caller path: no per-test CREATE DATABASE / DROP DATABASE, just
// a DELETE-based reset (see resetLeasedDatabase — NOT truncate: TRUNCATE
// rewrites a relfilenode per relation and measured 55x slower here) of a
// leased physical database that is returned to the pool (never dropped) on
// t.Cleanup. The returned string is kept for Qualified() compatibility;
// isolation happens at the database level, not via test schemas.
//
// Tests that genuinely need a fresh, untouched clone (e.g. tests that mutate
// DDL, or that must prove behavior on a virgin database) should call
// OpenFreshDatabase instead.
func Open(t *testing.T) (*sql.DB, string) {
	t.Helper()

	baseDSN := DSN(t)

	phaseStart := time.Now()
	ensureTemplateDatabase(t, baseDSN)
	tracePhase(t, "ensure_template", time.Since(phaseStart))

	phaseStart = time.Now()
	dbName := globalLeasePool.checkout(t, baseDSN, templateDBName, templateFingerprint16)
	tracePhase(t, "lease_checkout", time.Since(phaseStart))

	// The DELETE-based reset is measured in single-digit milliseconds (it
	// touches rows, not relfilenodes — see resetLeasedDatabase). 60s is a
	// pure hang-guard, not a budget.
	resetCtx, resetCancel := context.WithTimeout(context.Background(), 60*time.Second)
	phaseStart = time.Now()
	if err := resetLeasedDatabase(resetCtx, baseDSN, dbName); err != nil {
		resetCancel()
		globalLeasePool.release(dbName)
		t.Fatalf("reset leased database: %v", err)
	}
	resetCancel()
	tracePhase(t, "reset", time.Since(phaseStart))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	phaseStart = time.Now()
	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		globalLeasePool.release(dbName)
		t.Fatalf("open leased test db: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		globalLeasePool.release(dbName)
		t.Fatalf("ping leased test database %s: %v", dbName, err)
	}
	tracePhase(t, "first_ping", time.Since(phaseStart))

	t.Cleanup(func() {
		_ = db.Close()
		globalLeasePool.release(dbName)
	})

	return db, dbName
}

// OpenFreshDatabase returns a *sql.DB connected to a per-test isolated
// database physically cloned from the curated-baseline template, then DROPped
// in t.Cleanup. This is the pre-lease-pool Open behaviour, kept for callers
// that need a virgin database (DDL mutation, schema-lockdown assertions,
// anything that must not share a physical database with any other test).
func OpenFreshDatabase(t *testing.T) (*sql.DB, string) {
	t.Helper()

	baseDSN := DSN(t)
	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer adminDB.Close()

	phaseStart := time.Now()
	ensureTemplateDatabase(t, baseDSN)
	tracePhase(t, "ensure_template", time.Since(phaseStart))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}

	dbName := "metaldocs_test_" + randomSuffix(t)
	phaseStart = time.Now()
	// STRATEGY is pinned, not left to the server default, because the default is
	// a per-version decision this suite cannot afford to inherit silently.
	// FILE_COPY forces a checkpoint before and after the copy; on this storage
	// stack a forced checkpoint blocks on IPC:CheckpointDone for an unbounded
	// time. Measured against this template (12 MB, 3 alternating reps):
	// WAL_LOG 1.11/1.74/2.56s vs FILE_COPY 5.52/7.78/36.41s. The 36s outlier is
	// the point — FILE_COPY's cost has no ceiling here, it is not merely 5x.
	if _, err := adminDB.ExecContext(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s STRATEGY = WAL_LOG", quoteIdent(dbName), quoteIdent(templateDBName)),
	); err != nil {
		t.Fatalf("create isolated test database %s: %v", dbName, err)
	}
	tracePhase(t, "create_database", time.Since(phaseStart))

	phaseStart = time.Now()
	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping isolated test database %s: %v", dbName, err)
	}
	tracePhase(t, "first_ping", time.Since(phaseStart))

	t.Cleanup(func() {
		_ = db.Close()
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer dropCancel()
		dropDB, err := openDBWithDatabase(baseDSN, "postgres")
		if err != nil {
			return
		}
		defer dropDB.Close()

		cleanupStart := time.Now()
		if err := terminateBackends(dropCtx, dropDB, dbName); err != nil {
			t.Logf("terminate backends on %s: %v", dbName, err)
			return
		}
		tracePhase(t, "terminate_backends", time.Since(cleanupStart))

		cleanupStart = time.Now()
		if err := dropDatabaseOnly(dropCtx, dropDB, dbName); err != nil {
			t.Logf("drop isolated test database %s: %v", dbName, err)
		}
		tracePhase(t, "drop_database", time.Since(cleanupStart))
	})

	return db, dbName
}

func ensureTemplateDatabase(t *testing.T, baseDSN string) {
	t.Helper()
	templateDBOnce.Do(func() {
		templateDBErr = prepareTemplateDatabase(baseDSN)
	})
	if templateDBErr != nil {
		t.Fatalf("prepare template database: %v", templateDBErr)
	}
}

// prepareTemplateDatabase builds (or reuses) a content-addressed template
// database whose name embeds a fingerprint of every curated bootstrap input.
// The template is shared across test processes and across runs: it is rebuilt
// only when the schema fingerprint changes, so at most ONE template per schema
// version exists at a time. A stale-template retirement pass (under the same
// advisory lock) LOGICALLY retires templates from previous fingerprints — it
// does not physically drop them; see retireStaleTemplates for why. This
// replaces the old per-pid template, which was never dropped at process exit
// and leaked one DB per `go test` package-process (M3 close-out: 144 orphans
// after ~4h). Physical reclamation of retired templates is a separate,
// explicit, idle-only step: GCRetiredDatabases.
func prepareTemplateDatabase(baseDSN string) error {
	fingerprint, err := schemaFingerprint()
	if err != nil {
		return fmt.Errorf("compute schema fingerprint: %w", err)
	}
	fingerprint16 := fingerprint[:16]
	name := fmt.Sprintf("metaldocs_test_template_%s", fingerprint16)
	lockKey := advisoryKey(fingerprint)

	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		return fmt.Errorf("open admin db: %w", err)
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// Pin one connection: advisory locks are session-scoped, and DROP/CREATE
	// DATABASE must run outside a transaction (this conn stays autocommit).
	conn, err := adminDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("pin admin conn: %w", err)
	}
	defer conn.Close()

	// Serialize build across concurrent test processes (go test ./... runs one
	// process per package in parallel). Clones are NOT serialized — only build.
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return fmt.Errorf("acquire template build lock: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
	}()

	ready, err := templateIsReady(ctx, conn, name, fingerprint)
	if err != nil {
		return fmt.Errorf("probe template readiness: %w", err)
	}
	if !ready {
		// This DROP is under the global lock deliberately: it targets ONLY the
		// exact database name we are about to rebuild (a partial/previous build
		// of THIS fingerprint), so serializing it here is what prevents two
		// concurrent processes from racing CREATE DATABASE on the same name.
		// That is a different thing from dropping OTHER databases under this
		// lock (see retireStaleTemplates) — the latter is what wedged the
		// cluster, this is not.
		if err := buildTemplate(ctx, conn, baseDSN, name, fingerprint); err != nil {
			return err
		}
	}

	// Bound to one LOGICALLY-live template per fingerprint: mark every other
	// test template (previous schema versions) as retired. Never drops —
	// see retireStaleTemplates.
	if err := retireStaleTemplates(ctx, conn, name); err != nil {
		return fmt.Errorf("retire stale templates: %w", err)
	}

	templateDBName = name
	templateFingerprint16 = fingerprint16
	return nil
}

// templateIsReady reports whether a fully-built template of this fingerprint
// already exists. Readiness is recorded as a COMMENT on the database, readable
// via shobj_description WITHOUT connecting to the template — so a crash between
// CREATE DATABASE and the final COMMENT leaves no marker and forces a clean
// rebuild.
func templateIsReady(ctx context.Context, conn *sql.Conn, name, fingerprint string) (bool, error) {
	var marker sql.NullString
	err := conn.QueryRowContext(ctx,
		`SELECT shobj_description(oid, 'pg_database')
		   FROM pg_database WHERE datname = $1`, name).Scan(&marker)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return marker.Valid && marker.String == readyMarker(fingerprint), nil
}

func readyMarker(fingerprint string) string {
	return "metaldocs-testdb-ready:" + fingerprint
}

// buildTemplate drops any partial/previous database of this exact name, creates
// it fresh, applies the curated bootstrap, and records the readiness marker
// LAST so partial builds are never mistaken for ready.
func buildTemplate(ctx context.Context, conn *sql.Conn, baseDSN, name, fingerprint string) error {
	if err := dropDatabaseConn(ctx, conn, name); err != nil {
		return fmt.Errorf("drop stale template db: %w", err)
	}
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(name))); err != nil {
		return fmt.Errorf("create template db: %w", err)
	}

	templateDB, err := openDBWithDatabase(baseDSN, name)
	if err != nil {
		return fmt.Errorf("open template db: %w", err)
	}
	// Must be closed before any clone reads this template as a source, else
	// CREATE DATABASE ... TEMPLATE fails with "source database is being
	// accessed by other users". Closed on return, before the advisory unlock.
	defer templateDB.Close()

	if err := templateDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping template db: %w", err)
	}
	if err := ApplyCuratedBootstrap(ctx, templateDB); err != nil {
		return fmt.Errorf("apply curated bootstrap: %w", err)
	}
	if _, err := conn.ExecContext(ctx,
		fmt.Sprintf("COMMENT ON DATABASE %s IS %s", quoteIdent(name), quoteLiteral(readyMarker(fingerprint)))); err != nil {
		return fmt.Errorf("mark template ready: %w", err)
	}
	return nil
}

// retireStaleTemplates marks every metaldocs_test_template_* database other
// than keep as LOGICALLY retired by rewriting its COMMENT marker. It does
// NOT drop them.
//
// This function used to drop stale templates outright. That was the observed
// outage mechanism: it ran on the same pinned conn that holds the global
// template-build advisory lock (see prepareTemplateDatabase), so a
// DROP DATABASE that wedged on a forced checkpoint (measured >5 minutes on
// IPC:CheckpointStart on this Docker Desktop/WSL2 virtualized-storage host)
// blocked every other test process waiting on that same lock — a
// cluster-wide stall caused by run-start bookkeeping. So this function no
// longer drops anything, ever, under any lock. It only rewrites the
// COMMENT marker so templateIsReady (which matches the exact ready marker,
// unweakened) never mistakes a retired template for a live one. Physical
// space reclamation is deliberately deferred to GCRetiredDatabases, which is
// never called from this path and must be invoked explicitly, idle-only,
// without holding this global lock.
func retireStaleTemplates(ctx context.Context, conn *sql.Conn, keep string) error {
	// A stale-fingerprint template with live backends belongs to a concurrent
	// run on other code (another branch/worktree sharing this cluster —
	// HARNESS §9.3a). Retiring it would clear its ready marker and force that
	// run to rebuild its template mid-flight: two tracks would retire each
	// other's templates and thrash. Only untouched templates are retired, so
	// concurrent tracks leave each other alone.
	rows, err := conn.QueryContext(ctx,
		`SELECT d.datname, shobj_description(d.oid, 'pg_database')
		   FROM pg_database d
		  WHERE d.datname LIKE 'metaldocs_test_template_%'
		    AND d.datname <> $1
		    AND NOT EXISTS (
		          SELECT 1 FROM pg_stat_activity a WHERE a.datname = d.datname)`,
		keep)
	if err != nil {
		return err
	}
	type stale struct {
		name   string
		marker sql.NullString
	}
	var staleDBs []stale
	for rows.Next() {
		var s stale
		if err := rows.Scan(&s.name, &s.marker); err != nil {
			rows.Close()
			return err
		}
		staleDBs = append(staleDBs, s)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, s := range staleDBs {
		nameFP16 := strings.TrimPrefix(s.name, templateNamePrefix)
		// OWNERSHIP PROOF, not name-prefix. Retire only a database this factory
		// can prove it BUILT — one already carrying a recognised ready marker
		// (metaldocs-testdb-ready:<fp>) whose fingerprint matches the name. An
		// unmarked or foreign-marked prefix match is NOT ours to touch: writing
		// the retired marker onto it would MANUFACTURE the recognised marker
		// GCRetiredDatabases later trusts (db.go:576-579 → gcDropIfIdleAndUnclaimed),
		// laundering a database we never owned into a droppable one — turning
		// "never drop a database you cannot prove you own" into a lie one hop
		// upstream of the drop. Name-prefix is a namespace, not a title deed; the
		// marker is the deed. Legacy unmarked orphans are the hub sweep's job
		// (R1), never this path's. (GPT final-round finding 2.)
		if !s.marker.Valid {
			continue
		}
		fp, ok := strings.CutPrefix(s.marker.String, "metaldocs-testdb-ready:")
		if !ok {
			continue // foreign / already-retired / non-ready marker — not ours to retire
		}
		// The database name encodes only the first 16 hex of the fingerprint
		// (templateNamePrefix + fingerprint[:16]); the ready marker carries the
		// full fingerprint. Tie them on exactly those 16 shared hex — the marker
		// must be a ready marker whose fingerprint opens with the 16 hex this name
		// encodes. 16 is all the name can attest to, so 16 is the tie.
		if len(fp) < 16 || len(nameFP16) < 16 || fp[:16] != nameFP16[:16] {
			continue // ready marker's fingerprint does not match the database name
		}
		_, _ = conn.ExecContext(ctx, fmt.Sprintf(
			"COMMENT ON DATABASE %s IS %s", quoteIdent(s.name), quoteLiteral(retiredTemplateMarker(nameFP16)),
		))
	}
	return nil
}

// templateNamePrefix is the fixed prefix every content-addressed template
// database name carries; the suffix is the schema fingerprint16.
const templateNamePrefix = "metaldocs_test_template_"

// retiredTemplateMarker is the COMMENT ON DATABASE value written by
// retireStaleTemplates. It carries the template's own fingerprint16 (read
// back from its name, not re-derived) so GCRetiredDatabases can recompute the
// exact advisory lock key that guarded this template's build without having
// to trust or re-parse a previous ready marker.
func retiredTemplateMarker(fingerprint16 string) string {
	return "metaldocs-testdb-retired:" + fingerprint16
}

// schemaFingerprint hashes every curated bootstrap input (in load order) plus a
// hand-bumped salt, yielding a stable hex digest that changes iff the template
// contents would change.
func schemaFingerprint() (string, error) {
	root := repoRoot()
	paths, err := curatedBundlePaths(root)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	_, _ = io.WriteString(h, schemaFingerprintSalt+"\x00")
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("read sql bundle %s: %w", p, err)
		}
		rel, _ := filepath.Rel(root, p)
		_, _ = io.WriteString(h, filepath.ToSlash(rel)+"\x00")
		_, _ = h.Write(b)
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// advisoryKey derives a stable per-fingerprint 64-bit lock key. Collisions are
// harmless: they only serialize two distinct fingerprints' builds needlessly.
func advisoryKey(fingerprint string) int64 {
	var k uint64
	for i := 0; i < 8 && i < len(fingerprint); i++ {
		k = k<<8 | uint64(fingerprint[i])
	}
	return int64(k)
}

// terminateBackends and dropDatabaseOnly are split so the two costs can be
// timed independently: terminate is a catalog scan whose cost tracks concurrent
// session count, while DROP is physical file removal whose cost tracks database
// size. They move in opposite directions under parallelism, so a combined
// timing hides which one is being paid.
func terminateBackends(ctx context.Context, db *sql.DB, name string) error {
	_, err := db.ExecContext(ctx,
		`SELECT pg_terminate_backend(pid)
		   FROM pg_stat_activity
		  WHERE datname = $1
		    AND pid <> pg_backend_pid()`,
		name,
	)
	return err
}

func dropDatabaseOnly(ctx context.Context, db *sql.DB, name string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(name)))
	return err
}

// dropDatabaseConn is dropDatabase pinned to a single *sql.Conn, used ONLY by
// buildTemplate to drop a partial/previous database OF ITS OWN EXACT NAME
// before recreating it, on the same session that holds the global template
// build lock — serializing that specific drop is precisely why the lock
// exists. retireStaleTemplates (formerly the "sweeper") no longer drops
// anything; do not reuse this for dropping any OTHER database under the
// global lock — that is the mechanism that wedged the cluster.
func dropDatabaseConn(ctx context.Context, conn *sql.Conn, name string) error {
	if _, err := conn.ExecContext(ctx,
		`SELECT pg_terminate_backend(pid)
		   FROM pg_stat_activity
		  WHERE datname = $1
		    AND pid <> pg_backend_pid()`,
		name,
	); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(name))); err != nil {
		return err
	}
	return nil
}

// classifyGCCandidate is the pure decision core of GCRetiredDatabases: given a
// database name and its COMMENT marker, decide whether this package can prove
// ownership of a no-longer-live database (→ drop candidate + which advisory
// lock proves no live claimant) or must skip it. Extracted so the ownership
// guard test can falsify the decision without running the checkpoint-forcing
// drop loop on a hot cluster.
func classifyGCCandidate(name string, marker sql.NullString, currentFP16, currentPolicy string) (dropCandidate bool, lockKey int64) {
	if !marker.Valid {
		return false, 0
	}
	switch {
	case strings.HasPrefix(marker.String, "metaldocs-testdb-retired:"):
		fp16 := strings.TrimPrefix(marker.String, "metaldocs-testdb-retired:")
		return true, advisoryKey(fp16)
	case strings.HasPrefix(marker.String, "metaldocs-testdb-lease:"):
		rest := strings.TrimPrefix(marker.String, "metaldocs-testdb-lease:")
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			return false, 0
		}
		fp16, policy := parts[0], parts[1]
		if fp16 == currentFP16 && policy == currentPolicy {
			// Still current: this slot may be adopted by a live or future
			// process running today's code. Not stale — must not drop.
			return false, 0
		}
		return true, leaseSlotLockKey(name)
	default:
		return false, 0
	}
}

// GCRetiredDatabases physically drops databases this package can prove it
// owns — via a recognised COMMENT ON DATABASE marker — that are no longer
// live: logically-retired templates (retireStaleTemplates) and lease slots
// (lease_pool.go) from a stale fingerprint or reset-policy version.
// Classification is the pure decision in classifyGCCandidate; this function
// owns only the enumeration, locking, and drop side effects.
//
// It is NEVER invoked automatically. Nothing in Open, OpenFreshDatabase,
// prepareTemplateDatabase, or any init/TestMain path calls this. It exists to
// be run deliberately (e.g. from a maintenance script) when the cluster is
// otherwise idle, because DROP DATABASE forces a checkpoint on this
// virtualized-storage host and a single fsync has been measured at up to
// ~0.9s — running it opportunistically inside a hot test path is the exact
// mechanism that wedged the cluster before (see retireStaleTemplates).
//
// A database with no recognised marker is SKIPPED and reported, never
// dropped — "never drop a database you cannot prove you own" is a hard rule,
// not a style preference. Each surviving candidate is additionally verified
// idle (zero pg_stat_activity backends) and unclaimed (this call can take its
// advisory lock) immediately before its own drop; either check failing skips
// that one database with no force and no wait. Drops are serialized one at a
// time and never run while holding the global template-build lock — that is
// the whole point of separating this from prepareTemplateDatabase.
func GCRetiredDatabases(t *testing.T) (dropped []string, err error) {
	t.Helper()
	baseDSN := DSN(t)

	currentFingerprint, err := schemaFingerprint()
	if err != nil {
		return nil, fmt.Errorf("compute current schema fingerprint: %w", err)
	}
	currentFP16 := currentFingerprint[:16]
	currentPolicy := fmt.Sprintf("%d", resetPolicyVersion)

	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		return nil, fmt.Errorf("open admin db: %w", err)
	}
	defer adminDB.Close()

	ctx := context.Background()

	rows, queryErr := adminDB.QueryContext(ctx,
		`SELECT d.datname, shobj_description(d.oid, 'pg_database')
		   FROM pg_database d
		  WHERE d.datname LIKE 'metaldocs_test_template_%'
		     OR d.datname LIKE 'metaldocs_test_lease_%'`)
	if queryErr != nil {
		return nil, fmt.Errorf("list candidate databases: %w", queryErr)
	}

	type candidate struct {
		name    string
		lockKey int64
	}
	var candidates []candidate
	var skipped []string
	for rows.Next() {
		var name string
		var marker sql.NullString
		if scanErr := rows.Scan(&name, &marker); scanErr != nil {
			rows.Close()
			return nil, scanErr
		}
		dropCandidate, lockKey := classifyGCCandidate(name, marker, currentFP16, currentPolicy)
		if !dropCandidate {
			skipped = append(skipped, name)
			continue
		}
		candidates = append(candidates, candidate{name: name, lockKey: lockKey})
	}
	rows.Close()
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	for _, c := range candidates {
		ok, dropErr := gcDropIfIdleAndUnclaimed(ctx, adminDB, c.name, c.lockKey)
		if dropErr != nil {
			err = errors.Join(err, fmt.Errorf("gc %s: %w", c.name, dropErr))
			continue
		}
		if ok {
			dropped = append(dropped, c.name)
		}
	}
	if len(skipped) > 0 {
		t.Logf("GCRetiredDatabases: skipped %d database(s) with no actionable/recognised marker: %v", len(skipped), skipped)
	}
	return dropped, err
}

// gcDropIfIdleAndUnclaimed drops name only if, right now, it has zero
// pg_stat_activity backends AND this call can itself take name's advisory
// lock (proving no live process — template builder or lease-pool owner —
// currently claims it). Either check failing skips the drop; there is no
// force and no wait, because a positive here must be trustworthy enough to
// justify a checkpoint-forcing DROP DATABASE.
func gcDropIfIdleAndUnclaimed(ctx context.Context, adminDB *sql.DB, name string, lockKey int64) (bool, error) {
	var backends int
	if err := adminDB.QueryRowContext(ctx,
		`SELECT count(*) FROM pg_stat_activity WHERE datname = $1`, name).Scan(&backends); err != nil {
		return false, fmt.Errorf("check backends: %w", err)
	}
	if backends > 0 {
		return false, nil
	}

	conn, err := adminDB.Conn(ctx)
	if err != nil {
		return false, fmt.Errorf("pin conn for lock probe: %w", err)
	}
	defer conn.Close()

	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&locked); err != nil {
		return false, fmt.Errorf("probe advisory lock: %w", err)
	}
	if !locked {
		// A live process still owns this template/slot's lock — do not drop.
		return false, nil
	}
	defer func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
	}()

	if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(name))); err != nil {
		return false, fmt.Errorf("drop database: %w", err)
	}
	return true, nil
}

func openDBWithDatabase(dsn, dbName string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.Database = dbName
	// Connect-time search_path, not ALTER DATABASE ... SET search_path: ~9
	// test files currently run that ALTER as setup boilerplate, all to this
	// same value. Supplying it here makes those calls redundant (a later
	// slice deletes them) without mutating database-level state that a leased
	// database would then carry across tests.
	if cfg.RuntimeParams == nil {
		cfg.RuntimeParams = map[string]string{}
	}
	cfg.RuntimeParams["search_path"] = "public, metaldocs"
	return stdlib.OpenDB(*cfg), nil
}

// Qualified returns a fully-qualified runtime table/function name in the
// isolated database. The schema token is ignored because test isolation is at
// the database level.
func Qualified(_ string, object string) string {
	return quoteIdent(runtimeSchemaName(object)) + "." + quoteIdent(object)
}

func runtimeSchemaName(object string) string {
	if _, ok := metaldocsOwnedObjects[object]; ok {
		return "metaldocs"
	}
	return "public"
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return fmt.Sprintf("%x", b)
}

func quoteIdent(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
}

// quoteLiteral quotes a string for use as a SQL string literal (DDL like
// COMMENT ON DATABASE cannot use bind parameters).
func quoteLiteral(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

// repoRoot finds the repo root by walking up from this file.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find repo root")
		}
		dir = parent
	}
}

var metaldocsOwnedObjects = map[string]struct{}{
	"document_families":      {},
	"document_process_areas": {},
	"document_profiles":      {},
	"iam_group_members":      {},
	"iam_group_roles":        {},
	"iam_groups":             {},
	"iam_user_roles":         {},
	"iam_users":              {},
	"idempotency_keys":       {},
	"role_capabilities":      {},
}

func listSQLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") || strings.HasSuffix(e.Name(), "_down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	sort.Strings(files)
	return files, nil
}

// curatedBundlePaths returns the ordered list of SQL files the curated
// bootstrap applies: prerequisites, curated schema, product reference data,
// role/privilege posture, then the forward migration tail. It is the single
// source of both what ApplyCuratedBootstrap executes and what
// schemaFingerprint hashes, so the fingerprint can never drift from the actual
// template contents.
//
// The migration tail is legitimately empty after the 2026-07-29 baseline fold
// (archive/migrations/post-baseline-2026-07-fold/); listSQLFiles returns an
// empty slice for a directory that holds only README.md.
func curatedBundlePaths(root string) ([]string, error) {
	paths := []string{
		filepath.Join(root, "db", "prerequisites", "0001_extensions.sql"),
		filepath.Join(root, "db", "baseline", "0001_current_schema.sql"),
		filepath.Join(root, "db", "reference-data", "0001_product_reference_data.sql"),
	}

	// The grants stage is auto-discovered, not hand-listed: it mirrors
	// internal/platform/migrate/migrate.go's ApplyGrants, which reads every
	// *.sql file under db/grants via os.ReadDir and applies them in lexical
	// order on every real boot. Hand-listing filenames here would recreate
	// the exact hand-synced-enumeration defect this bundle's own doc comment
	// warns against (curatedBundlePaths is supposed to be *the* source of
	// truth) -- a new db/grants/000N_*.sql file must need zero edits to this
	// function to be picked up by both the fingerprint and the bootstrap,
	// same as it needs zero edits to ship to production. Lexical filename
	// order encodes the dependency (0000_identity_roles.sql creates the
	// roles that 0001_role_grants.sql then grants to; see the matching
	// ordering comment in deploy/compose/docker-compose.yml).
	grantFiles, err := listSQLFiles(filepath.Join(root, "db", "grants"))
	if err != nil {
		return nil, fmt.Errorf("list db grants: %w", err)
	}
	paths = append(paths, grantFiles...)

	migrationFiles, err := listSQLFiles(filepath.Join(root, "db", "migrations"))
	if err != nil {
		return nil, fmt.Errorf("list db migrations: %w", err)
	}
	return append(paths, migrationFiles...), nil
}

// ApplyCuratedBootstrap mirrors the official database-level baseline bootstrap:
// prerequisites, curated schema, product reference data, and forward tail.
func ApplyCuratedBootstrap(ctx context.Context, db *sql.DB) error {
	paths, err := curatedBundlePaths(repoRoot())
	if err != nil {
		return err
	}

	for _, path := range paths {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read sql bundle %s: %w", path, err)
		}
		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply sql bundle %s: %w", filepath.Base(path), err)
		}
	}

	// F5.7 T1: production provisions River's own schema (river_job,
	// river_leader, river_queue, etc.) via bootstrap.MigrateRiverSchema at
	// metaldocs-api startup (see internal/platform/bootstrap/jobs.go). The
	// curated bundles above never create those tables, so any integration
	// test asserting on river_job rows would fail with 42P01 without this.
	// Locally River lives in the default/empty schema ("").
	if err := bootstrap.MigrateRiverSchema(ctx, db, ""); err != nil {
		return fmt.Errorf("migrate river schema: %w", err)
	}
	return nil
}
