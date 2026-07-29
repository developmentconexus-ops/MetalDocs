// Command release-backfill is the ADR 0085 Stage C one-shot backfill: it gives
// an EXPLICIT allowlist of legacy approved documents the release prerequisites
// the coordinator needs (approval fact, materialization dispatch, evaluation).
//
// It is a `go run` tool, not a fifth binary: it ships in no Docker image, is
// exposed by no HTTP route, and holds no long-lived state. Precedent:
// scripts/api-lint, scripts/req-trace.
//
// Usage:
//
//	go run ./scripts/release-backfill -docs <uuid>[,<uuid>...]              # dry-run (default)
//	go run ./scripts/release-backfill -docs <uuid>[,<uuid>...] -dry-run=false
//
// The DSN comes from -dsn or, when that is empty, the DATABASE_URL environment
// variable. This tool never reads .env and never prints the DSN: every error it
// reports has the DSN redacted before it reaches stdout/stderr.
//
// -docs is REQUIRED and has no "all approved" mode by design. Which documents
// are safe to backfill is an operator judgement (quarantined documents are
// handled by the Etapa 5 release gate, not here), so the allowlist is always
// supplied explicitly and never inferred.
//
// Exit code is 1 if any document failed; 0 otherwise.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"metaldocs/internal/platform/db/postgres"
	riverjobs "metaldocs/internal/platform/jobs/river"
	"metaldocs/scripts/release-backfill/backfill"
)

func main() {
	docsFlag := flag.String("docs", "", "REQUIRED comma-separated document UUIDs to backfill (explicit allowlist; there is no 'all approved' mode)")
	dryRun := flag.Bool("dry-run", true, "plan only and write nothing; pass -dry-run=false to actually write")
	dsnFlag := flag.String("dsn", "", "Postgres DSN; when empty the DATABASE_URL environment variable is used (never read from .env, never printed)")
	flag.Parse()

	dsn := strings.TrimSpace(*dsnFlag)
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		fatal("no database DSN: pass -dsn or set DATABASE_URL")
	}

	docs := splitDocs(*docsFlag)
	if len(docs) == 0 {
		fatal("-docs is required: pass the explicit list of document UUIDs to backfill")
	}

	ctx := context.Background()

	database, err := postgres.Open(ctx, dsn)
	if err != nil {
		fatal("connect to postgres: " + redact(err.Error(), dsn))
	}
	defer func() { _ = database.Close() }()

	// SkipUnknownJobCheck: this process registers no workers — it only inserts
	// jobs for metaldocs-jobs / metaldocs-worker to execute.
	bundle, err := riverjobs.NewClientBundle(database, riverjobs.Config{
		Schema:              strings.TrimSpace(os.Getenv("METALDOCS_JOBS_RIVER_SCHEMA")),
		SkipUnknownJobCheck: true,
	}, nil)
	if err != nil {
		fatal("river client: " + redact(err.Error(), dsn))
	}

	deps, err := backfill.Wire(database, bundle.Client)
	if err != nil {
		fatal(redact(err.Error(), dsn))
	}

	mode := "DRY-RUN (nothing will be written)"
	if !*dryRun {
		mode = "APPLY (writes committed per document)"
	}
	fmt.Printf("release-backfill: %s, %d document(s)\n", mode, len(docs))

	var failed int
	for _, id := range docs {
		res := backfill.RunDocument(ctx, deps, id, *dryRun)
		if res.Err != nil {
			res.Err = fmt.Errorf("%s", redact(res.Err.Error(), dsn))
			failed++
		}
		fmt.Println(res.String())
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "release-backfill: %d of %d document(s) failed\n", failed, len(docs))
		os.Exit(1)
	}
}

func splitDocs(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// redact removes the DSN from any message before it is printed. Driver errors
// routinely echo the connection string back; this is the single chokepoint that
// stops one from reaching a terminal or a log.
func redact(msg, dsn string) string {
	if dsn == "" {
		return msg
	}
	return strings.ReplaceAll(msg, dsn, "<dsn-redacted>")
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "release-backfill: "+msg)
	os.Exit(1)
}
