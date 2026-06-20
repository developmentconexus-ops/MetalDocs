# Feature F9.4 — Evidence

> **Milestone:** M9  ·  **Feature:** `f9.4-noresponsemap-widen`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` — the `noresponsemap` analyzer flags any `map[string]<T>` response literal
> reaching a 2xx writer; §5b documents the widened scope.

## What was implemented

- `noresponsemap.go`: `isMapStringLiteral` now matches any `map[string]<T>` (key=="string", any value
  type). Scope, health.go exemption, and `//cilint:allow-responsemap` unchanged. Finding message + header
  doc updated to "map[string]<T>".
- Widening surfaced a 4th site — `finalizeDocument` — closed via `DocumentFinalizeResult {instance_id}`
  (no suppression). Test fakes' instance ids switched to real uuids.
- `api-contract.md` §5b widened (rule, Part A/B greps, mechanical-guard text, anti-evasion). Commit `2e3c8a8b`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — positive | `go test ./tools/cilint/internal/analyzers/ -run NoResponseMap` | `TestNoResponseMap_Positive_MapStringString` flags a `map[string]string` at a writer; PASS | real |
| TDD — negative | same | `TestNoResponseMap_Negative_NonResponseMapStringString` (map not reaching writer) PASS; typed/allow/exempt cases stay green | real |
| Class-closure demonstrated | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` pre-finalize-fix | exited 1 at `handler.go:636 finalizeDocument` (the widened gate caught the 4th `map[string]string`) | real |
| Static (build) | `GOFLAGS=-mod=mod go build ./...` | exit 0 | — |
| Gate (post-fix) | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | exit 0 — full repo, widened scope | real |
| Full suite | `GOFLAGS=-mod=mod go test ./...` | no failures | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `map[string]string` at a writer is flagged | yes | positive test |
| Non-response / exempt / typed not flagged | yes | negative + existing tests |
| Repo cilint exits 0 only after sites typed | yes | class-closure + post-fix gate rows |
| §5b documents widened scope | yes | api-contract.md §5b diff |

## Review disposition

- Spec-compliance review: PASS — widened to the class (any value type), scope/exemptions preserved.
- Code-quality review: PASS — 4th site closed by typing, not by adding a response-shaped exemption (§5b
  forbids that); single helper rename, callers updated.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
