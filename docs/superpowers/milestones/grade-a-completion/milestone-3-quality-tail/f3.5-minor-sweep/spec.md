# Feature F3.5 — Spec

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.5-minor-sweep`
> **Status:** Approved — 2026-06-16
> **Authored before any code.**

## Consumer contract

**Consumer:** future readers of the touched seams — every §7 Minor on the F3.5 boxed list is
either closed with a cite + commit, or carries a written defer trigger owned by name. Nothing is
silently skipped.

Boxed universe (4 items from report §7):

| Item | Decision | Action |
|------|----------|--------|
| Duplicate constructors `New`/`NewService` at `documents/application/service.go:127` | **Close** | Add `// Deprecated: use NewService.` to `New` |
| Hardcoded PT string `"Criacao do documento"` at `documents/approval/application/submit_service.go:233` | **Defer** | Written trigger: when a locale/i18n system is introduced or when the default title becomes a product-configurable field. Owner: product/backend. |
| 8× `tenantIDFromRequest` private copies (`fillin_handler.go:204` + 7 others per report; currently 3 definition sites at HEAD across `controlleddocuments`, `iam`, `taxonomy` delivery layers) | **Defer** | Written trigger: M4 module-boundaries milestone, where cross-module delivery-layer cleanup is in scope. Owner: M4 author. |
| SHA-1 in `stableID` at `security/application/service.go:179` | **Close** | Replace `sha1.New()` → `sha256.New()`, update import |

## Non-goals

- No i18n infrastructure, no locale config.
- No cross-module extraction of `tenantIDFromRequest` (M4 scope).
- No change to `New` callers (tests keep using it; deprecation comment is the signal).
- No change to the `sig_` ID format or its 16-char truncation.

## Validation Gate

| # | Criterion | Proof |
|---|-----------|-------|
| 1 | `New` has `// Deprecated:` comment | `grep -n 'Deprecated' internal/modules/documents/application/service.go` |
| 2 | `sha1` import gone from `security/application/service.go`; `sha256` used | `grep -n 'sha1\|sha256' internal/modules/security/application/service.go` |
| 3 | Build + tests green | `go build ./...` + `go test ./...` |
| 4 | Close-out table in `evidence.md` has one row per boxed item | inspection |

## Pre-spec investigation

| Item | Finding |
|------|---------|
| `New` callers in production | `grep 'application\.New\b' apps/` → 0 results. Only tests use `New`. |
| `NewService` callers | `apps/api/cmd/metaldocs-api/main.go` (production wiring). |
| `stableID` SHA-1 persistence risk | IDs not persisted server-side; ephemeral per-request. Hash algorithm change doesn't break stored data. |
| `tenantIDFromRequest` count at HEAD | 3 definition sites (4 files). Report had 8 — delta explained by M0–M2 changes and prior refactoring. Cross-module extraction still warranted but is M4 work. |
