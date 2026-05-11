# MetalDocs Wiki

> **Last verified:** 2026-05-11 (roadmap index added)
> **Purpose:** Single source of truth for codebase knowledge. Read this first — drill into folders only after.

## How to use this wiki

- **Humans:** Browse by folder.
- **AI agents:** Read this index, then `Glob wiki/**/*.md` to discover. Each doc has `Last verified:` + `Key files:` block at the top - use those file:line anchors instead of re-grepping.
- **Drift policy:** When changing code referenced by a doc, update the doc's `Last verified` stamp. Stale stamps = trust nothing in that doc until verified.

---

## Index

### Backlog (deferred / intentional stubs)
- **[backlog/roadmap.md](backlog/roadmap.md) — ordered refactor roadmap (11 sub-plans P0–P5, ~57 PRs). Read before starting a refactor session to know which Plan is next; update status when a Plan closes. (Last verified: 2026-05-11)**
- [backlog/templates_v2-refactor.md](backlog/templates_v2-refactor.md) — templates_v2 refactor backlog (16 rows: R-001..R-014 code-quality / authz / spec / envelope / schema items; R-100 retire `templates-v2.md` predecessor; R-101 rename module dir + route prefix) (Last verified: 2026-05-10)
- [backlog/library-screen.md](backlog/library-screen.md) — **7 deferred items** for `/documents` Library screen: ActivityPanel inbox + audit wiring, 3 mocked stat cards, Filtros panel, Exportar action. Each item has backend prereq + frontend steps. (Last verified: 2026-05-06)
- [backlog/novo-documento.md](backlog/novo-documento.md) — **6 deferred items** for the novo-documento wizard (`/documents-v2/new`): visibility enforcement, sequence preview, template versions, blank template, slot rollback, profile counts. (Last verified: 2026-05-07)
- [backlog/templates.md](backlog/templates.md) — **5 deferred items** for Templates List screen: `updated_at` field on `TemplateDTO`, `created_by` → display name, card gap delta (14px vs token), mobile tab clipping at 375px, `formatRelative` promotion to `lib/utils/`. (Last verified: 2026-05-08)
- [backlog/caixa-aprovacao.md](backlog/caixa-aprovacao.md) — **7 deferred items** for Caixa de Aprovação screen (`/approvals`): action button wiring (signoff/return/open-doc), heatmap real data, timeline click handlers, "Revisar →" cross-view nav, `view` type narrowing, eye icon, `approvalApi.ts` migration to `lib/api/client.ts`. (Last verified: 2026-05-08)
- [backlog/documento-publicado.md](backlog/documento-publicado.md) — **9 deferred items** for Documento Publicado screen (`/documents/:documentId`): revision list endpoint, relationship model, comments architecture, PDF download, coverage KPI, `area_code`/`profile_code`/`controlled_document_id` in `DocumentResponse`, "Iniciar revisão" mutation wiring. (Last verified: 2026-05-08)
- [backlog/template-editor.md](backlog/template-editor.md) — **7 deferred items** for Template Editor screen (`/templates-v2/:id/versions/:n`): `version-history` (list endpoint), `comments` (model gap), `outline-future-enhancements` (click-to-scroll, drag, panel persistence), `design-toolbar-parity` (Decision A wont-fix), `placeholder-catalog-panel-restyle` (tokens migration), `convergence-test-rewrite`, `submitForReview-error-codes`. (Last verified: 2026-05-10)
- [backlog/novo-template-wizard.md](backlog/novo-template-wizard.md) — **9 deferred items** for Template Creation Wizard (`/templates-v2/new`): `template-counts` aggregate endpoint, `chk-disabled` hardcode, `next-code-preview` endpoint, `key-generation` UX decision, `font-size-hero` token gap, `step3-docx-upload` / `step3-placeholder-extract` / `step3-editor-handoff` (3 Step-3 deferred items), `permissions-roles-api` / `permissions-area-counts` / `permissions-user-count` (3 Step-4 deferred items), `confirmacao-backend-submit` (Step-5 mocked submit); Steps 2–5 shipped; Step 5 submit mocked pending `key-generation` + API. (Last verified: 2026-05-10)
- [backlog/iam-refactor.md](backlog/iam-refactor.md) — IAM refactor backlog (14 rows, R-001 through R-014; R-014 merged) (Last verified: 2026-05-10)
- [backlog/documents-refactor.md](backlog/documents-refactor.md) — documents refactor backlog (9 rows: R-001..R-006 RFC-9457/spec-drift/tripwire/dupe-route/rename-tx/idempotency, R-008 capability namespace, R-009 placeholder FK, R-100 retire documents-v2.md stub) (Last verified: 2026-05-10)
- [backlog/auth-refactor.md](backlog/auth-refactor.md) — auth refactor backlog (12 rows, R-001 through R-012) (Last verified: 2026-05-10)
- [backlog/approval-refactor.md](backlog/approval-refactor.md) — approval refactor backlog (12 rows, R-001..R-012: RFC-9457 envelope, OpenAPI spec gap, substring classifier 500→409, deprecated PDF dispatcher, inbox snapshot drift, cancel/cutover tripwire audit, infra naming, document_v2_id suffix, NOT VALID FKs, undocumented symbols, dual GUC helpers, unwired iam AuthorizationService) (Last verified: 2026-05-10)
- [backlog/audit-refactor.md](backlog/audit-refactor.md) — audit refactor backlog (12 rows, R-001..R-012) (Last verified: 2026-05-10)
- [backlog/registry-refactor.md](backlog/registry-refactor.md) — registry refactor backlog (13 rows: R-001..R-012 code-quality / authz / audit / tripwire / tenant-scoping / error-envelope / spec-drift items; R-100 legacy `profile_sequence_counters` cleanup) (Last verified: 2026-05-11)
- [backlog/taxonomy-refactor.md](backlog/taxonomy-refactor.md) — taxonomy refactor backlog (rows R-001..R-016 one-to-one with T-001..T-016: tenant-header trust, global-families ADR, PATCH-families dispatcher, FamilyService govLogger, Profile/Area Create+Update govLogger, single-tier defense, TOCTOU race, RFC 9457 envelope, OpenAPI codegen migration, parallel audit sink, idempotency, pagination, family-code DB trigger, Go doc comments, redundant PK, area-hierarchy ADR) (Last verified: 2026-05-11)
- [backlog/editor-ui-eigenpal-refactor.md](backlog/editor-ui-eigenpal-refactor.md) — editor-ui-eigenpal refactor backlog (10 rows R-001..R-010 incl. 2 `maint:` rows; R-001 restore tarball, R-002 migrate `TemplateEditorPage` to adapter, R-003 rewrite wiring test, R-004..R-010 minor/maint cleanup) (Last verified: 2026-05-10)
- [backlog/editor-chrome-refactor.md](backlog/editor-chrome-refactor.md) — editor-chrome refactor backlog (9 rows R-001..R-009, one-to-one with T-001..T-009: autosave state widening, aria-live, selector guard, baseline tests, token gaps, typed style re-export, pointer-events contract, missing-ADR, slot truthy-collapse) (Last verified: 2026-05-10)

