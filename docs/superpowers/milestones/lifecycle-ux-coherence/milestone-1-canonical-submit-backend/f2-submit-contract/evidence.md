# Feature F2 — Evidence

> **Milestone:** 1 · **Feature:** `f2-submit-contract` · **Closed:** 2026-07-06
> **Contract:** `spec.md` (optionalize route_id/content_hash; add revision_title).

## What was implemented
- `contracts.SubmitRequest.Validate()` (`contracts/submit.go`): route_id optional (empty OK, present →
  UUID); content_hash optional (empty OK, present → 64 hex); `RevisionTitle` optional field added.
- `SubmitHandler` threads `RevisionTitle` into `application.SubmitRequest` (`submit_handler.go:76`).
- Handwritten contract aligned to the already-regenerated OpenAPI (dirty tree); spec not re-edited.
- Producer matches consumer contract: the FE editor's header-less body validates (empty → nil).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — validate-fails regression fixed | `go test ./internal/modules/documents/approval/http/...` | `ok` — `TestSubmitHandler/validate_fails` now sends malformed `route_id:"not-a-uuid"` → 400 (empty is valid post-F2) | **real** |
| Static | `go build ./...` | `EXIT 0` | — |
| Unit — optional/format matrix | `go test ./internal/modules/documents/approval/http/contracts/...` | `ok` | **real** |
| Runtime — empty body accepted | docker curl `POST /submit` body `{}` | 409 route-missing then **201** after route seeded — body accepted, no field-required 400 | **real** (docker) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| `SubmitRequest{}` all-empty `.Validate()` → nil | yes | contracts unit + live empty-body 201 |
| `{route_id:"not-a-uuid"}` → validation error | yes | `TestSubmitHandler/validate_fails` → 400 |
| `{content_hash:"xyz"}` → validation error | yes | contracts unit matrix |
| revision_title decoded + passed to application req | yes | `submit_handler.go:76` thread; happy-path handler test |
| build green | yes | build EXIT 0 |

## Review disposition
- Spec-compliance: PASS — handler/contract conform to on-disk OpenAPI (contract-first); no spec re-edit.
- Code-quality: PASS — the single test regression (empty route_id now valid) was fixed at root cause
  (moved the test's invalidity signal from "absent" to "malformed route_id"), not by weakening the assert.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | | |
