# Authz Tiers

> **Last verified:** 2026-06-29 (added "read-write tx" invariant to Common pitfalls — `authz.Require` may INSERT a bypass audit, so `DoReadOnly` is banned on authz-gated paths; new api-lint `authz-require-rw-tx` rule; 5 DoReadOnly sites fixed) | **Prior:** 2026-06-13 (Z-28 ADR 0022 Phase 6 wiki sync — capability model fully executed: typed const registry enforced by CI lint (`no-rawstring-capability`, `no-rolestring-in-delivery`, `authz-area-scope-binding`, `seed-registry-parity`, `wiki-capability-parity`); runtime scope binding complete (area-grade caps pass real areaCode, `"tenant"` sentinel banned for area-grade by AST guard); area_membership governance now in-tx via `LogTx` (Z-6, T-007 closed); registry at 29 caps (ADR 0022 Phase 10 minimization); `BypassSystem` fail-closed background-only bridge (CWE-269); Phase 13 CI net revived — blocking gate `0 blocking, 397 reported` on clean tree) | **Prior:** 2026-06-11 (adversarial-fix: tripwire anchor corrected to db/migrations/0231; View-cap enumeration corrected 4→6, added CapDocumentView + CapTemplateView) | **Prior:** 2026-06-10 (Stage-1 backend audit drift patch: capability count 27→29, line anchor :15→:88)
> **Scope:** Two authorization tiers in MetalDocs — HTTP middleware (tier 1) vs in-transaction area check (tier 2).
> **Out of scope:** Authentication (login/sessions) — see `wiki/references/local-dev-credentials.md`; Role/capability tables — see `wiki/modules/iam.md`.
> **Key files:**
> - `internal/modules/iam/application/capability_service.go:31` — tier-1 `CanDo`
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require`
> - `internal/modules/iam/authz/context.go:13` — `ErrActorContextMissing` / `ErrTenantContextMissing` typed errors; `MustActorID` at :21, `MustTenantID` at :34
> - `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql:36` — `enforce_capability_asserted()` function covering 12 tables (Plan 5 + hardening); trigger attachment follows inline in the same migration
> - `internal/modules/iam/domain/model.go:88` — `Capability` typed consts (29 total; ADR 0016 added `CapMetricsView`, `CapMembershipView`, `CapUserView`, `CapTaxonomyView`; Plan 5 follow-up added `CapControlledDocumentObsolete`/`CapControlledDocumentSupersede`; ADR 0022 Phase 10 minimized 33→29)
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
- **Special:** pass `areaCode = "tenant"` to skip area filter (only valid for tenant-grade caps; area-grade caps are banned from passing `"tenant"` by the `authz-area-scope-binding` CI guard)
- **Bypass:** `system_admin` role for the user (R1 — tenant-wide inheritance, no per-area row required; every bypass is audit-logged via `BypassAuditSink`)
- **Background bypass:** `BypassSystem` (background-only, fail-closed — returns `ErrBypassNotBackground` if the context lacks `WithBackgroundBypass`; HTTP request contexts can never reach it — CWE-269)

## When to use which

- **Route guards** (entry into HTTP handler): tier 1
- **Signoff, approval, area-scoped writes** (inside DB tx): tier 2
- **Both required** for area-scoped actions: middleware passes tier 1, then service layer enforces tier 2

## Capability scope classification (ADR 0022 Phases 2 + 7 + 8 + 12)

Each of the 29 registry capabilities is classified **tenant-grade** (`ScopeTenant`) or **area-grade** (`ScopeArea`) in `internal/modules/iam/domain/capability_scope.go`. This classification is CI-enforced: the `authz-area-scope-binding` AST guard (`scripts/api-lint`) bans `authz.Require(<areaGradeCap>, "tenant")` — any area-grade cap passed with the `"tenant"` literal is a red build. The `no-rawstring-capability` guard bans raw-string cap arguments to `authz.Require`; only typed consts are permitted.

**Area-grade (11):** `document.create`, `document.edit`, `document.submit`, `document.signoff`, `document.publish`, `document.obsolete`, `document.supersede`, `controlled_documents.create`, `controlled_documents.obsolete`, `controlled_documents.supersede`, `cap:membership.manage`.

**Tenant-grade (18):** all `*.view` caps, `template.*` lifecycle caps, `taxonomy.manage`, `user.manage`, `route.manage`, `metrics.view`, `audit.read`, `session.manage`.

## Modules with tier-2 coverage

| Module | tier-2 call sites | tripwire tables |
|---|---|---|
| documents | `CreateDocumentTx` (`cap:document.create`); `UpdateDocumentName`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive` (`cap:document.edit`); `LoadDocumentAreaCode` is the shared DB-derived area helper (one source of truth — `documents/application/document_area.go`) | `public.documents` (INSERT + UPDATE) |
| approval | `submit_service` (`cap:document.submit`), `signoff_service` (`cap:document.signoff`), `publish_service` (`cap:document.publish`), `obsolete_service` (`cap:document.obsolete`) | `approval_instances`, `approval_signoffs` |
| controlled-documents | `Create`, `CreateTx` (`cap:controlled_documents.create`); `changeStatus` (`cap:controlled_documents.obsolete`\|`cap:controlled_documents.supersede`) — area loaded from DB row before the check | `controlled_documents`, `cd_sequence_counters` |
| taxonomy | `FamilyRepository.Create/Update`, `ProfileRepository.Create/Update`, `AreaRepository.Create/Update` | `document_profiles`, `document_process_areas`, `document_families` |
| templates | `CreateTemplate`, `SubmitForReview`, `Review`, `Approve`, `PublishTemplateVersion`, `ArchiveTemplate` — all tenant-grade | `templates_template`, `templates_template_version` |
| iam | `UpsertUserAndAssignRole`, `ReplaceUserRoles` (`cap:user.manage`); `InsertTx`, `CloseActiveTx`, `GrantAtomicTx` (`cap:membership.manage`) — membership governance writes via `LogTx` in the same tx (Z-6, T-007 closed) | `iam_user_roles`, `user_process_areas` |

