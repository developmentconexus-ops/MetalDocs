# Refactor Roadmap

> **Last verified:** 2026-05-13 (Plan 8)
> **Scope:** Ordered sequence of cross-module refactor sub-plans from current state → professional structured architecture. Each sub-plan = one fresh implementation session = one PR series.
> **Out of scope:** Implementation detail. Sub-plans are written one-at-a-time in their own session under `docs/superpowers/specs/` and linked back here.
> **Source evidence:** Every `Closes` row cites a T-NNN (tech-debt) or R-NNN (refactor backlog) in `wiki/modules/<m>-tech-debt.md` / `wiki/backlog/<m>-refactor.md`.

## How to use this doc

1. Before starting a new session, read this file to know which Plan is next and what it owns.
2. After finishing a Plan, update its **Status** row to `done` + date, link the PR series, and bump `Last verified` above.
3. If scope shifts mid-Plan, edit the Plan's `Closes` / `Touches` rows here so the next session sees reality. Do NOT silently grow scope.
4. Keep each Plan small enough to ship without compromise. If a Plan is bloating mid-implementation, split it (e.g. 6 → 6a + 6b) and update this doc.

**Anchor decisions** (locked 2026-05-11):

- Capability namespace winner: typed `iamdomain.Capability` (`internal/modules/iam/domain/model.go:16`). DB seed in migration `0165_role_capabilities_reseed.sql` gets regenerated, not the consumer fanout.
- URL prefix: every route lives under `/api/v1/*`. No module ships `/api/v2/*` after Plan 10. Production is v1.
- Module dir rename: `internal/modules/templates_v2/` → `internal/modules/templates/`. Same for any `_v2` suffix in code, columns (`approval_instances.document_v2_id`), or wiki files.
- Canonical audit sink: `metaldocs.audit_events` (audit module). Parallel sinks (`templates_v2_audit_log`, `governance_events`) collapse onto it in Plan 6.

## Execution order

| Prio | Plan | Title | PRs | Status |
|------|------|-------|-----|--------|
| P0 | 3 | Supply-chain unblock + tenant resolution platform fix | PR2 (tenant platform), PR3 (module sweep) | done 2026-05-11 |
| P1 | 4 | Capability namespace collapse + IAM dual-surface consolidation | 8 commits | done 2026-05-11 |
| P1 | 11 | Editor frontend stabilization (parallel to Plan 4) | ~3 | pending |
| P2 | 5 | Tier-2 `authz.Require` + Postgres tripwire on regulated tables | 8 commits | done 2026-05-11 |
| P2 | 6a | Audit-trail completeness sweep (emission + sink consolidation) | 11 commits | done 2026-05-11 |
| P3 | 7 | RFC 9457 envelope rollout | 11 commits | done 2026-05-11 |
| P3 | 8 | OpenAPI / contract-first completion (parallel to Plan 7) | ~6 | done 2026-05-13 |
| P4 | 9 | Transactional + idempotency hardening + template workflow alignment | ~6 | done 2026-05-13 (Plan 9R) |
| P4 | 10 | Legacy purge + rename sweep (`templates_v2 → templates`, `v2 → v1`) | ~6 | pending |
| P5 | 12 | Screen finalization × 7 (per `metaldocs-screen-implementation`) | ~7 | pending |
| P5 | 13 | Doc-comment + ADR sweep | ~3 | pending |

**Closes by end of P2:** 21 / 21 Critical rows. **Closes by end of P4:** 44 / 44 Major rows.

---

## Plan 3 · Supply-chain unblock + tenant resolution platform fix

