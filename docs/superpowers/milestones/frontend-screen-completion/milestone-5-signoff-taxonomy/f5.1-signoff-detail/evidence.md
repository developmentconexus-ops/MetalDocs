> ⚠️ **SUPERSEDED BY ADR 0080 (2026-07-07, commit `0c96dfb2`) — history preserved, proof relocated.**
> The standalone Detalhe Signoff **cockpit** this document closes (`SignoffDetailPage.tsx` mounting
> `ControlledDocumentDetailPanel`, reusing `ApprovalTimelinePanel` + `SignoffDialog`, at route
> `/approvals/:documentId`) **no longer exists.** ADR 0080 ("single artifact destination") retired
> the cockpit pattern: `/approvals/:documentId` now redirects to `/documents/:id`, and the sign-off
> decision surface was **relocated into the mode-adaptive document workspace** (`DocumentWorkspacePage`,
> *approving* mode). The files this evidence proves against — `SignoffDetailPage.tsx`,
> `ControlledDocumentDetailPanel.tsx`, `SignoffDialog`, `ApprovalTimelinePanel`, and their tests — were
> deleted/forked by ADR 0080 + F2d.7.
>
> **F5.1's *objective* still holds** (a reviewer from the inbox reaches a real screen rendering live
> approval data and records a decision through the real sign-off endpoint — no mock, no dead-end); only
> its *implementation* changed. This record is **kept for history, not re-run.** The current-state proof
> lives in **`../f5.3-signoff-reconcile/evidence.md`** (reconciled 2026-07-10). Do not cite the gates
> below as current — they reference deleted files.
>
> ---

# F5.1 — Detalhe Signoff · Close-Out Evidence

> Feature: `f5.1-signoff-detail` · Milestone 5 (Detalhe Signoff + Taxonomy Admin restyle)
> Execution: `superpowers:subagent-driven-development` (fresh implementer + two-stage review
> — spec-compliance then code-quality — per task), this session, 2026-06-23.
> Branch: `main` (milestone work commits directly to `main` per established repo pattern;
> **not pushed** — awaiting operator HS-1).

## Commit trail (this feature)

Implementation spans `43343c8c` … `d75ee94d` on `main` (base before feature: `4f8e5c3b`).

| SHA | Task | What |
|-----|------|------|
| `43343c8c` | T1 | `commentPlainText` pure helper + test |
| `08320e3f` → `edb38606` | T2 | `useActiveDocumentContextQuery`; **fix**: re-keyed under `['approval']` so SignoffDialog invalidation refetches it |
| `9d68c596` → `cabf606f` | T3 | additive `autoOpenSignoff`/`initialSignoffDecision` props on `ControlledDocumentDetailPanel`; **fix**: `clearAllMocks` to honor one-shot spy contract |
| `58c3f882` → `3e75190c` | T4 | `SignoffDetailPage` cockpit; **fix**: explicit `content_hash` guard, context-error state, token bg, M4 (gate comments query) correctly skipped — hook has no `enabled` arg |
| `42cc7bbd` | T5 | register `/approvals/:documentId` route |
| `de19ad31` → `829db6b9` | T6 | inbox approve/reject navigates to cockpit, in-inbox modal retired; **fix**: drop dead `signoff` import/mock |
| `2ca9514c` | T7 | backlog row for deferred diff tab |
| `6a0da2fb`, `d75ee94d` | T8 | review fixes — `QK.approval.activeDocument`, wine tokens, full-bleed layout, NOTES audit, spacing/radius tokens, defer backlog |
| `f0cec18d` | T8 | this close-out evidence |
| `840154e0` | close-out | final whole-impl review fixes — complete tab ARIA triad (M4), cover context-error + populated-comments branches (M6) |

## Validation Gate (spec.md §"Validation Gate") — every row proven

