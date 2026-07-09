# Feature F2d.1 — Viewer Contract

> **Milestone:** 2d — Workflow Coherence FE + Viewer Contract  ·  **Folder:** `f1-viewer-contract`
> **Status:** Planning

## Source

- Milestone row: instance view DTO gains server-derived `viewer` block (`is_author`,
  `eligible_for_active_stage` = snapshot ∪ active delegation − SoD, `via_delegation_from`,
  `has_signed_active_stage`), computed in the view-read path; OpenAPI + regen; ADR.
- Governing spec: §4 A1 / §8.1 (Viewer-facts contract ADR). Consumer: `deriveWorkspaceMode` (F2d.3).
- Contract answers: `via_delegation_from` = object `{user_id, display_name}`; viewer always present
  when instance exists (`spec.md` interview rows 1–2).

## Plan

### Design (seam)

- **Pure domain fn** `domain.ViewerEligibility(viewerID, authorID string, active *StageInstance, delegations []Delegation, stages []StageInstance) ViewerFacts` — no I/O, fully unit-testable.
  - `ViewerFacts{ IsAuthor, EligibleForActiveStage, HasSignedActiveStage bool; ViaDelegationFromUserID string }` (empty string = none).
  - **Compose on the write-path's single eligibility source — do NOT re-implement delegation matching.**
    Reuse `domain.ResolveEligibleIdentity(viewerID, active.EligibleActorIDs, delegations) → (onBehalf, err)`
    (the exact primitive `decision_service` uses) and `domain.CheckSoD`. This is the anti-split-brain
    requirement — a second membership rule here would recreate the defect this milestone exists to kill.
  - Logic:
    - `IsAuthor = viewerID == authorID`.
    - `active == nil` ⇒ `Eligible=false, ViaDelegationFrom="", HasSigned=false` (terminal instance).
    - `onBehalf, err := ResolveEligibleIdentity(viewerID, active.EligibleActorIDs, delegations)` —
      `err==nil` ⇒ in pool (directly if `onBehalf==""`, via delegation from `onBehalf` otherwise).
    - `priorSignoffs =` this viewer's signoffs on stages **other than** `active`.
    - `passesSoD = CheckSoD(authorID, viewerID, onBehalf, priorSignoffs) == nil`.
    - `Eligible = err==nil && passesSoD`.
    - `ViaDelegationFromUserID = onBehalf` **iff** `Eligible && onBehalf != ""`, else `""`.
    - `HasSigned = ∃ signoff on active.Signoffs with ActorUserID == viewerID`.
  - Delegations passed in are already active-at-now (the repo predicate filters by window); the fn does not re-check time.
- **App method** `ReadService.LoadInstanceByDocumentForViewWithViewer(ctx, runner, tenantID, documentID) (*domain.Instance, ViewerFacts, error)`:
  - Same tx/body as `LoadInstanceByDocumentForView` (reuse — extract shared inner load, keep the old method as a thin wrapper for by-id/other callers).
  - After load + `requireInstanceVisible`, inside the SAME view tx:
    - `viewerID, _ := authz.MustActorID(ctx, tx)` — single source of viewer identity (the tx GUC).
    - `delegations, _ := s.repo.LoadActiveDelegationsFor(ctx, tx, tenantID, viewerID, s.now())` — plain SELECT, H-PRE-1 safe.
    - `vf = domain.ViewerEligibility(viewerID, loaded.SubmittedBy, loaded.Active(), delegations, loaded.Stages)`.
  - `application.ViewerFacts` mirrors the domain value (or reuse the domain type directly).
  - Clock: reuse the service's existing clock if present; else `time.Now().UTC()` via a `now()` helper (no new injected dep unless one already exists).
- **Handler** `GetInstanceByDocumentHandler`: call the new method; pass `&vf` to `mapInstanceResponse`.
- **`mapInstanceResponse(ctx, tenantID, inst, viewer *ViewerFacts)`**: when `viewer != nil`, build `contracts.ViewerFacts`; resolve delegator display name **off-tx** here via `h.displayNameReader.DisplayNames(ctx, tenantID, [delegatorID])` (same port as `resolveEligibleActorNames`, fallback to id). `GetInstanceHandler` (by-id) passes `nil` ⇒ `viewer` omitted.

### Contract (OpenAPI + generated)

- `api/openapi/v1/openapi.yaml` `ApprovalInstanceByDocumentResponse`: add required `viewer` object per `spec.md` block (is_author, eligible_for_active_stage, has_signed_active_stage required; via_delegation_from nullable object `{user_id, display_name}`).
- Regen: `oapi-codegen` (Go server types + FE TS types). **Never hand-edit generated files.**
- `contracts.InstanceResponse` gains `Viewer *ViewerFacts \`json:"viewer,omitempty"\`` ; `contracts.ViewerFacts` + `contracts.ViewerDelegationFrom` structs.

