# Feature F3.2 — Plan

> **Spec:** `spec.md` (approved 2026-06-16 before this plan was written).

## Tasks

### T1 — E2: Type `fanoutClient` param in `NewFreezeService`

**File:** `internal/modules/documents/application/freeze_service.go`

- Line 77: change `fanoutClient any` → `fanoutClient FanoutClient`
- Line 79: delete `fc, _ := fanoutClient.(FanoutClient)`
- Line 81 (struct init): change `fanout: fc` → `fanout: fanoutClient`

No callers change — both already pass `*fanout.Client` which satisfies the interface.

**Test strategy:** Existing `freeze_service_test.go` / `freeze_pin_test.go` / `freeze_idempotency_test.go`
already pass `*fakeFanoutClient` (which implements `FanoutClient`). They compile and pass without
modification — the type narrowing is backward-compatible for all existing callers.
Add a compile-time type guard: `var _ FanoutClient = (*fanout.Client)(nil)` in
`freeze_service.go` (package-scope sentinel, idiomatic Go, mirrors F3.1 pattern).

### T2 — E3: Remove dead `userID` from `ListDocumentComments`

Touch in this order (build-safe sequence):

1. **`internal/modules/documents/application/service.go:433`** — remove `userID` from signature;
   body unchanged (`s.repo.ListComments(ctx, tenantID, documentID)` already ignores it).

2. **`internal/modules/documents/delivery/http/handler.go:70`** — update `DocumentService` interface
   to match new 3-param signature.

3. **`internal/modules/documents/delivery/http/handler.go:1116`** — remove `userID` from call:
   `h.svc.ListDocumentComments(r.Context(), tenantID, docID)`.

4. **`internal/modules/documents/module_wrapper_test.go:75`** — update fake signature to 3 params.

5. **`internal/modules/documents/delivery/http/handler_test.go:206`** — update `fakeSvc` method.

6. **`internal/modules/documents/delivery/http/handler_comments_test.go:34`** — update
   `commentsStatefulSvc` method.

**Test strategy:** No new tests needed — the parameter was dead; removing it is a signature
simplification. Existing tests already pass `_` for `userID` (3 blanks `_, _, _`); after the
change they pass 2 blanks. `go build ./...` + `go test ./...` are the gates.

## Ordering

T1 first (self-contained, no interface change), then T2 (touches interface + 5 files).
Commit after each task to keep the diff readable.

## Files touched

| File | Change |
|------|--------|
| `internal/modules/documents/application/freeze_service.go` | T1: type param, delete assertion, add type guard |
| `internal/modules/documents/application/service.go` | T2: remove `userID` from signature |
| `internal/modules/documents/delivery/http/handler.go` | T2: interface + call site |
| `internal/modules/documents/module_wrapper_test.go` | T2: fake |
| `internal/modules/documents/delivery/http/handler_test.go` | T2: fake |
| `internal/modules/documents/delivery/http/handler_comments_test.go` | T2: fake |