### Bug Tracker
- [bugs/audit-2026-05-03.md](bugs/audit-2026-05-03.md) - **40 bugs in 5 groups (A–H)** pre-smoke deep audit 2026-05-03/04. Groups A–H mostly resolved.
- [bugs/audit-2026-05-04.md](bugs/audit-2026-05-04.md) - **13 bugs in 6 groups (I–N) + Group K (5 bugs)** QA pass 2026-05-04. J1/J2/M1 resolved 2026-05-05 (commits cb56e1e0..1cebea64). I/L/N resolved 2026-05-05 (branch chore/api-cleanup-sub-project-b). K resolved 2026-05-05 (6dc94759, 1b228d28, 7f4b7672; K2 wont-fix). (Last verified: 2026-05-05)

### Vision
- [vision/product-vision.md](vision/product-vision.md) - what MetalDocs is, problem it solves (stub, Last verified: 2026-05-01)
- [vision/target-users.md](vision/target-users.md) - quality engineers, ISO-bound orgs, document control roles (stub, Last verified: 2026-05-01)

### Architecture
- [architecture/system-overview.md](architecture/system-overview.md) - services, ports, end-to-end flow, `internal/platform/httpclient` (M1) (Last verified: 2026-05-05)
- [architecture/data-model.md](architecture/data-model.md) - Postgres tables, key relationships, document_families (global/is_active), metaldocs_app grants, two-query LIMIT/OFFSET+COUNT pattern (stub, Last verified: 2026-05-06)
- [architecture/tech-stack.md](architecture/tech-stack.md) - Go, React, Postgres, MinIO, Gotenberg, eigenpal (stub, Last verified: 2026-05-01)
- [architecture/deployment.md](architecture/deployment.md) - Docker compose, env vars, dev setup (stub, Last verified: 2026-05-01)
- **[architecture/frontend-structure.md](architecture/frontend-structure.md) - canonical frontend layout, routing, state, API, design-system rules; `lib/hooks/` added; `Avatar` color prop; comparison baseline for refactor reviews (Last verified: 2026-05-08)**
- **[architecture/api-contract.md](architecture/api-contract.md) - spec-as-source-of-truth via OpenAPI 3.0.3; oapi-codegen v2 backend codegen per module; openapi-typescript v7 frontend codegen; runtime enforcement gaps (unknown-fields, required-fields); DB invariant floor (migration 0183); CI drift guard; per-module migration status table (Last verified: 2026-05-08)**
- **[architecture/api-design-system.md](architecture/api-design-system.md) - API design system contract: RFC 9457 error envelope, cursor pagination, Stripe-model idempotency, two-tier authz (Postgres tripwire as real enforcer), list filtering, 5 CI lint rules; Plan 2 migrates handlers (Last verified: 2026-05-10)**
- **Skill:** `.claude/skills/metaldocs-screen-implementation/SKILL.md` — 6-phase workflow + per-screen `IMPLEMENTATION.md` worksheet for landing designed screens right the first time. Spec: `docs/superpowers/specs/2026-05-06-screen-implementation-skill-design.md`. (Last verified: 2026-05-06)

