# Tech Debt Register — templates_v2

> Companion to `wiki/modules/templates_v2.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/templates_v2-refactor.md`.

**Last verified:** 2026-05-12 (Plan 7)

## Items

### T-001 · Authz wired `nil` — every mutation bypasses capability assertion — CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/templates_v2/application/service.go:22` — `WithDB(db *sql.DB) *Service` builder added; `apps/api/cmd/metaldocs-api/main.go` wires `tv2Svc.WithDB(deps.SQLDB)` + real `capabilityService` `AuthzFunc`. `application/create.go:84` (`CreateTemplate`), `application/lifecycle.go:54` (`SubmitForReview`), `:125`/`:148` (`Review`), `:226` (`Approve`), `:345` (`PublishTemplateVersion`), `:406` (`ArchiveTemplate`) — all six mutation paths call `authz.Require` with the appropriate capability when `s.db != nil`. Migration `0188_tripwire_extend.sql:226-233` attaches `trg_require_cap_asserted` to `public.templates_v2_template` and `public.templates_v2_template_version`.
- **Observation (original):** `New(svc, authz)` accepted an `AuthzFunc` argument, then if `authz == nil` substituted a no-op callback. The composition root passed `nil`. None of the seven repo mutations was wrapped in `internal/platform/authz.Require`. No `metaldocs.asserted_caps` GUC tripwire was installed on `templates_v2_*` tables.
- **Evidence:** `_artifacts/02-flow-update-schema.md`, `_artifacts/02-flow-publish.md`, `_artifacts/04-persistence.md` §5 (7 tripwire violations), `_artifacts/03-deps.md` §3.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-001`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md` (decision exists; module deviation is the debt)

### T-002 · Cross-tenant version access — repo getters accept no tenant arg — PARTIALLY CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical → **partially resolved**
- **Surface (resolved):** `internal/modules/templates_v2/application/create.go:148` — `CreateNextVersion` now verifies `source.TemplateID != template.ID` and returns `domain.ErrNotFound` if the loaded version belongs to a different template. This closes the `CreateNextVersion` bypass path.
- **Surface (residual):** `internal/modules/templates_v2/repository/postgres.go` — `GetVersion(template_id, version_number)` and `GetVersionByID(version_id)` still have no `tenant_id` predicate. Callers that front these with `GetTemplate(tenant, template_id)` are safe; any future call site that forgets the gate is still vulnerable.
- **Observation (original):** Both repository getters select by version primary keys with no `tenant_id` predicate. `CreateNextVersion` skipped the tenant gate when cloning from `PublishedVersionID`.
- **Evidence:** `_artifacts/02-flow-update-schema.md`, `_artifacts/04-persistence.md` §1 (`templates_v2_template_version` schema), `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-002`
- **Linked ADR:** missing-ADR

### T-003 · `X-Tenant-ID` header trusted with `DevTenantID` fallback — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** `internal/modules/templates_v2/delivery/http/handler.go:83-84` (`tenantIDFromReq`).
- **Resolution (2026-05-11):** `tenantIDFromReq` now delegates to `tenant.FromContext(r.Context())` (Plan 3 module sweep). Tenant is injected into the request context by the auth middleware from the session-bound `tenant_id` (`auth_sessions.tenant_id`, migration 0184). The `X-Tenant-ID` header is stripped by auth middleware before reaching this handler. The residual risk is that T-001 (nil authz) still allows unauthenticated mutations, but tenant forging via header is closed.
- **Evidence:** `_artifacts/02-flow-list.md`, `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-003` (can be closed)
- **Linked ADR:** `wiki/architecture/tenant-context.md`

### T-004 · `PublishTemplateVersion` bypasses approval lifecycle (no SoD, no role check, no content_hash gate) — PARTIALLY CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical → **partially resolved**
- **Surface (resolved):** `internal/modules/templates_v2/application/lifecycle.go:320` — `content_hash != ""` guard added (`ErrContentHashMismatch` on empty hash). `lifecycle.go:325` — `domain.CheckSegregation("approver", ...)` now called (SoD gate). `lifecycle.go:345` — `authz.Require(CapTemplatePublish)` called inside a new tx when `s.db != nil`.
- **Surface (residual):** `PublishTemplateVersion` still does not verify `pending_approver_role` against `cmd.ActorUserID`'s roles (role-binding check). The `Approve` path enforces this; the direct publish path does not. ISO 9001 §7.5 identity traceability gap remains for `POST /publish`.
- **Observation (original):** Parallel publish path to `Service.Approve`. Transitions a version directly `draft → published` without SoD, role check, or `content_hash` gate.
- **Evidence:** `_artifacts/02-flow-publish.md`, `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-004`
- **Linked ADR:** missing-ADR

