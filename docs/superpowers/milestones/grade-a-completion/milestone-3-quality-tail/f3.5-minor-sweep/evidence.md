# Feature F3.5 — Evidence

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.5-minor-sweep`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md`.

## §7 Minor close-out table

| § 7 item | Location | Decision | Evidence / defer trigger |
|----------|----------|----------|--------------------------|
| Duplicate constructors `New`/`NewService` | `documents/application/service.go:127` | **Closed** | `// Deprecated: use NewService.` added. Commit `24843242`. |
| Hardcoded PT string `"Criacao do documento"` | `documents/approval/application/submit_service.go:233` | **Deferred** | Trigger: when locale/i18n system introduced or default title becomes product-configurable. Owner: product/backend. |
| 8× `tenantIDFromRequest` private copies | `controlleddocuments/delivery/http/routes.go:489`, `iam/delivery/http/routes_memberships.go:336`, `taxonomy/delivery/http/routes_profiles.go:260` (3 definition sites at HEAD) | **Deferred** | Trigger: M4 module-boundaries milestone — cross-module delivery-layer cleanup in scope there. Owner: M4 author. |
| SHA-1 dedup in `stableID` | `security/application/service.go:179` | **Closed** | `sha1.New()` → `sha256.New()`; `crypto/sha1` import removed. Commit `24843242`. |

Nothing silently skipped — all 4 boxed items have a row.

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| Gate 1: `Deprecated` comment on `New` | `grep -n 'Deprecated' internal/modules/documents/application/service.go` | `:127: // Deprecated: use NewService.` | — |
| Gate 2: sha1 gone, sha256 present | `grep -n 'sha1\|sha256' internal/modules/security/application/service.go` | `:14 "crypto/sha256"` `:180 sha256.New()` | — |
| Gate 3: build clean | `go build ./...` | `BUILD OK` | — |
| Gate 3: tests (force-fresh) | `go test -count=1 ./internal/modules/documents/application/... ./internal/modules/security/application/...` | both `ok` | fixture |

## Bounded defers

| Defer | Trigger | Owner |
|-------|---------|-------|
| Hardcoded PT string `"Criacao do documento"` | Locale/i18n system introduction or product-configurable default title | product/backend |
| `tenantIDFromRequest` cross-module extraction | M4 module-boundaries milestone | M4 author |