| Acceptance criterion | Proof | Outcome | Real vs fixture |
|----------------------|-------|---------|-----------------|
| Route `/approvals/:documentId` mounts the cockpit (left A4 + right decision panel) | `SignoffDetailPage.test.tsx` (4 tests) + `routes.tsx` registration | **PASS** — 4/4 | fixture (vitest, mocked query layer) |
| A4 embeds the **real** PDF from `GET /documents/{id}/view`; honest loading on `pdf_status` pending | `SignoffDetailPage.test.tsx` — "embeds the rendered PDF when ready" asserts `iframe[title="Pré-visualização do documento"].src === view url`; "honest A4 pending state" asserts "Gerando visualização do documento…" | **PASS** | fixture |
| Decision surface = mounted `ControlledDocumentDetailPanel` (not a new form); Assinar present for `under_review` | `SignoffDetailPage.test.tsx` — "mounts the decision panel (Assinar present)" | **PASS** | fixture |
| Recording a decision fires `signoff(documentId,…)` with `If-Match` + cache invalidation | reuse of existing `SignoffDialog`/panel — **no second `signoff(` call site added**. Grep `function SignoffDialog\|function ApprovalTimelinePanel` over `features/approval` → only the original component files | **PASS** — no fork | fixture + grep |
| Inbox approve/reject navigates to `/approvals/:documentId` (modal path removed) | `InboxPage.test.tsx` — "approve … navigates … decision=approve", "reject … decision=reject"; `dialogState` + `SignoffDialog` mount removed | **PASS** — 15/15 | fixture |
| Comentários tab live from `GET /documents/{id}/comments` | `SignoffDetailPage.test.tsx` comments branch (`useDocumentCommentsQuery`); `commentPlainText` flattens ProseMirror content | **PASS** | fixture |
| "Mudanças vs vX" honest deferred + backlog row with trigger | Grep `em breve\|TODO\|MOCK_\|FIXME` in `SignoffDetailPage.tsx` → **0 matches**; `wiki/backlog/detalhe-signoff.md` has the diff row + unblock trigger | **PASS** | grep + file |
| No forked timeline/decision/sign-off; generated types consumed directly | Grep (above) → only canonical files; `tsc --noEmit` clean | **PASS** | grep + `tsc` |
| Visual parity with `detalhe-signoff.html` | `frontend-screen-reviewer` (re-review after fixes) | **APPROVE** (with non-blocking nits) — see below | real (reviewer) |
| Architecture / maintainability | `frontend-code-reviewer` | **APPROVE** (with nits) — see below | real (reviewer) |
| Type + test health | `tsc --noEmit`; `vitest run` new + touched suites; broader approval suite | **PASS** — tsc EXIT=0; new+touched **33/33** (page suite 6/6 after M6); approval suite **105/105** | real |

## Reviewer verdicts (on record)

- **frontend-code-reviewer:** **APPROVE WITH NITS.** No Criticals. Majors all isolated (query-key
  central-registry, inbox imperative-fetch, panel >400 LOC, panel `window.prompt`) — the F5.1-owned
  one (query key) **fixed** (`QK.approval.activeDocument`, `6a0da2fb`); the rest are pre-existing
  shared-component debt, deferred with triggers (below). Praised: clean no-fork assembly, typed
  `decision` search-param pipeline, exemplary `commentPlainText` tests, guarded Node-v26 localStorage shim.
- **frontend-screen-reviewer:** initial **REQUEST CHANGES** → after fixes **APPROVE WITH NITS.**
  Verified resolved: `QK.approval.activeDocument` (key starts `'approval'`), wine tokens
  (`--brand` active tab, all fallbacks exact-matched to `tokens.css`), full-bleed layout,
  `:focus-visible`, complete Keep/Cut/Defer `NOTES.md`. Confirmed every deferral honestly recorded
  with a trigger. Remaining nit (spacing/radius tokens) **folded in** (`d75ee94d`, exact-match values
  only; genuine non-matches left raw rather than mapped to a wrong token).

## Honest defers (with triggers)

- **"Mudanças vs versão anterior" diff tab** — no document-diff backend exists; tab renders an honest
  explanation, **no fabricated diff**. Trigger + owner in `wiki/backlog/detalhe-signoff.md` (a
  `GET /documents/{id}/diff` endpoint). Per spec non-goal "no faked diff".
- **Design "Trilha" tab — CUT** (not missing): the approval timeline is already rendered by
  `ApprovalTimelinePanel` **inside** the mounted panel; a 4th tab would fork timeline rendering
  (forbidden). Recorded in `NOTES.md` with a re-enable path.
