# Tech Debt Register — templates

> Companion to `wiki/modules/templates.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/templates-refactor.md`.

**Last verified:** 2026-07-01 (Grade-A simplification register reconciliation — T-002 fully closed, GetVersion/GetVersionByID tenant-scoping residual gap confirmed fixed by commit 042d1504 (2026-05-25), was register drift since 2026-06-12 sync; T-008 closed, SEC-08 fail-fast resolver-registry guard confirmed at main.go:809-816 via commit fb0250e5) | **Prior:** 2026-06-12 (Wave 2.12 sync — db==nil dual-mode branches removed from all application services (single-mode); CreateTemplateTx areas/visibility/specific_areas columns dropped (migration 0236); nodualmode CI guard. Prior Wave 2 sync: typed CapTemplate* consts; upsertApprovalConfig fixed; publish tier-1 aligned; CompositionConfig deleted; AppendAuditTx in-tx. Prior: adversarial verification pass — corrected T-001 tripwire anchor to baseline DDL, added rename note for `templates_v2_*` → `templates_template*` history, fixed T-008 `ValidatePlaceholders` anchor to `schema.go:114`, fixed T-010 test file reference to `routes_contract_test.go:122`, fixed T-014 `schema.go:84` → `schema.go:114`, flagged stale line numbers in `_artifacts/02-flow-publish.md` in T-004 and T-007 evidence; prior: T-007 closed — transaction wrapping confirmed in code; 2026-06-10 Stage-1 backend audit drift patch; 2026-05-31 — fix/templates-publish-role-binding — T-004 fully closed; `PublishTemplateVersion` now enforces `pending_approver_role` binding parity with `Service.Approve`, denied attempts audited via `publish_forbidden_role`)

## Items

### T-001 · Authz wired `nil` — every mutation bypasses capability assertion — CLOSED 2026-05-11 (Plan 5)
- **Severity:** critical (closed)
- **Surface (resolved):** `internal/modules/templates/application/service.go:22` — `WithDB(db *sql.DB) *Service` builder added; `apps/api/cmd/metaldocs-api/main.go` wires `tv2Svc.WithDB(deps.SQLDB)` + real `capabilityService` `AuthzFunc`. `application/create.go`, `application/lifecycle.go`, and `application/autosave.go` now call `authz.Require` with the appropriate capability when `s.db != nil`; the 2026-05-17 repair added `template.edit` assertions around `SaveTemplateDraft` and `CommitAutosave` so DOCX import/autosave commits satisfy the templates table tripwire. The live schema (see **Rename note** below) attaches `trg_require_cap_asserted` to `public.templates_template` and `public.templates_template_version` (`db/baseline/0001_current_schema.sql:3797-3807`).
- **Observation (original):** `New(svc, authz)` accepted an `AuthzFunc` argument, then if `authz == nil` substituted a no-op callback. The composition root passed `nil`. None of the seven repo mutations was wrapped in `internal/platform/authz.Require`. No `metaldocs.asserted_caps` GUC tripwire was installed on `templates_*` tables.
- **Evidence:** `_artifacts/02-flow-update-schema.md`, `_artifacts/02-flow-publish.md`, `_artifacts/04-persistence.md` §5 (7 tripwire violations), `_artifacts/03-deps.md` §3.
- **Rename note:** Archive migrations `0120_templates_init.sql`, `0157_drop_editable_zones.sql`, and `0188_tripwire_extend.sql` all reference `templates_v2_template` and `templates_v2_template_version` — these are the pre-rename table names used during a v2 redesign phase. At some point the tables were renamed to `templates_template` and `templates_template_version`; the rename is reflected in the curated baseline (`db/baseline/0001_current_schema.sql:2188-2214`) and in the live `enforce_capability_asserted` CASE branches (`db/baseline/0001_current_schema.sql:477-486`). Migration 0188 attached the tripwire to the v2 names (`archive/migrations/0188_tripwire_extend.sql:226-234`); the authoritative production trigger bindings are the baseline DDL — readers should treat archive migration DDL for these tables as history, not as the live schema shape. For T-002 through T-013, all table name references to `templates_template` and `templates_template_version` refer to the current production names as defined in the baseline.
- **Linked backlog row:** `backlog/templates-refactor.md#R-001`
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md` (decision exists; module deviation is the debt)

### T-002 · Cross-tenant version access — repo getters accept no tenant arg — CLOSED 2026-07-01 (fully closed; residual gap fixed 2026-05-25, commit 042d1504)
- **Severity:** critical → **resolved**
- **Surface (resolved 2026-05-11 Plan 5):** `internal/modules/templates/application/create.go:148` — `CreateNextVersion` now verifies `source.TemplateID != template.ID` and returns `domain.ErrNotFound` if the loaded version belongs to a different template. This closed the `CreateNextVersion` bypass path.
- **Surface (residual gap, now resolved):** `internal/modules/templates/repository/postgres.go:266-275` (`GetVersion`) and `:287-296` (`GetVersionByID`) both now take a `tenantID` parameter and join `templates_template_version v JOIN templates_template t ON t.id = v.template_id WHERE ... AND t.tenant_id = $N::uuid` — cross-tenant reads by primary key alone are no longer possible. All call sites pass a tenant ID: `application/autosave.go:38,60,84,115`, `application/create.go:126,135`, `application/lifecycle.go:29,117,206,340`, `application/queries.go:25,50`, `application/schema.go:28`, `delivery/http/routes_query.go:97,137,176` — verified 2026-07-01, no caller omits the tenant argument.
- **Observation (original):** Both repository getters selected by version primary keys with no `tenant_id` predicate. `CreateNextVersion` skipped the tenant gate when cloning from `PublishedVersionID`.
- **Resolution:** The residual repo-getter gap was closed by commit `042d1504` ("fix(templates): tenant scope version and audit reads", 2026-05-25) — predates the register's last "partially closed" sync (2026-06-12), so this was register drift rather than an open gap.
- **Evidence:** `internal/modules/templates/repository/postgres.go:266-296`; commit `042d1504`; `_artifacts/02-flow-update-schema.md`, `_artifacts/04-persistence.md` §1 (`templates_template_version` schema), `_artifacts/05-industry.md` IP-008 (may need separate refresh by wiki-curator).
- **Linked backlog row:** `backlog/templates-refactor.md#R-002` (can be closed)
- **Linked ADR:** missing-ADR