### Modules (one per backend module / frontend feature)
- **[modules/templates_v2.md](modules/templates_v2.md)** - **Arc42 + C4 living doc** — templates_v2 backend: authoring lifecycle (`draft → in_review → approved → published → obsolete`), 20 HTTP routes, placeholder catalog enforcement, SoD probing, MinIO presigned upload/download, `documents` downstream snapshot contract; §8.8 placeholder catalog enforcement; §8.7 SoD / author identity surface; companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-10)
- [modules/templates_v2-tech-debt.md](modules/templates_v2-tech-debt.md) - templates_v2 tech-debt register (14 items: 4 Critical / 6 Major / 4 Minor — T-001 authz not applied, T-002 tenant guard gap, T-003 SoD gap on SubmitForReview, T-004 SoD gap on Publish, T-005 legacy error envelope, T-006 partial codegen, T-007 parallel audit sink, T-008 resolver registry unwired, plus 6 additional items) (Last verified: 2026-05-10)
- [modules/templates-v2.md](modules/templates-v2.md) - (predecessor — retire pending R-100) frontend-heavy doc: template authoring screens, List screen, Creation wizard Steps 1–5, `TemplateEditorPage`, `EditorChrome` wiring (Last verified: 2026-05-10)
- `modules/templates_v2/_artifacts/00-context.md` … `06-selfreview.md` — research artifacts (index-only; do not link from other docs)
- [modules/frontend-primitives.md](modules/frontend-primitives.md) - generic `components/ui/` primitives: `SelectableCard` (forwardRef card button, `role="radio"`, idle/selected/disabled CSS states) + `useRovingRadioGroup` hook (ARIA radiogroup roving-tabIndex, orientation config, programmatic focus) (Last verified: 2026-05-10)
- [modules/documents.md](modules/documents.md) - **Arc42 + C4 living doc** — document instance lifecycle (`draft → under_review → approved → published → superseded|obsolete`); §5 C4 Container view; §6 finalize trace (trigger `enforce_snapshot_on_submit_trg`); §8.1 two-tier authz + Postgres tripwire; §8.7 placeholder snapshot + freeze; `CreateDocumentTx` port; `internal/modules/documents/`, table `public.documents`; codegen bootstrap only (ADR 0012); companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-10)
- [modules/documents-tech-debt.md](modules/documents-tech-debt.md) - documents tech-debt register (10 items: T-001 RFC 9457 envelope major; T-002 spec/handler drift critical; T-003 missing tripwire on documents table major; T-004 duplicate route registration minor; T-005 rename audit outside tx major; T-006 finalize idempotency gap major; T-007 audit port latent minor; T-008 capability namespace straddle minor; T-009 placeholder FK wrong target major; T-010 snapshot trigger semantics minor) (Last verified: 2026-05-10)
- **[modules/novo-documento-wizard.md](modules/novo-documento-wizard.md)** - wizard state machine (`wizardReducer`, `clampStep`, `canAdvance`), step components, `WizardFooter` + `DocPaperPreview` primitives, visibility sub-controls, `resolveQueryError`, `STALE_FIVE_MINUTES`, `QK.templates.byProfile` (Last verified: 2026-05-07)
- **[modules/taxonomy.md](modules/taxonomy.md)** - **Arc42 + C4 living doc** — taxonomy: document families (global), profiles, areas; 16 HTTP routes under `/api/v2/taxonomy/*`; per-tenant profile + area scoping; deactivation guards; 5 Critical / 5 Major / 6 Minor tech-debt items; companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-11)
- [modules/taxonomy-tech-debt.md](modules/taxonomy-tech-debt.md) - taxonomy tech-debt register (16 items: T-001..T-005 Critical — tenant header trusted, global families no ADR, PATCH families dispatcher bypass, FamilyService no govLogger, Profile/Area Create+Update no govLogger; T-006..T-010 Major — single-tier defense, TOCTOU race, legacy envelope, no OpenAPI spec, parallel audit sink; T-011..T-016 Minor) (Last verified: 2026-05-11)
- `modules/taxonomy/_artifacts/00-context.md` … `06-selfreview.md` — research artifacts (index-only; do not link from other docs)
- **[modules/approval.md](modules/approval.md)** - Arc42 + C4 living architecture doc — 16-route sign-off chain; SoD, J1 eligibility, quorum, transactional outbox, 4-layer defense-in-depth authz; §6 sequence diagrams for Submit/Signoff/Inbox; §11 pointers to tech-debt register (2C/4M/6m). Supersedes prior stub. (Last verified: 2026-05-10)
- [modules/approval-tech-debt.md](modules/approval-tech-debt.md) - approval tech-debt register (12 items: T-001 RFC 9457 envelope Critical; T-002 doc-scoped routes absent from OpenAPI Critical; T-003 looksLikeValidationError 500 Major; T-004 deprecated PDF dispatcher path Major; T-005 inbox two-query drift Major; T-006 cancel/cutover tripwire gap Major; T-007..T-012 Minor) (Last verified: 2026-05-10)
- [modules/render-fanout.md](modules/render-fanout.md) - DOCX -> PDF rendering, substitution engine (stub, Last verified: 2026-05-01)
- **[modules/iam.md](modules/iam.md)** - IAM module — capabilities, roles, area memberships (Arc42 + C4); covers tier-1 `CapabilityService.CanDo`, tier-2 `authz.Require`, Postgres tripwire, area-scoped authz, group grants, tenant-scoped repositories (Group B), migrations 0002–0170 (Last verified: 2026-05-10)
- [modules/iam-tech-debt.md](modules/iam-tech-debt.md) - IAM tech-debt register (12 items, T-001 through T-012) (Last verified: 2026-05-10)
- **[modules/auth.md](modules/auth.md)** - auth module — session cookie authn, bcrypt password storage, per-account lockout, HMAC-signed opaque session tokens, ManagedUser admin ops (Arc42 + C4); 2 Critical / 3 Major / 7 Minor tech-debt items; companion registers linked below (Last verified: 2026-05-10)
- [modules/auth-tech-debt.md](modules/auth-tech-debt.md) - auth tech-debt register (12 items: T-001 LegacyHeader bypass Critical; T-002 audit-trail gap Critical; T-003 RFC 9457 envelope Major; T-004 non-atomic CreateUser Major; T-005 IP rate-limit absent Major; T-006..T-012 Minor) (Last verified: 2026-05-10)
- **[modules/audit.md](modules/audit.md)** - Arc42 + C4 living doc — append-only event sink (`metaldocs.audit_events`), `Writer`/`Reader` port+adapter, `GET /api/v1/audit/events`; consumers: iam, documents, auth (gap); 2 Critical / 4 Major / 6 Minor tech-debt items; companion register + refactor backlog linked below (Last verified: 2026-05-10)
- [modules/audit-tech-debt.md](modules/audit-tech-debt.md) - audit tech-debt register (12 items: T-001 unauthenticated list endpoint Critical; T-002 legacy error envelope Critical; T-003..T-006 Major; T-007..T-012 Minor) (Last verified: 2026-05-10)
- **[modules/editor-ui-eigenpal.md](modules/editor-ui-eigenpal.md)** - **Arc42 + C4 living doc** — eigenpal Anti-Corruption Layer (`packages/editor-ui/`): 3-value `EditorMode`, mode-gated `templatePlugin`, 1500ms autosave debounce + `inFlightRef` guard, `computeSidebarModel` pure path; 1 Critical (T-001 tarball absent) / 2 Major / 5 Minor debt items; one production consumer (`DocumentEditorPage`); companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-10)
- [modules/editor-ui-eigenpal-tech-debt.md](modules/editor-ui-eigenpal-tech-debt.md) - editor-ui-eigenpal tech-debt register (8 items: T-001 vendored tarball absent Critical; T-002 `TemplateEditorPage` bypasses adapter Major; T-003 stale wiring test Major; T-004..T-008 Minor — dormant `OutlinePlugin` export, `mergefieldPlugin` naming, mode-flip autosave race, missing `templatePlugin`-gate ADR, missing wrapper-only ACL ADR) (Last verified: 2026-05-10)
- [backlog/editor-ui-eigenpal-refactor.md](backlog/editor-ui-eigenpal-refactor.md) - editor-ui-eigenpal refactor backlog (10 rows R-001..R-010 incl. 2 `maint:` rows; R-001 restore tarball, R-002 migrate `TemplateEditorPage` to adapter, R-003 rewrite wiring test, R-004..R-010 minor/maint cleanup) (Last verified: 2026-05-10)
- **[modules/editor-chrome.md](modules/editor-chrome.md)** - **Arc42 + C4 living doc** — shared toolbar overlay primitive for eigenpal-based pages; slot API (`left/center/right/alert`), 6 exported symbols, 17 eigenpal `:global` CSS overrides, `VersionBadge` + `AutosaveStatus` parts; 2 consumers (`TemplateEditorPage`, `DocumentEditorPage`); 0 Critical / 4 Major / 5 Minor debt items; companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-10)
- [modules/editor-chrome-tech-debt.md](modules/editor-chrome-tech-debt.md) - editor-chrome tech-debt register (9 items: 0 Critical / 4 Major / 5 Minor — T-001 autosave 4-vs-7 state collapse Major; T-002 `AutosaveStatus` missing `aria-live` Minor; T-003 eigenpal `:global` selector fragility Major; T-004 zero test coverage Major; T-005 tokens drift Minor; T-006 weakly-typed style re-export Minor; T-007 pointer-events discipline Minor; T-008 missing-ADR Minor; T-009 slot truthy-collapse Minor) (Last verified: 2026-05-10)
- [backlog/editor-chrome-refactor.md](backlog/editor-chrome-refactor.md) - editor-chrome refactor backlog (9 rows R-001..R-009, one-to-one with T-001..T-009) (Last verified: 2026-05-10)
- [modules/search.md](modules/search.md) - cross-module search; v2 reader JOINs `controlled_documents` to populate `DocumentCode` (stub, Last verified: 2026-05-01)
- **[modules/registry.md](modules/registry.md)** - **Arc42 + C4 living doc** — registry: CD catalog, atomic CD + first-revision create (`POST /api/v2/controlled-documents`), per-(tenant, profile, area) sequence numbering (`{PROFILE}-{AREA}-{NNN}`), lifecycle (`active → obsolete | superseded`), `DocumentInitializer` port, idempotency-key replay; 2 Critical / 6 Major / 4 Minor tech-debt items; companion tech-debt register + refactor backlog linked below (Last verified: 2026-05-11)
- [modules/registry-tech-debt.md](modules/registry-tech-debt.md) - registry tech-debt register (12 items: T-001 lifecycle PUTs lack in-module authz Critical; T-002 audit-trail gap on Obsolete/Supersede Critical; T-003 legacy error envelope Major; T-004 tier-3 tripwire absent Major; T-005 tenant query-arg only Major; T-006 GetActiveDocument no authz Major; T-007 OpenAPI spec/handler drift Major; T-008 cross-module audit sink Major; T-009..T-012 Minor) (Last verified: 2026-05-11)
- `modules/registry/_artifacts/00` … `06` — research artifacts (index-only; do not link from other docs)

