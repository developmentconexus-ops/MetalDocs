# Feature F8.6 — Evidence (H-D/H-G gate-scope widening + mechanical CI guard)

> **Milestone:** 8  ·  **Feature:** `f8.6-gate-scope-widening`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20, operator-confirmed health exemption). Plan: `plan.md`.
> **Durable decision:** recorded as the **mission §8 scope amendment** (chosen over a standalone ADR).
> **Commit:** recorded at commit time below.

## What was implemented (gate/doc/CI only — no product behavior change)

1. **Mechanical guard — `noresponsemap` analyzer** ([`tools/cilint/internal/analyzers/noresponsemap.go`](../../../../../tools/cilint/internal/analyzers/noresponsemap.go)) —
   flags a `map[string]any` composite literal reaching a 2xx body writer (`writeJSON` / `writeFillInJSON` /
   `WriteJSON`) on any registered-route package (`*/delivery/http/`, `documents/approval/http/`,
   `iam/presence/`, `platform/observability/`). **Laundering-resistant**: a per-function first pass collects
   idents bound to a `map[string]any` literal, so the built-then-written-local pattern
   (`page := map[string]any{...}; writeJSON(w, 200, page)`) — invisible to the historical Grep A — is caught.
   Non-response maps (audit payloads, command FormData, security Evidence, declared-dynamic metrics) never
   reach a writer, so they are not flagged. Registered in `RunAll`, so the existing `invariants.yml` cilint
   job enforces it — no new workflow. Exemptions: `noResponseMapExemptFiles` (health.go) + inline
   `//cilint:allow-responsemap`.
2. **§5b widened** ([`wiki/architecture/api-contract.md`](../../../../../wiki/architecture/api-contract.md)) —
   Part A/B grep path globs widened from `internal/modules/*/delivery/http/` to the full public-route surface;
   mechanical-guard paragraph added; health + F8.2 declared-dynamic-metrics named exemptions added to the
   allowlist; Last-verified bumped.
3. **mission §8 scope amendment** ([`mission.md`](../../../../../docs/superpowers/milestones/grade-a-completion/mission.md)) —
   restates H-D (full surface) and H-G (any cross-module owned-table read), records the exemptions, and points
   the terminal re-audit at the widened greps + the CI guard.
4. **Drive-by (pre-existing, CLAUDE.md §4)** — added the M4c `tests/integration/testdb/` seed harness to the
   `txownership` `allowedTxPackages`. `seedWithCaps` legitimately owns a tx (set_config asserted-caps +
   governed write on one conn) — same seed-harness category as `internal/test/`. Unrelated to H-D; fixed
   because it falsely red-lit the shared cilint gate this guard rides on.

## Consumer contract satisfied

The H-D/H-G class definitions + grep commands now cover **every package that registers a public route**, not
just `internal/modules/*/delivery/http/`. A response-literal `map[string]any` anywhere on the registered
surface fails the gate mechanically. This is the **root-cause fix** for the 4-times-missed Contract/API
dimension: presence/metrics/search evaded every prior path-scoped sweep; the widened scope + AST guard make
the class non-regressable.

## Verification

| Check (spec Validation Gate) | Command / proof | Result | Real vs fixture |
|------------------------------|-----------------|--------|-----------------|
| §5b + §8 cover the full public surface | read amended §5b (widened globs + exemptions) and §8 scope-amendment | done | real |
| Widened Part A at HEAD (post F8.1–F8.5) | literal §5b Part A grep | **= 2 recorded-exempt `health.go` lines only; 0 elsewhere** | real |
| Widened Part B survivors = allowlist only | literal §5b Part B grep | every survivor on allowlist (audit/recordAudit, command FormData, security Evidence, declared-dynamic metrics, 2 comments); **0 response literals** | real |
| Widened H-G at HEAD (post F8.3) | `grep document_profiles` outside taxonomy | only comments + test seed; **0 cross-module reads** | real |
| Mechanical guard = 0 over repo | `go run ./tools/cilint ./internal/...` (noresponsemap) | **0** | real |
| Guard FAILS on a planted violation | planted `writeJSON(w,200,map[string]any{})` in `audit/delivery/http/` | **flagged** `[noresponsemap] … H-D`; clean after removal | real |
| Analyzer unit tests | `go test ./tools/cilint/...` | `ok` (8 noresponsemap cases: direct, laundered local, alias, typed-negative, non-response-negative, exempt-health, allow-directive, out-of-scope) | real |
| Full gate green | `go run ./tools/cilint ./...` | **exit 0, 0 findings** (after testdb drive-by) | real |
| Build | `go build ./...` | exit 0 | — |

## Acceptance vs spec Validation Gate

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| §5b + §8 cover the full public surface | yes | amended sections |
| Widened H-D grep at HEAD = 0 (post F8.1–F8.5) | yes | Part A = exempt-only; mechanical guard = 0 |
| Widened H-G grep at HEAD = 0 (post F8.3) | yes | only comments/test-seed |
| CI guard fails on a planted violation, passes clean | yes | planted-violation proof |

## Review disposition

- Spec-compliance review: PASS — both doc contract (§5b/§8) AND mechanical guard delivered (interview Q2);
  full surface incl. presence/metrics/approval (Q1); health recorded as explicit exemption per operator (Q3).
  Non-goals honored — **no source-behavior change** (analyzer + docs + allowlist only).
- Code-quality review: PASS — AST guard is laundering-resistant (built-then-written locals) and precise
  (semantic = literal→2xx-writer, not a dumb text match that would false-positive on the dozens of legit
  non-response maps); exemptions named in code + docs, not silent; rides the existing CI job.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `noresponsemap` does not follow a map literal laundered through a helper-function return | The two real evasions (alias writers, built-then-written local) are covered; a cross-function return-laundering path has no current instance and would still trip Part B's whole-package `map[string]any` grep at audit time | Trigger: a future audit Part B survivor that is a response literal returned from a helper → extend the analyzer's dataflow |