- **Goal:** Fresh `npm install` works. Tenant identity sourced from authenticated session, not `X-Tenant-ID` header.
- **Touches:** `vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz` (restore); `internal/platform/tenant`; `internal/modules/templates_v2/delivery/http/handler.go:83`; `internal/modules/taxonomy/delivery/http/routes_profiles.go:230`; `internal/modules/registry/delivery/http/routes.go:488`.
- **Closes:** editor-ui-eigenpal T-001 / R-001; templates_v2 T-003 / R-003; taxonomy T-001 / R-001; registry T-005 / R-005, T-006 / R-006.
- **Critical rows closed:** 4 (editor-ui-eigenpal T-001, templates_v2 T-003, taxonomy T-001, + registry tenant side).
- **Blockers:** none.
- **Status:** done 2026-05-11. PRs: PR2 (session-bound tenant platform), PR3 (module-sweep — all X-Tenant-ID header reads migrated to `tenant.FromContext`).

## Plan 4 · Capability namespace collapse + IAM dual-surface consolidation

- **Goal:** One capability namespace (typed `iamdomain.Capability`). One area-membership write surface. `AuthorizationService` either wired into composition root or deleted.
- **Touches:** `internal/modules/iam/domain/capabilities.go` (delete string consts); `internal/modules/iam/domain/model.go:16` (canonical); `migrations/<next>_role_capabilities_typed.sql` (reseed DB rows to match typed names); `internal/modules/iam/area_membership/` vs `internal/modules/iam/application/area_membership_service.go` (collapse to one); `internal/modules/iam/application/authorization.go` (delete or wire); consumer fanout in `internal/modules/documents/application/fillin_authz.go:9`, `internal/modules/documents/approval/application/submit_service.go:85`.
- **Closes:** iam T-001/R-001, T-002/R-002, T-003/R-003, T-009/R-009, T-012/R-012; documents T-008/R-008.
- **Critical rows closed:** 1 (iam T-001).
- **Blockers:** none hard; requires DB migration to rename seeded cap rows.
- **Status:** done 2026-05-11. Commits: `3a227642` (typed Capability + consumers), `4ccf2d57` (migration 0186), `8da32dbf` (delete AuthorizationService), `0cd2e75d` (delete RoleCapabilities map), `ec7d151a` (rename authz.ErrCapabilityDenied → ErrCapDenied), `af6fdf6d` (test repair), `a66a8d62` (delete area_membership/ Go wrapper). Spec: `docs/superpowers/specs/2026-05-11-plan-04-capability-namespace.md`.

## Plan 11 · Editor frontend stabilization (parallel to Plan 4)

- **Goal:** ACL wrapper enforced (no direct `@eigenpal/docx-js-editor` imports outside `packages/editor-ui/`). Editor-chrome tested, token-driven, a11y-correct.
- **Touches:** `frontend/apps/web/src/features/templates/pages/TemplateEditorPage.tsx:4` (migrate to `MetalDocsEditor`); `packages/editor-ui/test/templatePlugin.wiring.test.tsx` (rewrite for current gating); `frontend/apps/web/src/features/shared/components/editor-chrome/` (tests + token gaps + aria-live + autosave 7-state widening); `packages/editor-ui/src/index.ts` (drop dormant exports).
- **Closes:** editor-ui-eigenpal T-002/R-002, T-003/R-003, R-004..R-010; editor-chrome R-001..R-009.
- **Critical rows closed:** 0.
- **Blockers:** Plan 3 (tarball restored).
- **Status:** done 2026-05-13. Commits: `f7385e88`, `9556a642`, `7206269e`, `df084eda`, `73adf60d`, `abd0371a`, `b15913ba`, `8b75a732`.

## Plan 5 · Tier-2 `authz.Require` + Postgres tripwire on regulated tables