### T-003 · `X-Tenant-ID` header trusted with `DevTenantID` fallback — **RESOLVED Plan 3**
- **Severity:** critical → **resolved**
- **Surface:** `internal/modules/templates/delivery/http/handler.go:83-84` (`tenantIDFromReq`).
- **Resolution (2026-05-11):** `tenantIDFromReq` now delegates to `tenant.FromContext(r.Context())` (Plan 3 module sweep). Tenant is injected into the request context by the auth middleware from the session-bound `tenant_id` (`auth_sessions.tenant_id`, migration 0184). The `X-Tenant-ID` header is stripped by auth middleware before reaching this handler. The residual risk is that T-001 (nil authz) still allows unauthenticated mutations, but tenant forging via header is closed.
- **Evidence:** `_artifacts/02-flow-list.md`, `_artifacts/05-industry.md` IP-008.
- **Linked backlog row:** `backlog/templates-refactor.md#R-003` (can be closed)
- **Linked ADR:** `wiki/architecture/tenant-context.md`

### T-004 · `PublishTemplateVersion` bypasses approval lifecycle (no SoD, no role check, no content_hash gate) — CLOSED 2026-05-31 (`fix/templates-publish-role-binding`)
- **Severity:** critical (closed)
- **Surface (resolved 2026-05-11 Plan 5):** `internal/modules/templates/application/lifecycle.go` — `content_hash != ""` guard, `domain.CheckSegregation("approver", ...)` SoD gate, `authz.Require(CapTemplatePublish)` capability assertion inside the same tx.
- **Surface (resolved 2026-05-31):** `lifecycle.go` `PublishTemplateVersion` now calls `version.RoleBindingFor(VersionStatusPublished)` (new helper on `domain/version.go`) and returns `domain.ErrForbiddenRole` → RFC 9457 `code: "forbidden_role"` 403 when the actor's roles do not satisfy `pending_approver_role`. Denied attempts emit `AuditPublishForbiddenRole` to the canonical audit sink. `Service.Approve` was refactored to consume the same helper — single source of truth for role-binding semantics.
- **Observation (original):** Parallel publish path to `Service.Approve`. Transitioned a version directly `draft → published` without SoD, role check, or `content_hash` gate.
- **Evidence:** `_artifacts/02-flow-publish.md` (**note:** line numbers in this artifact are stale — the artifact references `lifecycle.go:265` as the `PublishTemplateVersion` entry point, but the function is declared at `internal/modules/templates/application/lifecycle.go:373`; artifact requires a separate refresh), `_artifacts/05-industry.md` IP-004; `go test ./internal/modules/templates/...` PASS (race build unavailable on this host — no gcc); `internal/modules/templates/application/lifecycle_publish_role_test.go` (3-case table-driven service test); `internal/modules/templates/delivery/http/routes_lifecycle_test.go::TestPublishTemplateVersion_ForbiddenRoleRFC9457`.
- **Linked backlog row:** `backlog/templates-refactor.md#R-004` (closed)
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-005 · Legacy error envelope — RFC 9457 Problem+JSON not adopted — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/templates/delivery/http/handler.go:92-93` (`writeErr`) — body is now `problem.Write(w, problem.New(status, code, message))`. All call sites that use `writeErr` (including `MapErr` consumers at `:109`) emit `application/problem+json`. The legacy `{"error":{"code":"...","message":"..."}}` shape is gone.
- **Observation (original):** Module emitted `{"error":{"code":"...","message":"..."}}` for all non-2xx responses. `internal/platform/problem` existed from Plan 2 but templates had not been migrated.
- **Evidence:** `_artifacts/05-industry.md` IP-001, `_artifacts/01-surface.md` §1c.
- **Linked backlog row:** `backlog/templates-refactor.md#R-005` (merged Plan 7 2026-05-11, commit `bbe3933b`)
- **Linked ADR:** `wiki/architecture/api-design-system.md`

