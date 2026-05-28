# Plan: Frontend Screen QA Roadmap (autonomous, per-screen, fresh-session)

## Summary
Operationalize an autonomous, per-screen QA sweep of the entire MetalDocs React frontend after the Go backend quality-bar refactor. Each user-facing screen gets its own branch and a fresh-session `spawn_task` QA run grounded in the canonical QA operating system. Primary risk being hunted: frontend ↔ contract ↔ runtime ↔ wiki **drift** introduced by the backend refactor.

## User Story
As the MetalDocs maintainer, I want every frontend screen QA'd in isolation by a disciplined autonomous agent, so that backend-refactor-induced drift, broken workflows, and misleading UI states are caught and classified before release.

## Problem → Solution
Backend refactored across all modules (tenant isolation, authz GUC ordering, contract reshape, migrations 0211-0214) → frontend may now expect behavior the backend no longer owns. Solution: branch-per-screen + fresh-session autonomous QA, each following the 5-truths / 7-gates / closed-loop model, producing evidence + classified findings.

## Metadata
- **Complexity**: Large (orchestration across ~20 screens)
- **Source PRD**: N/A (free-form roadmap request)
- **Estimated Screens**: 20 distinct QA targets across 4 priority tiers
- **Base branch for all screen branches**: `main` (at squash `5d9e46ba0`)

---

## Grounding (mandatory reading for every QA run)

| Priority | File | Why |
|---|---|---|
| P0 | `CLAUDE.md` | Read order, skill routing, mandatory gates, evidence rule, hard-stop rule |
| P0 | `wiki/quality/qa-operating-system.md` | 5 truths, 7 gates, severity, classification, closed-loop, hard-stops |
| P0 | `wiki/quality/screen-qa-checklist.md` | Default per-screen QA pass |
| P1 | `wiki/quality/backend-api-qa-checklist.md` | When a screen exposes contract/route drift |
| P1 | `wiki/quality/workflow-async-qa-checklist.md` | Approval / distribution / render-fanout screens |
| P1 | `wiki/quality/release-closeout-checklist.md` | Per-screen close-out |
| P1 | `wiki/index.md` → domain index → owning `wiki/modules/<module>.md` | Screen's owning-module truth (do not re-grep) |
| P2 | `wiki/architecture/frontend-structure.md` | Canonical FE layout rules (no legacy paths) |
| P2 | `wiki/architecture/api-contract.md` | Contract-first / codegen drift rules |

**Startup truth (every run):** `.\scripts\start-api.ps1` (API :8081), frontend dev/preview :4173. Login `admin / AdminMetalDocs123!`. Use `corepack pnpm` (pnpm not on PATH; bash lacks `pnpm.cmd`).

---

## Screen Inventory (QA targets)

Route paths from `features/*/routes.tsx`. Owning module from `wiki/modules/`. QA class drives which checklist leads.

