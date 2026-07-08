# Milestone 2c — Approval Screen FE

> **Program:** approval-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` §8 (C1–C5)
> **Status:** Validator PASS (2026-07-07, `qa/milestone-qa.md`) — pending operator HS-1
> **Authored:** 2026-07-07 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M2c is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains
> **no execution steps** — the "how" of each feature lives in that feature's `plan.md`
> (seeded from `docs/superpowers/plans/2026-07-07-m2c-approval-screen-fe.md`). The
> end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Objective

After M2c, a reviewer or approver opens a single document destination — the **editor shell** —
and the surface adapts to the workflow: a **review** stage opens the doc in eigenpal *suggesting*
mode with a suggestion+comment sidebar and verdict CTAs; an **approval** stage opens *viewing*
(read-only) with a decision CTA that signs with meaning-of-signature. The standalone approval
cockpit page, its writable review canvas (the W2 mutation vector), the duplicated approval
timeline, the "dados desatualizados" polling banner, and every autosave hook in the approval
feature are **gone**. The author whose document returns `changes_requested` opens the editor to a
"mudanças solicitadas" panel and cannot re-submit until every tracked change and comment is
resolved — so the buffer is clean and the backend freeze markup-gate (F0) passes. `/approvals` is
the one worklist: real filters (stage kind, due, doc type), due-date sort, teaching empty state,
deep-link into the cockpit in the right mode.

**Quality bar moved:** the M2b backend kernel gains its consuming FE screen, and the approval
feature is scrubbed of the two structural defects M2b's backend closed at the contract level but
the old FE still embodied — (a) a writable editor session on an approval surface (W2), re-measured
by `grep -r "useDocumentSession\|useDocumentAutosave" frontend/apps/web/src/features/approval` →
**zero**; (b) two competing approval timelines, re-measured by the single-timeline DOM assertion in
F4's tests. Program terminal acceptance = this screen live-QA-verified on the validator-passed M2b
backend.

## Appetite

- **Appetite:** 9 features (F0 backend prerequisite + F1–F8 FE), reusing the existing document
  editor shell and the eigenpal adapter — no new editor engine, no new design system.
- **Rabbit holes (do not chase):**
  - **Delegation admin UI** — backend delegations (ADR 0077) exist; worklist/cockpit already honor
    delegated eligibility server-side. No FE surface unless the operator asks. *Reason: no consumer
    yet; server-side eligibility already covers the walkthrough.*
  - **Oversee dashboard** — oversee is a `scope=oversee` param on the existing worklist, not a new
    screen. *Reason: wrong milestone; a dashboard is unbudgeted scope with no ratified spec row.*
  - **Parallel-stage / DAG routing (W12)** — serial routing covers standard eQMS. *Reason: deferred
    in spec §10 with a written trigger (first concurrent-stage customer requirement).*
  - **Re-theming the author editor** — the author path keeps its current sidebar and visuals
    unchanged; F3 only extracts a shell around it. *Reason: zero-consumer-impact refactor must stay
    visually inert for authors.*

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0 | `f0-backend-contract-closure` | M2b carried defers: wire `InstanceStatus` enum gains `changes_requested`; route-create/update stage schema gains `stage_kind` (enum `review\|approval`, default `approval`); markup gate wired into `executeFreeze` (loads current docx via the existing published blob port, `ScanForUnresolvedMarkup` → 409 problem+json); Go + FE api-types regen. Consumer: FE api-types (`components['schemas']`) + the approval `executeFreeze` service. | `go build ./... && go test ./internal/modules/documents/approval/...` PASS incl. `TestInstanceStatusWireEnumComplete` (wire enum covers all 5 domain statuses) and a freeze integration test: dirty docx (`w:ins`) → 409, clean → hash pinned. Regenerated `frontend/apps/web/src/lib/api-types/index.d.ts` contains `stage_kind` + `changes_requested`. |
| F1 | `f1-editor-ui-track-changes` | Extend `MetalDocsEditorRef` (`packages/editor-ui`) with `getTrackedChanges`/`acceptChange`/`rejectChange`/`acceptAllChanges`/`rejectAllChanges`/`removeCommentMark` as thin passthroughs to the verified vendor API, plus `onTrackedChangesChange` prop. Consumer: F4 `SuggestionList` + F6 `RequestedChangesPanel`. | `npm run test:docx-v2` + `npm run typecheck:docx-v2` PASS incl. mount-with-tracked-insertion test: `getTrackedChanges().length > 0`, shape `{revisionId, author, type, excerpt}`, `acceptChange(id)` → length 0. |
| F2 | `f2-approval-api-layer` | `approvalApi.reviewVerdict` (POST review-verdict + `Idempotency-Key` + `If-Match`), `useReviewVerdictMutation`, `useInboxQuery` filter params (`stageKind`/`dueBefore`/`scope`), `useDelegationsMutations`. Generated `components['schemas']` types only (ADR 0035). Consumer: F4 decision footer, F5 worklist filters. | `approvalApi.test.ts` PASS: review-verdict request hits the contracted URL with `If-Match` + UUID `Idempotency-Key`; mutation invalidates `['approval','instance',documentId]` + `['approval','inbox']`. No hand-written body DTO (grep). |
| F3 | `f3-cockpit-document-shell` | Extract `DocumentShell` from `DocumentEditorPage` (chrome + editor mount + sidebar/header slots). `ApprovalCockpitPage` mounts it; mode derived from instance DTO (`stage_kind==='review'` & eligible → `review`/suggesting; `approval` → `readonly`/viewing; oversee/ineligible → readonly observer). Delete `ReviewDocumentCanvas` (+test). Consumer: reviewer/approver/oversee actors on `/approvals/:documentId`. | `ApprovalCockpitPage.test.tsx` PASS: approval mode mounts no autosave hook (jest.mock spy), review mode passes `mode="review"`, 404 `not_found.instance_not_visible` renders the not-found screen (not a toast). `grep -r useDocumentSession frontend/apps/web/src/features/approval` → zero. Author `DocumentEditorPage.test.tsx` stays green (visually inert). |
| F4 | `f4-approval-sidebar-ia` | `ApprovalSidebar` composition: `StageContextHeader` (etapa N/M, pool, `due_at` chip, quorum), single `ApprovalTimeline` (duplicate band deleted), `IntegrityDisclosure` (collapsed "Conteúdo verificado ✓ · detalhes" → hash/etag inside), mode-aware `DecisionFooter` (review: ready + solicitar-mudanças dialog; approval: aprovar-e-assinar w/ 21 CFR meaning line + rejeitar), `SuggestionList` (review cards from `onTrackedChangesChange` + comments). Consumer: `ApprovalCockpitPage` slot. | `ApprovalSidebar.test.tsx` + `DecisionFooter.test.tsx` PASS: review mode → suggestion cards + verdict CTAs, no password field; approval mode → sign+reject, no verdict CTAs; request-changes dialog submit disabled until comment non-empty; exactly ONE timeline in DOM; hash hidden until disclosure expanded; footer sticky (visible without scroll); `due_at` relative PT-BR + overdue state. |
| F5 | `f5-worklist-single-destination` | `InboxFilters` (stage_kind `Todos\|Revisão\|Aprovação`, due `Todas\|Vence em 7 dias\|Atrasadas`, oversee toggle gated on a cached `scope=oversee` probe), wired to `useInboxQuery`; due-date ascending sort; teaching empty state. Consumer: reviewer/approver/oversee actor landing on `/approvals`. | `InboxPage.test.tsx` PASS: filters map to query params; due-ascending default sort; empty-state teaching copy present; item click navigates `approvals/:documentId`; overdue badge on past `due_at`. |
| F6 | `f6-author-request-changes-panel` | `RequestedChangesPanel` mounted on `DocumentEditorPage` when active instance `status==='changes_requested'`: per-change accept/reject (`acceptChange`/`rejectChange`), per-comment resolve; re-submit blocked while tracked changes or unresolved comments remain; on re-submit `removeCommentMark` each resolved comment, flush save, then `approvalApi.submit`. Consumer: author whose doc returned from `changes_requested`. | `RequestedChangesPanel.test.tsx` PASS: lists tracked changes + unresolved comments with actions; re-submit disabled while `getTrackedChanges().length>0 \|\| unresolvedComments>0`; on re-submit `removeCommentMark` called per resolved comment → save flush → `submit` (ordered); backend 409 markup gate renders problem detail, not generic toast. |
| F7 | `f7-visual-a11y-polish` | `/impeccable audit` scoped to `features/approval/**` + new shell/sidebar files: contrast ≥4.5:1, visible focus on CTAs/filters/disclosure, `prefers-reduced-motion`, no slate tokens, loading/empty/error state per surface. Consumer: end user (visual + a11y). | `grep -r "slate-" frontend/apps/web/src/features/approval` → zero; impeccable audit findings closed; wine `--brand #6b1f2a` tokens only. |
| F8 | `f8-close` | Full suites + mandatory live QA walkthrough on the real stack + dispatch `milestone-validator`. Consumer: operator (HS-1 gate). | `make test` + `npm run typecheck:docx-v2` + `go build ./... && go test ./...` green; live QA walkthrough evidence recorded (route review→approval, suggesting, request_changes, author panel, clean buffer, freeze+markup gate 409, sign with meaning, publish, visibility matrix, oversee, cancel-with-reason); `milestone-validator` verdict written. NOT pushed. |

For each feature, "what to validate" is objectively checkable — a named test, a grep that returns
zero, a route responding with the contracted shape, a runtime behavior observed in live QA.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for M2c:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored (producer matches consumer: FE consumes the
   regenerated api-types F0 produced; F4/F6 consume the F1 adapter ref API).
2. **Workflow-class QA checklist** — frontend screen QA: `wiki/quality/qa-operating-system.md` +
   the frontend structure rules in `wiki/architecture/frontend-structure.md`; F0's backend slice
   additionally checked against `wiki/architecture/backend-api-structure.md` + `api-contract.md`.
3. **Regression** — M2b (`milestone-2b-approval-kernel-backend`) still passes its gate: re-run
   `go test ./internal/modules/documents/approval/...` and the M2b QA-cited suites from clean state
   (F0 edits the same module).
4. **Quality-bar / root-cause check** — the two structural defects are re-measured, not
   symptom-patched: `grep -r "useDocumentSession\|useDocumentAutosave" frontend/apps/web/src/features/approval`
   → zero (writable-session W2 vector gone at the root, not hidden behind a readonly flag); the
   single-timeline DOM assertion passes (duplicate band deleted, not CSS-hidden).
5. **No unplanned scope** — anything beyond the F0–F8 table is recorded with rationale; the
   rabbit-hole list above is the scope-drift baseline.
6. **Live QA is real-stack, honestly labeled** — the F8 walkthrough is executed against the running
   API + FE (not fixtures); fixture-only proof is labeled as such and does not count as the
   walkthrough.

## Dependencies & constraints

- **Depends on:** M2b passed (validator PASS + HS-1 approved 2026-07-07; ADRs 0074–0077). Backend
  landed: review-verdict endpoint, `stage_kind`/`due_at`/`frozen_content_hash`/`signature_meaning`
  DTOs, visibility 404 `not_found.instance_not_visible`, delegations, caps `approval.review` /
  `approval.oversee`. F0 closes the three carried defers before FE consumes them.
- **Quality goals (ranked):** 1) **contract truth** (generated api-types only; no hand-written body
  consumers — ADR 0035); 2) **structural correctness** (kill the writable-session vector + duplicate
  timeline at the root, not by flag); 3) **eQMS integrity UX** (no-fallback integrity display,
  meaning-of-signature, clean-buffer discipline) — ahead of visual polish, which is F7 and last.
- **Architectural constraints (hard rules the validator can fail on):**
  - Contract-first: routes/DTOs change only via `api/openapi` + `oapi-codegen`; FE consumes the
    regenerated `api-types` — no hand-written DTO body consumers (ADR 0035).
  - No-fallback principle: integrity-critical reads (`frozen_content_hash`, signature payloads) fail
    closed with a typed error; the integrity disclosure never substitutes a placeholder hash.
  - AuthZ = capabilities, never roles: FE eligibility is server-derived from the instance DTO /
    `scope=oversee` probe; never client-side role reasoning.
  - Canonical test frameworks per class: vitest for FE, testdb factory for F0 DB integration.
  - Startup only via PowerShell scripts (`.\scripts\start-api.ps1`); never read/expose `.env`.
  - Tokens: wine 100% on approval screens; PT-BR sentence case; WCAG AA.
- **Risks:**
  - *FE vitest pnpm junction drift* (memory `fe-node-modules-junction-drift`) — mitigation: if vitest
    breaks on module resolution, run a complete `pnpm install` (the real fix), not config patching.
  - *F0 blob-read path for the markup gate* — mitigation: reuse the existing published blob port used
    by render/export (locate by grep, runtime-truth), never a bespoke reader; if no such port is
    cleanly reusable → HS-2 (boundary), stop and report.
  - *eigenpal adapter behavior under review mode* — mitigation: F1 lands with a mount test against a
    real tracked-insertion fixture before F4/F6 consume it; compile ≠ works (M2b lesson).

## Applicable hard-stops

Default catalog HS-1..HS-6 in force. What trips each here:

- **HS-1** — M2c boundary: operator review gate at close; no merge/push and no program close-out
  without explicit approval. (Program terminal acceptance sits behind this gate.)
- **HS-2** — a fix implies redesign outside the boundary: e.g. the markup gate needs a blob-read
  path that doesn't exist as a reusable published port, or the shell extraction forces a change to
  the cross-module document session contract. Stop; report boundary + minimum prerequisite plan.
- **HS-3** — a prerequisite boundary fails: build broken, API not runnable, auth-session broken,
  target route missing, or contract/generated drift after regen. Repair the prerequisite, rerun the
  failed checkpoint, resume.
- **HS-4** — `milestone-validator` returns FAIL: open the named fix feature, re-run its lifecycle,
  re-dispatch the validator. Milestone stays active.
- **HS-5** — program terminal acceptance (live-QA screen on validator-passed backend) misses its
  bar: bounded remediation micro-feature, re-run acceptance; operator decides continue vs replan.
- **HS-6** — scope drift / off-plan discovery mid-milestone (e.g. a temptation from the rabbit-hole
  list gets pulled in): stop, surface the deviation, replan before continuing.
