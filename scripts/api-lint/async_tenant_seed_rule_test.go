package main

import (
	"path/filepath"
	"testing"
)

// F3.2 T3 — ASYNC-TENANT-SEED (validation-contract.md §2.3, spec.md PG-3).
//
// checkAsyncTenantSeed AST-scans the worker/jobs handler package roots
// (asyncHandlerRoots) for a DB write (Exec/ExecContext with a string-literal
// INSERT/UPDATE/DELETE) against a table in async-tenant-tables.txt whose
// enclosing function does not also call SeedTxTenant/SeedTxIdentity, unless
// the site is on async-tenant-seed-allowlist.txt.

// POSITIVE proof: the real repo tree must be 0 violations — the five F3.2
// seed sites all call SeedTxTenant in the same function as their write, and
// the one same-file handler-local exception (notifications fanout's
// insertRow) is allowlisted (category E).
func TestAsyncTenantSeed_CleanTreeGreen(t *testing.T) {
	got, err := checkAsyncTenantSeed(repoRoot(t), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 0 {
		t.Fatalf("clean tree: want 0 ASYNC-TENANT-SEED violations, got %d: %+v", n, got)
	}
}

// NEGATIVE proof (PG-3): an unseeded worker write against a tenant-scoped
// table, in a fixture rooted under one of the scanned async handler roots
// (internal/platform/worker), must fire exactly 1 violation naming
// file:line+table. This is the throwaway RED case.
func TestAsyncTenantSeed_UnseededWorkerWriteFiresNamed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	strayFile := filepath.Join("internal", "platform", "worker", "stray_runner.go")
	writeFile(t, dir, strayFile, `package worker

import (
	"context"
	"database/sql"
)

func handleStray(ctx context.Context, tx *sql.Tx, tenantID, docID string) error {
	_, err := tx.ExecContext(ctx, "UPDATE documents SET status='x' WHERE tenant_id=$1 AND id=$2", tenantID, docID)
	return err
}
`)

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 1 {
		t.Fatalf("unseeded write: want 1 violation, got %d: %+v", n, got)
	}
	found := false
	for _, v := range got {
		if v.Rule != "ASYNC-TENANT-SEED" {
			continue
		}
		if filepath.ToSlash(v.File) == filepath.ToSlash(filepath.Join(dir, strayFile)) && v.Line == 9 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected violation naming %s:9, got %+v", strayFile, got)
	}
}

// GREEN after fix: the same fixture with SeedTxTenant added to the same
// function must return to 0 violations (RED->GREEN pairing PG-3 requires).
func TestAsyncTenantSeed_AfterSeedingGreen(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	fixedFile := filepath.Join("internal", "platform", "worker", "fixed_runner.go")
	writeFile(t, dir, fixedFile, `package worker

import (
	"context"
	"database/sql"

	"metaldocs/internal/modules/iam/authz"
)

func handleFixed(ctx context.Context, tx *sql.Tx, tenantID, docID string) error {
	if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, "UPDATE documents SET status='x' WHERE tenant_id=$1 AND id=$2", tenantID, docID)
	return err
}
`)

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 0 {
		t.Fatalf("seeded fixture: want 0 violations, got %d: %+v", n, got)
	}
}

// A write outside any async handler root (e.g. a sync API-path repository)
// must never be flagged — the F3.1 TxRunner chokepoint covers that seam via
// ctx auto-seed, not manual SeedTxTenant calls, so scanning it here would be
// a false positive by construction (contract §2.3 scope).
func TestAsyncTenantSeed_SyncPathOutsideHandlerRootsNeverFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	syncFile := filepath.Join("internal", "modules", "documents", "repository", "sync_repo.go")
	writeFile(t, dir, syncFile, `package repository

import (
	"context"
	"database/sql"
)

func UpdateDocumentName(ctx context.Context, tx *sql.Tx, tenantID, docID, name string) error {
	_, err := tx.ExecContext(ctx, "UPDATE documents SET name=$3 WHERE tenant_id=$1 AND id=$2", tenantID, docID, name)
	return err
}
`)

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 0 {
		t.Fatalf("sync-path file outside handler roots: want 0 violations, got %d: %+v", n, got)
	}
}

// A write against a table absent from async-tenant-tables.txt (not tenant-
// scoped / no RLS) must never be flagged even when unseeded.
func TestAsyncTenantSeed_NonTenantTableNeverFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	strayFile := filepath.Join("internal", "platform", "worker", "system_table_runner.go")
	writeFile(t, dir, strayFile, `package worker

import (
	"context"
	"database/sql"
)

func sweepLeases(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "DELETE FROM river_leader WHERE expires_at < now()")
	return err
}
`)

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 0 {
		t.Fatalf("non-tenant table: want 0 violations, got %d: %+v", n, got)
	}
}

// An allowlisted site (file:line present in async-tenant-seed-allowlist.txt)
// must not be flagged even though it is unseeded.
func TestAsyncTenantSeed_AllowlistedSiteNotFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	allowedFile := filepath.Join("internal", "modules", "render", "fanout", "claim_repo.go")
	writeFile(t, dir, allowedFile, `package fanout

import (
	"context"
	"database/sql"
)

func claimAndTouch(ctx context.Context, tx *sql.Tx, tenantID, docID string) error {
	_, err := tx.ExecContext(ctx, "UPDATE documents SET status='claimed' WHERE tenant_id=$1 AND id=$2", tenantID, docID)
	return err
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-seed-allowlist.txt"),
		allowedFile+":9  # (A) outbox claim step, cross-tenant by design\n")

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED"); n != 0 {
		t.Fatalf("allowlisted site: want 0 violations, got %d: %+v", n, got)
	}
}

// A stale allowlist entry (line no longer matches any live unseeded write)
// must be reported so the list cannot rot, mirroring
// seed-chokepoint-allowlist-stale / tripwire-allowlist-stale.
func TestAsyncTenantSeed_StaleAllowlistEntryFires(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "documents\n")

	cleanFile := filepath.Join("internal", "platform", "worker", "clean_runner.go")
	writeFile(t, dir, cleanFile, `package worker

import "context"

func doNothing(ctx context.Context) error {
	return nil
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-seed-allowlist.txt"),
		cleanFile+":99  # stale — no longer matches any call site\n")

	got, err := checkAsyncTenantSeed(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "ASYNC-TENANT-SEED-ALLOWLIST-STALE"); n != 1 {
		t.Fatalf("stale allowlist entry: want 1 violation, got %d: %+v", n, got)
	}
}