> **Capability reference convention (ADR 0022 Phase 5):** an enforcement claim —
> a statement that a capability is actually checked — is written `` `cap:<name>` ``
> (e.g. `` `cap:membership.manage` ``). `scripts/api-lint`'s `wiki-capability-parity`
> rule binds every `cap:` marker in this doc to the Go registry (`validCapabilities`),
> so a name that does not exist (a future `membership.grant`-style drift) is a red
> build. Illustrative or deferred mentions (e.g. `doc.create` above, `route.view`
> below) are unmarked prose and intentionally not checked.

## Tier-1 rule authoring rules

Source: F-001 audit (`wiki/references/qa-runs/plans-f001-f002.md`). The Tier-1 declarative table at `apps/api/cmd/metaldocs-api/permissions.go` is hand-authored. The following invariants apply to every new row and are CI-enforced by `TestPermissionsTable_NoMethodlessWriteShadowing` (`apps/api/cmd/metaldocs-api/permissions_test.go`).

1. **Never declare a write-grade cap on GET.** GET rows must point at a View-grade cap (`*.view`, `audit.read`). Write-grade caps (`*.manage`, `*.submit`, `*.create`, `*.edit`, `*.approve`, `*.publish`, `*.signoff`, `*.obsolete`, `*.supersede`, `*.review`) must NEVER appear on a GET row.
2. **Never omit `method` on a prefix that has any write verbs.** A methodless row matches every HTTP method (rules scan top-down, first match wins) and silently shadows any per-verb intent declared elsewhere for the same prefix. Always declare per-verb rows on writable prefixes.
3. **Every writable resource declares at least one read row and at least one write row.** The IAM users block at `permissions.go:101-116` is the canonical pattern: one `{method: GET, ..., cap: <readCap>}` plus one row per write verb with the Manage/Submit cap.
4. **Read caps live in the registry.** Do NOT invent caps inline. If a needed View-grade cap is missing from `internal/modules/iam/domain/model.go`, STOP and file an ADR (precedent: ADR 0016). The six current View caps are `CapDocumentView`, `CapTemplateView`, `CapTaxonomyView`, `CapMembershipView`, `CapUserView`, and `CapMetricsView` (all defined at `internal/modules/iam/domain/model.go:88–119`). A fifth — `CapRouteView` — is proposed for the approval route catalogue read path in [ADR 0018](../decisions/0018-approval-route-lifecycle.md) §6 and deferred to the F-001 follow-up.
5. **Tier-2 read calls match Tier-1 caps.** When a repository read calls `authz.Require(ctx, tx, <cap>, "tenant")`, the cap MUST match the Tier-1 GET row's cap for the same resource. Mismatch causes a 403 inside the handler after a 200 from the middleware — silent partial denial.

