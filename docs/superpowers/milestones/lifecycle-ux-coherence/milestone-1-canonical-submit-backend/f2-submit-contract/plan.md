# Feature F2 — Plan

> **Milestone:** 1 — canonical-submit-backend · **Folder:** `f2-submit-contract`
> **Status:** Done

## Source
- Milestone row: *Submit contract makes route_id/content_hash optional and adds revision_title; REV≥1
  requiredness stays downstream in the service.*
- Governing-spec reference: §1.3; ADR 0073 §2. OpenAPI spec + generated code already on disk (dirty tree).

## Plan

1. **Align handwritten contract** (`contracts/submit.go`) to the already-regenerated OpenAPI:
   - `RouteID` optional — empty OK; present → UUID-validated.
   - `ContentHash` optional — empty OK; present → 64-hex-validated.
   - Add `RevisionTitle` (optional string).
2. **Thread `RevisionTitle`** through `SubmitHandler` → `application.SubmitRequest.RevisionTitle`
   (`submit_handler.go`).
3. **Do not re-edit the spec** — it is route truth on disk (dirty tree, absorbed); handler/contract
   conform to it.

### Files touched
- `internal/modules/documents/approval/http/contracts/submit.go` (Validate + RevisionTitle field)
- `internal/modules/documents/approval/http/submit_handler.go` (thread RevisionTitle)

### Test strategy
- **Unit** — `contracts` optional/format matrix: empty-all → nil; malformed route_id → err; malformed
  content_hash → err. Plus the handler-level `submit_handler_test.go` "validate fails" case updated to
  a present-but-malformed route_id (empty is now valid post-F2).
- `go build ./...` green.

### Ordering
Contract Validate → RevisionTitle field → handler thread → unit tests → fix handler test regression.

## Execution notes
Built inline (spike), retro-formalized. The one test regression this caused — `TestSubmitHandler/
validate_fails` returning 201 instead of 400 because empty route_id became valid — was fixed by
switching that case to a malformed `"route_id":"not-a-uuid"` body (root cause: the test's invalidity
signal moved from "absent" to "malformed").
