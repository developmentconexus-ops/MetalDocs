# R10 Rebaseline Cockpit — Design

> **Status:** ACTIVE STAGING / APPROVED SUPPORT-ARTIFACT DESIGN  
> **Approved:** 2026-08-19  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Product implementation:** BLOCKED

## Goal

Create one self-contained visual orientation cockpit that lets the operator recover MetalDocs R10 context in minutes: what the platform is converging toward, what has been ratified, where the program is now, what remains before implementation, and the exact prompt to start a fresh session.

## Authority law

The cockpit is **never architecture authority**. It must visibly state:

> **VISUAL ORIENTATION / PROJECTION — NOT ARCHITECTURE AUTHORITY.**

The sole stage/status/next-action authority remains `wiki/architecture/r10-technical-architecture.md`. Durable decisions remain in `wiki/`. The cockpit is updated only after those authorities change.

## Delivery

- Create: `docs/operator/r10-rebaseline-cockpit.html`
- Standalone HTML/CSS/vanilla JavaScript.
- No framework, package, build step, CDN, webfont, remote image or runtime dependency.
- Works from `file://` and ordinary static hosting.
- Responsive desktop/mobile layout and print stylesheet.
- Dark technical visual language: graphite/black base; green closed; amber active; gray blocked/future; red warning/rejected; blue information/mechanism.
- Last-synchronized date displayed as `2026-08-19`.
- Do not embed a commit SHA that becomes stale on the next cockpit commit; new-session prompt must require fresh PR/HEAD revalidation.

## Required sections

1. **You Are Here / Executive State**
   - `T1→T7 CLOSED / OPERATOR-RATIFIED`
   - `T8-A ACTIVE / CURRENT TECHNICAL CENSUS + REMEASUREMENT NEXT`
   - `T8-B→T12 NOT OPEN`
   - `IMPLEMENTATION BLOCKED`
   - No fake completion percentage or ETA.

2. **Path to Implementation**
   - Product Contract / GCR / ownership / T1→T7 complete.
   - T8-A→T8-H, T9, T10, T11, T12, Integrated Whole-R10 GCR, independent/cold review, final operator ratification, explicit implementation authorization.
   - Progress visualization must distinguish semantic architecture from physical realization / validation / transition / execution planning.

3. **What MetalDocs Is Becoming**
   - Semantic owners: Authentication, Organization, Authorization, Controlled Documents, Audit supporting.
   - Mechanisms visually separate from semantic authority: managed content/storage, River/jobs, rendering, search, backup/restore, OIDC/Keycloak, HTTP/OpenAPI.

4. **Launch / Launch+ / Future**
   - Launch Core summary.
   - Launch+: Distribution/Read&Acknowledge; Periodic Review.
   - Future: Dossier, Evidence, Retention/Hold/disposition, WORM/records, governed export, external repositories, Training/LMS, generic change control, pooled tenancy, CRDT.
   - Explicit message: deferred does not mean forgotten.

5. **Decision Memory — T1→T7**
   - Expandable cards with concise ratified outcomes.
   - T7 must state `NO HISTORICAL BUSINESS MIGRATION REQUIRED FOR LAUNCH` and current data DEV/TEST/THROWAWAY.

6. **Do Not Accidentally Reopen**
   - Standalone Artifact owner; parallel Template lifecycle; generic BPM; dormant ETL/import platform; DEV history preservation; other removed/deferred Launch complexity.
   - Reopen only for a concrete failure mode, named consumer, authority contradiction, source evidence, or regulatory/contractual requirement.

7. **Current T8-A Workbench**
   - Explain T8-A as census/disposition, not target topology design.
   - Disposition vocabulary: PRESERVE / REFINE / REHOME / REWRITE / DELETE / CURRENT-STATE ONLY / SUPERSEDED.
   - Evidence vocabulary: CURRENT-PROVEN / LAST-REPRODUCED / STALE/SUPERSEDED / UNKNOWN/REMEASURE.
   - Legacy Disposition Matrix ready to grow as census results appear.
   - Current-proven high-level evidence: one Go module; mixed Go + React/TS + SQL + Node/TS; API/worker/jobs/DOCX-renderer roots; tools/verify SSOT; CI delegates PR verification to tools/verify.
   - Old exact Aug-09 mechanical metrics, when shown, must be labeled LAST-REPRODUCED / REMEASURE.

8. **Scheduled Open Decisions**
   - T8-B backend packages; T8-C internal contracts; T8-D persistence; T8-E executable wire contract; T8-F frontend; T8-G runtime/deploy; T8-H coherence; T9 proof; T10 transition; T11 execution graph; T12 adversarial readiness.
   - Distinguish `not decided yet because owned by a future gate` from `unknown`.

9. **Authority Navigator**
   - Read order and relative links/paths for AGENTS, Method, current handoff, router, Product Contract, T1→T7, Registry amendments, post-T6 program, TRRB, active T8-A staging.
   - Current implementation/code/schema/OpenAPI/deploy/tests shown as evidence, not target authority.

10. **Architecture Journey / Why We Are Here**
    - Product Contract → T1→T6 → implementation-readiness challenge → post-T6 restructure → TRRB → T7 source clarification → no historical migration → T7 closed → T8-A active.

11. **Fresh Session Prompt**
    - One-click copy button.
    - Prompt must bootstrap from repo authority, require fresh PR/HEAD revalidation, include current stage/gates, T7 decision, active T8-A objective, implementation block, and prohibit using the cockpit/chat as authority.

## Interaction requirements

- Sticky section navigation.
- Search over decision/authority/workbench content.
- Filter controls for Closed / Active / Future-Blocked information.
- Expand/collapse decision cards.
- Copy new-session prompt with Clipboard API plus fallback that works under `file://` when possible.
- Copy authority-chain control.
- Print button / `window.print()`.
- Respect `prefers-reduced-motion`.
- Usable keyboard focus states and semantic buttons.

## Non-goals

- No product runtime code.
- No automatic authority parser/generator.
- No live GitHub/API calls.
- No attempt to calculate a misleading numeric implementation-readiness percentage.
- No second source of truth.

## Reopen / maintenance law

Update cockpit only after router/durable authority changes. A stale cockpit is corrected to the router; the router is never corrected to match the cockpit.
