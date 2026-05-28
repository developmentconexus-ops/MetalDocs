// Package migrate applies post-baseline forward SQL files from a configured
// migrations directory. It is not responsible for fresh database bootstrap;
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
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrateLockKey); err != nil {
		return fmt.Errorf("migrate: acquire advisory lock: %w", err)
	}
	defer func() {
		if _, err := conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, migrateLockKey); err != nil && retErr == nil {
			retErr = fmt.Errorf("migrate: release advisory lock: %w", err)
		}
	}()

	applied, err := loadApplied(ctx, conn)
	if err != nil {
		return fmt.Errorf("migrate: load schema_migrations: %w", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}
	type file struct {
		version, path, name string
	}
	var files []file
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		m := versionRE.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		files = append(files, file{version: m[1], path: filepath.Join(dir, e.Name()), name: e.Name()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	skipped, ran := 0, 0
	for _, f := range files {
		if applied[f.version] {
			skipped++
			continue
		}
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
		ran++
	}
	log.Info("migrations done", "applied_now", ran, "already_applied", skipped, "total", len(files))
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
	defer rows.Close()
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
