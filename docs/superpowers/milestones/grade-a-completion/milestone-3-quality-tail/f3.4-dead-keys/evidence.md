# Feature F3.4 — Evidence

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.4-dead-keys`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.

## What was implemented

Deleted `internal/platform/objectstore/template_keys.go` (two exported functions: `TemplateDocxKey`,
`TemplateSchemaKey`) and `internal/platform/objectstore/template_keys_test.go` (their only tests).
Zero callers updated — none existed.

Commit: `07d8424e fix(f3.4): remove dead TemplateDocxKey/TemplateSchemaKey exports (E5)`

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| Gate 1: 0 Go refs | `grep -RIn 'TemplateDocxKey\|TemplateSchemaKey' --include='*.go' .` (via Grep tool) | No files found | — |
| Gate 2: build clean | `go build ./...` | `BUILD OK` | — |
| Gate 3: objectstore tests | `go test -count=1 ./internal/platform/objectstore/...` | `ok` | fixture |

## Acceptance vs spec Validation Gate

| # | Criterion | Met? | Evidence |
|---|-----------|------|----------|
| 1 | Functions gone, 0 Go refs | yes | Grep returns no files |
| 2 | Build clean | yes | `go build ./...` OK |
| 3 | Tests green | yes | objectstore package PASS |

## Bounded defers

| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| Wiki docs still reference `template_keys.go` (`wiki/backend/platform/data-layer.md`, `wiki/backend/_artifacts/`) | Wiki updates are post-M3 wiki-curator sweep — not a correctness gap | wiki-curator dispatch after M3 milestone-validator PASS |
