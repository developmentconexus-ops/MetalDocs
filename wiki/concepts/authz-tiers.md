# Authz Tiers

> **Last verified:** 2026-05-11
> **Scope:** Two authorization tiers in MetalDocs — HTTP middleware (tier 1) vs in-transaction area check (tier 2).
> **Out of scope:** Authentication (login/sessions) — see `wiki/references/local-dev-credentials.md`; Role/capability tables — see `wiki/modules/iam.md`.
> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` — tier-1 `CanDo`
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require`
> - `internal/modules/iam/authz/context.go:13` — `ErrActorContextMissing` / `ErrTenantContextMissing` typed errors; `MustActorID` at :21, `MustTenantID` at :34
> - `migrations/0188_tripwire_extend.sql:18` — extended `enforce_capability_asserted()` function covering 12 tables (Plan 5); trigger attachment at lines :186–233
> - `internal/modules/iam/domain/model.go:15` — `Capability` typed consts (20 total including `CapRegistryObsolete`/`CapRegistrySupersede` added Plan 5)
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

## Modules with tier-2 coverage (as of Plan 5)

| Module | tier-2 call sites | tripwire tables |
|---|---|---|
| documents | `CreateDocumentTx`, `UpdateDocumentName`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive` | `public.documents` (INSERT + UPDATE) |
| approval | `submit_service` (doc.submit), `signoff_service` (doc.signoff) | `approval_instances`, `approval_signoffs` |
| registry | `Create`, `CreateTx` (registry.create); `changeStatus` (registry.obsolete\|supersede) | `controlled_documents`, `cd_sequence_counters` |
| taxonomy | `FamilyRepository.Create/Update`, `ProfileRepository.Create/Update`, `AreaRepository.Create/Update` | `document_profiles`, `document_process_areas`, `document_families` |
| templates | `CreateTemplate`, `SubmitForReview`, `Review`, `Approve`, `PublishTemplateVersion`, `ArchiveTemplate` | `templates_template`, `templates_template_version` |
| iam | `UpsertUserAndAssignRole`, `ReplaceUserRoles` (user.manage); `Insert`, `CloseActive`, `GrantAtomic` (membership.manage) | `iam_user_roles`, `user_process_areas` |

## Common pitfalls

- Forgetting to set `metaldocs.actor_id`/`metaldocs.tenant_id` GUCs before calling `authz.Require` → returns typed sentinel errors `authz.ErrActorContextMissing` or `authz.ErrTenantContextMissing` (defined in `internal/modules/iam/authz/context.go:13`). GUC helpers use `current_setting(..., true)` (`missing_ok=true`), so Postgres does not panic on unset GUCs — the helper returns the typed error instead. Set via `SET LOCAL metaldocs.actor_id = '<userID>'` at start of tx.
- Assigning a tenant role via IAM admin UI does NOT grant area access. Area grants live in `user_process_areas`.

## See also

- [`wiki/modules/auth.md`](../modules/auth.md) — canonical auth module doc; §8.1 covers the session-enforcement layer that is upstream of both authz tiers; middleware at `internal/modules/auth/delivery/http/middleware.go:47` injects `iamdomain.WithAuthContext` so tier-1 and tier-2 checks have an actor in context
- [`wiki/modules/iam.md`](../modules/iam.md) — full Arc42 + C4 doc for `internal/modules/iam`; two live authz tiers documented in §8.1 (AuthorizationService deleted in Plan 4 — T-003 closed; Plan 5 expanded tier-2 to all IAM-owned write surfaces)
- [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) — ADR rationale for the two-tier design