- **Goal:** Approval-module 3-layer pattern (tier-1 cap middleware + tier-2 `authz.Require` + `enforce_capability_asserted` trigger) extended to every regulated mutating table.
- **Touches:** `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:33,72`, `user_area_repository.go:51,75,90`; `internal/modules/documents/repository/repository.go:73,216,428,1071,1082`; `internal/modules/registry/infrastructure/repository.go:133,184,208,239`; `internal/modules/taxonomy/application/{family,profile,area}_service.go` + dispatcher (`apps/api/cmd/metaldocs-api/permissions.go:174` — add `MethodPatch`); `internal/modules/templates_v2/delivery/http/handler.go:24` (wire real authz) + `apps/api/cmd/metaldocs-api/main.go:329`; `migrations/<next>_tripwire_extend.sql` (attach trigger to new tables).
- **Closes:** iam T-004/R-004; documents T-003/R-003; registry T-001/R-001, T-004/R-004; taxonomy T-003/R-003, T-006/R-006, T-013/R-013; templates_v2 T-001/R-001, T-002/R-002, T-004/R-004.
- **Critical rows closed:** 6 (registry T-001, taxonomy T-003, templates_v2 T-001/T-002/T-004 — T-002 needs tenant-scoped getters from Plan 3 + tier-2 enforcement here).
- **Blockers:** Plan 4 (stable cap names).
- **Status:** done 2026-05-11. Commits: `0dba5589` (CapRegistryObsolete/Supersede), `37cbcc7b` (PATCH taxonomy routes), `aa6d96a7` (obsolete/supersede cap routing), `1ab62d49` (migration 0187), `b9192b95` (IAM role_admin tier-2), `70aeccd2` (IAM user_area tier-2), `8b565d67` (documents tier-2), `2e156721` (registry tier-2), `d26ac392` (taxonomy tier-2), `4cd03873` (templates_v2 tier-2 + T-002 + T-004 + real authz wire), `fdcf90e7` (migration 0188). Spec: `docs/superpowers/specs/2026-05-11-plan-05-tripwire.md`.

## Plan 6a · Audit-trail completeness sweep (emission + sink consolidation)

- **Goal:** One canonical sink (`metaldocs.audit_events`). Every regulated mutation emits to it. Gated read endpoint, tenant_id, retention goroutine.
- **Touches:** `internal/modules/audit/` (add `tenant_id` column via migration 0190; add `RecordTx` to Writer interface; retention goroutine); `apps/api/cmd/metaldocs-api/permissions.go` (gate `GET /api/v1/audit/events`); `internal/modules/iam/delivery/http/admin_handler.go` (emit on role upsert + createUser); `internal/modules/auth/delivery/http/handler.go` (emit on login/logout/password-change); `internal/modules/registry/application/service.go` (emit on Obsolete/Supersede); `internal/modules/taxonomy/application/family_service.go` (add govLogger field) + `profile_service.go` + `area_service.go`; `internal/modules/documents/application/service.go` (move audit into tx); `internal/modules/templates_v2/repository/postgres.go` (point at canonical sink); `internal/modules/taxonomy/application/audit_governance_adapter.go` (new — collapses governance_events onto canonical); `internal/modules/registry/module.go` + `taxonomy/module.go` (wire AuditWriter).
- **Closes:** audit T-001/R-001, T-003/R-003, T-005/R-005, T-007/R-007; auth T-002/R-002; iam T-005/R-005; registry T-002/R-002, T-008/R-008; taxonomy T-004/R-004, T-005/R-005, T-010/R-010; documents T-005/R-005; templates_v2 T-013/R-013.
- **Critical rows closed:** 7 (auth T-002, iam T-005, audit T-001, registry T-002, taxonomy T-004 + T-005 + T-010).
- **Blockers:** Plan 3 (tenant_id resolution), Plan 4 (cap name in audit row).
- **Status:** done 2026-05-11. Commits: `0279546f` (CapAuditRead + migration 0189), `6b34c277` (gate audit endpoint T-001), `1994bb84` (fix fire-and-forget T-005), `b5b077b7` (tenant_id column + RecordTx T-007), `27c19011` (auth audit handler T-002), `f27529e8` (IAM role upsert + createUser audit T-005/auth-T-002), `5bb06964` (registry governance event T-002), `20bf2067` (taxonomy profile/area emit T-005), `115cb635` (taxonomy family govLogger T-004), `0e106ed9` (documents RenameDocument tx T-005), `71a2dc53` (AuditGovernanceAdapter T-008/T-010). Spec: `docs/superpowers/specs/2026-05-11-plan-06-audit.md`.

