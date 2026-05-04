# Authz Tiers

> **Last verified:** 2026-05-03
> **Scope:** Two authorization tiers in MetalDocs — HTTP middleware (tier 1) vs in-transaction area check (tier 2).
> **Out of scope:** Authentication (login/sessions) — see `wiki/references/local-dev-credentials.md`; Role/capability tables — see `wiki/modules/iam-rbac.md`.
> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` — tier-1 `CanDo`
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require`
> - `internal/modules/iam/authz/context.go:13` — `ErrActorContextMissing` / `ErrTenantContextMissing` typed errors; `MustActorID` at :21, `MustTenantID` at :34
> See ADR `wiki/decisions/0007-two-tier-authz.md` for the decision rationale.

MetalDocs has **two authorization tiers**.

## Tier 1 — Tenant Capability

- **Where:** HTTP middleware (`internal/modules/iam/delivery/http/middleware.go`)
- **Service:** `CapabilityService.CanDo(ctx, userID, tenantID, capability)`
- **Tables:** `iam_user_roles` JOIN `role_capabilities`
- **Use:** "Can user X do `doc.create` in tenant T?"
- **Bypass:** `system_admin` role in tenant T

## Tier 2 — Area Grant

- **Where:** service layer, inside transactions (`internal/modules/iam/authz/authz.go`)
- **Service:** `authz.Require(ctx, tx, capability, areaCode)`
- **Tables:** `user_process_areas` JOIN `role_capabilities`
- **Use:** "Can user X sign for area QA-01?"
- **Special:** pass `areaCode = "tenant"` to skip area filter
- **Bypass:** `system_admin` role for the user

## When to use which

- **Route guards** (entry into HTTP handler): tier 1
- **Signoff, approval, area-scoped writes** (inside DB tx): tier 2
- **Both required** for area-scoped actions: middleware passes tier 1, then service layer enforces tier 2

## Common pitfalls

- Forgetting to set `metaldocs.actor_id`/`metaldocs.tenant_id` GUCs before calling `authz.Require` → returns typed sentinel errors `authz.ErrActorContextMissing` or `authz.ErrTenantContextMissing` (defined in `internal/modules/iam/authz/context.go:13`). GUC helpers use `current_setting(..., true)` (`missing_ok=true`), so Postgres does not panic on unset GUCs — the helper returns the typed error instead. Set via `SET LOCAL metaldocs.actor_id = '<userID>'` at start of tx.
- Assigning a tenant role via IAM admin UI does NOT grant area access. Area grants live in `user_process_areas`.
