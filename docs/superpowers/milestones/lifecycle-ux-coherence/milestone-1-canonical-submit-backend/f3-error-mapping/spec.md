# F3 — Map submit domain sentinels to RFC 9457 responses

> **Milestone:** M1 · **Findings:** 3 · **Status:** spec approved (ADR 0073-ratified)
> **Approved:** 2026-07-06 — from ADR 0073 §1 + finalize handler's prior mappings (handler.go:766-778).

## Consumer contract

The submit HTTP boundary (`approvalhttp.MapErrorToResponse`) must translate the sentinels the
canonical submit path now returns (previously finalize-only) into typed `problem+json`, never 500:

| Sentinel | Status | Rationale (matches prior finalize mapping) |
|---|---|---|
| `application.ErrRevisionTitleRequired` | 422 | REV≥1 missing title = friendly business-rule rejection (mirrors reason-for-change 422). |
| `docsdomain.ErrDocumentNotDraft` | 409 | illegal-state-for-write (finalize used 409 StateTransitionInvalid). |
| `docsdomain.ErrProfileNotConfigured` | 400 | actionable request problem (finalize used 400 ValidationError). |
| `docsdomain.ErrApprovalRouteMissing` | 409 | no active route (finalize used 409 ApprovalRouteMissing). |

Each maps to a typed module-local `problem.Code` (dot-notation taxonomy). RFC 9457 body.

## Non-goals
- No new sentinels; no change to existing reason/idempotency/if-match mappings.
- No status changes to already-mapped errors.

## Validation Gate
- Table test over `MapErrorToResponse`: each sentinel → its status + a non-`internal.unknown` code.
- REV≥1 empty-title submit integration → 422 typed (proven in F1 integration).
- `go build ./...` + `go test ./...` green.
