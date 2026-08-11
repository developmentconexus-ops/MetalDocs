// Package migrate applies post-baseline forward SQL files from a configured
// migrations directory (Apply), plus the ledger-less privilege/role grants
// stage (ApplyGrants). It is not responsible for fresh database bootstrap;
// curated baseline bootstrap is owned by scripts/dev-bootstrap-baseline.ps1.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsApplicableSQLFile reports whether name is a SQL file this package's
// discovery functions (Apply, ApplyGrants) treat as applicable: a *.sql file
// that is not a *_down.sql rollback script. Exported so
// tests/integration/testdb's bootstrap-bundle discovery
// (tests/integration/testdb/db.go's listSQLFiles) can share this exact
// filter instead of maintaining an independently drifting copy -- PR #110
// review finding: ApplyGrants and listSQLFiles used to disagree on
// *_down.sql handling, so a future grants rollback file could run in
// production while staying invisible to the test bootstrap and its schema
// fingerprint.
func IsApplicableSQLFile(name string) bool {
	return strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, "_down.sql")
}

var versionRE = regexp.MustCompile(`^(\d{4})_`)
var trailingTxnTokenRE = regexp.MustCompile(`(?is)\bcommit\b\s*;?\s*$`)

const migrateLockKey int64 = 0x4D444D4947528000

// Apply runs every *.sql file in dir whose version (the leading 4-digit prefix)
// is not present in public.schema_migrations. Files are applied in lexical
// order. Each file is executed as a single Exec; files must be self-contained
// SQL (BEGIN/COMMIT and ledger insert included).
// TODO: fail closed when a migration file omits explicit BEGIN/COMMIT so mixed transactional semantics cannot slip through.
func Apply(ctx context.Context, db *sql.DB, dir string, log *slog.Logger) (retErr error) {
	if log == nil {
		log = slog.Default()
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrateLockKey); err != nil {
		return fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}
	defer func() {
		// Detached on purpose: the advisory lock must still be released if ctx
		// was canceled mid-migration, so this uses WithoutCancel (not
		// Background) — it keeps request-scoped values (e.g. trace id) while
		// dropping only the cancellation/deadline.
		if _, err := conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrateLockKey); err != nil && retErr == nil {
			retErr = fmt.Errorf("migrate: release advisory lock: %w", err)
		}
	}()

	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return fmt.Errorf("migrate: load schema_migrations: %w", err)
	}
	files, err := collectMigrationFiles(dir)
	if err != nil {
		return err
	}

	skipped, ran := 0, 0
	for _, f := range files {
		if applied[f.version] {
			skipped++
			continue
		}
		if err := applyMigrationFile(ctx, conn, log, f); err != nil {
			return err
		}
		ran++
	}
	log.Info("migrations done", "applied_now", ran, "already_applied", skipped, "total", len(files))
	return nil
}

type migrationFile struct {
	version, path, name string
}

// collectMigrationFiles lists dir for *.sql files matching the leading
// 4-digit version prefix, returning them sorted in lexical (application)
// order.
func collectMigrationFiles(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir %q: %w", dir, err)
	}
	var files []migrationFile
	for _, e := range entries {
		if e.IsDir() || !IsApplicableSQLFile(e.Name()) {
			continue
		}
		m := versionRE.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		files = append(files, migrationFile{version: m[1], path: filepath.Join(dir, e.Name()), name: e.Name()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files, nil
}

// applyMigrationFile reads, validates, and executes a single migration file
// as one Exec, per Apply's self-contained-SQL contract.
func applyMigrationFile(ctx context.Context, conn *sql.Conn, log *slog.Logger, f migrationFile) error {
	body, err := os.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("read %s: %w", f.name, err)
	}
	if err := requireExplicitTransactionGuard(string(body)); err != nil {
		return fmt.Errorf("apply %s: %w", f.name, err)
	}
	log.Info("applying migration", "version", f.version, "file", f.name)
	if _, err := conn.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("apply %s: %w", f.name, err)
	}
	return nil
}

