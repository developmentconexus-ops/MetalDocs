# Feature F2.4 (T2) — Cockpit Approver-Only: Evidence

> **Milestone:** 2 · **Task:** `T2` · **Closed:** 2026-07-07
> **Finding:** 7 · **Rules:** R1 (author submits from authoring context), R5.

## What was implemented
The document approval cockpit no longer offers any submit affordance — submitting for
review lives exclusively on the document editor. Removed, across three layers:

- **`documents/lib/approvalWorkflow.ts`** — `TransitionPolicy.actions.submit` field deleted;
  every `TRANSITION_POLICY[state].actions` entry dropped `submit` (draft no longer grants it).
- **`documents/adapters/useDocumentApprovalArtifact.ts`** — deleted the `submit` action push,
  the `openSubmit` handler in `DocumentApprovalHandlers`, the cold-submit `'"v0"'` `etagCache`
  seed on the 404 branch, and the now-unused `etagCache` import. Cockpit emits only
  cancel → publish + the DecisionPanel signoff.
- **`approval/pages/SignoffDetailPage.tsx`** — removed `showSubmit` state, `openSubmit` handler
  wiring, and the `showSubmit`/`onCloseSubmit` props to the extras slot.
- **`approval/components/DocumentApprovalExtras.tsx` (+ `.module.css`)** — deleted the inherited
  route-picker (`listRoutes`, gated on `route.manage` authors lack) and the cold
  `submit({route_id, content_hash})` block. Extras = integrity / lock / timeline only.

## Verification

| Check | Command / action | Result |
|-------|------------------|--------|
| No submit in policy | grep `submit` in `approvalWorkflow.ts` `TRANSITION_POLICY` | **NONE** |
| No cold-submit seed | grep `"v0"` / `etagCache` in adapter | **NONE** |
| FE suite | `make test` | 751 pass |
| Touched suites | `useDocumentApprovalArtifact.test.tsx` | PASS (submit assertions removed) |

### LIVE QA (preview :4173, real API)
- **Draft cockpit** — PO-RH-003 (draft): cockpit rendered **zero** submit affordances
  (no route-picker, no "Submeter" button). Screenshot captured.
- **under_review cockpit** — showed only sign / reject / cancel + timeline; no submit.

## Acceptance vs spec

| Criterion | Met? | Evidence |
|-----------|------|----------|
| Cockpit renders no submit affordance in draft | yes | live draft screenshot (zero submit) |
| Route-picker + cold-submit removed | yes | `DocumentApprovalExtras` diff |
| `TRANSITION_POLICY` loses `submit` | yes | `approvalWorkflow.ts` diff |
| Cockpit = sign/reject/cancel/timeline | yes | live under_review cockpit |

## Review disposition
- Spec-compliance: PASS — R1 held (submit only from editor); removed the `route.manage`-gated
  picker authors could never use anyway (dead affordance).
- Code-quality: PASS — deletions clean, no dangling `etagCache`/`openSubmit` references; tests updated.

## Bounded defers
_None._