| # | Screen | Route(s) | Page component | Owning module / wiki | QA class | Tier |
|---|---|---|---|---|---|---|
| 1 | Login | `/login` | `auth/pages/LoginPage.tsx` | auth | screen + authz | P0 |
| 2 | Password change | `change-password` | `auth/pages/PasswordChangeRoutePage.tsx` | auth | screen + authz | P0 |
| 3 | Workspace shell | layout wrapper | `shell/WorkspaceShell.tsx` | shell / frontend-primitives | screen (nav/authz gating) | P0 |
| 4 | Dashboard | `/` index | `dashboard/pages/DashboardPage.tsx` | dashboard | screen | P0 |
| 5 | New Document Wizard | `documents/new` | `documents/pages/NewDocumentWizardPage.tsx` | novo-documento-wizard | screen (re-verify only) | P1 |
| 6 | Documents Hub | `documents`, `documents/mine`, `documents/recent` | `documents/pages/DocumentsHubPage.tsx` | documents | screen | P1 |
| 7 | Library | `documents/all`, `documents/area/:areaCode`, `documents/type/:profileCode` | `documents/pages/LibraryPage.tsx` | documents | screen | P1 |
| 8 | Document Editor | `documents/:documentId/edit` | `documents/pages/DocumentEditorPage.tsx` (+RoutePage) | documents + editor-ui-eigenpal + editor-chrome | screen (heavy) | P1 |
| 9 | Controlled Documents list/explorer | `controlled-documents` | `controlled-documents/pages/ControlledDocumentsPage.tsx` / `ControlledDocumentsExplorerPage.tsx` | controlled-documents | screen | P1 |
| 10 | Controlled Document detail | `controlled-documents/:controlledDocumentId` | `controlled-documents/ControlledDocumentDetailPage.tsx` | controlled-documents | screen | P1 |
| 11 | Templates list | `templates` | `templates/pages/TemplatesListRoutePage.tsx` | templates | screen | P1 |
| 12 | Template Wizard | `templates/new` | `templates/pages/TemplateWizardPage.tsx` | templates | screen | P1 |
| 13 | Template Editor | `templates/:templateId/versions/:versionNum` | `templates/pages/TemplateEditorRoutePage.tsx` | templates + editor | screen (heavy) + workflow (publish/version) | P1 |
| 14 | Approval Inbox | `approvals` | `approval/pages/InboxPage.tsx` | approval | workflow-async + screen | P2 |
| 15 | Approval Route Admin | `approval-routes` | `approval/pages/RouteAdminPage.tsx` | approval | screen | P2 |
| 16 | Document Published view | `documents/doc/:documentId` | `documents/pages/DocumentPublishedPage.tsx` | documents | screen | P2 |
| 17 | Document Distribution | `distribution` | `documents/pages/DocumentDistributionPage.tsx` | documents + render-fanout | workflow-async + screen | P2 |
| 18 | Admin Center (IAM) | `admin` | `iam/pages/AdminCenterPage.tsx` | iam | screen + authz | P3 |
| 19 | Area Membership Admin | `admin/memberships` | `iam/pages/AreaMembershipAdminRoutePage.tsx` | iam | screen + authz | P3 |
| 20 | Taxonomy Admin | `admin/taxonomy` | `taxonomy/pages/TaxonomyAdminRoutePage.tsx` | taxonomy | screen | P3 |
| 21 | Audit | `audit` | `audit/pages/AuditPage.tsx` | audit | screen + authz | P3 |
| 22 | Notifications | `notifications` | `notifications/pages/NotificationsPage.tsx` | notifications | screen | P3 |
| 23 | Operations | `operations` | `operations/pages/OperationsPage.tsx` | platform/operations | screen | P3 |
| 24 | Content Builder | `content-builder` | `content-builder/pages/ContentBuilderPage.tsx` | content-builder | screen | P3 |

---

## Prioritization & Ordering

Rationale: prerequisites first (auth/shell gate everything), then the document lifecycle spine (highest user value + highest contract-drift surface area), then async workflows, then governance/admin surfaces.

### Tier P0 — Prerequisite gates (run first, sequential)
Auth/session/shell failures invalidate every downstream QA run.
- `qa/auth-login` (1)
- `qa/auth-password-change` (2)
- `qa/workspace-shell` (3)
- `qa/dashboard` (4)

### Tier P1 — Document & template lifecycle spine (run after P0 green; parallelizable)
- `qa/documents-new-wizard` (5) — light re-verify; just synced 2026-05-28
- `qa/documents-hub` (6)
- `qa/documents-library` (7)
- `qa/documents-editor` (8) — heavy; editor ACL + eigenpal
- `qa/controlled-documents-list` (9)
- `qa/controlled-documents-detail` (10)
- `qa/templates-list` (11)
- `qa/templates-wizard` (12)
- `qa/templates-editor` (13) — heavy; version/publish lifecycle

### Tier P2 — Async / workflow surfaces (workflow-async checklist leads)
- `qa/approval-inbox` (14)
- `qa/approval-route-admin` (15)
- `qa/documents-published` (16)
- `qa/documents-distribution` (17)

### Tier P3 — Governance / admin surfaces
- `qa/iam-admin-center` (18)
- `qa/iam-area-membership` (19)
- `qa/taxonomy-admin` (20)
- `qa/audit` (21)
- `qa/notifications` (22)
- `qa/operations` (23)
- `qa/content-builder` (24)

**Branch naming convention:** `qa/<kebab-screen-slug>`, branched from `main`. One screen per branch. QA findings + fixes land on that branch; PR per screen back to `main`.

---

## Per-Screen Workflow (what each spawned run does)

Each fresh `spawn_task` session runs the closed loop from `qa-operating-system.md`:

