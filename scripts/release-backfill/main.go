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
//	go run ./scripts/release-backfill -docs <uuid> -repair-only -dry-run=false
//
// -repair-only is the operator-ratified expurgo pass for documents whose frozen
// artifacts are INVALID (blank frozen.docx / final.pdf produced before the
// frozen-body fix). It re-materializes from the SAME current revision with a
// fresh lineage pin and touches no approval fact — integrity restoration, not
// history mutation. Artifact keys are deterministic per revision, so the fresh
// render overwrites the invalid bytes in place.
//
// Part of that expurgo is freeing the keys that would otherwise swallow the
// re-dispatch, on BOTH layers and BOTH legs:
//
//   - the TERMINAL (dispatched/failed) staging rows in
//     metaldocs.materialize_dispatch_outbox and metaldocs.pdf_dispatch_outbox,
//     which dedupe on a generation-aware unique index; and
//   - the TERMINAL (published/dead-lettered) metaldocs.outbox_events rows on the
//     two delivery idempotency keys those dispatch jobs publish on.
//
// Leaving either layer in place makes the repair a successful-looking no-op —
// that is not a theory, it is what a live run did. Anything still in flight
// (pending/processing staging rows, unpublished events) is never deleted: the
// repair refuses instead. Every purged id is printed on the outcome line, and
// dry-run prints the same set as a plan. See backfill's package doc for the full
// swallow stack and the literal write set.
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
	repairOnly := flag.Bool("repair-only", false, "re-materialize a frozen, post-approval document from its current revision (invalid-artifact expurgo); purges the terminal materialize/pdf staging rows AND the terminal delivery outbox_events that would swallow the re-dispatch, records NO approval fact and enqueues NO release evaluation")
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
	pass := "BACKFILL (approval fact + materialization + evaluation)"
	if *repairOnly {
		pass = "REPAIR-ONLY (re-materialization only; approval facts untouched)"
	}
	fmt.Printf("release-backfill: %s, %s, %d document(s)\n", pass, mode, len(docs))

	opts := backfill.Options{RepairOnly: *repairOnly, DryRun: *dryRun}

	var failed int
	for _, id := range docs {
		res := backfill.RunDocument(ctx, deps, id, opts)
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