### T-006 · OpenAPI / handler drift — 12 of 20 routes hand-rolled, not in spec — CLOSED 2026-05-16 (Plan 12.4)
- **Severity:** major (closed)
- **Surface (resolved):** `api/openapi/v1/openapi.yaml` and `api/openapi/v1/partials/templates.yaml` include the mounted template route set, including `/api/v1/templates/placeholder-catalog`; `internal/modules/templates/api/api.gen.go` and `frontend/apps/web/src/lib/api-types/index.d.ts` were regenerated. `Handler.Register` mounts generated wrapper methods for the template routes and delegates some generated methods to existing internal handler bodies.
- **Observation (original):** Twelve routes were mounted by `Handler.Register` but absent from the OpenAPI spec consumed by oapi-codegen.
- **Evidence:** Plan 12.4 diff; `scripts/check-module-contract-sync.ps1 -Module templates` PASS; `go generate ./internal/modules/templates/api/...`; `pnpm gen:api`.
- **Linked backlog row:** `backlog/templates-refactor.md#R-006` (closed for route/spec/generated coverage; strict-server cleanup can be tracked separately if desired)
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-007 · Multi-step publish + obsolete + audit not transactional — CLOSED (date unknown; confirmed resolved 2026-06-11)
- **Severity:** major (closed)
- **Surface (resolved):**
  - `internal/modules/templates/application/lifecycle.go:442-480` (`PublishTemplateVersion`) — `BeginTx` at line 442, all mutations via `*Tx` repo variants (`ObsoletePreviousPublishedTx`, `UpdateTemplateTx` ×2, `UpdateVersionTx`, `CreateVersionTx`, `AppendAuditTx`), `Commit` at line 478. No `db != nil` guard — `PublishTemplateVersion` requires a live DB and unconditionally uses the transaction path.
  - `internal/modules/templates/application/lifecycle.go:265-319` (`Approve` accept branch) — `BeginTx` at line 266, `Commit` at line 296, same `*Tx` repo variants. Falls back to non-Tx variants only when `s.db == nil` (unit-test mode with no real DB).
  - `internal/modules/templates/application/create.go:62-126` (`CreateTemplate`) — `BeginTx` at line 63, `Commit` at line 98, same pattern.