### Files touched

| File | Change |
|---|---|
| `internal/modules/documents/approval/domain/viewer.go` (new) | `ViewerFacts` + pure `ViewerEligibility(...)` |
| `internal/modules/documents/approval/domain/viewer_test.go` (new) | unit tests, every branch |
| `internal/modules/documents/approval/application/read_service.go` | `LoadInstanceByDocumentForViewWithViewer` + shared inner load |
| `internal/modules/documents/approval/http/contracts/instance_read.go` | `Viewer` field + `ViewerFacts`/`ViewerDelegationFrom` |
| `internal/modules/documents/approval/http/get_instance_handler.go` | `mapInstanceResponse` viewer param + off-tx delegator name |
| `internal/modules/documents/approval/http/doc_approval_handler.go` | by-document handler wires new method |
| `api/openapi/v1/openapi.yaml` | `viewer` schema |
| generated Go/TS | via `oapi-codegen` (not hand-edited) |
| `wiki/decisions/NNNN-viewer-facts-contract.md` (new) | ADR |
| real-DB test file (approval http/application integration) | 7 viewer scenarios |

### Test strategy (TDD — failing first)

1. **Domain unit** (`viewer_test.go`) — pure fn, table-driven: author; snapshot actor; delegate; already-signed; observer; author-also-approver (author-rule excludes); no active stage. RED first.
2. **Real-DB** (testdb factory) — seed instance + stages + delegation + signoffs; call the app method through the tx; assert the `ViewerFacts` per scenario. Delegator display-name resolution asserted at the handler/mapper level (off-tx).
3. **Contract/build** — `oapi-codegen` regen clean; `go build ./...`; grep gate: no hand-written `body.data.viewer` consumer.

### Ordering

domain fn+test (RED→GREEN) → app method + real-DB test → contracts + OpenAPI + regen → handler wiring → ADR → full `go build ./...` + targeted suites.

## Execution notes

- **Domain + app + contract + handler** built per plan. `ViewerFacts` + pure `ViewerEligibility` in
  `domain/viewer.go`; `LoadInstanceByDocumentForViewWithViewer` in `read_service.go`; `viewer` block in
  OpenAPI + regenerated types; off-tx delegator display-name resolution in the mapper (H-PRE-1).
- **Emergent scope — visibility convergence (not in original plan, surfaced by the Delegate scenario).**
  Building `TestViewerBlock_Delegate` exposed a second defect of the same split-brain class: the
  instance **visibility** gate `requireInstanceVisible` (ADR 0075/F8) predated delegation (ADR 0077/F9)
  and never learned it — a delegate could sign the stage but got `ErrInstanceNotVisible` → 404 loading
  the instance. Ran `/developing-new-work` before touching it → system-impact analysis
  `docs/superpowers/analysis/2026-07-08-delegation-aware-visibility-system-impact.md`, **🟡 Yellow**
  (proceed; ADR amendment + one named risk carried). Converged `requireInstanceVisible` onto the single
  primitive: author fast-path → `domain.CheckEligibility` → on miss `LoadActiveDelegationsFor` +
  `domain.ResolveEligibleIdentity`; deleted the hand-rolled membership loop. AS-1/AS-2/AS-3 all clean
  (§2/§10 of the analysis). ADR 0078 amended to record the convergence.
- **Test-harness root-cause fixes (drive-by, canonical-framework-compliant):**
  - `ctxWithIdentity` (`read_service_tenant_grade_view_integration_test.go`) seeded only the iam ctx key,
    never the platform actor key the TxRunner chokepoint reads → every real-DB test in that file failed
    with `metaldocs.actor_id GUC not set`. Added `tenant.WithActorID`. Pre-existing latent defect.
  - `seedSignoff` (`read_service_viewer_facts_integration_test.go`) inserted via a raw tx without
    asserting caps → `ErrCapabilityNotAsserted` on the `approval_signoffs` tripwire. Added
    `testdb.SetCapsOnTx(... document.signoff ...)` before insert.
- **Locked constraints carried forward (analysis §10):** never a second membership/SoD rule; no new
  capability / no write-path change; ADR 0078 amended (done); worklist/inbox SQL parity → **M3**;
  F8 deny-regression is part of this feature's gate (below).
- **Validation Gate — real DB (`-tags integration`, isolated DB per test; drop-teardown ~60–480s/test
  on this box, so split into two runs, `-timeout 45m`):**
  - `TestViewerBlock_*` — **7/7 PASS** (`GO_TEST_EXIT=0`, 567s), including `TestViewerBlock_Delegate`
    (the convergence closure).
  - `TestLoadInstance_*` / `TestLoadActiveInstanceByDocument_*` — 10 F8 deny+grant regression tests —
    _(result recorded in evidence.md)_.
- `go build ./...` = 0, `go vet ./internal/modules/documents/approval/...` = 0, approval-package unit
  suites all `ok`.