// ApplyGrants runs every *.sql file in dir, in lexical order, on every startup —
// unconditionally, with no schema_migrations ledger row and no skip check.
//
// The grants stage (db/grants) re-homes the privilege effects the curated
// baseline cannot carry (pg_dump --no-privileges emits no ACLs, and role
// creation is cluster-global so it is never dumped at all — see
// db/grants/0001_role_grants.sql). Those files were previously applied ONLY at
// fresh bootstrap (compose initdb.d, scripts/dev-bootstrap-baseline.ps1,
// tests/integration/testdb), so a long-lived volume never saw a later edit to
// the privilege posture. Running them here closes that gap: every file is
// idempotent by construction (guarded DO blocks + GRANT/REVOKE, which are
// themselves set-semantics), so replay is a no-op and a ledger row would only
// re-open the "edit never lands" hole.
//
// A missing or empty dir is a fatal error, not a silent pass: an API booting
// without the privilege stage is exactly the failure this function exists to
// prevent (no-fallback principle).
func ApplyGrants(ctx context.Context, db *sql.DB, dir string, log *slog.Logger) (retErr error) {
	if log == nil {
		log = slog.Default()
	}
	// Resolve the stage before touching the database so a packaging mistake
	// (image missing db/grants, wrong METALDOCS_GRANTS_DIR) fails fast and
	// unambiguously rather than mid-connection.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read grants dir %q: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !IsApplicableSQLFile(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	if len(names) == 0 {
		return fmt.Errorf("grants dir %q contains no .sql files; the privilege stage is mandatory", dir)
	}
	sort.Strings(names)

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate: acquire connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Same advisory lock as Apply: the grants stage runs immediately before it,
	// and two replicas booting together must not interleave DDL/ACL work.
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrateLockKey); err != nil {
		return fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}
	defer func() {
		// Detached on purpose: see the matching comment in Apply — the lock
		// must be released even if ctx was canceled, so use WithoutCancel
		// rather than Background to keep carrying request-scoped values.
		if _, err := conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, migrateLockKey); err != nil && retErr == nil {
			retErr = fmt.Errorf("migrate: release advisory lock: %w", err)
		}
	}()

	for _, name := range names {
		// #nosec G304 -- name is drawn from os.ReadDir(dir)'s own directory listing (line 122), and dir is the operator-configured grants directory (METALDOCS_GRANTS_DIR / deploy config) resolved once at process boot, never per-request; no request-scoped or tenant-scoped value reaches this path.
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if err := requireExplicitTransactionGuard(string(body)); err != nil {
			return fmt.Errorf("apply grants %s: %w", name, err)
		}
		log.Info("applying grants stage", "file", name)
		if _, err := conn.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply grants %s: %w", name, err)
		}
	}
	log.Info("grants stage done", "files", len(names))
	return nil
}

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func loadApplied(ctx context.Context, db queryer) (map[string]bool, error) {
	out := map[string]bool{}
	rows, err := db.QueryContext(ctx, `SELECT version FROM public.schema_migrations`)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			// Table may not exist on a brand-new DB; treat as empty.
			return out, nil
		}
		return nil, fmt.Errorf("load applied: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan migration row: %w", err)
		}
		out[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schema_migrations versions: %w", err)
	}
	return out, nil
}

func requireExplicitTransactionGuard(body string) error {
	normalized := trimLeadingSQLComments(body)
	if !strings.HasPrefix(strings.ToUpper(normalized), "BEGIN") {
		return errors.New("migration must start with explicit BEGIN")
	}
	if !trailingTxnTokenRE.MatchString(normalized) {
		return errors.New("migration must end with explicit COMMIT")
	}
	return nil
}

func trimLeadingSQLComments(body string) string {
	remaining := strings.TrimSpace(strings.TrimPrefix(body, "\uFEFF"))
	for {
		switch {
		case strings.HasPrefix(remaining, "--"):
			if idx := strings.IndexByte(remaining, '\n'); idx >= 0 {
				remaining = strings.TrimSpace(remaining[idx+1:])
				continue
			}
			return ""
		case strings.HasPrefix(remaining, "/*"):
			if idx := strings.Index(remaining, "*/"); idx >= 0 {
				remaining = strings.TrimSpace(remaining[idx+2:])
				continue
			}
			return ""
		default:
			return remaining
		}
	}
}
