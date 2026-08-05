package main

import (
	"path/filepath"
	"testing"
)

// TestWireNullableOmitempty covers F1.2 Part B's acceptance cases: the Go-
// side guard that closes the gap SHAPE-NULLABLE-NOT-REQUIRED cannot see
// (that rule is spec-only; this one cross-checks the hand-written Go wire
// structs against the spec's nullable-and-required response properties).
//
// specPath is shared across cases: testdata/shape_nullable_required_good.openapi.yaml
// declares Widget.published_version as nullable AND required, reachable from
// the GET /widgets response — exactly the shape this rule hunts for on the
// Go side.
func TestWireNullableOmitempty(t *testing.T) {
	t.Parallel()
	specPath := filepath.Join("testdata", "shape_nullable_required_good.openapi.yaml")

	t.Run("fires_on_reintroduced_drift", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// A hand-written response struct under an http/ path, carrying
		// omitempty on a pointer field whose JSON name is nullable-and-
		// required on Widget — the exact class of bug Part A fixed by hand.
		writeFile(t, dir, "internal/modules/widgets/http/contracts/widget.go", `package contracts

type WidgetResponse struct {
	ID                string  `+"`json:\"id\"`"+`
	PublishedVersion *string `+"`json:\"published_version,omitempty\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if n := countRule(got, "WIRE-NULLABLE-OMITEMPTY"); n != 1 {
			t.Fatalf("want exactly 1 WIRE-NULLABLE-OMITEMPTY violation, got %d\nfull got=%+v", n, got)
		}
	})

	t.Run("clean_tree_zero_after_fix", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// The fixed shape: omitempty dropped, key always present.
		writeFile(t, dir, "internal/modules/widgets/http/contracts/widget.go", `package contracts

type WidgetResponse struct {
	ID                string  `+"`json:\"id\"`"+`
	PublishedVersion *string `+"`json:\"published_version\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("want 0 violations on the fixed shape, got %d\nfull got=%+v", len(got), got)
		}
	})

	t.Run("request_suffix_exempt", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// A request-only struct legitimately keeps omitempty (a client omits
		// an optional field on the way in) — exempted by the "...Request"
		// naming convention, mirroring the spec-side requestOnly exemption
		// SHAPE-NULLABLE-NOT-REQUIRED already applies.
		writeFile(t, dir, "internal/modules/widgets/http/contracts/widget.go", `package contracts

type WidgetCreateRequest struct {
	ID                string  `+"`json:\"id\"`"+`
	PublishedVersion *string `+"`json:\"published_version,omitempty\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("want 0 violations on a *Request struct, got %d\nfull got=%+v", len(got), got)
		}
	})

	t.Run("outside_http_path_not_scanned", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// Same drift shape, but not under an "http" path segment (a domain/
		// application DTO) — out of this rule's documented scope.
		writeFile(t, dir, "internal/modules/widgets/domain/widget.go", `package domain

type WidgetSnapshot struct {
	ID                string  `+"`json:\"id\"`"+`
	PublishedVersion *string `+"`json:\"published_version,omitempty\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("want 0 violations outside an http/ path, got %d\nfull got=%+v", len(got), got)
		}
	})

	t.Run("generated_file_exempt", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// oapi-codegen output already emits the correct tag from the spec by
		// construction; *.gen.go is excluded like every other code-side rule.
		writeFile(t, dir, "internal/modules/widgets/http/api/api.gen.go", `package api

type WidgetResponse struct {
	ID                string  `+"`json:\"id\"`"+`
	PublishedVersion *string `+"`json:\"published_version,omitempty\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("want 0 violations on a generated file, got %d\nfull got=%+v", len(got), got)
		}
	})

	t.Run("non_pointer_field_not_flagged", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// A non-pointer field cannot marshal an explicit null regardless of
		// omitempty, so it is outside this rule's target failure mode.
		writeFile(t, dir, "internal/modules/widgets/http/contracts/widget.go", `package contracts

type WidgetResponse struct {
	ID                string `+"`json:\"id\"`"+`
	PublishedVersion string `+"`json:\"published_version,omitempty\"`"+`
}
`)

		got, err := checkWireNullableOmitempty(specPath, dir)
		if err != nil {
			t.Fatalf("checkWireNullableOmitempty: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("want 0 violations on a non-pointer field, got %d\nfull got=%+v", len(got), got)
		}
	})
}

// TestWireNullableOmitempty_LiveRepoCleanTree proves the real repo's
// post-F1.2-Part-A state is green: the approval module's hand-written wire
// structs (the only ones in the tree) all carry the corrected tags.
func TestWireNullableOmitempty_LiveRepoCleanTree(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	specPath := filepath.Join(repoRoot, "api", "openapi", "v1", "openapi.yaml")
	// modulesRoot is the full repo root, matching the real production
	// invocation (`api-lint -strict <spec> .`) exactly — not a narrower
	// internal/modules-only walk.
	got, err := checkWireNullableOmitempty(specPath, repoRoot)
	if err != nil {
		t.Fatalf("checkWireNullableOmitempty: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 violations on the live repo tree, got %d\nfull got=%+v", len(got), got)
	}
}
