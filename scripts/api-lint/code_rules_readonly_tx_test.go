package main

import (
	"go/token"
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