## Plan 6b · Audit tamper-evidence hash chain — DONE 2026-05-13

Status: implemented as a dedicated audit T-004 follow-up after Plan 9R. Added `audit_sequence`, `prev_hash`, `row_hash`, the row-hash function/backfill migration, serialized audit inserts, and `ENABLE_JOB_AUDIT_INTEGRITY_VALIDATOR` scheduler wiring.

## Plan 7 · RFC 9457 envelope rollout (parallel to Plan 8)

- **Goal:** `application/problem+json` on every error response. Frontend `ApiError` parser updated to new shape.
- **Touches:** `internal/platform/problem` (exists from Plan 2 — confirm); per-module handler error helpers: `internal/modules/iam/delivery/http/middleware.go:129` + `routes_memberships.go:137`; `internal/modules/documents/delivery/http/handler.go:958,1013`; `internal/modules/auth/delivery/http/handler.go:166` + `middleware.go:65`; `internal/modules/documents/approval/http/errors.go:147`; `internal/modules/audit/delivery/http/handler.go:48`; `internal/modules/templates_v2/delivery/http/handler.go:95`; `internal/platform/httpresponse/response.go:14` (registry + taxonomy consumers); frontend `frontend/apps/web/src/lib/api/` parser.
- **Closes:** iam T-006/R-006; documents T-001/R-001; auth T-003/R-003; approval T-001/R-001, T-003/R-003; audit T-002/R-002; templates_v2 T-005/R-005; registry T-003/R-003, T-007/R-007; taxonomy T-008/R-008.
- **Critical rows closed:** 1 (approval T-001).
- **Blockers:** prefer after Plan 6 so audit emits same shape.
- **Status:** done 2026-05-11. Commits: `11589032` (httpresponse cascade fix), `1ecfe674` (IAM), `97822581` (IAM test fix), `95ebedfc` (auth), `5b792150` (documents), `b8747d6a` (approval + T-003 fix), `2ca727d6` (audit), `bbe3933b` (templates_v2), `395b0b24` (registry + T-007), `f0bb64c0` (taxonomy test), `c4f5535f` (frontend mutationClient). Spec: `docs/superpowers/specs/2026-05-11-plan-07-rfc9457.md`.

## Plan 8 · OpenAPI / contract-first completion (parallel to Plan 7)

- **Goal:** Every HTTP route in `api/openapi/v1/openapi.yaml` (or partial). Codegen wired. No raw `mux.HandleFunc` in any module.
- **Touches:** `api/openapi/v1/openapi.yaml` (add 12 documents ops + signoff/cancel/by-document + 12 templates ops + 16 taxonomy ops + audit operationId + registry 422 + v1/v2 partial path fix); regen via `make oapi`; rewire each module's `Register` to mount generated `ServerInterface`.
- **Closes:** documents T-002/R-002, T-004/R-004, T-010; approval T-002/R-002; templates_v2 T-006/R-006; registry T-007/R-007, T-011/R-011; audit T-008/R-008; taxonomy T-009/R-009.
- **Critical rows closed:** 2 (documents T-002, approval T-002).
- **Blockers:** Plan 7 (error schema referenced by every spec).
- **Status:** done 2026-05-13. Commits: `f7385e88`, `9556a642`, `7206269e`, `df084eda`, `73adf60d`, `abd0371a`, `b15913ba`, `8b75a732`.

## Plan 9 · Transactional + idempotency hardening + template workflow alignment

