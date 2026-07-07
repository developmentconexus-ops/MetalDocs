# Milestone 2 — FE Surface Ownership

> **Program:** `docs/superpowers/milestones/lifecycle-ux-coherence/`
> **Governing spec:** `docs/superpowers/specs/2026-07-06-lifecycle-ux-coherence-design.md`
> **Findings:** 6, 7, 8, 13, 14
> **Precondition:** M1 PASSED (HS-1 approved 2026-07-06).

## Objective (outcome)

Exactly **one** submit implementation per artifact kind, rendered **only** on the
author's surface. The document editor is the sole document-submit trigger; the
template editor is the sole template-submit trigger. The approval cockpit becomes
**approver-only** (sign / reject / cancel / timeline). Binding model rules: R1
(author submits from authoring context), R5 (one implementation, N entry points).

## Binding constraints

- **R5** — no second submit path survives. One client fn per kind.
- **Contract-first** — request/response types come from `lib/api-types`; no hand-rolled DTOs.
- **Frontend-structure conventions** — feature-sliced; API through `lib/api`; TanStack Query for server state.
- **YAGNI §4** — no new tracking screen, no speculative submit UI.
- **Global-max** — any redesign beyond M2 scope → HS-2 stop + surface, do not patch around.

## Runtime baseline (verified in code, 2026-07-06)

- Two document submit impls: `documents.ts:44 submitDocumentForReview` (manual
  apiFetch If-Match/idempotency) **and** `approval/api/approvalApi.ts:63 submit()`
  (via `mutationClient` — etag/If-Match/on412). → collapse to `approvalApi.submit`.
- `mutationClient.mutate` accepts explicit `ifMatch` opt → editor keeps its OCC
  `"v${revision_version}"` while adopting the canonical client.
- Cockpit `DocumentApprovalExtras.tsx` holds an inherited route-picker (`listRoutes`,
  gated `route.manage` authors lack) + cold `submit({route_id, content_hash})`.
- `useDocumentApprovalArtifact.ts:133` seeds `'"v0"'` etag for cold submit; `:174`
  pushes a `submit` action; `TRANSITION_POLICY.draft.submit=true` (`approvalWorkflow.ts:45`).
- Editor already migrated to `/submit` (M1 dirty tree) but lacks: `reason_for_change`
  collection for REV≥1, `isSubmitting` guard, correct "submeter" strings, success toast.
- Backend REV≥1 requires `revision_title` + `reason_for_change`; `reason_category`
  optional enum. 422 codes: `validation.reason_for_change_required`,
  `validation.reason_category_invalid`, `validation.revision_title_required`.
- Template cockpit `TemplateApprovalRoute.tsx:77-79` + `templateApprovalActions.ts`
  draft case emit a duplicate submit trigger.

## Features

| # | Slug | Findings | Outcome |
|---|---|---|---|
| F2.1 | editor-submit-unify | 8 | Single document-submit client (`approvalApi.submit`); `submitDocumentForReview` deleted; all callers on one path. |
| F2.2 | editor-reason-for-change | 6 | REV≥1 submit dialog collects `revision_title` + `reason_for_change` (required) + `reason_category` (optional); REV0 unchanged; 422 typed codes surfaced. |
| F2.3 | editor-polish | 13 | `isSubmitting` guard (no double-submit); "finalizar" strings → "submeter"; success feedback styled as success. |
| F2.4 | cockpit-approver-only | 7 | `TRANSITION_POLICY` loses `submit`; route-picker + `'"v0"'` seed + cold-submit deleted from `DocumentApprovalExtras` + adapter + `SignoffDetailPage`. Cockpit = sign/reject/cancel/timeline. |
| F2.5 | template-single-trigger | 14 | Draft submit action removed from `templateApprovalActions` + `TemplateApprovalRoute`; template editor is sole template-submit surface. |

F2.1–F2.3 are one cohesive editor change (same file/flow) executed as one implementer task (**T1**).
F2.4 (**T2**) and F2.5 (**T3**) are independent. Sequential execution, two-stage review each.

## Definition of done

- One submit impl per kind; no second path (`submitDocumentForReview` gone; template
  draft submit gone from cockpit).
- Editor REV≥1 collects reason_for_change (+ optional category); REV0 title-only path intact.
- Editor polish: submit guard + correct strings + success feedback.
- Cockpit renders no submit affordance in `draft`.
- `make test` (FE) green; vitest green for touched suites.
- LIVE QA: author submit fresh draft REV0 + REV≥1 from editor; cockpit shows no submit
  in draft; proof captured.
- milestone-validator PASS → `qa/milestone-qa.md`.
- Commits local (never pushed). HS-1 operator gate before M3.
