package main

import (
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoReadOnlyTxOptions_FlagsReadOnlyLiteral proves the guard actually
// fires (RED) on a deliberately-added sql.TxOptions{ReadOnly: true} literal
// before anyone trusts it as the replacement for the retired
// checkAuthzRequireRWTx (A5.2, Lane E issue #92). Covers both the addressed
// (&sql.TxOptions{...}) and value forms, and confirms field order does not
// matter.
func TestNoReadOnlyTxOptions_FlagsReadOnlyLiteral(t *testing.T) {
	dir := t.TempDir()

	// (1) offender: addressed literal, the shape BeginTx actually takes.
	writeFile(t, dir, "internal/modules/x/repository.go",
		"package x\nimport \"database/sql\"\nfunc f(db *sql.DB) { _, _ = db.Begin() ; _ = &sql.TxOptions{ReadOnly: true} }\n")
	// (2) offender: value literal, field order reversed — still must fire.
	writeFile(t, dir, "internal/modules/y/repository.go",
		"package y\nimport \"database/sql\"\nvar opts = sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: true}\n")
	// (3) clean: ReadOnly explicitly false is not a violation.
	writeFile(t, dir, "internal/modules/z/repository.go",
		"package z\nimport \"database/sql\"\nvar opts = sql.TxOptions{ReadOnly: false}\n")
	// (4) clean: no ReadOnly field at all.
	writeFile(t, dir, "internal/modules/w/repository.go",
		"package w\nimport \"database/sql\"\nvar opts = sql.TxOptions{}\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 2 {
		t.Fatalf("no-readonly-tx-options: want exactly 2 violations (files x and y), got %d\nfull got=%+v", n, got)
	}
}

// TestNoReadOnlyTxOptions_AliasedImportFlagged proves the guard resolves the
// file's own import alias for database/sql instead of hard-coding the
// identifier "sql" — a CodeRabbit/Codex finding on PR #120 (2026-08-11):
// import dbsql "database/sql"; &dbsql.TxOptions{ReadOnly: true} silently
// passed the pre-fix guard, reopening the exact read-only-tx gap this rule
// exists to close for any file that aliases the import.
func TestNoReadOnlyTxOptions_AliasedImportFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/modules/x/repository.go",
		"package x\nimport dbsql \"database/sql\"\nfunc f(db *dbsql.DB) { _, _ = db.Begin() ; _ = &dbsql.TxOptions{ReadOnly: true} }\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 1 {
		t.Fatalf("no-readonly-tx-options: want exactly 1 violation for aliased import, got %d\nfull got=%+v", n, got)
	}
}

// TestNoReadOnlyTxOptions_QualifiedAliasFlagged proves the guard follows the
// composite literal's canonical type through an alias exported by another
// package. The package p spelling is deliberately not TxOptions.
func TestNoReadOnlyTxOptions_QualifiedAliasFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module fixture\n\ngo 1.26.5\n")
	writeFile(t, dir, filepath.Join("internal/modules/p", "options.go"),
		"package p\nimport \"database/sql\"\ntype Opts = sql.TxOptions\n")
	writeFile(t, dir, filepath.Join("internal/modules/x", "repository.go"),
		"package x\nimport p \"fixture/internal/modules/p\"\nfunc f() { _ = &p.Opts{ReadOnly: true} }\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 1 {
		t.Fatalf("qualified alias: want exactly 1 violation, got %d\nfull got=%+v", n, got)
	}
}

// TestNoReadOnlyTxOptions_LocalAliasFlagged proves the same canonical-type
// match for a local alias whose source name comes from a dot import.
func TestNoReadOnlyTxOptions_LocalAliasFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/modules/x/repository.go",
		"package x\nimport . \"database/sql\"\ntype Opts = TxOptions\nvar _ = &Opts{ReadOnly: true}\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 1 {
		t.Fatalf("local alias: want exactly 1 violation, got %d\nfull got=%+v", n, got)
	}
}