## Common pitfalls

- Forgetting to set `metaldocs.actor_id`/`metaldocs.tenant_id` GUCs before calling `authz.Require` → returns typed sentinel errors `authz.ErrActorContextMissing` or `authz.ErrTenantContextMissing` (defined in `internal/modules/iam/authz/context.go:13`). GUC helpers use `current_setting(..., true)` (`missing_ok=true`), so Postgres does not panic on unset GUCs — the helper returns the typed error instead. Set via `SET LOCAL metaldocs.actor_id = '<userID>'` at start of tx.
- Assigning a tenant role via IAM admin UI does NOT grant area access. Area grants live in `user_process_areas`.
- **`authz.Require` MUST run in a read-WRITE transaction — never `DoReadOnly`.** The `system_admin`/`BypassSystem` short-circuit audits the bypass in-tx via an `INSERT` into `audit_events` (ADR 0022 Phase 11 F8, fail-closed). A Postgres `READ ONLY` tx rejects that INSERT, so an authz-gated read opened with `DoReadOnly` works for ordinary actors but 500s the moment the caller is a `system_admin` (latent, actor-dependent). Open authz-gated reads with `TxRunner.Do` (or `BeginTx(ctx, nil)`). The `SeedTxIdentity` GUCs are RO-safe (`set_config(..., true)`), but the Require grant path is not. Enforced statically by the api-lint `authz-require-rw-tx` rule (`scripts/api-lint/code_rules.go`): no `DoReadOnly` closure may call a tier-2 require. (Surfaced 2026-06-29 by the tokens dictionary read 500; the 5 deviating sites — tokens Get/GetByName/List, documents view_service + reconstruct_service — were switched to `Do`.)

## See also

- [`wiki/modules/auth.md`](../modules/auth.md) — canonical auth module doc; §8.1 covers the session-enforcement layer that is upstream of both authz tiers; middleware at `internal/modules/auth/delivery/http/middleware.go:47` injects `iamdomain.WithAuthContext` so tier-1 and tier-2 checks have an actor in context
- [`wiki/modules/iam.md`](../modules/iam.md) — full Arc42 + C4 doc for `internal/modules/iam`; two live authz tiers documented in §8.1 (AuthorizationService deleted in Plan 4 — T-003 closed; Plan 5 expanded tier-2 to all IAM-owned write surfaces)
- [`wiki/decisions/0007-two-tier-authz.md`](../decisions/0007-two-tier-authz.md) — ADR rationale for the two-tier design
- [`wiki/concepts/approval-routes.md`](approval-routes.md) — plain-language overview of the approval route catalogue (Tier-1 read split / `CapRouteView` deferral cross-referenced from §"Tier-1 rule authoring rules" rule 4)
- [`wiki/decisions/0018-approval-route-lifecycle.md`](../decisions/0018-approval-route-lifecycle.md) — route lifecycle ADR; §6 specifies the deferred Tier-1 `route.view` capability