- **Observation (original):** Publish and Approve emitted 3–5 independent `*sql.DB.ExecContext` calls with no wrapping transaction. Repo methods accepted only `context.Context` — no `WithTx` variant existed. Partial failures left DB inconsistent.
- **Resolution:** `*Tx` variants added to `Repository`; all three multi-step mutation paths (`PublishTemplateVersion`, `Approve` accept, `CreateTemplate`) now wrap their statements in a single `*sql.Tx`. The `AuditObsoleted` emission gap (original observation) is not separately tracked here — see `_artifacts/02-flow-publish.md` for current audit coverage.
- **Evidence:** `internal/modules/templates/application/lifecycle.go:442-480` (`PublishTemplateVersion` tx — `BeginTx` at 442, `Commit` at 478), `lifecycle.go:265-296` (`Approve` accept branch — `BeginTx` at 266, `Commit` at 296), `internal/modules/templates/application/create.go:62-98` (create tx). Note: `_artifacts/02-flow-publish.md` cites `lifecycle.go:265` as the `PublishTemplateVersion` entry — this is incorrect; that line is inside the `Approve` accept branch. The artifact's line numbers are stale and require a separate refresh.
- **Linked backlog row:** `backlog/templates-refactor.md#R-007` (can be closed)
- **Linked ADR:** missing-ADR

### T-008 · `ResolverRegistryReader` wired `nil` — `PHComputed.resolver_key` validation skipped — CLOSED 2026-07-01 (fail-fast guard, commit fb0250e5, SEC-08)
- **Severity:** major (closed)
- **Surface (resolved):** `apps/api/cmd/metaldocs-api/main.go:798-823` (`buildTemplatesModule`) — a dedicated `templatesResolverReg := resolvers.NewRegistry()` is built and populated via `resolvers.RegisterBuiltins(...)` (`:807-808`), then a fail-fast guard at `:814-816` returns a startup error (`"templates resolver registry is nil or empty; resolver_key validation would be silently skipped (SEC-08 / T-008)"`) if the registry is nil or empty; the non-nil registry is passed into `templatesapp.New(..., templatesResolverReg)` at `:818`, satisfying the variadic `resolvers ResolverRegistryReader` parameter (`internal/modules/templates/application/service.go:11`). Validation site unchanged: `internal/modules/templates/application/schema.go:114` (`ValidatePlaceholders`).
- **Observation (original):** The placeholder catalog gate enforces the 7-token `PHType` enum unconditionally. For `PHType == PHComputed`, the resolver_key string is intended to be checked against `ResolverRegistryReader.HasResolver(key)`. The register originally claimed the composition root omitted the variadic, leaving the registry reader nil and silently skipping the check.
- **Resolution:** Commit `fb0250e5` ("fix(security/SEC-08,SEC-12,SEC-07): fail-fast resolver registry + honest 501s in memory mode") notes the registry wiring itself was already present and the original T-008 claim was stale; it adds the explicit boot-time fail-fast guard so a nil/empty registry now hard-stops startup instead of silently permitting arbitrary `resolver_key` values to propagate into published templates.
- **Evidence:** `apps/api/cmd/metaldocs-api/main.go:809-816`; commit `fb0250e5`; `_artifacts/02-flow-update-schema.md`, `_artifacts/03-deps.md` §3 (original observation, now superseded).
- **Linked backlog row:** `backlog/templates-refactor.md#R-008` (can be closed)
- **Linked ADR:** missing-ADR