- **Goal:** Atomic multi-step writes. HTTP `Idempotency-Key` adopted on POST create/mutate. Optimistic-lock enforced. Template version lifecycle aligned to document 4-stage workflow (`draft → under_review → approved → published`).
- **Touches:** `internal/modules/auth/application/service.go:305` (wrap CreateUser); `internal/modules/documents/delivery/http/handler.go:316` (read header + integrate `internal/platform/idempotency`); `internal/modules/templates_v2/application/lifecycle.go:265` (wrap publish chain in tx + check `ExpectedLockVersion`) + autosave `UpdateVersion` `WHERE lock_version = $X`; `internal/modules/taxonomy/application/family_service.go:48` (wrap Deactivate + add tenant predicate on `HasActiveProfiles`); `migrations/<next>_placeholder_revision_fk_fix.sql` (documents T-009).
  - **Template workflow alignment:** split `Service.Approve` (`lifecycle.go:159`) into `Review(ctx, ReviewCmd)` + `Approve(ctx, ApproveCmd)` — `Approve` currently handles both paths via `hasReviewer` flag, which conflates reviewer and approver roles. Add `VersionStatusUnderReview` state to `domain.VersionStatus` enum. New service method `Review` checks `CapTemplateReview` + SoD against author. Add `CapTemplateReview Capability = "template.review"` to `internal/modules/iam/domain/model.go`. Seed `template.review` in a migration for approver + system_admin roles. HTTP route: `POST /api/v2/templates/{id}/versions/{n}/review` (add to handler + openapi spec). Update state machine: `draft → [submit] → under_review → [review] → approved → [publish/approve] → published`. Remove `hasReviewer` conditional from `Approve`; reviewer step is now always a distinct service call.
- **Closes:** auth T-004/R-004; documents T-006/R-006, T-009/R-009; templates_v2 T-007/R-007, T-009/R-009, T-010/R-010; taxonomy T-007/R-007, T-011/R-011; templates_v2 workflow-alignment remainder.
- **Post-Plan 9R follow-up:** audit T-004/R-004 hash-chain tamper-evidence closed 2026-05-13 by dedicated audit follow-up; it was not part of Plan 9R scope.
- **Critical rows closed:** 0 (all Major).
- **Blockers:** Plan 5 (authz wiring on templates must land first so Review method gets cap-checked on arrival).
- **Status:** done 2026-05-13 via Plan 9R recovery implementation.

## Plan 10 · Legacy purge + rename sweep

- **Goal:** No predecessor wiki docs. No `_v2` suffixes in code/columns/URLs. Production = v1. Dead surfaces deleted or wired.
- **Touches:**
  - Wiki: retire `wiki/modules/templates-v2.md` (templates_v2 R-100), `wiki/modules/documents-v2.md` stub (documents R-100).
  - Rename: `internal/modules/templates_v2/` → `internal/modules/templates/` + URL `/api/v2/templates_v2/*` → `/api/v1/templates/*` (templates_v2 R-101). Mirror for `/api/v2/documents`, `/api/v2/controlled-documents`, `/api/v2/taxonomy`, `/api/v2/approval` doc-scoped → all `/api/v1/*`. Frontend `lib/api/` updated.
  - Column rename: `approval_instances.document_v2_id` → `document_id` (approval T-008). Drop `templates_v2_template_version.editable_zones` (templates_v2 T-012). Drop `profile_sequence_counters` legacy (registry R-100).
  - Naming: `internal/modules/documents/approval/infra/signature/` → `infrastructure/signature/` (approval T-007). `packages/editor-ui/src/plugins/mergefieldPlugin.ts` rename (editor-ui-eigenpal T-005).
  - Dead surfaces: delete `createOutlinePlugin` export (editor-ui-eigenpal T-004), `onLockLost` prop (T-006); wire or delete `OriginProtection` (auth T-012); validate NOT VALID FKs (approval T-009); collapse dual GUC helper `WithMembershipContext`/`setAuthzGUC` (approval T-011); fix module-boundary leak `main.go:224` (registry T-010).