### T-005 · Legacy error envelope — RFC 9457 Problem+JSON not adopted — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/templates_v2/delivery/http/handler.go:92-93` (`writeErr`) — body is now `problem.Write(w, problem.New(status, code, message))`. All call sites that use `writeErr` (including `MapErr` consumers at `:109`) emit `application/problem+json`. The legacy `{"error":{"code":"...","message":"..."}}` shape is gone.
- **Observation (original):** Module emitted `{"error":{"code":"...","message":"..."}}` for all non-2xx responses. `internal/platform/problem` existed from Plan 2 but templates_v2 had not been migrated.
- **Evidence:** `_artifacts/05-industry.md` IP-001, `_artifacts/01-surface.md` §1c.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-005` (merged Plan 7 2026-05-11, commit `bbe3933b`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-006 · OpenAPI / handler drift — 12 of 20 routes hand-rolled, not in spec
- **Severity:** major
- **Surface:** `internal/modules/templates_v2/api/api.gen.go:954-961` (8 generated routes) vs `internal/modules/templates_v2/delivery/http/handler.go:39-61` (20 routes registered).
- **Observation:** Twelve routes — `POST .../versions`, `PUT .../schema`, `POST .../{submit,review,approve}`, `POST .../autosave/{presign,commit}`, `POST .../archive`, `PUT .../approval-config`, `GET .../docx-url`, `GET .../audit`, `GET v2/placeholder-catalog` — are mounted by `Handler.Register` but absent from the OpenAPI spec consumed by oapi-codegen. Generated typed surfaces (`*RequestObject`, `*ResponseObject`) cover only 8 routes. Downstream tooling (frontend codegen, contract tests) sees half the surface. Same drift class as `documents` T-002.
- **Evidence:** `_artifacts/01-surface.md` §1a + §1b, `_artifacts/05-industry.md` IP-005.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-006`
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md` (decision exists; module deviation is the debt)

### T-007 · Multi-step publish + obsolete + audit not transactional
- **Severity:** major
- **Surface:** `internal/modules/templates_v2/application/lifecycle.go:265` (`PublishTemplateVersion`); also `Service.Approve` (lifecycle.go:159), `Service.CreateTemplate` (`application/create.go:30`).
- **Observation:** Publish emits 3–5 independent `*sql.DB.ExecContext` calls (`ObsoletePreviousPublished` → `UpdateTemplate` → `UpdateVersion` → `AppendAudit` → `CreateNextVersion`). No `pgx.Tx` wraps the sequence. Repo methods take `context.Context` only — there is no `WithTx(tx)` variant on `Repository`. Partial failure between any two statements leaves DB inconsistent: e.g. previous version marked obsolete but new version not flipped to published, or published flip lands without an audit row. Concurrent publish on the same template has a race window where two versions can briefly co-exist as `published` before `obsolete`-on-next-write resolves it. `AuditObsoleted` constant exists in `domain/audit.go:7` but is never written for the obsolete side-effect.
- **Evidence:** `_artifacts/02-flow-publish.md`.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-007`
- **Linked ADR:** missing-ADR

### T-008 · `ResolverRegistryReader` wired `nil` — `PHComputed.resolver_key` validation skipped
- **Severity:** major
- **Surface:** `internal/modules/templates_v2/application/service.go:11` (`New` constructor accepts variadic `resolvers ResolverRegistryReader`); `apps/api/cmd/metaldocs-api/main.go:328` (does not pass the variadic). Validation site: `internal/modules/templates_v2/application/schema.go:84` (`ValidatePlaceholders`).
- **Observation:** The placeholder catalog gate enforces the 7-token `PHType` enum unconditionally. For `PHType == PHComputed`, the resolver_key string is intended to be checked against `ResolverRegistryReader.HasResolver(key)`. Composition root omits the variadic, leaving the registry reader nil. `ValidatePlaceholders` short-circuits the resolver check when reader is nil. A template author can save a schema with arbitrary `resolver_key` strings; the value propagates into every document instantiated from the published version (per `wiki/modules/documents.md §8.7` snapshot path). Template-injection blast radius is module-wide.
- **Evidence:** `_artifacts/02-flow-update-schema.md`, `_artifacts/03-deps.md` §3.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-008`
- **Linked ADR:** missing-ADR

### T-009 · No `Idempotency-Key` support on POST routes
- **Severity:** major
- **Surface:** `internal/modules/templates_v2/delivery/http/handler.go` (no header parsing); `internal/modules/templates_v2/application/create.go:30` (`CreateTemplate`); `internal/modules/templates_v2/application/lifecycle.go:265` (`PublishTemplateVersion`).
- **Observation:** No route reads `Idempotency-Key`. Replays of `POST /api/v2/templates`, `POST /publish`, `POST /submit`, `POST /approve` either succeed twice (creating duplicate audit rows / duplicate template rows when `(tenant_id, key)` differs) or fail with `ErrInvalidStateTransition` after the first transition lands. `internal/platform/idempotency/` exists from Plan 2 but is not consumed here. Network retries by clients (mobile, browser refresh on slow API) produce inconsistent observable state.
- **Evidence:** `_artifacts/05-industry.md` IP-002.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-009`
- **Linked ADR:** missing-ADR