### T-009 · Idempotency replay coverage incomplete on generated POST routes
- **Severity:** major
- **Surface:** generated mutation wrappers under `internal/modules/templates/delivery/http/routes_generated.go` and `h.idempotent(...)` coverage; application mutations in `internal/modules/templates/application/create.go` and `internal/modules/templates/application/lifecycle.go`.
- **Observation:** Plan 12.4 requires `Idempotency-Key` in the OpenAPI contract for `POST /api/v1/templates`, routes the wizard create path through the generated typed wrapper, and verifies HTTP 201 create with the header present. Replay semantics still need a focused audit across generated POST mutations (`/templates`, `/publish`, `/submit`, `/review`, `/approve`) to prove same-key retries return the first result and do not duplicate audit/state changes.
- **Evidence:** `_artifacts/05-industry.md` IP-002; Plan 12.4 runtime smoke for `POST /api/v1/templates`; generated API contract requiring `Idempotency-Key` on create.
- **Linked backlog row:** `backlog/templates-refactor.md#R-009`
- **Linked ADR:** missing-ADR

### T-010 · Optimistic locking field carried but never enforced on autosave — CLOSED 2026-05-31 (fix/templates-schema-occ-lock)
- **Severity:** major (closed)
- **Surface (resolved):**
  - `internal/modules/templates/application/autosave.go` (`SaveTemplateDraft`) calls `UpdateVersionDraftCAS` / `UpdateVersionDraftCASTx` (closed 2026-05-17).
  - `internal/modules/templates/application/schema.go` (`UpdateSchemas`) now accepts `ExpectedLockVersion` and calls `UpdateVersionSchemaCAS` / `UpdateVersionSchemaCASTx` (closed 2026-05-31).
  - `internal/modules/templates/repository/postgres.go` adds `UpdateVersionSchemaCAS`/`Tx` mirroring the draft CAS pattern: `WHERE id = $1 AND lock_version = $2` returning `ErrStaleLockVersion` on miss.
  - `api/openapi/v1/openapi.yaml` PUT `/api/v1/templates/{id}/versions/{n}/schema` request body now requires `expected_lock_version` (legacy `expected_content_hash` removed from the schema-PUT contract). `VersionDTO.lock_version` exposed.
  - `internal/modules/templates/delivery/http/routes_schema.go` rejects requests missing `expected_lock_version` with 400 and surfaces 412 `stale_lock_version` on CAS miss.
  - `frontend/apps/web/src/features/templates/hooks/useTemplateSchemas.ts` holds lockVersion, sends `expected_lock_version`, raises `staleConflict` on 412 instead of silent overwrite; `refetch` clears it.
- **Surface (residual):** legacy `/autosave/commit` carries only `expected_content_hash`, not `expected_lock_version`; it is hash-gated and tripwire-protected, but not multi-tab lock-version protected. Tracked separately (not under T-010 — the original unverified-field gap is closed for every route that takes a `lock_version` field).
- **Observation:** Schema PUT was hard-coding `expected_content_hash: ''` from the FE; the server treated empty as "skip CAS", so two tabs editing placeholders concurrently last-write-wins with no audit signal. Switching that path to lock_version CAS closes the silent overwrite class for the schema editor. Eigenpal's DOCX import path still uses `/autosave/commit` and is content-hash gated — separate cleanup outside this branch.
- **Evidence:** `_artifacts/02-flow-update-schema.md`; `TestUpdateSchemas_StaleLockVersion`; `internal/modules/templates/delivery/http/routes_contract_test.go:122` (`TestUpdateTemplateSchema_StaleLockVersion_412`); `useTemplateSchemas` vitest stale-conflict spec; preview drive (two-tab editor) — second tab surfaces stale-lock alert, first save survives until explicit refetch.
- **Linked backlog row:** `backlog/templates-refactor.md#R-010` (close)
- **Linked ADR:** missing-ADR

### T-011 · `ListTemplates` is unbounded — no LIMIT / OFFSET / cursor
- **Severity:** minor
- **Surface:** `internal/modules/templates/repository/postgres.go:114-127` (`ListTemplates`).
- **Observation:** Query applies `LIMIT $3 OFFSET $4` populated from `ListFilter.Limit` / `ListFilter.Offset`. LIMIT/OFFSET pagination is present; keyset/cursor pagination is not. Plan 2 cursor primitive (`feat(pagination): cursor primitive with sort + filter_hash validation`, commit 7effa430) exists but is not consumed here. At large offset values performance degrades (full scan to skip rows). Severity remains minor.
- **Evidence:** `_artifacts/02-flow-list.md`, `_artifacts/05-industry.md` IP-003.
- **Linked backlog row:** `backlog/templates-refactor.md#R-011`
- **Linked ADR:** missing-ADR