1. **Gate 0 — Scope truth**: confirm route ownership, owning module wiki, contract surface. Establish startup + auth truth (`start-api.ps1`, login, target route reachable).
2. **Gate 1 — Implementation truth**: `corepack pnpm tsc --noEmit -p tsconfig.build.json` + targeted `corepack pnpm exec vitest run <screen test>` green before QA.
3. **Gate 3 — Product QA truth**: drive the screen as the acting role(s) through the `screen-qa-checklist`. For async screens add `workflow-async-qa-checklist` (request accepted → state persisted → async owner executed → user-visible truth). Use the Playwright/preview tooling for runtime proof.
4. **Gate 2/4 — Review + Root cause**: classify every finding (`runtime prerequisite` / `shared contract prerequisite` / `module-local` / `screen-local` / `wiki-memory drift` / `workflow gap` / `architecture contradiction` / `defer`) + severity (`critical/high/medium/low`). Fix by family, not symptom.
5. **Hard-stop rule**: if a fix needs API redesign / cross-module authz change / storage redesign / workflow semantic redesign / large coordinated rewrite → STOP, report boundary + minimum prerequisite plan. Do not symptom-patch.
6. **Gate 5/6 — Regression + Evidence**: targeted regression for touched surface; record validation commands, runtime proof, persisted/API proof, classified findings, explicit bounded defers. Update owning `wiki/modules/<module>.md` + `Last verified:` when code truth changed; dispatch `wiki-curator` if structural.

---

## spawn_task Brief Template

Each screen's autonomous QA is launched in a **fresh session** with a self-contained brief. Fill `<…>` from the inventory row.

```
title: QA <Screen> — autonomous closed-loop screen QA
cwd: C:\Users\leandro.theodoro\Documents\MetalDocs

prompt:
You are running an autonomous, fresh-session QA pass on ONE MetalDocs frontend screen,
following the canonical QA operating system. Do not assume prior context.

SCREEN: <Screen name>
ROUTE(S): <route paths>
PAGE COMPONENT: <features/.../Page.tsx>
OWNING MODULE WIKI: wiki/modules/<module>.md
QA CLASS: <screen | screen+authz | workflow-async+screen>
BRANCH: qa/<slug> (branch from main; commit findings + fixes here; open a PR per screen)

MANDATORY READ ORDER (do not re-grep what the wiki already states):
1. CLAUDE.md (read order, skill routing, mandatory gates, evidence rule, hard-stop rule)
2. wiki/quality/qa-operating-system.md (5 truths, 7 gates, severity, classification, closed loop, hard stops)
3. wiki/quality/screen-qa-checklist.md  (+ workflow-async-qa-checklist.md if QA class is async)
4. The owning module wiki + its -tech-debt page
5. wiki/architecture/frontend-structure.md + api-contract.md if you touch FE/contract

SKILL ROUTING (mandatory): metaldocs-frontend for FE work; metaldocs-tanstack-query for query/cache;
metaldocs-backend-api for contract/route drift; runtime-contract-prereq if a prerequisite boundary fails.

STARTUP TRUTH: .\scripts\start-api.ps1 (API :8081). Frontend preview :4173. Login admin / AdminMetalDocs123!.
Use `corepack pnpm` (pnpm is NOT on PATH; bash cannot see pnpm.cmd). tsc: corepack pnpm tsc --noEmit -p tsconfig.build.json.

EXECUTE THE CLOSED LOOP:
- Gate 0 scope truth: confirm route/runtime/contract ownership + auth/session truth before QA.
- Gate 1: tsc + targeted vitest for this screen green first.
- Gate 3 product QA: drive the screen as the acting role(s) per the checklist — initial load, happy path,
  empty, validation, loading/disabled/submitting, server/network failure, stale/refresh/return-nav,
  authz per role, optimistic→persisted settle, misleading copy/badges, linked-nav destinations, refresh re-entry.
  For async class: split proof into request-accepted / state-persisted / async-owner-executed / user-visible-final-truth.
- Classify EVERY finding by family + severity. Fix by root-cause family, not symptom.
- HARD STOP if the fix needs API redesign / cross-module authz change / storage redesign / workflow
  semantic redesign / large coordinated rewrite — report boundary + minimum prerequisite plan instead.
- Gate 5/6: targeted regression + record evidence (commands, runtime proof, persisted/API proof, classified
  findings, explicit bounded defers). Update wiki/modules/<module>.md Last verified when code truth changed.

PRIME RISK TO HUNT: frontend ↔ contract ↔ runtime ↔ wiki DRIFT from the Go backend quality-bar refactor
(tenant isolation, authz GUC ordering, contract reshape, migrations 0211-0214).

DELIVERABLE: a per-screen QA report — screen/route/role exercised, gate-by-gate result, classified findings
(family+severity) with disposition (fixed / hard-stop / bounded defer + wiki link), evidence, and the
verification commands you ran. Close ONLY with evidence; "looks good" is not closure.
```

