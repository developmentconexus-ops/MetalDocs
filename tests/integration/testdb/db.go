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
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// testNamespace is a fixed UUID v5 namespace for deterministic fixture IDs.
var testNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

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

// Open returns a *sql.DB connected to the test database and a unique schema
// for this test. The schema is dropped automatically at test cleanup.
func Open(t *testing.T) (*sql.DB, string) {
	t.Helper()

	db, err := sql.Open("pgx", DSN(t))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("integration DB unreachable: %v", err)
	}

	schema := "metaldocs_test_" + randomSuffix(t)
	for _, name := range []string{metaSchemaName(schema), publicSchemaName(schema)} {
		if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+quoteIdent(name)); err != nil {
			t.Fatalf("create test schema %s: %v", name, err)
		}
	}

	ApplyMigrations(t, db, schema)

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+quoteIdent(publicSchemaName(schema))+" CASCADE")
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+quoteIdent(metaSchemaName(schema))+" CASCADE")
	})
	return db, schema
}

// Qualified returns a fully-qualified table/function name in the test schema.
func Qualified(schema, object string) string {
	return quoteIdent(qualifiedSchemaName(schema, object)) + "." + quoteIdent(object)
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

var (
	schemaPattern = regexp.MustCompile(`(?i)\bSCHEMA\s+(metaldocs|public)\b`)
	dotPattern    = regexp.MustCompile(`\b(metaldocs|public)\.`)
	searchPathEq  = regexp.MustCompile(`(?i)(search_path\s*=\s*)(.+)`)
	searchPathTo  = regexp.MustCompile(`(?i)(search_path\s+to\s+)(.+)`)
)

var preservedSettings = []string{
	"metaldocs.actor_id",
	"metaldocs.asserted_caps",
	"metaldocs.bypass_authz",
	"metaldocs.cancel_in_progress",
	"metaldocs.verified_capability",
}

func rewriteSchema(sqlText, schema string) string {
	metaSchema := quoteIdent(metaSchemaName(schema))
	publicSchema := quoteIdent(publicSchemaName(schema))
	out := sqlText
	placeholders := make(map[string]string, len(preservedSettings))
	for idx, setting := range preservedSettings {
		placeholder := fmt.Sprintf("__TESTDB_GUC_%d__", idx)
		placeholders[placeholder] = setting
		out = strings.ReplaceAll(out, setting, placeholder)
	}

	out = dotPattern.ReplaceAllStringFunc(out, func(match string) string {
		if strings.HasPrefix(strings.ToLower(match), "metaldocs.") {
			return metaSchema + "."
		}
		return publicSchema + "."
	})
	out = schemaPattern.ReplaceAllStringFunc(out, func(match string) string {
		lower := strings.ToLower(match)
		if strings.Contains(lower, "metaldocs") {
			return "SCHEMA " + metaSchema
		}
		return "SCHEMA " + publicSchema
	})
	out = rewriteSearchPaths(out, metaSchema, publicSchema)
	for placeholder, setting := range placeholders {
		out = strings.ReplaceAll(out, placeholder, setting)
	}
	return out
}

func rewriteSearchPaths(sqlText, metaSchema, publicSchema string) string {
	replaceList := func(list string) string {
		parts := strings.Split(list, ",")
		for i, part := range parts {
			trimmed := strings.TrimSpace(part)
			unquoted := strings.Trim(trimmed, `' "`)
			switch strings.ToLower(unquoted) {
			case "metaldocs":
				parts[i] = metaSchema
			case "public":
				parts[i] = publicSchema
			default:
				parts[i] = trimmed
			}
		}
		return strings.Join(parts, ", ")
	}

	out := searchPathEq.ReplaceAllStringFunc(sqlText, func(match string) string {
		parts := searchPathEq.FindStringSubmatch(match)
		return parts[1] + replaceList(parts[2])
	})
	out = searchPathTo.ReplaceAllStringFunc(out, func(match string) string {
		parts := searchPathTo.FindStringSubmatch(match)
		return parts[1] + replaceList(parts[2])
	})
	return out
}

func metaSchemaName(base string) string {
	return base + "_meta"
}

func publicSchemaName(base string) string {
	return base + "_public"
}

var metaldocsOwnedObjects = map[string]struct{}{
	"acquire_lease":         {},
	"governance_events":     {},
	"grant_area_membership": {},
	"iam_users":             {},
	"idempotency_keys":      {},
	"job_leases":            {},
	"role_capabilities":     {},
}

func qualifiedSchemaName(base, object string) string {
	if _, ok := metaldocsOwnedObjects[object]; ok {
		return metaSchemaName(base)
	}
	return publicSchemaName(base)
}

type sqlBundle struct {
	path    string
	rewrite bool
}

func listSQLFiles(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read sql dir %s: %v", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") || strings.HasSuffix(e.Name(), "_down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, e.Name()))
	}
	sort.Strings(files)
	return files
}

// ApplyMigrations mirrors the curated bootstrap flow used by runtime startup.
func ApplyMigrations(t *testing.T, db *sql.DB, schema string) {
	t.Helper()

	root := repoRoot()
	bundles := []sqlBundle{
		{path: filepath.Join(root, "db", "prerequisites", "0001_extensions.sql"), rewrite: false},
		{path: filepath.Join(root, "db", "baseline", "0001_current_schema.sql"), rewrite: true},
		{path: filepath.Join(root, "db", "reference-data", "0001_product_reference_data.sql"), rewrite: true},
	}

	for _, path := range listSQLFiles(t, filepath.Join(root, "db", "migrations")) {
		bundles = append(bundles, sqlBundle{path: path, rewrite: true})
	}

	for _, bundle := range bundles {
		sqlBytes, err := os.ReadFile(bundle.path)
		if err != nil {
			t.Fatalf("read sql bundle %s: %v", bundle.path, err)
		}
		sqlText := string(sqlBytes)
		if bundle.rewrite {
			sqlText = rewriteSchema(sqlText, schema)
		}
		if _, err := db.Exec(sqlText); err != nil {
			t.Fatalf("apply sql bundle %s: %v", filepath.Base(bundle.path), err)
		}
	}
}
