// Command problem-codes-dump emits the canonical RFC 9457 Problem code catalog
// from the registry itself, and writes the artifacts that have to agree with it.
//
// ADR 0089, step 12. The predecessor (scripts/dump-error-codes.go) scraped the
// source tree with five regexes for `problem.Code = "literal"` shapes. That was
// the only tool it could have been while a code was a bare string: there was
// nothing to ask. Now every code is created by problem.Register at package init,
// so the catalog is a runtime value — this command imports the registering
// packages, asks problem.Registrations(), and writes what it is told.
//
// The scraper could only ever approximate. It needed an isNoise() denylist to
// drop file paths and HTTP methods that happened to match its code regex, it
// silently missed any code not written in one of the five shapes it knew (eight
// taxonomy codes were invisible to it for exactly this reason — annex §2.6b),
// and it could not report a status because a status did not exist anywhere near
// the literal. None of those failure modes are reachable from a registry.
//
// The one thing this design requires is that every registering package be linked
// into this binary — a package nobody imports never runs its init and its codes
// never appear. That is not left to vigilance: scripts/api-lint's
// PROBLEM-DUMP-IMPORT rule AST-scans for problem.Register calls and fails the
// build when a registering package is missing from the import block below.
//
//	go run ./cmd/problem-codes-dump          # write the artifacts
//	go run ./cmd/problem-codes-dump -check   # fail if they are stale (CI)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"metaldocs/internal/platform/problem"

	// Registering packages. Blank-imported for their init side effect: each one
	// declares its module's problem.Code vars, and a package that is not linked
	// in contributes nothing to problem.Registrations(). Kept in sync by the
	// PROBLEM-DUMP-IMPORT api-lint rule, not by memory.
	_ "metaldocs/internal/modules/approval/http"
	_ "metaldocs/internal/modules/controlleddocuments/delivery/http"
	_ "metaldocs/internal/modules/documents/delivery/http"
	_ "metaldocs/internal/modules/taxonomy/delivery/http"
	_ "metaldocs/internal/modules/templates/delivery/http"
	_ "metaldocs/internal/modules/tokens/delivery/http"
)

const (
	feJSONPath = "frontend/apps/web/src/lib/api/error-codes.generated.json"
	wikiPath   = "wiki/references/problem-codes.md"

	regenCmd = "go run ./cmd/problem-codes-dump"
)

func main() {
	check := flag.Bool("check", false, "verify the artifacts are current instead of writing them; exit 1 if any is stale")
	flag.Parse()

	regs := problem.Registrations()
	if len(regs) == 0 {
		// Unreachable in practice, but a zero catalog would silently truncate
		// both artifacts to empty and pass -check on the next run.
		fmt.Fprintln(os.Stderr, "problem-codes-dump: registry is empty; no registering package was linked in")
		os.Exit(1)
	}
	sort.Slice(regs, func(i, j int) bool { return regs[i].Code.String() < regs[j].Code.String() })

	artifacts := map[string][]byte{
		feJSONPath: renderFEJSON(regs),
		wikiPath:   renderWiki(regs),
	}

	stale := make([]string, 0, len(artifacts))
	for _, path := range sortedKeys(artifacts) {
		want := artifacts[path]
		if *check {
			got, err := os.ReadFile(filepath.FromSlash(path))
			if err != nil || !bytes.Equal(got, want) {
				stale = append(stale, path)
			}
			continue
		}
		if err := os.WriteFile(filepath.FromSlash(path), want, 0o644); err != nil { // #nosec G306 -- path iterates the fixed feJSONPath/wikiPath constants; both are committed generated artifacts (FE snapshot + wiki table) that must stay git/build readable at 0644, no secret material.
			fmt.Fprintf(os.Stderr, "problem-codes-dump: write %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}

	if *check {
		if len(stale) > 0 {
			fmt.Fprintf(os.Stderr,
				"problem-codes-dump: stale artifact(s):\n  %s\n\nThe Problem code registry changed and these were not regenerated. Run:\n  %s\n",
				strings.Join(stale, "\n  "), regenCmd)
			os.Exit(1)
		}
		fmt.Printf("problem-codes-dump: %d codes, all artifacts current\n", len(regs))
		return
	}
	fmt.Printf("problem-codes-dump: %d codes\n", len(regs))
}

// renderFEJSON writes the snapshot the frontend coverage test reads
// (errorMessages.coverage.test.ts), which fails the frontend build when a
// backend code has no PT-BR string — and, in the other direction, when
// errorMessages.ts still maps a code the backend stopped emitting.
//
// Field-level codes (problem.RegisterField) are deliberately absent: they name a
// per-field validation reason inside Problem.errors[], not a Problem.code, and
// the frontend renders them from the field context rather than the code map.
func renderFEJSON(regs []problem.Registration) []byte {
	codes := make([]string, 0, len(regs))
	for _, r := range regs {
		codes = append(codes, r.Code.String())
	}
	payload := map[string]any{
		"$comment": "Generated by " + regenCmd + " from the problem.Register catalog. Do not edit by hand.",
		"codes":    codes,
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "problem-codes-dump: marshal: %v\n", err)
		os.Exit(1)
	}
	return append(body, '\n')
}

// renderWiki writes the human-facing catalog: one table per semantic family, in
// the family order declared by problem.Families(), each row carrying the code,
// the HTTP status it resolves to, the owning module, and the source site that
// registered it.
//
// It is generated rather than maintained because a hand-written code table is
// the same hand-synced-enumeration defect class ADR 0089 exists to remove: it
// drifts the moment someone adds a code without opening the wiki, and a drifted
// table is worse than no table — it is confidently wrong.
func renderWiki(regs []problem.Registration) []byte {
	byFamily := map[string][]problem.Registration{}
	for _, r := range regs {
		byFamily[r.Family] = append(byFamily[r.Family], r)
	}

	var b strings.Builder
	b.WriteString("# Problem codes (RFC 9457)\n\n")
	b.WriteString("<!-- GENERATED by `" + regenCmd + "` from the problem.Register catalog. Do not edit by hand. -->\n\n")
	b.WriteString("Every `problem+json` body the API emits carries a `code` from this table. ")
	b.WriteString("The vocabulary is closed at the type level: `problem.Code` has an unexported field, so a ")
	b.WriteString("code that is not registered here cannot be constructed — a bare string literal is a compile ")
	b.WriteString("error, not a review finding (ADR 0089).\n\n")
	b.WriteString("Families are **semantic, never module-named**. A module name on a public wire contract leaks ")
	b.WriteString("internal topology to clients and outlives the module that coined it. Status comes from the ")
	b.WriteString("registration, so the same code cannot answer two different statuses on two different routes.\n\n")
	b.WriteString(fmt.Sprintf("**%d codes across %d families.**\n", len(regs), len(byFamily)))

	for _, f := range problem.Families() {
		rows := byFamily[f.Name]
		if len(rows) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("\n## `%s` — default %d\n\n", f.Name, f.DefaultStatus))
		b.WriteString("| Code | Status | Module | Declared at |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, r := range rows {
			status := fmt.Sprintf("%d", r.DefaultStatus)
			if r.DefaultStatus != f.DefaultStatus {
				// A row that departs from its family default carries the reason
				// RegisterWithStatus made mandatory, so the table shows WHY
				// rather than leaving a reader to assume a mistake.
				status = fmt.Sprintf("**%d** — %s", r.DefaultStatus, r.StatusOverrideReason)
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | %s | `%s` |\n",
				r.Code.String(), status, r.Module, r.DeclaredAt))
		}
	}
	return []byte(b.String())
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
