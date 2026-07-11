# F5.3 — Signoff Reconcile · Close-Out Evidence

> Feature: `f5.3-signoff-reconcile` · Milestone 5 · 2026-07-10.
> HS-4 remediation of the milestone-validator FAIL on F5.1. **Verify-only** (no rebuild, no behavior
> change). Fresh worktree `lucid-heisenberg-d6c94f`, branch `claude/lucid-heisenberg-d6c94f`.
> **Not pushed** — awaiting operator HS-1.
> Orchestrator authorization (2026-07-10) conditions honored: (i) old F5.1 evidence annotated
> superseded, not deleted; (ii) current surface re-proven with runnable tests + 2 reviewer APPROVEs;
> (iii) zero behavior change — real defects → separate chip.

## Root cause (why F5.1 FAILed the validator)

F5.1 closed against a standalone approval **cockpit** (`SignoffDetailPage.tsx` mounting
`ControlledDocumentDetailPanel`, reusing `ApprovalTimelinePanel` + `SignoffDialog`) at
`/approvals/:documentId`. The parallel `approval-remediation` M2d program then shipped **ADR 0080**
("single artifact destination", `0c96dfb2`, 2026-07-07): the cockpit route became a redirect and the
decision surface moved into the mode-adaptive document workspace; F2d.7 deleted the cockpit files.
F5.1's evidence still pointed at the deleted files → validator C2/C4/C5/C6 FAIL. **Same supersession
class as F5.2/FE-14, but never reconciled.** The signoff *function is live* (relocated) and *0 backend
changed*, so the fix is reconciliation, not a rebuild.

## What changed in this feature (the entire diff)

| Change | File | Nature |
|---|---|---|
| Superseded banner (history kept) | `f5.1-signoff-detail/evidence.md` | docs |
| Objective + status + F5.3 row reconciled to ADR-0080 truth | `milestone.md` | docs |
| New consumer-contract spec | `f5.3-signoff-reconcile/spec.md` | docs |
| This evidence | `f5.3-signoff-reconcile/evidence.md` | docs |
| **New** characterization test re-proving `POST /signoff` + If-Match + 412→stale | `features/approval/hooks/useSignoffMutation.test.tsx` | test-only |
| Removed orphan (last dead trace of the deleted cockpit) | `features/approval/pages/SignoffDetailPage.module.css` (deleted) | dead-asset removal |

**No product-source `.tsx`/`.ts` behavior was touched.** The only code artifacts are a new test and
the removal of an unreferenced CSS file (grep-confirmed: no import; the sole surviving `SignoffDetailPage`
mention is a code *comment* in `documents/lib/documentSignoffDecision.ts`). Zero behavior change; 0 backend.

## Acceptance (spec.md) — every row proven

| Criterion | Proof | Outcome |
|---|---|---|
| F5.1 evidence carries superseded banner + pointer (history preserved) | Banner prepended atop `f5.1-signoff-detail/evidence.md`; original content untouched below the `---` | **PASS** |
| Milestone objective/status reflect the workspace surface | `milestone.md` objective bullet + status block + new F5.3 row reconciled | **PASS** |
| Current surface re-proven by runnable tests | `pnpm vitest run` over the 6 current-surface suites → **81/81** (see below) | **PASS** |
| New `useSignoffMutation.test.tsx` asserts the real endpoint contract | 3/3: (a) `signoff('doc-1', {…,content_hash:'sha256:abc'}, {ifMatch:'"v3"'})`; (b) 412 `ApprovalError`→`SignoffError{kind:'stale',stale:true}`; (c) invalid-signature→`SignoffError` | **PASS** |
| Orphan cockpit CSS removed; no ref to deleted cockpit files | `git rm SignoffDetailPage.module.css`; `grep SignoffDetailPage.module` → 0 imports | **PASS** |
| `tsc --noEmit` clean; zero behavior change; 0 backend | `tsc` EXIT=0; diff is docs+test+dead-CSS only; no Go source touched | **PASS** |
| Both reviewers APPROVE the current surface | `frontend-screen-reviewer` + `frontend-code-reviewer` — see "Reviewer verdicts" | **PASS** (recorded below) |

## Gate commands (re-runnable)