---

## NOT Building
- New product features or screens (this is QA of existing surfaces, not net-new build).
- Backend redesigns — those are hard-stops to report, not work to perform inside a screen QA run.
- A shared E2E harness rewrite — use existing Playwright/preview + vitest tooling.
- Cross-screen refactors — each run is bounded to its screen + owning module boundary.

---

## Step-by-Step Tasks (orchestration)

### Task 1: Create P0 branches
- **ACTION**: From `main`, create `qa/auth-login`, `qa/auth-password-change`, `qa/workspace-shell`, `qa/dashboard`.
- **VALIDATE**: `git branch --list 'qa/*'` shows the 4 branches; each points at `main` HEAD.

### Task 2: Spawn P0 QA runs (sequential)
- **ACTION**: One `spawn_task` per P0 screen using the brief template. Run sequentially — gate failures here block downstream tiers.
- **GOTCHA**: If auth/shell QA finds a runtime/contract prerequisite failure, fix it before spawning P1 (downstream runs assume working auth + shell).
- **VALIDATE**: each P0 run returns a closed report with evidence; no unresolved `critical`/`high` authz finding.

### Task 3: Create + spawn P1 (lifecycle spine)
- **ACTION**: Create `qa/*` branches for screens 5-13; spawn (parallel allowed except the two heavy editors run focused).
- **VALIDATE**: per-screen reports; drift findings classified; heavy editors get explicit authz + persistence proof.

### Task 4: Create + spawn P2 (async/workflow)
- **ACTION**: Branches + spawns for 14-17 using workflow-async checklist as lead.
- **VALIDATE**: each async report shows the 4-part proof split.

### Task 5: Create + spawn P3 (governance/admin)
- **ACTION**: Branches + spawns for 18-24.
- **VALIDATE**: authz-sensitive admin screens (IAM, audit, taxonomy) show role-based allow/deny proof.

---

## Validation Commands

### Per-screen static
```
corepack pnpm tsc --noEmit -p tsconfig.build.json     # from frontend/apps/web
```
EXPECT: zero type errors.

### Per-screen targeted test
```
corepack pnpm exec vitest run <path to screen test>   # from frontend/apps/web
```
EXPECT: all tests pass.

### Runtime
```
.\scripts\start-api.ps1        # API :8081
# frontend preview :4173, login admin / AdminMetalDocs123!
```
EXPECT: target route reachable as acting role; screen-qa-checklist paths exercised.

---

## Acceptance Criteria
- [ ] All 24 screens have a `qa/<slug>` branch + a fresh-session QA report.
- [ ] Every finding classified by family + severity with disposition.
- [ ] Hard-stops reported with boundary + minimum prerequisite plan (not symptom-patched).
- [ ] Evidence recorded per run (commands + runtime + persisted/API proof).
- [ ] Owning `wiki/modules/<module>.md` `Last verified:` bumped where code truth changed.

## Risks
| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Backend contract drift cascades across many screens | High | High | P0/P1 ordering surfaces shared-contract prerequisites early; classify as `shared contract prerequisite` and fix at the owning boundary once |
| Autonomous run symptom-patches instead of root-cause | Medium | High | Brief enforces classification + hard-stop rule + evidence-before-closure |
| Parallel runs touch shared FE primitives and conflict | Medium | Medium | Each run bounded to its screen + module; shared-primitive changes are hard-stop/defer, not inline |
| GitHub Actions billing block prevents CI on screen PRs | High (current) | Medium | Local green evidence (tsc + vitest + runtime proof) is the gate until billing restored |

## Notes
- CI is currently **billing-blocked** (jobs fail in ~3s, "spending limit needs to be increased"). Until resolved, per-screen PR gating relies on local evidence per the QA operating system's runtime-wins + evidence rules.
- `novo-documento` (screen 5) was already synced to runtime truth 2026-05-28 (`/documents/new`, visibility company+area + blank-template closed); its run is a light re-verify, not a full first pass.
- Base for all branches: `main` @ `5d9e46ba0` (PR #18 squash).
```