#### documents snapshot note

Snapshot columns (`placeholder_schema_snapshot`, etc.) are populated at document creation by `application.SnapshotService`, wired via `documents.Dependencies.SnapshotReader`/`SnapshotWriter`. The `enforce_snapshot_on_submit_trg` trigger enforces these are non-NULL before draft -> under_review.

`public.documents_v2` was the W1 scaffold table (migration 0103); dropped by migration 0168. Use `public.documents` for all queries.

### Concepts (cross-cutting)
- [concepts/placeholders.md](concepts/placeholders.md) - **CRITICAL:** fixed 7-token catalog, substitution at freeze; composition system deprecated 2026-04-27; links to templates_v2 §8.8 for backend enforcement (Last verified: 2026-05-10)
- [concepts/token-syntax.md](concepts/token-syntax.md) - `{name}` vs `{{uuid}}` - why it matters (Last verified: 2026-05-10)
- [concepts/controlled-documents.md](concepts/controlled-documents.md) - code generation (`{profile}-{area}-NNN`, 3-digit), atomic create endpoint (`POST /api/v2/controlled-documents`), preview endpoint, idempotency-key requirement, revision endpoint (Last verified: 2026-05-07)
- [concepts/iso-segregation.md](concepts/iso-segregation.md) - why submitter cannot approve own submit; SoD enforcement points + cross-ref to error-ux (stub, Last verified: 2026-05-04)
- [concepts/freeze-and-hashing.md](concepts/freeze-and-hashing.md) - content_hash, values_hash, schema_hash, immutability (stub, Last verified: 2026-05-01)
- [concepts/authz-tiers.md](concepts/authz-tiers.md) - two-tier authz model: tier-1 CapabilityService (HTTP middleware) vs tier-2 authz.Require (in-tx area check); GUC setup, pitfalls; See also: iam.md §5.4 (AuthorizationService, third surface) + iam-tech-debt.md T-003 (Last verified: 2026-05-10)
- [concepts/error-ux.md](concepts/error-ux.md) - shared `apiFetch`/`ApiError`/auth-bus/`resolveErrorMessage`; E2 SoD dialog states, E3 finalize toast, E4 global 401 interceptor, `signoff.not_eligible` 403 code (J1) (Last verified: 2026-05-05)
- [concepts/design-workflow-audit.md](concepts/design-workflow-audit.md) - audit AI-generated `design-source/` mockups vs real document states, RBAC, personas before implementing; record Keep/Cut/Defer in screen NOTES.md (Last verified: 2026-05-06)
- [concepts/css-leakage-offenders.md](concepts/css-leakage-offenders.md) - global CSS rules that clobber component styles; known offenders list + override patterns; required reading before adding `<textarea>` or custom-height inputs (Last verified: 2026-05-09)