- **Rich-header kicker / deadline / author row — CUT**: depend on stage-count / due-date / submitter
  data not exposed by the consumed queries; rendering them would fabricate values (spec non-goal
  "no fabricated header fields"). Recorded in `NOTES.md` with the data trigger.
- **Pre-existing shared-component debt** (SignoffDialog mojibake; panel `window.prompt`; panel
  non-`resolveErrorMessage` catches; panel raw-hex CSS; panel `<span class=label>` a11y; panel
  >400 LOC; `useDocumentPdfStatus` `useEffect`-polling; inbox imperative `getActiveDocumentContext`):
  surfaced by reviewers, **not introduced by F5.1**, and live in shared tested components the plan's
  no-fork mandate (HS-2) reserves for a follow-up that owns them. F5.1 only **mounted/additively
  extended** them. All recorded with triggers + owners in `wiki/backlog/detalhe-signoff.md` for the
  operator's HS-1 decision.

## Live end-to-end gap (honest)

A real end-to-end sign-off (click Aprovar → `signoff` → state transition) was **not** executed: the
dev DB is empty (no seeded `under_review` document) — consistent with the spec's stated gating. Per
spec, this path is covered by the **reviewer drive + the reused, separately-tested `SignoffDialog`
suite (13/13 in the approval run)**, not by a fabricated value. No invented runtime data was recorded.

## Environment repair (disclosed, not a feature change)

Commit `de19ad31` also touched `vitest.config.ts` + `vitest.setup.ts`: a **guarded** Node v26
`localStorage`/`sessionStorage` shim. Root cause: Node v26 defines these globals as `undefined`
(experimental Storage API), shadowing vitest's jsdom re-population; `InboxPage`'s import chain
(`approvalApi` → `client.ts` → `apiTrace.ts`) reads `localStorage` at module-load, so its suite could
not run without the patch (earlier F5.1 suites don't hit that import chain, which is why only this one
needed it). The patch is a **no-op** where storage already works (`typeof === 'undefined' || null`
guard). Verified regression-free: approval suite **103/103** + unrelated localStorage suites
(`LoginPage`, `AppRoot`) **16/16** pass.

## Gate commands (re-runnable)

```
cd frontend/apps/web && pnpm.cmd tsc --noEmit          # EXIT=0
cd frontend/apps/web && pnpm.cmd vitest run \
  src/features/approval/lib/commentPlainText.test.ts \
  src/features/approval/components/ControlledDocumentDetailPanel.test.tsx \
  src/features/approval/pages/SignoffDetailPage.test.tsx \
  src/features/approval/pages/InboxPage.test.tsx        # 31/31
cd frontend/apps/web && pnpm.cmd vitest run src/features/approval   # 103/103
```

## Final whole-implementation review (close-out gate)

`frontend-code-reviewer` over the full feature diff (`4f8e5c3b..f0cec18d`): **APPROVE**, no Criticals.
Majors M1/M2/M3/M5 = pre-existing shared-component / codebase-wide debt, correctly attributed in
`wiki/backlog/detalhe-signoff.md` with triggers (no regression introduced). Two net-new F5.1 gaps —
**M4** (incomplete ARIA tab pattern — `role="tablist"`/`tab` present but no `tabpanel`/`aria-controls`
linkage) and **M6** (uncovered `contextQuery.isError` + populated-comments branches) — were **fixed,
not deferred** (`840154e0`): full `tab`→`tabpanel` ARIA triad + `aria-labelledby` + `tabIndex={0}`;
two added tests (context-error `role="alert"`, populated Comentários via `commentPlainText`). Re-verified:
tsc EXIT=0; `SignoffDetailPage` suite **6/6**; approval suite **105/105**, 0 failures.

## Disposition

All Validation-Gate rows **PASS**; per-feature reviewers + final whole-impl reviewer **APPROVE**;
the two net-new gaps the final review surfaced are **fixed and re-verified**; honest defers recorded
with triggers; no fabricated data; no fork; scope held to consumer-side assembly + additive props.
**Ready for the HS-1 operator gate.** Not pushed.
