//go:build integration
// +build integration

package testdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
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
)

// testNamespace is a fixed UUID v5 namespace for deterministic fixture IDs.
var testNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

var (
	templateDBOnce sync.Once
	templateDBName = fmt.Sprintf("metaldocs_test_template_%d", os.Getpid())
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

	ensureTemplateDatabase(t, baseDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}

	dbName := "metaldocs_test_" + randomSuffix(t)
	if _, err := adminDB.ExecContext(
		ctx,
		fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", quoteIdent(dbName), quoteIdent(templateDBName)),
	); err != nil {
		t.Fatalf("create isolated test database %s: %v", dbName, err)
	}

	db, err := openDBWithDatabase(baseDSN, dbName)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatalf("ping isolated test database %s: %v", dbName, err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer dropCancel()
		dropDB, err := openDBWithDatabase(baseDSN, "postgres")
		if err != nil {
			return
		}
		defer dropDB.Close()
		if err := dropDatabase(dropCtx, dropDB, dbName); err != nil {
			t.Logf("drop isolated test database %s: %v", dbName, err)
		}
	})

	return db, dbName
}

func ensureTemplateDatabase(t *testing.T, baseDSN string) {
	t.Helper()
	templateDBOnce.Do(func() {
		templateDBErr = rebuildTemplateDatabase(baseDSN)
	})
	if templateDBErr != nil {
		t.Fatalf("prepare template database: %v", templateDBErr)
	}
}

func rebuildTemplateDatabase(baseDSN string) error {
	adminDB, err := openDBWithDatabase(baseDSN, "postgres")
	if err != nil {
		return fmt.Errorf("open admin db: %w", err)
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := dropDatabase(ctx, adminDB, templateDBName); err != nil {
		return fmt.Errorf("drop stale template db: %w", err)
	}
	if _, err := adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdent(templateDBName))); err != nil {
		return fmt.Errorf("create template db: %w", err)
	}

	templateDB, err := openDBWithDatabase(baseDSN, templateDBName)
	if err != nil {
		return fmt.Errorf("open template db: %w", err)
	}
	defer templateDB.Close()

	if err := templateDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping template db: %w", err)
	}
	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer bootstrapCancel()
	if err := ApplyCuratedBootstrap(bootstrapCtx, templateDB); err != nil {
		return fmt.Errorf("apply curated bootstrap: %w", err)
	}
	return nil
}

func dropDatabase(ctx context.Context, db *sql.DB, name string) error {
	if _, err := db.ExecContext(ctx,
		`SELECT pg_terminate_backend(pid)
		   FROM pg_stat_activity
		  WHERE datname = $1
		    AND pid <> pg_backend_pid()`,
		name,
	); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(name))); err != nil {
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
	"acquire_lease":          {},
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
	"job_leases":             {},
	"role_capabilities":      {},
}

type sqlBundle struct {
	path string
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

// ApplyCuratedBootstrap mirrors the official database-level baseline bootstrap:
// prerequisites, curated schema, product reference data, and forward tail.
func ApplyCuratedBootstrap(ctx context.Context, db *sql.DB) error {
	root := repoRoot()
	bundles := []sqlBundle{
		{path: filepath.Join(root, "db", "prerequisites", "0001_extensions.sql")},
		{path: filepath.Join(root, "db", "baseline", "0001_current_schema.sql")},
		{path: filepath.Join(root, "db", "reference-data", "0001_product_reference_data.sql")},
	}

	migrationFiles, err := listSQLFiles(filepath.Join(root, "db", "migrations"))
	if err != nil {
		return fmt.Errorf("list db migrations: %w", err)
	}
	for _, path := range migrationFiles {
		bundles = append(bundles, sqlBundle{path: path})
	}

	for _, bundle := range bundles {
		sqlBytes, err := os.ReadFile(bundle.path)
		if err != nil {
			return fmt.Errorf("read sql bundle %s: %w", bundle.path, err)
		}
		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("apply sql bundle %s: %w", filepath.Base(bundle.path), err)
		}
	}
	return nil
}
