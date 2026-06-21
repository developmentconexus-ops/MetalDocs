# Discovery Brief: Frontend Screen Completion

> **Mission slug:** `frontend-screen-completion`  ·  **Type:** enhancement (with embedded greenfield slices)
> **Date:** 2026-06-21  ·  **Branch:** main
> **Agents / models used:** 4× sonnet Explore analysts (route inventory · documents/templates completeness · ops/approval/admin completeness · design-source coverage) + 1× main-session skeptic pass (endpoint-existence verification by OpenAPI grep + router read).
> This is the **evidence base** the mission stands on. Every claim in `mission.md` traces to a finding here.

## Method

Four read-only Explore agents swept the frontend in parallel on 2026-06-21, after the Grade-A backend program closed (HEAD `d477e9f0`). The main session then ran a skeptic pass on the load-bearing claims (which screens are blocked on missing backend) by grepping `api/openapi/v1/openapi.yaml` and reading `src/app/AppRouter.tsx` + the conflicting `routes.tsx` files directly.

| Agent / lens | Scope swept | Verified how |
|--------------|-------------|--------------|
| Route inventory | `src/app/AppRouter.tsx` + every `features/*/routes.tsx` | Read all router files; enumerated mounted routes; flagged legacy redirects + duplicate index |
| Documents/templates completeness | `features/documents/pages`, `features/templates`, `features/dashboard`, `features/content-builder` | Per-page read: query wiring vs mock literals; TODO/"em breve" markers; token usage |
| Ops/approval/admin completeness | `features/operations`, `audit`, `approval`, `iam`, `taxonomy`, `notifications`, `feature-flags` | Per-screen read: data source, placeholder markers, styling system |
| Design-source coverage | `design-source/**`, `wiki/implementation/screen-redesign-tracker.md` | NOTES.md status extraction; matched slugs → implemented pages |
| Skeptic (main session) | `api/openapi/v1/openapi.yaml`, `AppRouter.tsx`, `dashboard/routes.tsx`, `operations/routes.tsx` | grep for `/stats`, `notification`, `distribution/fanout/coverage`; confirmed duplicate `index: true` |

**Verified vs assumed:** the per-page completeness verdicts and the endpoint-existence findings are **verified** (real files read/grepped this session). The "likely cut" verdict on `alternativas-inicio-caixa` / `catalogo-slots` is **assumed** (no NOTES, no route, no product intent on record) — operator confirmed CUT in Phase 2.

**Skeptic outcome:** the headline backend-blocker finding survived — Notifications and Distribuição fanout/coverage have **no** OpenAPI endpoint; Dashboard's needs (`/documents/stats`, `/iam/kpi`, `/audit/events`, approval inbox) **all exist**. One agent's "biblioteca = not implemented, likely cut" claim was **downgraded**: `biblioteca` is the library design source and already shipped as `LibraryPage` at `/documents`.

## Findings

