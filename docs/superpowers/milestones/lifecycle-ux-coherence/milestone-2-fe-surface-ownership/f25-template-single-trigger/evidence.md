# Feature F2.5 (T3) — Template Single Submit Trigger: Evidence

> **Milestone:** 2 · **Task:** `T3` · **Closed:** 2026-07-07
> **Finding:** 14 · **Rules:** R1 (author submits from authoring context), R5.

## What was implemented
The template approval cockpit no longer emits a draft submit action — submitting a
template version for review lives solely in the template editor.

- **`templates/lib/templateApprovalActions.ts`** — deleted the `draft` case (the
  `canSubmit` gate + `submit` action push), dropped `runSubmit` from
  `TemplateApprovalHandlers`, and removed the `canSubmit` import. Cockpit builds actions
  only for `under_review` (review/reject) and `approved` (publish/reject).
- **`templates/pages/TemplateApprovalRoute.tsx`** — deleted the `runSubmit` handler and the
  `submitForReview` import; draft versions now yield no cockpit actions.

## Verification

| Check | Command / action | Result |
|-------|------------------|--------|
| No draft submit in cockpit | grep `canSubmit`/`runSubmit`/`submitForReview` in `templates/lib` + `pages/TemplateApprovalRoute.tsx` | **NONE** |
| FE suite | `make test` | 751 pass |
| Touched suites | `TemplateApprovalRoute.test.tsx`, `templateApprovalActions.test.ts`, `useTemplateApprovalArtifact.test.tsx` | PASS (submit cases removed) |

> `submitForReview` remains exported from `templates/api/templates.ts` — it is the client
> the **template editor** calls (the surviving single trigger). Only the cockpit's duplicate
> entry point was removed, per R5 (one impl, N entry points).

## LIVE QA note
No template versions were seeded in the dev DB this session, so the draft-cockpit state
was not driven live. Covered by the unit suites above (draft case now emits no `submit`
action; route no longer wires `runSubmit`). Bounded defer below.

## Acceptance vs spec

| Criterion | Met? | Evidence |
|-----------|------|----------|
| Draft submit action removed from cockpit | yes | `templateApprovalActions.ts` + route diff |
| Template editor is sole template-submit surface | yes | `submitForReview` client retained, only cockpit caller removed |
| Tests green | yes | 751 pass; template suites updated |

## Review disposition
- Spec-compliance: PASS — R1/R5 held; the editor's `submitForReview` client is untouched.
- Code-quality: PASS — no dangling `runSubmit`/`canSubmit` references; switch simplified.

## Bounded defers
| Defer | Why bounded | Trigger / owner |
|-------|-------------|-----------------|
| Draft-cockpit "no submit" not driven live (no seeded template versions) | unit suites assert the draft case emits no submit action + route drops the handler | drive live when a draft template version exists; owner templates cockpit |