### Tests
- **[tests/system-acceptance-test.md](tests/system-acceptance-test.md)** — full manual end-to-end acceptance run for regulatory-grade QMS; Groups A–E regression coverage, Routines A0–G, pass/fail rubric; E12 anchor updated for atomic create (Last verified: 2026-05-07) **Use this for pre-release validation.**

### Workflows (end-to-end flows)
- **[workflows/user-onboarding.md](workflows/user-onboarding.md)** - full user journey, non-technical: taxonomy -> template -> profile binding -> CD -> fill-in -> approval -> freeze -> PDF (Last verified: 2026-05-04) **Read for conceptual context; use tests/system-acceptance-test.md for the click-by-click run.**
- [workflows/template-authoring.md](workflows/template-authoring.md) - create -> edit schema -> submit -> approve (stub, Last verified: 2026-05-01)
- [workflows/document-fillin.md](workflows/document-fillin.md) - pick CD -> wizard -> editor -> fill placeholders (stub, Last verified: 2026-05-01)
- [workflows/approval.md](workflows/approval.md) - submit, route, signoffs, idempotency; atomic finalize+submit, eligible_actor_ids fix, PostgresSignoffIdempStore; cross-ref to error-ux (Last verified: 2026-05-04)
- [workflows/freeze-and-fanout.md](workflows/freeze-and-fanout.md) - approve -> freeze -> fanout -> PDF artifact

