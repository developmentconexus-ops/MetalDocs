# F0 — Backend contract closure (M2b carried defers)

> **Milestone:** M2c approval-screen-fe · **Consumer:** FE api-types (`components['schemas']`) + the route-create admin flow + the approval read path.
> **Status:** Approved (contract) — 2026-07-07. Approval line below.

## Consumer contract (what downstream requires, defined before producer)

1. **Wire `InstanceStatus` covers every domain status.** The FE, reading an instance whose
   `status === 'changes_requested'` (reached when a review verdict is `request_changes`), must
   receive a parseable value. Contract: `contracts.IsValidInstanceStatus("changes_requested") == true`
   and openapi `ApprovalInstanceByDocumentResponse.status` enum includes `changes_requested`.
2. **Route-create can specify `stage_kind`.** Without it, only approval-kind stages are creatable,
   so a review→approval route (needed for the live-QA walkthrough and the whole review UX) cannot
   exist. Contract: `StageRequest.stage_kind` (enum `review|approval`, optional, empty→approval at
   persistence) accepted on create/update; echoed on the list-routes read; persisted to
   `approval_route_stages.stage_kind`.
3. **Regen alignment.** Go server types + `frontend/apps/web/src/lib/api-types/index.d.ts`
   regenerated so the FE consumes generated DTOs only (ADR 0035), never hand-written bodies.

## Non-goals

- The freeze docx-markup gate (see Decision below — removed, not in F0).
- Per-stage capability enforcement (still advisory; a separate ADR per ADR 0022 §scope).
- Any new migration — `approval_route_stages.stage_kind` (migration 0286) and the domain layer
  (`InstanceChangesRequested`, `StageKind`, review-verdict) already landed in M2b.

## Interview record (B1.5) — freeze markup gate feasibility

| # | Question | Finding (runtime truth, file:line) | Decision |
|---|----------|-----------------------------------|----------|
| 1 | Does a docx-byte source exist at freeze time to feed `ScanForUnresolvedMarkup`? | No. Content during review = form-data JSON (`documents/domain/model.go:36`); `ContentHashAtSubmit` hashes FormData, not docx (`submit_service.go:158`). Docx only materialized externally by render svc, addressed by S3 key (`render/fanout/client.go:16`). No byte-read port exists — only key-based presign (`export_service.go:23`). | Gate has nothing to call. |
| 2 | Do reviewer suggestions/comments ever exist as docx `w:ins`/`w:del` server-side? | No. Comments persist as structured JSON (`documents/domain/comment.go:10`, `document_comments`). Suggestions are an eigenpal client transform over form-data. A render from form-data is markup-free by construction. Whole-repo grep for tracked-change XML outside `markup_gate.go` → nothing. | The scan checks a state that cannot occur. |
| 3 | Can bytes be fetched inside `executeFreeze` (in-tx, row `FOR UPDATE`)? | No. Precedent forbids network/blob I/O in a tx: `reconstruct_service.go:32` closes its tx before the render round-trip. Approval only ever calls documents-module **metadata** helpers in-tx, never bytes/render. | In-tx fetch is an architectural regression. |
| 4 | Is the freeze integrity invariant already enforced without the scan? | Yes. `executeFreeze` already blocks on `HasUnresolvedInstanceComments` (`freeze.go:50`); the canonical hash chain pins `FrozenContentHash` and echo-verifies at signoff (no-fallback); F6 clean-buffer resolves suggestions client-side before re-submit. | Triad is complete. |

## Decision (HS-2 → path A, operator-surfaced 2026-07-07)

The docx-XML markup gate is **infeasible and misdirected** in MetalDocs' form-data content model.
Wiring it would lock in a bespoke in-tx blob reader (a real regression) to detect a state that
cannot occur — the local-maximum trap. The global-maximum freeze-integrity structure (hash-attested
final content + resolved comment threads + clean client buffer) is **already in place**.

**Resolution:** `ScanForUnresolvedMarkup` + its test deleted as misdirected dead code (built in M2b
as an unwired "bounded defer"). The genuine future work — a *server-authoritative* suggestion-
resolution freeze gate (model suggestions as structured resolvable state like comments) — is
registered as a bounded defer in the program README with an explicit trigger. F0 is re-scoped to
contracts 1–3 above.

## Validation Gate

- `go build ./...` clean; `go test ./internal/modules/documents/approval/...` PASS.
- New test `TestInstanceStatusWireEnumComplete` (contracts): all 5 domain statuses valid on the wire.
- New test: route-create with `stage_kind: review` persists a review-kind stage; invalid kind → 422.
- Regenerated `api-types/index.d.ts` contains `stage_kind` and `changes_requested`.
- `grep -rn ScanForUnresolvedMarkup internal` → zero (dead code gone).

## Approval

- **Contract approved:** 2026-07-07 (main session, per ratified master plan §F0 as amended by the
  HS-2 decision above; operator surfaced and holds the HS-1 close gate).