```
cd frontend/apps/web && pnpm.cmd tsc --noEmit                                   # EXIT=0
cd frontend/apps/web && pnpm.cmd vitest run \
  src/features/approval/hooks/useSignoffMutation.test.tsx \
  src/features/documents/pages/DocumentWorkspacePage.test.tsx \
  src/features/approval/components/sidebar/__tests__/DecisionFooter.test.tsx \
  src/features/documents/components/workspace/WorkspaceSidebar.test.tsx \
  src/features/approval/lib/signoffErrors.test.ts \
  src/features/approval/pages/InboxPage.test.tsx                                # 81/81
grep -rn "SignoffDetailPage.module" frontend/apps/web/src                       # 0 (orphan gone)
```

Result captured this session: `tsc` **EXIT=0**; the 6 suites **81/81** (useSignoffMutation 3, DocumentWorkspacePage 26,
DecisionFooter 12, WorkspaceSidebar 8, signoffErrors 12, InboxPage 20).

## Honest disclosures (not masked)

- **Live end-to-end signoff (click Aprovar → real `POST /signoff` → transition) not driven this
  session** — same empty-dev-DB seed gap F5.1 originally disclosed (no seeded `under_review` doc with an
  eligible non-author approver). The path is covered by the added `useSignoffMutation.test.tsx` (endpoint
  + If-Match contract) and `DocumentWorkspacePage.test.tsx` (panel render/gate/preselect), not by a
  fabricated runtime value. No invented data recorded.
- **Pre-existing coverage gaps beyond the added test** (`ArtifactDecisionPanel` has no direct unit test;
  `approvalApi.test.ts` still doesn't exercise `signoff()` at the transport layer) are **not F5.1/F5.3
  defects** — surfaced honestly; if the operator wants them closed they are a separate coverage chip, not
  in-scope for this verify-only reconcile (condition iii).

## Reviewer verdicts (on record)

- **frontend-screen-reviewer: APPROVE** — no Criticals, no Majors. Independently reproduced the diff
  scope (`git status` — zero product `.tsx`/`.ts` touched), `tsc` EXIT=0, the 6 suites **81/81** (exact
  per-file split matched), and traced the full runtime path routes.tsx→DocumentWorkspacePage→
  WorkspaceSidebar→DecisionFooter→ApprovalModeFooter→ArtifactDecisionPanel→useSignoffMutation→
  signoffErrors against spec.md file-for-file. Confirmed ADR 0080's own "Closure" section corroborates
  the current-surface map (reconciliation matches the ADR record, not an invented narrative), the
  superseded-banner pattern is correct (history preserved verbatim below the fold), and the live-e2e
  seed-gap is honestly disclosed (mirrors F5.1, not a regression). Minors: evidence "Reviewer verdicts"
  placeholder (this section — now filled); pre-existing coverage gaps correctly deferred to a separate
  chip per anti-circle.
- **frontend-code-reviewer: APPROVE** — no Criticals, no Majors. Independently ran
  `pnpm vitest run useSignoffMutation.test.tsx` → 3/3 and `tsc --noEmit` → exit 0; confirmed the test is a
  real characterization test (mock via `importOriginal` stubs only the `approvalApi.signoff` network
  boundary, leaving `ApprovalError`/`SignoffError`/`mapSignoffError` as real code paths, so it exercises
  the actual `content_hash`-from-closure body, the actual `"v3"` If-Match template, and the actual
  412→`stale` classification), that the CSS deletion is safe (grep 0 refs; surviving `SignoffDetailPage`
  mention is a doc-comment), and that the reconciliation does not overclaim. Minors (both addressed/
  deferred): test-3 assertion strengthened to `toMatchObject({ kind: 'bad_password' })` this session
  (re-ran → 3/3 green); pre-existing `ArtifactDecisionPanel`/`approvalApi.test.ts` gaps deferred to a
  separate coverage chip per anti-circle.

## Disposition

F5.1's stale evidence is reconciled to ADR-0080 runtime truth: history preserved with a superseded
banner, the current workspace signoff surface re-proven by 81/81 runnable tests (incl. a new guard on
the `POST /signoff` + If-Match contract), the orphan cockpit CSS removed, `tsc` clean, zero behavior
change, 0 backend regressions. **Ready for the M5 milestone-validator re-run**, then the HS-1 operator
gate. Not pushed.
