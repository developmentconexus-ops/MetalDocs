# Feature F3 — Plan

> **Milestone:** 1 — canonical-submit-backend · **Folder:** `f3-error-mapping`
> **Status:** Done

## Source
- Milestone row: *Map the submit-path domain sentinels (previously finalize-only) to typed RFC 9457
  problem+json — never 500.*
- Governing-spec reference: ADR 0073 §1; prior finalize mappings (handler.go:766-778, pre-removal).

## Plan

1. **Extend `MapErrorToResponse`** (`http/errors.go`) with the four sentinels the in-tx submit path
   now surfaces, each → its finalize-era status + a typed dot-notation `problem.Code`:
   - `application.ErrRevisionTitleRequired` → 422
   - `docsdomain.ErrDocumentNotDraft` → 409
   - `docsdomain.ErrProfileNotConfigured` → 400
   - `docsdomain.ErrApprovalRouteMissing` → 409
2. **Keep** existing reason/idempotency/if-match/stale mappings untouched.

### Files touched
- `internal/modules/documents/approval/http/errors.go` (four new sentinel arms)

### Test strategy
- **Unit table test** over `MapErrorToResponse`: each sentinel → expected status + non-`internal.unknown`
  code.
- **Integration confirmation** (F1 suite) — REV≥1 empty-title submit → 422 typed; no-route → 409.
- **Live QA** — no-route submit → clean 409 `state.approval_route_missing` (not 500); duplicate → 409.
- `go build ./...` + `go test ./...` green.

### Ordering
Add mapping arms → unit table test → confirm via F1 integration + live QA.

## Execution notes
Built inline (spike), retro-formalized. Mapping arms carry inline comments citing the finalize-era
status each mirrors, so the "same contract, new entrypoint" intent is legible at the boundary.