### T-012 · Dead `editable_zones` jsonb column persists despite ADR 0002 zone purge
- **Severity:** minor
- **Surface:** `migrations/0120_templates_init.sql` (column declared NOT NULL); `migrations/0157_drop_editable_zones.sql` (drop migration whose effect on the live DDL is not visible per Phase 4 §6).
- **Observation:** `templates_template_version.editable_zones jsonb NOT NULL` was retained for backward compat at the zone-purge cutover (per ADR 0002 — referenced from `domain/version.go:40` comment "legacy from zone-purge era"). `0157_drop_editable_zones.sql` exists but the column is still present in the table inheritance from migration 0120. Either the drop migration silently no-ops on this lineage, or it was deferred — Phase 4 flagged the drift. Latent: column is written-and-ignored; no read site consumes it. Cleanup deferred.
- **Evidence:** `_artifacts/04-persistence.md` §6.
- **Linked backlog row:** `backlog/templates-refactor.md#R-012`
- **Linked ADR:** `wiki/decisions/0002-zone-purge.md` (decision exists; column persistence is the residual debt) — missing-ADR for the deferral itself.

### T-013 · Module-local `templates_audit_log` parallel to canonical `metaldocs.audit_events` — WRITE path CLOSED 2026-05-11 (Plan 6a); READ path divergence unacknowledged
- **Severity:** minor (write path closed; read/write sink divergence is a residual gap)
- **Surface:** `migrations/0120_templates_init.sql` (CREATE `templates_audit_log`); `internal/modules/templates/repository/postgres.go:631-639` (`AppendAudit` now delegates to `r.audit.Record` — canonical `metaldocs.audit_events` via `auditdomain.Writer`); `internal/modules/templates/repository/postgres.go:676` (`ListAudit` still reads from `templates_audit_log`).
- **Observation:** Plan 6a closed the write side: `AppendAudit` delegates to the canonical `auditdomain.Writer` (injected via `WithAudit`). However, `ListAudit` (`repository/postgres.go:676`) still reads from `templates_audit_log`. New audit events written after Plan 6a land in `metaldocs.audit_events`; historical events and the `GET /api/v1/templates/{id}/audit` read path still serve from the old local table. The two sinks are now diverged in direction (writes: canonical; reads: local) rather than being unified. Compliance queries consuming `ListAudit` see only the pre-Plan-6a history. Decision to leave the read path on the legacy sink was not documented.
- **Evidence:** `_artifacts/04-persistence.md` §1, §6; `repository/postgres.go:631-639` (write path); `repository/postgres.go:676-683` (read path).
- **Linked backlog row:** `backlog/templates-refactor.md#R-013`
- **Linked ADR:** missing-ADR

### T-014 · Exported symbols lack Go doc comments
- **Severity:** minor
- **Surface:** all files under `internal/modules/templates/{domain,application,delivery,repository}/`.
- **Observation:** Per `_artifacts/01-surface.md` §3, every exported type, function, method, and constant in the module is undocumented (no leading `// SymbolName ...` doc comment). `golint` / `revive` exported-rule would flag the module wholesale. Reader of `Service.PublishTemplateVersion` must read the body to learn it skips SoD; reader of `ResolverRegistryReader` must read `schema.go:114` (`ValidatePlaceholders`) to learn it gates resolver_key. Hexagonal layout itself also lacks an ADR (`domain/application/delivery/repository` split is convention-only, same as `documents` and `auth`).
- **Evidence:** `_artifacts/01-surface.md` §3.
- **Linked backlog row:** `backlog/templates-refactor.md#R-014`
- **Linked ADR:** missing-ADR

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: 100% (per T-014)
- Operations missing C4 placement: 0 / 22
- Cross-deps missing in §5/§8: 0 / 17
- State transitions missing in §6: 0 / 9
- Decisions without ADR link: 8 / 8
