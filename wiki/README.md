# MetalDocs Wiki

> **Last verified:** 2026-05-09
> **Purpose:** Single source of truth for codebase knowledge. Read this first — drill into folders only after.

## How to use this wiki

- **Humans:** Browse by folder.
- **AI agents:** Read this index, then `Glob wiki/**/*.md` to discover. Each doc has `Last verified:` + `Key files:` block at the top - use those file:line anchors instead of re-grepping.
- **Drift policy:** When changing code referenced by a doc, update the doc's `Last verified` stamp. Stale stamps = trust nothing in that doc until verified.

---

## Index

### Backlog (deferred / intentional stubs)
- [backlog/library-screen.md](backlog/library-screen.md) — **7 deferred items** for `/documents` Library screen: ActivityPanel inbox + audit wiring, 3 mocked stat cards, Filtros panel, Exportar action. Each item has backend prereq + frontend steps. (Last verified: 2026-05-06)
- [backlog/novo-documento.md](backlog/novo-documento.md) — **6 deferred items** for the novo-documento wizard (`/documents-v2/new`): visibility enforcement, sequence preview, template versions, blank template, slot rollback, profile counts. (Last verified: 2026-05-07)
- [backlog/templates.md](backlog/templates.md) — **5 deferred items** for Templates List screen: `updated_at` field on `TemplateDTO`, `created_by` → display name, card gap delta (14px vs token), mobile tab clipping at 375px, `formatRelative` promotion to `lib/utils/`. (Last verified: 2026-05-08)
- [backlog/caixa-aprovacao.md](backlog/caixa-aprovacao.md) — **7 deferred items** for Caixa de Aprovação screen (`/approvals`): action button wiring (signoff/return/open-doc), heatmap real data, timeline click handlers, "Revisar →" cross-view nav, `view` type narrowing, eye icon, `approvalApi.ts` migration to `lib/api/client.ts`. (Last verified: 2026-05-08)
- [backlog/documento-publicado.md](backlog/documento-publicado.md) — **9 deferred items** for Documento Publicado screen (`/documents/:documentId`): revision list endpoint, relationship model, comments architecture, PDF download, coverage KPI, `area_code`/`profile_code`/`controlled_document_id` in `DocumentResponse`, "Iniciar revisão" mutation wiring. (Last verified: 2026-05-08)
- [backlog/novo-template-wizard.md](backlog/novo-template-wizard.md) — **2 deferred items + 4 stub steps** for Template Creation Wizard (`/templates-v2/new`): `template-counts` aggregate endpoint, `chk-disabled` hardcode; Steps 2–5 (Identidade/Estrutura/Permissões/Confirmação) not yet implemented. (Last verified: 2026-05-09)

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
- **Skill:** `.claude/skills/metaldocs-screen-implementation/SKILL.md` — 6-phase workflow + per-screen `IMPLEMENTATION.md` worksheet for landing designed screens right the first time. Spec: `docs/superpowers/specs/2026-05-06-screen-implementation-skill-design.md`. (Last verified: 2026-05-06)

### Modules (one per backend module / frontend feature)
- [modules/templates-v2.md](modules/templates-v2.md) - template authoring, schemas, versioning, approval; List screen (`/templates-v2`) wired to real API with tab filter + loading/error/empty states; `WorkspaceHeroHeader tone="flat"` pattern; `TabBar` WAI-ARIA a11y; `TemplateAuthorPage` consumes `EditorChrome`; Creation wizard (`/templates-v2/new`) Step 1 (Escopo/profile picker) implemented; Steps 2–5 stub (Last verified: 2026-05-09)
- [modules/documents.md](modules/documents.md) - document instances, Library screen (`/documents`), DocumentPublishedPage (`/documents/:documentId`, Phases 0–3c), novo-documento wizard (`/documents-v2/new`, 4-step, atomic single-call create), editing flow, session model, API; `libraryStatus.ts` status-meta, `visibilityMeta.ts` SSOT, `documentDetailMeta.ts` date-formatter SSOT, `DocumentVersionTimeline`, `useDocumentDetailQuery`, `useApprovalInstanceQuery`, `asApiError` wrapper, `AuthorCell`, debounced search, `keepPreviousData`/`staleTime`; backend `internal/modules/documents/`, table `public.documents`; codegen bootstrap only — handler migration deferred (see `backlog/contract-first-followups.md`); migration 0183 adds `name NOT NULL + CHECK` (Last verified: 2026-05-08)
- **[modules/novo-documento-wizard.md](modules/novo-documento-wizard.md)** - wizard state machine (`wizardReducer`, `clampStep`, `canAdvance`), step components, `WizardFooter` + `DocPaperPreview` primitives, visibility sub-controls, `resolveQueryError`, `STALE_FIVE_MINUTES`, `QK.templates.byProfile` (Last verified: 2026-05-07)
- [modules/taxonomy.md](modules/taxonomy.md) - document families (global), profiles, areas; CRUD routes, scoping distinction, deactivation guards (Last verified: 2026-05-02)
- [modules/approval.md](modules/approval.md) - approval routes, signoffs, ISO segregation, eligibility enforcement (J1), idempotency store, SoD error states (E2), known gaps E4; Caixa de Aprovação inbox UI (`/approvals`): `InboxStack`/`InboxTimeline` two-view pattern, mock fallback strategy, keyboard nav, view persistence (Last verified: 2026-05-08)
- [modules/render-fanout.md](modules/render-fanout.md) - DOCX -> PDF rendering, substitution engine (stub, Last verified: 2026-05-01)
- [modules/iam-rbac.md](modules/iam-rbac.md) - capabilities, roles (viewer/editor/author/approver/system_admin) + process-area roles (signer/area_admin/qms_admin), DB-backed CanDo, area-scoped authz.Require, group grants, tenant-scoped role_provider + role_admin_repository (Group B), migration 0162-0166 + 0169 + 0170; migration 0164 visibility note corrected (Last verified: 2026-05-07)
- [modules/editor-ui-eigenpal.md](modules/editor-ui-eigenpal.md) - eigenpal integration layer, controlled package, plugin wiring (Last verified: 2026-05-06)
- [modules/editor-chrome.md](modules/editor-chrome.md) - shared toolbar overlay primitive for eigenpal-based pages; slot API (`left/center/right/alert`), `VersionBadge`, `AutosaveStatus`, eigenpal CSS overrides, design-token coverage; consumed by `TemplateAuthorPage` + `DocumentEditorPage` (Last verified: 2026-05-06)
- [modules/search.md](modules/search.md) - cross-module search; v2 reader JOINs `controlled_documents` to populate `DocumentCode` (stub, Last verified: 2026-05-01)