| # | Finding (with citation) | Severity / kind | Confidence | Proposed home |
|---|-------------------------|-----------------|------------|---------------|
| 1 | Stale tracker: `wiki/implementation/screen-redesign-tracker.md` last updated 2026-05-08; ~half its rows wrong (claims Editor/Publicado "not started" though both ship) | major (truth drift) | verified | M0 / F0.1 |
| 2 | Duplicate `index: true` route — both `dashboard/routes.tsx:5` and `operations/routes.tsx:5` declare the home index; Operations stub is shadowed dead code | major (routing bug) | verified | M0 / F0.2 |
| 3 | `OperationsPage` renders `OperationsCenter` with empty arrays + noop handlers, zero API wiring, legacy CSS (`operations/pages/OperationsPage.tsx`) | major (dead stub) | verified | M0 / F0.3 |
| 4 | `AuditPage` is an identical copy of `OperationsPage` — same empty `OperationsCenter`, no API (`audit/pages/AuditPage.tsx`) | major (dead stub dup) | verified | M0 / F0.3 |
| 5 | `alternativas-inicio-caixa`, `catalogo-slots` — design-source slugs with no NOTES, no route, no product intent | minor (unspecced) | assumed→operator-confirmed CUT | M0 / F0.4 (cut list) |
| 6 | `DashboardPage` (home `/`) ships `MOCK_STATS` + `MOCK_ACTIVITY` hardcoded; only the approval-inbox query is real; redesign tokens already applied | major (mock data on home) | verified | M1 |
| 7 | Dashboard data sources all exist: `/documents/stats` (`openapi.yaml:2587`), `/iam/kpi`, `/audit/events`, approval inbox — Dashboard is wire-only, no backend work | enabler | verified | M1 |
| 8 | `DocumentDistributionPage` — canvas/hero ready but all KPI cards illustrative ("Dados ilustrativos · Em breve"), 4 disabled CTAs; no fanout API | major (stub + missing BE) | verified | M2 |
| 9 | **No** distribution/fanout/coverage endpoint in OpenAPI (only unrelated `/security/mfa-coverage`) → Distribuição needs new backend | blocker | verified | M2 (BE feature) |
| 10 | `NotificationsPage` is an empty stub; `notifications.ts` returns empty arrays, stream subscription is noop, legacy CSS | major (stub + missing BE) | verified | M3 |
| 11 | **No** `notification` endpoint anywhere in OpenAPI → Notifications needs new backend | blocker | verified | M3 (BE feature) |
| 12 | `DocumentPublishedPage` — core lifecycle real (doc detail, approval chain, revision history); 3 TODOs + 6 "em breve" placeholders: PDF download, fanout/coverage, file size/page count, confidentiality, related docs, inline comments | major (backlog stubs) | verified | M4 / F4.1 |
| 13 | `documento-obsoleto` NOTES: "Phase 0 Audit not started"; it is the `obsolete={true}` variant of `PublicadoV5`; no variant logic in `DocumentPublishedPage` | feature (net-new variant) | verified | M4 / F4.2 |
| 14 | `detalhe-signoff` design-source dir exists, no NOTES, no implemented page; approval panels (`ApprovalTimelinePanel`, `ControlledDocumentDetailPanel`, `SignoffDialog`) exist but no standalone sign-off detail screen | feature (net-new) | verified | M5 / F5.1 |
| 15 | `TaxonomyAdminPage` — functional + API-wired (`/api/v1/taxonomy/*`) but uses inline styles, NOT redesign tokens (off the design system) | minor (legacy styling) | verified | M5 / F5.2 |
| 16 | Screen-impl workflow is real: `wiki/architecture/frontend-structure.md` documents `metaldocs-screen-implementation`; design-source slugs carry `IMPLEMENTATION.md` + `artifacts/phase2..4`; reviewer agents `frontend-screen-reviewer` + `frontend-code-reviewer` exist | enabler | verified | all milestones (gate) |

## Constraints & risks surfaced

- **Contract-first regen order (HS-3 hazard).** New backend endpoints (M2 fanout, M3 notifications) MUST follow spec → OpenAPI → codegen → FE types, matching the Grade-A api-lint discipline. FE may not hand-author response types.
- **6 backend CI guards must stay green.** gitleaks · platformboundary · PostCommitAudit · chain-order · nosqltxindomain · nodualmode. Any backend work in M2/M3/M4 inherits these gates + `api-lint -strict` = 0.
- **ADR required for new capabilities/endpoints** (governance rule: PRs cite REQ IDs; new surface needs an ADR). M2/M3 each likely mint an ADR.
- **Advisory-lock hazard (H-PRE-1):** any new authz-recording read stays off lock-holding atomic tx — applies if notifications/fanout touch authz.
- **FE/BE/DB boundary + skill routing** (CLAUDE.md): backend features route through the backend QA checklist; FE screens through the screen-impl workflow + both reviewer agents.
- **Blast radius:** the M0 index-route fix touches the app's landing route — low code blast but high visibility; validate the resolved `/` renders the intended page.

## Open questions for the operator

Resolved in Phase 2 (Locked Decisions D1–D3):
- **D1 — backend-blocked screens:** FULL-STACK — missing endpoints become Grade-A backend milestones. *(operator-approved)*
- **D2 — per-screen gate:** BOTH reviewers (screen + code) APPROVE + tests green. *(operator-approved)*
- **D3 — unspecced slugs:** CUT `alternativas-inicio-caixa` + `catalogo-slots`; `biblioteca` already shipped. *(operator-approved)*

Deferred to the Phase-4 mission.md gate (proposed, operator confirms): milestone **sequencing** (D5) and the Operations/Audit disposition detail (delete vs designate) inside M0.

## Coverage statement

**Swept:** every `features/*` screen with a route, every `design-source/*` slug with a NOTES/HTML/JSX design file, the full OpenAPI path set for the three data-gap screens, and the router config.

**Not swept (no silent caps):**
- Backend endpoint *internals* for the wire-only screens (M1) — confirmed the routes exist, did not re-verify their response payloads field-by-field; the per-feature spec gate does that downstream.
- Deep visual pixel-parity of the already-DONE screens — out of scope; this mission completes partial/stub/missing screens, it does not re-review shipped ones.
- `content-builder` internals — `ContentBuilderPage` is a thin wrapper delegating to a shared view; classified "deferred/wrapper", not a target screen. If it surfaces as a real gap during execution → HS-6.
- The two HTML-only design mocks slated for CUT were not deeply read (no implementation intent).