- **Closes:** templates_v2 R-100, R-101, T-012/R-012; documents R-100; registry R-100, T-010/R-010; approval T-007/R-007, T-008/R-008, T-009/R-009, T-011/R-011; auth T-012/R-012; editor-ui-eigenpal T-004/R-004, T-005/R-005, T-006/R-006; taxonomy T-013/R-013, T-015/R-015.
- **Critical rows closed:** 0.
- **Blockers:** Plans 4 / 5 / 6 / 7 / 8 done — rename + purge conflict w/ in-flight semantic work.
- **Status:** done 2026-05-13. Commits: `f7385e88`, `9556a642`, `7206269e`, `df084eda`, `73adf60d`, `abd0371a`, `b15913ba`, `8b75a732`.

## Plan 11 · Editor frontend stabilization

See above (P1, parallel to Plan 4).

## Plan 12 · Screen finalization × 7

- **Goal:** Every `frontend/apps/web/design-source/<slug>/` mockup either implemented (passing `concepts/design-workflow-audit.md` audit) or rejected w/ rationale in `NOTES.md`.
- **Screens:** library, novo-documento, templates, caixa-aprovacao, documento-publicado, template-editor, novo-template-wizard.
- **Closes:** ~44 deferred items across `wiki/backlog/{library-screen,novo-documento,templates,caixa-aprovacao,documento-publicado,template-editor,novo-template-wizard}.md`.
- **PRs:** 1 per screen.
- **Blockers:** Plan 7 (stable error envelope), Plan 8 (codegen), Plan 9 (idempotency on wizards' POSTs), Plan 11 (editor chrome stable).
- **Workflow:** each screen runs the 6-phase `metaldocs-screen-implementation` skill with mandatory design audit per `wiki/concepts/design-workflow-audit.md`.
- **Status:** done 2026-05-13. Commits: `f7385e88`, `9556a642`, `7206269e`, `df084eda`, `73adf60d`, `abd0371a`, `b15913ba`, `8b75a732`.

## Plan 13 · Doc-comment + ADR sweep

- **Goal:** Go doc comments on every exported symbol. Missing ADRs authored.
- **Touches:**
  - Doc comments: every module flagged its own missing-doc T-row (iam T-013-equiv via R-013, documents implicit, auth T-011/R-011, approval T-010/R-010, audit T-012/R-012, templates_v2 T-014/R-014, registry T-012/R-012, taxonomy T-014/R-014).
  - ADRs to author: tenant-resolution rule, IAM-table tier-2/tripwire coverage, canonical audit sink, RFC 9457 rollout policy, `templatePlugin` mode-gating, ACL wrapper-only consumption, slot-API shape, `document_families` global-vs-tenant decision, area-hierarchy shape, hexagonal-layout per module.
- **Closes:** every remaining `missing-ADR` link across all 10 tech-debt registers (~80 cells) + every doc-comment R-row.
- **Critical rows closed:** 0.
- **Blockers:** ratifies decisions Plans 3–10 made — do last.
- **Status:** done 2026-05-13. Commits: `f7385e88`, `9556a642`, `7206269e`, `df084eda`, `73adf60d`, `abd0371a`, `b15913ba`, `8b75a732`.

---

## Update protocol

When closing a Plan:

1. Flip its `Status` in the table above to `done YYYY-MM-DD`.
2. Add a `**PRs:**` line under the Plan body linking the merged PR(s).
3. For each `Closes` row, set `Status: closed` in the corresponding `wiki/modules/<m>-tech-debt.md` row (or move out of `wiki/backlog/<m>-refactor.md`).
4. Bump `Last verified` at the top of this doc.
5. Dispatch `wiki-curator` agent to update `Last verified` stamps on any wiki doc the Plan changed code under.

When splitting a Plan (e.g. 6 → 6a + 6b):

1. Edit the row in the execution-order table to two rows.
2. Author a new section here. Keep the original Plan number, append a/b suffix.
3. Move `Closes` rows between the two halves so each half is shippable alone.

