# Feature F3.4 — Spec

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.4-dead-keys`
> **Status:** Approved — 2026-06-16
> **Authored before any code.**

## Consumer contract

**Consumer:** the objectstore key catalog — exported functions declare a contract; if no production
caller exists, the contract is a lie that misleads future readers.

After this feature:
- `TemplateDocxKey` and `TemplateSchemaKey` no longer exist in `internal/platform/objectstore/`.
- `grep -RIn 'TemplateDocxKey\|TemplateSchemaKey' --include='*.go' .` returns 0 production refs
  (test/fixture refs removed with the functions themselves).
- `go build ./...` clean; `go test ./...` green.

No replacement or alignment needed — the live key schema is owned by `TemplatesPresigner`
(separate file in the same package), which uses a different format and has active callers.

## Non-goals

- No change to any other file in `internal/platform/objectstore/`.
- No change to `TemplatesPresigner` or any live key helper.
- No wiki update in this feature — wiki refs are a follow-up for the wiki-curator agent after M3 closes.

## Validation Gate

| # | Criterion | Command / proof |
|---|-----------|-----------------|
| 1 | Functions gone | `grep -RIn 'TemplateDocxKey\|TemplateSchemaKey' --include='*.go' .` → 0 matches |
| 2 | Build clean | `go build ./...` |
| 3 | Tests green | `go test ./...` |

## Pre-spec investigation (R2 risk)

| Question | Finding |
|----------|---------|
| Any production caller (direct)? | No — Grep of all `.go` files returns only definition + own test file. |
| Any string-form reference (reflection / key build)? | No — discovery-brief §E5 + architecture-re-audit 2026-06-15 §8 both confirm zero production callers. Grep of `TemplateDocxKey\|TemplateSchemaKey` across all file types returns docs/wiki only. |
| Is there a live consumer using the same S3 key format? | `TemplatesPresigner` uses a different format — it is an independent, active code path. Deleting these functions does not affect it. |
| R2 verdict | Safe to delete. |
