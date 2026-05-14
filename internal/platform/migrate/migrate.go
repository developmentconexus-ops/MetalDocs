// Package migrate applies post-baseline forward SQL files from a configured
// migrations directory. It is not responsible for fresh database bootstrap;
// curated baseline bootstrap is owned by scripts/dev-bootstrap-baseline.ps1.
package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var versionRE = regexp.MustCompile(`^(\d{4})_`)

// Apply runs every *.sql file in dir whose version (the leading 4-digit prefix)
// is not present in public.schema_migrations. Files are applied in lexical
// order. Each file is executed as a single Exec — files must be self-contained
// SQL (BEGIN/COMMIT and ledger insert included).
func Apply(ctx context.Context, db *sql.DB, dir string, log *slog.Logger) error {
	if log == nil {
		log = slog.Default()
	}
	applied, err := loadApplied(ctx, db)
	if err != nil {
		return fmt.Errorf("load schema_migrations: %w", err)
	}
	// High-water mark: only apply versions strictly greater than the max
	// recorded version. Pre-ledger history (migrations applied before
	// schema_migrations existed) must not be re-run on startup.
	maxApplied := ""
	for v := range applied {
		if v > maxApplied {
			maxApplied = v
		}
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
		if applied[f.version] || (maxApplied != "" && f.version <= maxApplied) {
			skipped++
			continue
		}
		body, err := os.ReadFile(f.path)
		if err != nil {
			return fmt.Errorf("read %s: %w", f.name, err)
		}
		log.Info("applying migration", "version", f.version, "file", f.name)
		if _, err := db.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", f.name, err)
		}
		ran++
	}
	log.Info("migrations done", "applied_now", ran, "already_applied", skipped, "total", len(files))
	return nil
}

func loadApplied(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	out := map[string]bool{}
	rows, err := db.QueryContext(ctx, `SELECT version FROM public.schema_migrations`)
	if err != nil {
		// Table may not exist on a brand-new DB; treat as empty.
		return out, nil
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}
