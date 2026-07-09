# Feature F2d.3 — Spec

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Folder:** `f3-workspace-mode-machine`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-09 / operator (milestone pre-approved HS-1 2026-07-08; contract
> fully pinned by the milestone F2d.3 row + F2d.1/F2d.2 ADR-shaped types — no product-fork remained,
> see Interview record) — *no implementation begins until this line is filled.*

> Feature contract, approved **before any code**. Defines what the feature must do and how it is
> proven — not how it is built (that is `plan.md`). Validator judges against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Mode enum + which sub-derivation is subject-agnostic? | **none needed** — pinned verbatim by the milestone F2d.3 row: modes = `author-editing \| author-waiting \| author-changes-requested \| reviewing \| approving \| observing \| lifecycle`; the **stage-mode** helper (`reviewing/approving/observing` from `stage_kind` + `viewer.eligible_for_active_stage`) is subject-agnostic; the **lifecycle** sub-derivation (author modes + `lifecycle`) is subject-specific. |
| 2 | Source signal for `author-changes-requested` (the milestone says "keyed on document status", but which status)? | **code-grounded correction, not a product fork.** A `request_changes` verdict "immediately collapses the instance to `changes_requested` and reverts the document to **draft**" (OpenAPI note, generated types line 1734). So the document is `draft` in BOTH fresh-draft and post-reject cases — document status alone cannot split them. The distinguishing signal is `instance.status === 'changes_requested'` (`ApprovalInstanceByDocumentResponse.status`, generated types line 3348). The lifecycle sub-derivation therefore keys on document status **plus** `instance.status` for the changes-requested split. Recorded here as the precise contract. |
| 3 | Precedence when `is_author` and an active stage coexist? | **none needed** — SoD author-exclusion (F2d.1) guarantees an author is never `eligible_for_active_stage` on their own document. `is_author`-first and stage-first therefore converge; the selector uses **`is_author`-first** for clarity (author always sees the lifecycle/author lens, never `reviewing`/`approving`). |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** the single working screen `DocumentWorkspace` (F2d.5) and `DecisionFooter` — they
  render ONE mode's panels/CTAs. They call `deriveWorkspaceMode(doc, instance, viewer)` and switch on
  the returned `WorkspaceMode` string. No consumer re-derives eligibility.
- **Contract:** a pure, side-effect-free function

  ```ts
  type WorkspaceMode =
    | 'author-editing' | 'author-waiting' | 'author-changes-requested'
    | 'reviewing' | 'approving' | 'observing' | 'lifecycle';

  // stage-mode sub-derivation — SUBJECT-AGNOSTIC (survives M3 kernel/templates reuse):
  //   only stage_kind + the server eligibility bool. No document/subject knowledge.
  function deriveStageMode(
    stageKind: 'review' | 'approval' | undefined,
    eligibleForActiveStage: boolean,
  ): 'reviewing' | 'approving' | 'observing';

  // top-level selector — subject-specific lifecycle wrapper around the agnostic helper:
  function deriveWorkspaceMode(
    doc:      { status: DocumentStatus },
    instance: ApprovalInstanceByDocumentResponse | null,
    viewer:   ApprovalInstanceByDocumentResponse['viewer'] | null,
  ): WorkspaceMode;
  ```

  **Behaviour (every branch is a named unit test in the Gate):**
  - `viewer.is_author === true` → lifecycle/author sub-derivation:
    - `instance.status === 'changes_requested'` → `author-changes-requested` (document is `draft`; the
      screen shows the request-changes panel *and* allows editing).
    - `doc.status === 'draft'` → `author-editing`.
    - `doc.status === 'under_review'` (instance `in_progress`) → `author-waiting`.
    - `doc.status ∈ {approved, scheduled, published, superseded, obsolete}` → `lifecycle`.
  - `viewer.is_author === false` **and** an active stage exists
    (`instance.stages.find(s => s.status === 'active')`) →
    `deriveStageMode(activeStage.stage_kind, viewer.eligible_for_active_stage)`:
    - `review` + eligible → `reviewing`; `approval` + eligible → `approving`; not eligible (either kind)
      → `observing`.
  - otherwise (non-author, no active stage): `doc.status ∈ terminal set` → `lifecycle`; else → `observing`.
  - `viewer == null` or `instance == null` (no approval yet): a draft the caller authored →
    `author-editing`; otherwise `observing`. (Draft is the only status reachable with no instance.)
  - **`via_delegation_from` does not change the mode** — eligibility is already folded into
    `eligible_for_active_stage` server-side; the delegation object is surfaced by the screen (F2d.5),
    not consumed here.
