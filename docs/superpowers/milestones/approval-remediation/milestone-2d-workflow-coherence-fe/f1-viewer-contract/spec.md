# Feature F2d.1 — Spec

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Folder:** `f1-viewer-contract`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-08 / operator (milestone pre-approved + interview answers below) — *no implementation begins until this line is filled.*

> Feature contract, approved **before any code**. Defines what the feature must do and how it is
> proven — not how it is built (that is `plan.md`). Validator judges against this file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | `viewer.via_delegation_from` shape — id-only vs object? | **Object `{user_id, display_name}`** (nullable). Delegator display name resolved off-tx via the existing `displayNameReader.DisplayNames` port (already used by `resolveEligibleActorNames`), H-PRE-1 safe. Enables the brief §6 badge with zero extra FE lookup. (operator 2026-07-08) |
| 2 | Emit `viewer` when the instance has no active stage (published/cancelled/terminal)? | **Always present when the instance exists.** `is_author` stays real; `eligible_for_active_stage=false`, `via_delegation_from=null`, `has_signed_active_stage=false` when no active stage. One shape, selector branches cleanly. (operator 2026-07-08) |
| 3 | SoD depth for `eligible_for_active_stage`? | Full `CheckSoD` — author-exclusion (`actorUserID≠author`, `onBehalfOf≠author`) **and** cross-stage double-sign (`actorUserID ∉ priorSignoffs`), priorSignoffs = this viewer's signoffs on OTHER stages of this instance. Reflects `sod.go:38-51` exactly; NO reviewer≠approver rule added. (grounded in code — no operator ambiguity) |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):** `deriveWorkspaceMode(doc, instance, viewer)` (F2d.3) — the single FE selector; the
  single screen (F2d.5) renders facts from it. The generated TS DTO type is the contract surface.
- **Contract:** `ApprovalInstanceByDocumentResponse` (`api/openapi/v1/openapi.yaml:6436`) gains a
  **required** `viewer` object (present whenever the instance is returned):

  ```yaml
  viewer:
    type: object
    required: [is_author, eligible_for_active_stage, has_signed_active_stage]
    properties:
      is_author:                 { type: boolean }   # viewer == instance.submitted_by
      eligible_for_active_stage: { type: boolean }   # snapshot ∪ active delegation − SoD, vs active stage
      has_signed_active_stage:   { type: boolean }   # viewer already recorded a signoff on the active stage
      via_delegation_from:                            # non-null only when eligibility is ONLY via delegation
        nullable: true
        type: object
        required: [user_id, display_name]
        properties:
          user_id:      { type: string }
          display_name: { type: string }
  ```
  Semantics:
  - `eligible_for_active_stage` = there IS an active stage AND (viewer ∈ `stage.eligible_actor_ids`
    OR ∃ active `Delegation` where `delegatorID ∈ stage.eligible_actor_ids`) AND `CheckSoD` passes
    (author-exclusion + cross-stage). No active stage ⇒ `false`.
  - `via_delegation_from` = the delegator granting eligibility, resolved to `{user_id, display_name}`,
    **only** when the viewer is NOT directly in the snapshot (pure-delegation eligibility). Directly
    eligible ⇒ `null` even if a delegation also exists. Multiple qualifying delegators ⇒ the first
    active one (deterministic by `delegatorID` ascending).
  - `has_signed_active_stage` independent of `eligible` (a signed approver stays `eligible=true`,
    `has_signed=true`; the selector uses both).
- **Source of truth for the contract:** `api/openapi/v1/openapi.yaml` → `oapi-codegen` generated
  Go + TS types. No hand-written `body.data.X` consumers (ADR 0035).

## What this feature implements

Server computes viewer eligibility truth in the instance view-read path and projects it onto the DTO:
- `read_service.go:LoadInstanceByDocumentForView` (or its handler mapper `mapInstanceResponse`,
  `get_instance_handler.go:49`) gains the viewer computation. The viewer identity = the authenticated
  caller (already in ctx for the CapDocumentView gate).
- Delegation input via `LoadActiveDelegationsFor(ctx, tx, tenantID, viewerID, now)` — plain SELECT,
  H-PRE-1 safe.
- SoD via `CheckSoD(instance.SubmittedBy, viewerID, onBehalfOf, priorSignoffs)`.
- Delegator display name via the existing `displayNameReader.DisplayNames` port (off-tx), reusing the
  `resolveEligibleActorNames` machinery.
- OpenAPI edit + `oapi-codegen` regen (Go + TS). New required `viewer` object as above.

## Non-goals (mandatory)

- No kernel write-path change (freeze/signoff/verdict/submit services untouched).
- No new SoD rule — reflect exactly `sod.go` (author-exclusion + cross-stage double-sign for
  signoffs only; NO reviewer≠approver rule).
- `verdicts[]` (by-instance verdict history) is **F2d.2**, not here. Stage-actor display names
  already exist (`resolveEligibleActorNames`) — untouched here.
- FE selector + screen are F2d.3 / F2d.5.
- No migration (no schema change; read-path only).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| viewer correct: **author** (`is_author=true`, `eligible=false` by author-rule) | real-DB test `TestViewerBlock_Author` (testdb factory) | real |
| viewer correct: **snapshot actor on active stage** (`eligible=true`, `via_delegation_from=null`) | `TestViewerBlock_SnapshotActor` | real |
| viewer correct: **delegate** (`eligible=true`, `via_delegation_from={id,name}`) | `TestViewerBlock_Delegate` | real |
| viewer correct: **already-signed** (`eligible=true`, `has_signed_active_stage=true`) | `TestViewerBlock_AlreadySigned` | real |
| viewer correct: **observer** (not author, not eligible → all false/null) | `TestViewerBlock_Observer` | real |
| viewer correct: **author-who-is-also-an-approver** — excluded by author-rule (`eligible=false`) | `TestViewerBlock_AuthorAlsoApprover` | real |
| viewer present on terminal instance (published) with all-false | `TestViewerBlock_NoActiveStage` | real |
| no in-tx display-name lookup (H-PRE-1) | code review + off-tx call-site assertion | real |
| regen produces no hand-written DTO consumer; build clean | `oapi-codegen` regen + `go build ./...` + grep gate | real |

> TDD: write the failing real-DB test first, then implement to green.

## ADR needed?

- [x] Durable decision → **Viewer-facts contract ADR** (governing spec §8.1): the server-derived
  `viewer` block is the single eligibility truth for clients; the FE is forbidden from deriving
  eligibility. viewer facts are *display* truth — enforcement stays tier-2 `authz.Require` in-tx; the
  DTO never gates the server (spec §10). Link: [`wiki/decisions/0078-viewer-facts-contract.md`](../../../../../wiki/decisions/0078-viewer-facts-contract.md).