// TestNoReadOnlyTxOptions_LocalAliasChainFlagged proves the prefilter does not
// require a selector in the composite literal type. A local alias can hide a
// database/sql.TxOptions alias behind more than one package boundary, so the
// ReadOnly field must be enough to send the literal through go/types.
func TestNoReadOnlyTxOptions_LocalAliasChainFlagged(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module fixture\n\ngo 1.26.5\n")
	writeFile(t, dir, filepath.Join("internal/modules/p", "options.go"),
		"package p\nimport \"database/sql\"\ntype Opts = sql.TxOptions\n")
	writeFile(t, dir, filepath.Join("internal/modules/x", "repository.go"),
		"package x\nimport p \"fixture/internal/modules/p\"\ntype LocalOpts = p.Opts\nvar _ = &LocalOpts{ReadOnly: true}\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 1 {
		t.Fatalf("local alias chain: want exactly 1 violation, got %d\nfull got=%+v", n, got)
	}
}

func TestNoReadOnlyTxOptions_UnresolvableFileFailsClosed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "internal/modules/x/repository.go")
	writeFile(t, dir, filepath.Join("internal/modules/x", "repository.go"),
		"package x\nimport missing \"example.invalid/missing\"\nvar _ = missing.Opts{ReadOnly: true}\n")

	_, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("unresolvable candidate: want an error naming %s, got %v", path, err)
	}
}

// TestNoReadOnlyTxOptions_DotImportedDatabaseSQLFlagged names the bare
// TxOptions{ReadOnly: true} AST shape produced by a dot import. The test file
// is intentionally the production-shaped offender; the sibling _test.go
// fixture proves the existing test-file scope exclusion stays silent.
func TestNoReadOnlyTxOptions_DotImportedDatabaseSQLFlagged(t *testing.T) {
	dir := t.TempDir()
	offender := filepath.Join("internal/modules/x", "repository.go")
	excluded := filepath.Join("internal/modules/x", "repository_test.go")
	shadowed := filepath.Join("internal/modules/x", "repository_shadow.go")
	writeFile(t, dir, offender,
		"package x\nimport . \"database/sql\"\nfunc f(db *DB) { _, _ = db.Begin() ; _ = TxOptions{ReadOnly: true} }\n")
	writeFile(t, dir, excluded,
		"package x\nimport . \"database/sql\"\nvar opts = TxOptions{ReadOnly: true}\n")
	writeFile(t, dir, shadowed,
		"package x\nimport . \"database/sql\"\nvar _ = LevelDefault\nfunc g() { type TxOptions struct { ReadOnly bool }; _ = TxOptions{ReadOnly: true} }\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if n := countRule(got, "no-readonly-tx-options"); n != 1 {
		t.Fatalf("dot-imported TxOptions: want exactly 1 violation for %s, got %d\nfull got=%+v", offender, n, got)
	}
	if got[0].File != filepath.Join(dir, offender) {
		t.Fatalf("dot-imported TxOptions: want violation for %s, got %+v", offender, got)
	}
	for _, violation := range got {
		if violation.File == filepath.Join(dir, excluded) {
			t.Fatalf("dot-imported TxOptions: NotWant a violation for excluded fixture %s, got %+v", excluded, violation)
		}
		if violation.File == filepath.Join(dir, shadowed) {
			t.Fatalf("dot-imported TxOptions: NotWant a violation for shadowed fixture %s, got %+v", shadowed, violation)
		}
	}
}

// TestNoReadOnlyTxOptions_TestFileExempt proves _test.go fixtures (e.g. a
// future unit test for a hypothetical read-only helper) are not flagged —
// mirrors every other single-file AST rule's test exemption in this package.
func TestNoReadOnlyTxOptions_TestFileExempt(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/modules/x/repository_test.go",
		"package x\nimport \"database/sql\"\nvar opts = sql.TxOptions{ReadOnly: true}\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("_test.go must be exempt: want 0 violations, got %d\nfull got=%+v", len(got), got)
	}
}

// TestNoReadOnlyTxOptions_CleanTreeZero proves the real repo is green today:
// zero production sql.TxOptions{ReadOnly: true} literals exist anywhere in
// the tree (confirmed by grep before writing this rule).
func TestNoReadOnlyTxOptions_CleanTreeZero(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/platform/db/runner.go",
		"package db\nimport \"database/sql\"\nfunc do(db *sql.DB) { _, _ = db.BeginTx(nil, nil) }\n")

	got, err := checkNoReadOnlyTxOptions(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkNoReadOnlyTxOptions: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("clean tree: want 0 violations, got %d\nfull got=%+v", len(got), got)
	}
}