- **Source of truth for the contract:** the milestone F2d.3 row + the generated types
  `frontend/apps/web/src/lib/api-types/index.d.ts` (`viewer` block line 3359, `status` enum line 3348,
  `ApprovalStageInstanceResponse.stage_kind` line 3399 / `.status` line 3392) — ADR 0078/0079 shapes.
  The selector reads server facts only; it never re-implements eligibility.

## What this feature implements

- **New pure module** `frontend/apps/web/src/features/approval/lib/workspaceMode.ts` exporting
  `WorkspaceMode`, the subject-agnostic `deriveStageMode`, and `deriveWorkspaceMode`. No React, no I/O.
- **Replaces** the correct-but-duplicated `resolveEditorMode` (`ApprovalCockpitPage.tsx:41-55`) and the
  broken status/policy-only `signoffOffered` (`useDocumentApprovalArtifact.ts:205-206`) as the single
  source of action/mode truth. (Deletion of the old call sites lands with their consumers in F2d.5/F2d.7;
  this feature ships the selector + its tests and rewires `signoffOffered`'s derivation to the viewer
  truth where it is read today, removing the status-only gate.)
- **`TRANSITION_POLICY` shrinks to document-lifecycle actions only** (`approvalWorkflow.ts:42-80`): the
  `signoff` action leaves the policy (it is a stage action, now derived from the mode + viewer);
  `cancelInstance` and `publishOrSchedule` (document-lifecycle) stay.

## Non-goals (mandatory)

- **No new server contract, no OpenAPI change** — pure FE consumer of the F2d.1/F2d.2 shapes.
- **No screen/component build** — that is F2d.5. This feature ships only the selector + unit tests
  (and the `TRANSITION_POLICY`/`signoffOffered` derivation rewire that removes the status-only gate).
- **No change to document-lifecycle capability gates** (`canPublish`/`canRevise`/`isObsolete`/
  `isPublished`/`isApproved` in `useDocumentArtifact.ts` — those are *lifecycle* gates keyed on
  document status by design, not stage-eligibility; they are out of the grep-gate scope and stay).
- **No `via_delegation_from` mode influence** (surfaced by F2d.5, not the selector).
- **No React-query / fetch-state change** — that is F2d.4.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `author-editing` — is_author, draft, no instance | `deriveWorkspaceMode › author-editing` | real (pure fn) |
| `author-waiting` — is_author, under_review, instance in_progress | `… › author-waiting` | real |
| `author-changes-requested` — is_author, `instance.status==='changes_requested'` (doc draft) | `… › author-changes-requested` | real |
| `reviewing` — non-author, active review stage, eligible | `… › reviewing` | real |
| `approving` — non-author, active approval stage, eligible | `… › approving` | real |
| `approving via delegation` — eligible via `via_delegation_from`, mode still `approving` | `… › approving via delegation` | real |
| `observing` — non-author, active stage, NOT eligible | `… › observing (ineligible on active stage)` | real |
| `observing` — non-author, draft / no active stage | `… › observing (non-author draft)` | real |
| `lifecycle` — approved/published terminal doc status | `… › lifecycle (terminal status)` | real |
| SoD-excluded author never `approving` — is_author dominates | `… › author dominates over stage` | real |
| subject-agnostic helper is pure `stage_kind`+bool (no doc arg) | `deriveStageMode › maps kind×eligible` (helper called with no document) | real |
| no FE stage-eligibility derivation outside the selector | grep gate: no `signoffOffered`-style status/policy-only stage gate remains; `signoff` removed from `TRANSITION_POLICY` | real |
| build/types clean | `npm run typecheck` (or `make test` scope) + `make test` for the selector suite | real |

> TDD: write the failing `workspaceMode.test.ts` branch cases first, then implement to green.

## ADR needed?

- [x] No **new** durable decision — the governing decision ("FE renders server-derived facts, never
  derives eligibility client-side") is already **ADR 0078** (F2d.1). This feature is the FE-side
  application of that contract; the code-grounded `changes_requested`-signal clarification is recorded
  in this spec's Interview record, not a separate ADR.