### T-010 · Optimistic locking field carried but never enforced on autosave
- **Severity:** major
- **Surface:** `internal/modules/templates_v2/application/autosave.go` (`SaveTemplateDraftCmd.ExpectedLockVersion` field) → `internal/modules/templates_v2/repository/postgres.go:236` (`UpdateVersion`).
- **Observation:** Command struct has `ExpectedLockVersion int64`, set by HTTP handler from request body. Service does not compare it against the row's current lock_version before `UpdateVersion`. Repo's UPDATE statement does not include `WHERE lock_version = $X`. Two concurrent autosaves on the same draft from different editors produce last-write-wins silently. No `412 Precondition Failed` ever returns. Eigenpal autosave fires on every keystroke flush; multi-tab editing data loss is observable.
- **Evidence:** `_artifacts/02-flow-update-schema.md`.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-010`
- **Linked ADR:** missing-ADR

### T-011 · `ListTemplates` is unbounded — no LIMIT / OFFSET / cursor
- **Severity:** minor
- **Surface:** `internal/modules/templates_v2/repository/postgres.go:88` (`ListTemplates`).
- **Observation:** Query selects every row from `templates_v2_template` matching the filter without LIMIT or cursor. Single-tenant deployments with low template counts (current production) hide the gap. Latent at multi-tenant scale or a tenant with high template churn. Plan 2 cursor primitive (`feat(pagination): cursor primitive with sort + filter_hash validation`, commit 7effa430) exists but is not consumed here.
- **Evidence:** `_artifacts/02-flow-list.md`, `_artifacts/05-industry.md` IP-003.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · Dead `editable_zones` jsonb column persists despite ADR 0002 zone purge
- **Severity:** minor
- **Surface:** `migrations/0120_templates_v2_init.sql` (column declared NOT NULL); `migrations/0157_drop_editable_zones.sql` (drop migration whose effect on the live DDL is not visible per Phase 4 §6).
- **Observation:** `templates_v2_template_version.editable_zones jsonb NOT NULL` was retained for backward compat at the zone-purge cutover (per ADR 0002 — referenced from `domain/version.go:40` comment "legacy from zone-purge era"). `0157_drop_editable_zones.sql` exists but the column is still present in the table inheritance from migration 0120. Either the drop migration silently no-ops on this lineage, or it was deferred — Phase 4 flagged the drift. Latent: column is written-and-ignored; no read site consumes it. Cleanup deferred.
- **Evidence:** `_artifacts/04-persistence.md` §6.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-012`
- **Linked ADR:** `wiki/decisions/0002-zone-purge.md` (decision exists; column persistence is the residual debt) — missing-ADR for the deferral itself.

### T-013 · Module-local `templates_v2_audit_log` parallel to canonical `metaldocs.audit_events` — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** minor (closed)
- **Surface:** `migrations/0120_templates_v2_init.sql` (CREATE `templates_v2_audit_log`); `internal/modules/templates_v2/repository/postgres.go:318` (`AppendAudit` writes only to the local sink).
- **Observation:** MetalDocs has a canonical audit sink at `metaldocs.audit_events` (per `wiki/modules/audit.md`). templates_v2 writes to a module-local `templates_v2_audit_log` table instead. Two sinks of record means downstream queries (compliance export, forensic timeline) must union both — and a future canonical-sink consumer that does not know about the local sink will silently miss every templates_v2 event. Decision to fork was undocumented.
- **Evidence:** `_artifacts/04-persistence.md` §1, §6.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-013`
- **Linked ADR:** missing-ADR

### T-014 · Exported symbols lack Go doc comments
- **Severity:** minor
- **Surface:** all files under `internal/modules/templates_v2/{domain,application,delivery,repository}/`.
- **Observation:** Per `_artifacts/01-surface.md` §3, every exported type, function, method, and constant in the module is undocumented (no leading `// SymbolName ...` doc comment). `golint` / `revive` exported-rule would flag the module wholesale. Reader of `Service.PublishTemplateVersion` must read the body to learn it skips SoD; reader of `ResolverRegistryReader` must read `schema.go:84` to learn it gates resolver_key. Hexagonal layout itself also lacks an ADR (`domain/application/delivery/repository` split is convention-only, same as `documents` and `auth`).
- **Evidence:** `_artifacts/01-surface.md` §3.
- **Linked backlog row:** `backlog/templates_v2-refactor.md#R-014`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 100% (per T-014)
- Operations missing C4 placement: 0 / 20
- Cross-deps missing in §5/§8: 0 / 17
- State transitions missing in §6: 0 / 9
- Decisions without ADR link: 8 / 8