### Decisions (ADRs)
- [decisions/0001-eigenpal-adoption.md](decisions/0001-eigenpal-adoption.md) - why we picked eigenpal over CKEditor/BlockNote; affected modules section links templates_v2 (Last verified: 2026-05-10)
- [decisions/0002-zone-purge.md](decisions/0002-zone-purge.md) - why we removed editable zones (2026-04-25)
- [decisions/0003-token-syntax-migration.md](decisions/0003-token-syntax-migration.md) - plan to move from `{{uuid}}` -> `{name}` (stub, Last verified: 2026-05-01)
- [decisions/0007-two-tier-authz.md](decisions/0007-two-tier-authz.md) - accept two distinct authz tiers (CapabilityService vs authz.Require); J2 amendment: `document.create` wired via `NewCapabilityChecker`, `permissiveAuthzChecker` removed (Last verified: 2026-05-05)
- [decisions/0008-placeholder-fixed-catalog.md](decisions/0008-placeholder-fixed-catalog.md) - replace user-fill placeholders with fixed 7-token computed catalog (2026-04-26)
- [decisions/0011-cd-atomic-create.md](decisions/0011-cd-atomic-create.md) - atomic CD create + per-area 3-segment numbering + idempotency-key adoption; deletes legacy two-screen wizard flow (2026-05-07)
- [decisions/0012-contract-first-api.md](decisions/0012-contract-first-api.md) - adopt spec-as-source-of-truth via oapi-codegen; root cause of `documents.name` empty-name bug; migration scope; residual risks for non-migrated modules (2026-05-08)

