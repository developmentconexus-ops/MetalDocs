//go:build integration
// +build integration

package testdb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	templateDBErr  error
)

// DeterministicID returns a deterministic UUID v5 based on test name + suffix.
func DeterministicID(t *testing.T, suffix string) string {
	t.Helper()
	return uuid.NewSHA1(testNamespace, []byte(t.Name()+":"+suffix)).String()
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

// Open returns a *sql.DB connected to a per-test isolated database cloned from
// a curated-baseline template. The returned string is kept for Qualified()
// compatibility; isolation happens at the database level, not via test schemas.
func Open(t *testing.T) (*sql.DB, string) {
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
	if _, err := adminDB.ExecContext(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", quoteIdent(dbName), quoteIdent(templateDBName)),
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
// version exists at a time. A stale-template sweep (under the same advisory
// lock) drops templates from previous fingerprints, so orphan
// metaldocs_test_template_* databases cannot accumulate. This replaces the old
// per-pid template, which was never dropped at process exit and leaked one DB
// per `go test` package-process (M3 close-out: 144 orphans after ~4h).
func prepareTemplateDatabase(baseDSN string) error {
	fingerprint, err := schemaFingerprint()
	if err != nil {
		return fmt.Errorf("compute schema fingerprint: %w", err)
	}
	name := fmt.Sprintf("metaldocs_test_template_%s", fingerprint[:16])
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
		if err := buildTemplate(ctx, conn, baseDSN, name, fingerprint); err != nil {
			return err
		}
	}

	// Bound to one template per fingerprint: drop every other test template
	// (previous schema versions) that has no active connections. Best-effort —
	// individual drop failures never fail the run.
	if err := sweepStaleTemplates(ctx, conn, name); err != nil {
		return fmt.Errorf("sweep stale templates: %w", err)
	}

	templateDBName = name
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

// sweepStaleTemplates drops every metaldocs_test_template_* database other than
// keep that has no active connections. This bounds templates to one per schema
// fingerprint even as the schema evolves across runs. Individual drops are
// best-effort: another process may claim a template between the SELECT and the
// DROP, and that race must not fail the run.
func sweepStaleTemplates(ctx context.Context, conn *sql.Conn, keep string) error {
	rows, err := conn.QueryContext(ctx,
		`SELECT d.datname
		   FROM pg_database d
		  WHERE d.datname LIKE 'metaldocs_test_template_%'
		    AND d.datname <> $1
		    AND NOT EXISTS (
		          SELECT 1 FROM pg_stat_activity a WHERE a.datname = d.datname)`,
		keep)
	if err != nil {
		return err
	}
	var stale []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		stale = append(stale, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, n := range stale {
		_, _ = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(n)))
	}
	return nil
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

// dropDatabaseConn is dropDatabase pinned to a single *sql.Conn, used by the
// template builder/sweeper which must terminate + drop on the same session that
// holds the advisory lock.
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

func openDBWithDatabase(dsn, dbName string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.Database = dbName
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
	"governance_events":      {},
	"grant_area_membership":  {},
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
// then the forward migration tail. It is the single source of both what
// ApplyCuratedBootstrap executes and what schemaFingerprint hashes, so the
// fingerprint can never drift from the actual template contents.
func curatedBundlePaths(root string) ([]string, error) {
	paths := []string{
		filepath.Join(root, "db", "prerequisites", "0001_extensions.sql"),
		filepath.Join(root, "db", "baseline", "0001_current_schema.sql"),
		filepath.Join(root, "db", "reference-data", "0001_product_reference_data.sql"),
	}
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
