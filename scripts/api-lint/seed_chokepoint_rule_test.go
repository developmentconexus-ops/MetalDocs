package main

import (
	"path/filepath"
	"testing"
)

// F3.1 T5 — SEED-CHOKEPOINT (validation-contract.md §1.5, spec.md PG-3).
//
// checkSeedChokepoint AST-scans for authz.SeedTxIdentity( call expressions.
// Any occurrence outside the chokepoint file(s), outside the declared
// allowlist, and outside _test.go is a violation naming file:line.

// POSITIVE proof: the real, post-collapse repo tree must be 0 violations —
// every remaining SeedTxIdentity call site is either the definition itself
// (internal/modules/iam/authz/context.go), a _test.go fixture, or listed in
// scripts/api-lint/seed-chokepoint-allowlist.txt.
func TestSeedChokepoint_CleanTreeGreen(t *testing.T) {
	got, err := checkSeedChokepoint(repoRoot(t), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT"); n != 0 {
		t.Fatalf("clean tree: want 0 SEED-CHOKEPOINT violations, got %d: %+v", n, got)
	}
}

// NEGATIVE proof (PG-3): a stray, non-allowlisted authz.SeedTxIdentity call in
// an application-service-shaped file must fire exactly 1 violation naming
// file:line. Removing it must return to 0 (asserted via a second run against
// a clean copy of the same fixture dir).
func TestSeedChokepoint_StrayCallFiresNamed(t *testing.T) {
	dir := t.TempDir()

	// Minimal go.mod-less fixture: checkSeedChokepoint walks the tree by file
	// name, not by build graph, so a syntactically valid .go file is enough
	// (mirrors checkTripwireArmDrift's fixture-root convention).
	strayFile := filepath.Join("internal", "modules", "fakemodule", "application", "stray_service.go")
	writeFile(t, dir, strayFile, `package application

import (
	"context"
	"database/sql"

	authz "metaldocs/internal/modules/iam/authz"
)

func doSomething(ctx context.Context, tx *sql.Tx, tenantID, actorID string) error {
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
		return err
	}
	return nil
}
`)

	got, err := checkSeedChokepoint(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT"); n != 1 {
		t.Fatalf("stray call: want 1 violation, got %d: %+v", n, got)
	}
	found := false
	for _, v := range got {
		if v.Rule != "SEED-CHOKEPOINT" {
			continue
		}
		if filepath.ToSlash(v.File) == filepath.ToSlash(filepath.Join(dir, strayFile)) && v.Line == 11 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected violation naming %s:11, got %+v", strayFile, got)
	}
}

// GREEN after removal: the same fixture dir with the stray call deleted must
// be 0 violations (the negative-then-positive pairing PG-3 requires).
func TestSeedChokepoint_AfterRemovalGreen(t *testing.T) {
	dir := t.TempDir()
	strayFile := filepath.Join("internal", "modules", "fakemodule", "application", "clean_service.go")
	writeFile(t, dir, strayFile, `package application

import "context"

func doSomethingElse(ctx context.Context) error {
	return nil
}
`)

	got, err := checkSeedChokepoint(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT"); n != 0 {
		t.Fatalf("clean fixture: want 0 violations, got %d: %+v", n, got)
	}
}

// The chokepoint file itself (internal/platform/db/runner.go) must never be
// flagged even though it references the seed SQL inline — it does not call
// authz.SeedTxIdentity(, so this is really a "no false positive" pin, but
// stated explicitly since it's the load-bearing exemption.
func TestSeedChokepoint_ChokepointFileNeverFlagged(t *testing.T) {
	got, err := checkSeedChokepoint(repoRoot(t), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, v := range got {
		if v.Rule != "SEED-CHOKEPOINT" {
			continue
		}
		if filepath.Base(v.File) == "runner.go" {
			t.Fatalf("chokepoint file must never be flagged: %+v", v)
		}
	}
}

// A _test.go file calling authz.SeedTxIdentity directly (as many unit tests
// legitimately do to set up fixtures) must never be flagged.
func TestSeedChokepoint_TestFilesExempt(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join("internal", "modules", "fakemodule", "application", "stray_service_test.go")
	writeFile(t, dir, testFile, `package application

import (
	"context"
	"database/sql"
	"testing"

	authz "metaldocs/internal/modules/iam/authz"
)

func TestSomething(t *testing.T) {
	var ctx context.Context
	var tx *sql.Tx
	_ = authz.SeedTxIdentity(ctx, tx, "t", "a")
}
`)

	got, err := checkSeedChokepoint(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT"); n != 0 {
		t.Fatalf("_test.go fixture: want 0 violations, got %d: %+v", n, got)
	}
}

// An allowlisted site (file:line present in seed-chokepoint-allowlist.txt)
// must not be flagged.
func TestSeedChokepoint_AllowlistedSiteNotFlagged(t *testing.T) {
	dir := t.TempDir()
	allowedFile := filepath.Join("internal", "modules", "fakemodule", "infrastructure", "grant_repository.go")
	writeFile(t, dir, allowedFile, `package infrastructure

import (
	"context"
	"database/sql"

	authz "metaldocs/internal/modules/iam/authz"
)

func grantAccess(ctx context.Context, tx *sql.Tx, tenantID, grantedByActor string) error {
	if err := authz.SeedTxIdentity(ctx, tx, tenantID, grantedByActor); err != nil {
		return err
	}
	return nil
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "seed-chokepoint-allowlist.txt"),
		allowedFile+":11  # category A — grantedByActor is a stored distinct actor, not ctx actor\n")

	got, err := checkSeedChokepoint(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT"); n != 0 {
		t.Fatalf("allowlisted site: want 0 violations, got %d: %+v", n, got)
	}
}

// A stale allowlist entry (line no longer matches any live SeedTxIdentity
// call) must be reported so the list cannot rot, mirroring
// tripwire-allowlist-stale.
func TestSeedChokepoint_StaleAllowlistEntryFires(t *testing.T) {
	dir := t.TempDir()
	cleanFile := filepath.Join("internal", "modules", "fakemodule", "infrastructure", "clean_repository.go")
	writeFile(t, dir, cleanFile, `package infrastructure

import "context"

func doNothing(ctx context.Context) error {
	return nil
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "seed-chokepoint-allowlist.txt"),
		cleanFile+":99  # stale — no longer matches any call site\n")

	got, err := checkSeedChokepoint(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SEED-CHOKEPOINT-ALLOWLIST-STALE"); n != 1 {
		t.Fatalf("stale allowlist entry: want 1 violation, got %d: %+v", n, got)
	}
}