### References
- [references/eigenpal-spike.md](references/eigenpal-spike.md) - pointer to spike repo + key findings (T1-T8)
- [references/eigenpal-controlled-package.md](references/eigenpal-controlled-package.md) - current vendored EigenPal package contract for MetalDocs
- [references/environment-setup.md](references/environment-setup.md) - local dev: compose, migrations, seed (stub, Last verified: 2026-05-01)
- [references/how-to-run-tests.md](references/how-to-run-tests.md) - Go tests, frontend vitest, e2e playwright (stub, Last verified: 2026-05-01)
- [references/local-dev-startup.md](references/local-dev-startup.md) - **START HERE** - PS script, port, credentials, common mistakes
- [references/local-dev-credentials.md](references/local-dev-credentials.md) - admin login details, DB access
- [references/oapi-codegen.md](references/oapi-codegen.md) - how to regenerate, vendor-mode `GOFLAGS=-mod=mod` requirement, add a new module, include-tags filter (Last verified: 2026-05-08)

### Glossary
- [GLOSSARY.md](GLOSSARY.md) - placeholder, zone (deprecated), fanout, freeze, eigenpal, controlled doc, profile, etc.

---

## Conventions

**Filename:** kebab-case, descriptive. ADRs prefix with 4-digit number.

**File header (every doc):**
```markdown
# Title

> **Last verified:** YYYY-MM-DD
> **Scope:** what this covers
> **Out of scope:** what it doesn't (link to where it does)
> **Key files:**
> - `path/to/file.go:42` - anchor description
> - `path/to/other.tsx:115` - anchor description
```

**Cross-refs:** Use full path + line numbers. Example: `internal/modules/templates_v2/application/schema.go:42`.

**Length:** Hard cap ~300 lines. Split if longer.

**Code blocks:** Always with language tag for highlighting.