#### documents snapshot note

Snapshot columns (`placeholder_schema_snapshot`, etc.) are populated at document creation by `application.SnapshotService`, wired via `documents.Dependencies.SnapshotReader`/`SnapshotWriter`. The `enforce_snapshot_on_submit_trg` trigger enforces these are non-NULL before draft -> under_review.

`public.documents_v2` was the W1 scaffold table (migration 0103); dropped by migration 0168. Use `public.documents` for all queries.

### Concepts (cross-cutting)
- [concepts/placeholders.md](concepts/placeholders.md) - **CRITICAL:** fixed 7-token catalog, substitution at freeze; composition system deprecated 2026-04-27 (Last verified: 2026-04-27)
- [concepts/token-syntax.md](concepts/token-syntax.md) - `{name}` vs `{{uuid}}` - why it matters
- [concepts/controlled-documents.md](concepts/controlled-documents.md) - code generation (`{profile}-{area}-NNN`, 3-digit), atomic create endpoint (`POST /api/v2/controlled-documents`), preview endpoint, idempotency-key requirement, revision endpoint (Last verified: 2026-05-07)
- [concepts/iso-segregation.md](concepts/iso-segregation.md) - why submitter cannot approve own submit; SoD enforcement points + cross-ref to error-ux (stub, Last verified: 2026-05-04)
- [concepts/freeze-and-hashing.md](concepts/freeze-and-hashing.md) - content_hash, values_hash, schema_hash, immutability (stub, Last verified: 2026-05-01)
- [concepts/authz-tiers.md](concepts/authz-tiers.md) - two-tier authz model: tier-1 CapabilityService (HTTP middleware) vs tier-2 authz.Require (in-tx area check); GUC setup, pitfalls (Last verified: 2026-05-03)
- [concepts/error-ux.md](concepts/error-ux.md) - shared `apiFetch`/`ApiError`/auth-bus/`resolveErrorMessage`; E2 SoD dialog states, E3 finalize toast, E4 global 401 interceptor, `signoff.not_eligible` 403 code (J1) (Last verified: 2026-05-05)
- [concepts/design-workflow-audit.md](concepts/design-workflow-audit.md) - audit AI-generated `design-source/` mockups vs real document states, RBAC, personas before implementing; record Keep/Cut/Defer in screen NOTES.md (Last verified: 2026-05-06)

### Tests
- **[tests/system-acceptance-test.md](tests/system-acceptance-test.md)** — full manual end-to-end acceptance run for regulatory-grade QMS; Groups A–E regression coverage, Routines A0–G, pass/fail rubric; E12 anchor updated for atomic create (Last verified: 2026-05-07) **Use this for pre-release validation.**

### Workflows (end-to-end flows)
- **[workflows/user-onboarding.md](workflows/user-onboarding.md)** - full user journey, non-technical: taxonomy -> template -> profile binding -> CD -> fill-in -> approval -> freeze -> PDF (Last verified: 2026-05-04) **Read for conceptual context; use tests/system-acceptance-test.md for the click-by-click run.**
- [workflows/template-authoring.md](workflows/template-authoring.md) - create -> edit schema -> submit -> approve (stub, Last verified: 2026-05-01)
- [workflows/document-fillin.md](workflows/document-fillin.md) - pick CD -> wizard -> editor -> fill placeholders (stub, Last verified: 2026-05-01)
- [workflows/approval.md](workflows/approval.md) - submit, route, signoffs, idempotency; atomic finalize+submit, eligible_actor_ids fix, PostgresSignoffIdempStore; cross-ref to error-ux (Last verified: 2026-05-04)
- [workflows/freeze-and-fanout.md](workflows/freeze-and-fanout.md) - approve -> freeze -> fanout -> PDF artifact

### Decisions (ADRs)
- [decisions/0001-eigenpal-adoption.md](decisions/0001-eigenpal-adoption.md) - why we picked eigenpal over CKEditor/BlockNote
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
