package main

import (
	"go/token"
	"testing"
)

// --- authz-require-rw-tx --------------------------------------------------

const rwTxImports = "package pkg\n" +
	"import (\n\t\"context\"\n\t\"database/sql\"\n)\n" +
	"type runner struct{}\n" +
	"func (runner) Do(ctx context.Context, fn func(*sql.Tx) error) error { return fn(nil) }\n" +
	"func (runner) DoReadOnly(ctx context.Context, fn func(*sql.Tx) error) error { return fn(nil) }\n" +
	"var authz = struct{ Require func(context.Context, *sql.Tx, string, string) error }{}\n"

func TestAuthzRequireRWTx_CleanWhenDo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/svc.go", rwTxImports+
		"func f(r runner) error {\n"+
		"\treturn r.Do(nil, func(tx *sql.Tx) error { return authz.Require(nil, tx, \"c\", \"a\") })\n"+
		"}\n")
	got, err := checkAuthzRequireRWTx(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-require-rw-tx"); n != 0 {
		t.Fatalf("Do path: want 0 violations, got %d: %+v", n, got)
	}
}

func TestAuthzRequireRWTx_BitesOnDirectRequire(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/svc.go", rwTxImports+
		"func f(r runner) error {\n"+
		"\treturn r.DoReadOnly(nil, func(tx *sql.Tx) error { return authz.Require(nil, tx, \"c\", \"a\") })\n"+
		"}\n")
	got, err := checkAuthzRequireRWTx(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-require-rw-tx"); n != 1 {
		t.Fatalf("DoReadOnly+authz.Require: want 1 violation, got %d: %+v", n, got)
	}
}

// The require seam (lowercase s.require) must also be caught — name-based scans
// would otherwise miss the tokens-style application-service indirection.
func TestAuthzRequireRWTx_BitesOnRequireSeam(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/svc.go", rwTxImports+
		"type svc struct{ require func(context.Context, *sql.Tx, string, string) error }\n"+
		"func (s svc) f(r runner) error {\n"+
		"\treturn r.DoReadOnly(nil, func(tx *sql.Tx) error { return s.require(nil, tx, \"c\", \"a\") })\n"+
		"}\n")
	got, err := checkAuthzRequireRWTx(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-require-rw-tx"); n != 1 {
		t.Fatalf("DoReadOnly+require seam: want 1 violation, got %d: %+v", n, got)
	}
}

// A DoReadOnly closure that does only pure reads (no require) is legitimate.
func TestAuthzRequireRWTx_CleanPureRead(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pkg/svc.go", rwTxImports+
		"func f(r runner) error {\n"+
		"\treturn r.DoReadOnly(nil, func(tx *sql.Tx) error { _, e := tx.ExecContext(nil, \"SELECT 1\"); return e })\n"+
		"}\n")
	got, err := checkAuthzRequireRWTx(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n := countRule(got, "authz-require-rw-tx"); n != 0 {
		t.Fatalf("pure-read DoReadOnly: want 0 violations, got %d: %+v", n, got)
	}
}
