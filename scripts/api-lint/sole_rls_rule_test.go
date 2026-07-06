package main

import (
	"path/filepath"
	"testing"
)

// M7 F7.4 §4.3 — SOLE-RLS-ASYNC-READ.
//
// checkSoleRLSAsyncRead AST-scans the worker/jobs handler package roots
// (asyncHandlerRoots) for a Query/QueryContext/QueryRow/QueryRowContext call
// whose first string-literal SQL argument is a SELECT against a table in
// async-tenant-tables.txt with no explicit tenant_id/actor_tenant_id token in
// the SQL text, unless the site is on sole-rls-read-allowlist.txt.

// POSITIVE proof: the real repo tree must be 0 violations after M6 F6.4 / F7.3
// predicate work (the review-due ports now carry explicit tenant_id
// predicates).
func TestSoleRLSAsyncRead_CleanTreeGreen(t *testing.T) {
	got, err := checkSoleRLSAsyncRead(repoRoot(t), true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 0 {
		t.Fatalf("clean tree: want 0 SOLE-RLS-ASYNC-READ violations, got %d: %+v", n, got)
	}
}

// NEGATIVE proof (required): a tenant-scoped SELECT with no tenant predicate,
// in a fixture rooted under an async handler root, must fire exactly 1
// violation naming file:line+table.
func TestSoleRLSAsyncRead_UnpredicatedReadFiresNamed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	strayFile := filepath.Join("internal", "platform", "worker", "stray_reader.go")
	writeFile(t, dir, strayFile, `package worker

import (
	"context"
	"database/sql"
)

func handleStrayRead(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM metaldocs.tenant_lifecycle_jobs WHERE status='x'")
	if err != nil {
		return err
	}
	return rows.Close()
}
`)

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 1 {
		t.Fatalf("unpredicated read: want 1 violation, got %d: %+v", n, got)
	}
	found := false
	for _, v := range got {
		if v.Rule != "SOLE-RLS-ASYNC-READ" {
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

// POSITIVE: the same query WITH an explicit tenant_id predicate must not be
// flagged.
func TestSoleRLSAsyncRead_PredicatedReadNotFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	fixedFile := filepath.Join("internal", "platform", "worker", "fixed_reader.go")
	writeFile(t, dir, fixedFile, `package worker

import (
	"context"
	"database/sql"
)

func handleFixedRead(ctx context.Context, db *sql.DB, tenantID string) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM metaldocs.tenant_lifecycle_jobs WHERE tenant_id = $1 AND status='x'", tenantID)
	if err != nil {
		return err
	}
	return rows.Close()
}
`)

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 0 {
		t.Fatalf("predicated read: want 0 violations, got %d: %+v", n, got)
	}
}

// A read outside any async handler root (e.g. a sync API-path repository)
// must never be flagged — covered by the F3.1 TxRunner chokepoint instead.
func TestSoleRLSAsyncRead_SyncPathOutsideHandlerRootsNeverFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	syncFile := filepath.Join("internal", "modules", "documents", "repository", "sync_repo.go")
	writeFile(t, dir, syncFile, `package repository

import (
	"context"
	"database/sql"
)

func ListJobs(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM tenant_lifecycle_jobs WHERE status='x'")
	if err != nil {
		return err
	}
	return rows.Close()
}
`)

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 0 {
		t.Fatalf("sync-path file outside handler roots: want 0 violations, got %d: %+v", n, got)
	}
}

// A read against a table absent from async-tenant-tables.txt (not
// tenant-scoped / no RLS) must never be flagged even when unpredicated.
func TestSoleRLSAsyncRead_NonTenantTableNeverFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	strayFile := filepath.Join("internal", "platform", "worker", "system_table_reader.go")
	writeFile(t, dir, strayFile, `package worker

import (
	"context"
	"database/sql"
)

func listLeases(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM river_leader WHERE expires_at < now()")
	if err != nil {
		return err
	}
	return rows.Close()
}
`)

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 0 {
		t.Fatalf("non-tenant table: want 0 violations, got %d: %+v", n, got)
	}
}

// An allowlisted site (file:line present in sole-rls-read-allowlist.txt) must
// not be flagged even though it is unpredicated.
func TestSoleRLSAsyncRead_AllowlistedSiteNotFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	allowedFile := filepath.Join("internal", "modules", "render", "fanout", "claim_reader.go")
	writeFile(t, dir, allowedFile, `package fanout

import (
	"context"
	"database/sql"
)

func scanClaims(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id FROM tenant_lifecycle_jobs WHERE status='claimed'")
	if err != nil {
		return err
	}
	return rows.Close()
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "sole-rls-read-allowlist.txt"),
		allowedFile+":9  # sanctioned cross-tenant scan\n")

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ"); n != 0 {
		t.Fatalf("allowlisted site: want 0 violations, got %d: %+v", n, got)
	}
}

// A stale allowlist entry (line no longer matches any live unpredicated read)
// must be reported so the list cannot rot, mirroring
// ASYNC-TENANT-SEED-ALLOWLIST-STALE.
func TestSoleRLSAsyncRead_StaleAllowlistEntryFires(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "async-tenant-tables.txt"), "tenant_lifecycle_jobs\n")

	cleanFile := filepath.Join("internal", "platform", "worker", "clean_reader.go")
	writeFile(t, dir, cleanFile, `package worker

import "context"

func doNothing(ctx context.Context) error {
	return nil
}
`)
	writeFile(t, dir, filepath.Join("scripts", "api-lint", "sole-rls-read-allowlist.txt"),
		cleanFile+":99  # stale — no longer matches any call site\n")

	got, err := checkSoleRLSAsyncRead(dir, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "SOLE-RLS-ASYNC-READ-ALLOWLIST-STALE"); n != 1 {
		t.Fatalf("stale allowlist entry: want 1 violation, got %d: %+v", n, got)
	}
}
