package main

import (
	"go/token"
	"testing"
)

// TestPaginationCodec_FlagsStdEncodingOutsideCursor is the Family 2 · B2 guard:
// any non-generated source naming base64.StdEncoding is a violation, while the
// canonical cursor.go and generated *.gen.go files (oapi-codegen gzip) are
// exempt. This keeps the three list endpoints on one URL-safe cursor dialect.
func TestPaginationCodec_FlagsStdEncodingOutsideCursor(t *testing.T) {
	dir := t.TempDir()

	// (1) offender: a regular module file reaching for the std dialect.
	writeFile(t, dir, "internal/modules/x/cursor_helper.go",
		"package x\nimport \"encoding/base64\"\nvar _ = base64.StdEncoding.EncodeToString(nil)\n")
	// (2) exempt: the canonical codec file (allow-listed defensively).
	writeFile(t, dir, "internal/platform/pagination/cursor.go",
		"package pagination\nimport \"encoding/base64\"\nvar _ = base64.StdEncoding\n")
	// (3) exempt: generated code legitimately gunzips embedded specs with std.
	writeFile(t, dir, "internal/modules/x/api/api.gen.go",
		"package api\nimport \"encoding/base64\"\nvar _ = base64.StdEncoding.DecodeString\n")
	// (4) clean: the URL-safe dialect is fine anywhere.
	writeFile(t, dir, "internal/modules/x/ok.go",
		"package x\nimport \"encoding/base64\"\nvar _ = base64.RawURLEncoding.EncodeToString(nil)\n")

	got, err := checkPaginationCodec(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkPaginationCodec: %v", err)
	}
	if n := countRule(got, "pagination-codec"); n != 1 {
		t.Fatalf("pagination-codec: want exactly 1 violation (only the module helper), got %d\nfull got=%+v", n, got)
	}
}

// TestPaginationCodec_CleanTreeZero proves the real repo's post-B2 state is
// green: with cursor.go on RawURLEncoding and only *.gen.go using std, the rule
// returns nothing.
func TestPaginationCodec_CleanTreeZero(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "internal/platform/pagination/cursor.go",
		"package pagination\nimport \"encoding/base64\"\nvar _ = base64.RawURLEncoding\n")
	writeFile(t, dir, "internal/modules/x/api/api.gen.go",
		"package api\nimport \"encoding/base64\"\nvar _ = base64.StdEncoding.DecodeString\n")

	got, err := checkPaginationCodec(dir, token.NewFileSet())
	if err != nil {
		t.Fatalf("checkPaginationCodec: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("clean tree: want 0 violations, got %d\nfull got=%+v", len(got), got)
	}
}
