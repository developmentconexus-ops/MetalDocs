# Phase 0 — Context Load (iam)

**Date:** 2026-05-10
**Module path:** `internal/modules/iam/`
**Existing wiki coverage:** `wiki/modules/iam-rbac.md` (Last verified 2026-05-03) — predecessor stub. Will be superseded by `wiki/modules/iam.md` (broader scope: tenant capabilities + area memberships + authz tiers + admin surface + GUC context + caching).

## Relevant wiki docs (read)

| Doc | Last verified | Why it matters |
|---|---|---|
| `wiki/README.md` | 2026-05-10 | Index format; module entry currently named `iam-rbac.md` line 51 |
| `wiki/modules/iam-rbac.md` | 2026-05-03 | Predecessor; lists capability matrix, migrations 0162–0170, StaticAuthorizer removal |
| `wiki/concepts/authz-tiers.md` | 2026-05-03 | Tier-1 vs tier-2 split, GUC pitfalls, typed errors |
| `wiki/concepts/iso-segregation.md` | 2026-05-04 | SoD overlay (lives in approval module, but IAM is the scaffolding) |
| `wiki/architecture/api-design-system.md` | 2026-05-10 | Two-tier authz + Postgres tripwire = "real enforcer"; 5 CI lint rules |
| `wiki/decisions/0007-two-tier-authz.md` | 2026-05-10 | ADR; J2 amendment (CapabilityChecker adapter); codegen rejected amendment (lint+tripwire) |

## Relevant ADRs

- **0007 — Two-tier authz** (accepted 2026-05-03; amended 2026-05-05 J2; amended 2026-05-10 codegen rejected). Anchors tier-1 `CapabilityService.CanDo` + tier-2 `authz.Require`. Tripwire trigger `enforce_capability_asserted` (migration 0142b lines 138–172) backs both via `metaldocs.asserted_caps` GUC.

No other ADRs touch IAM directly. 0011 (CD atomic create) consumes tier-2 via signoff; 0012 (contract-first API) constrains codegen but does not author authz.

## Industry Patterns Index — relevant rows

| id | topic | relevance |
|---|---|---|
| IP-004 | authz | Defense-in-depth: edge + in-tx + DB constraint → maps 1:1 to tier-1 + tier-2 + tripwire trigger |
| IP-008 | tenancy | row-level `tenant_id` + scoped indexes → `iam_user_roles` Group B fix |
| IP-001 | errors | RFC 9457 Problem envelope → tier-1 returns 403 Problem `metaldocs.authz.forbidden` |
| IP-006 | migrations | forward-only → migrations 0162–0170 ordering |

IP-002 (idempotency), IP-003 (pagination), IP-005 (oapi-codegen), IP-007 (observability) not central to IAM core — may surface only in §11 / admin endpoints / tech-debt.

## Existing file inventory (from Glob)

- `domain/` — 9 files: capabilities.go, context.go, errors.go, model.go, port.go, role_capabilities.go (+ test + integration_test), user_area.go
- `application/` — 9 files: area_membership_service (+ test), authorization.go (+ bench + test), capability_service.go, cached_role_provider.go, dev_role_provider.go, admin_service.go, startup.go
- `area_membership/` — 2 files: area_membership.go (+ test)
- `authz/` — 5 files: authz.go (+ test + bypass_test), context.go (+ test)
- `delivery/http/` — 4 files: admin_handler.go, middleware.go (+ test), routes_memberships.go (+ contract_test)
- `infrastructure/memory/` — role_admin_repository.go
- `infrastructure/postgres/` — 3 files: role_admin_repository.go (+ test), role_provider.go (+ test), user_area_repository.go
- `integration_test.go`

39 source files total (incl. tests). Multi-package layered: `domain` / `application` / `delivery` / `infrastructure` / horizontal `authz` + `area_membership`.

## Open questions (one batch, then proceed)

These are flagged so we can recall later — none block Phase 1.

1. **Doc filename** — Skill prescribes `wiki/modules/iam.md`. Existing wiki entry is `iam-rbac.md` line 51. **Plan:** publish new `iam.md`, retire `iam-rbac.md` (delete file + replace README index entry in Phase 7). Confirm with user before deleting.
2. **`area_membership/` package vs `application/area_membership_service.go`** — two surfaces for area membership. Need Phase 1 surface scan to disambiguate (DTO/value type package vs service?).
3. **`authz/` standalone package** — sibling of `delivery/`, not under `application/`. Whether it counts as a tier-2 "service" or a thin helper — Phase 1 + Phase 2 trace will tell.
4. **Tripwire trigger ownership** — migration `0142b_role_capabilities_v2_enforce.sql` lives in global migrations. Is it owned by `iam` or by `migrations/`? ADR 0007 treats it as IAM's enforcement floor. Phase 4 to confirm.
5. **Admin handler scope** — `delivery/http/admin_handler.go` + `application/admin_service.go` — admin-only role mgmt API. Need its operationIDs in Phase 1 to map authz capability (`user.manage` / `route.manage` from existing matrix).
6. **Tier-3 candidate** — Postgres tripwire is technically a 3rd tier (DB-layer enforcement) on top of tier-1/tier-2. Existing ADR 0007 + concept doc both name only two tiers. Phase 6 may surface this as a documentation tighten (not new debt).

## Coverage scope frozen

`internal/modules/iam/` (in-repo Go module). Cross-module IAM consumers (`approval`, `documents`, `templates`, wiring) appear in §3 (cross-deps) but are NOT documented here — their own module docs own them.

Out of scope:
- Authentication (login, sessions, JWT issuance) — handled by `internal/modules/auth/` (separate module).
- `internal/platform/tenant/` (DevTenantID sentinel) — referenced, not owned.

## Proceeding

Phase 1 dispatch next (Codex surface scan) with module path `internal/modules/iam` and artifact path `wiki/modules/iam/_artifacts/01-surface.md`.
