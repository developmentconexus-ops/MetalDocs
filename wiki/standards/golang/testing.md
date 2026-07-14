---
extends:
  - ~/.claude/rules/ecc/golang/testing.md
evidence:
  - docs/superpowers/plans/2026-05-03-group-b-authz-cleanup.md
  - go.mod
enforced_by:
  - go test -race ./...
  - manual-review
---
# Testing

## No Mock DB Rule

`DATA-DOG/go-sqlmock` is banned for new tests. It remains in `go.mod` for legacy tests, but new persistence and transaction behavior tests use real Postgres through testcontainers or the CI-provided Postgres service.

Reason: prior design plans used sqlmock heavily, and the project now treats mock/real divergence as a backend quality risk. CI enforcement is real Postgres plus `go test -race`; reviewers enforce no new `sqlmock.New` call sites.

## Table-Driven Tests

Use anonymous struct slices with named fields. Failure messages include all inputs.

```go
cases := []struct {
	method string
	path   string
	public bool
}{
	{http.MethodGet, "/api/v1/health/live", true},
}
for _, tc := range cases {
	got := fn(tc.method, tc.path)
	if got != tc.public {
		t.Errorf("fn(%q, %q) = %v, want %v", tc.method, tc.path, got, tc.public)
	}
}
```

Shared assertion helpers call `t.Helper()`.

## Test Naming

Use `Test<FunctionName>_<Scenario>`, for example `TestBeginReplay_EmptyActorReturnsError`.

## testdata Fixtures

SQL fixtures live in package-local `testdata/`. JSON and YAML request bodies also live there.

## No Shared Mutable State

Tests must not depend on execution order. Package-level mutable state is forbidden unless guarded and reset with `t.Cleanup`.

## Race Detection

CI runs `go test -race ./...`. Data races are merge blockers.

## testsupport Package

Shared helpers live under `internal/testsupport/`. Do not add generic test helpers to `internal/platform/` or module packages.

## Context Extractor Tests

For context extractor packages, cover nil/empty context, populated context, whitespace, and trim behavior. Mirror `internal/platform/authn/context_test.go`.
