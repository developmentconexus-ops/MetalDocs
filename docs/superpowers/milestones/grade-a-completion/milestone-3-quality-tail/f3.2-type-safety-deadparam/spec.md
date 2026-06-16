# Feature F3.2 — Spec

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.2-type-safety-deadparam`
> **Status:** Approved — 2026-06-16
> **Authored before any code.**

## Consumer contract

**Consumers:**

1. **Every caller of `NewFreezeService`** (`apps/api/cmd/metaldocs-api/main.go:791`,
   `apps/worker/cmd/metaldocs-worker/main.go:105`) — receives a typed constructor that rejects a
   wrong-type argument at compile time instead of silently yielding a nil fanout at runtime.

2. **Every caller of `Service.ListDocumentComments`** (delivery handler
   `internal/modules/documents/delivery/http/handler.go:1116`) — the method signature drops the
   `userID` parameter that was always ignored; callers pass one fewer argument and the interface
   contract is honest about what the method actually needs.

**What the consumer observes after this feature:**

- `docapp.NewFreezeService(...)` last parameter type is `FanoutClient` (not `any`). Passing a
  non-`FanoutClient` value is a **compile error**, not a silent nil.
- `Service.ListDocumentComments` signature is `(ctx, tenantID, documentID string)` — three
  parameters, no `userID`. All implementors of the `DocumentService` interface in `handler.go:70`
  match this shape.
- Whole-repo build succeeds (`go build ./...`).
- Whole-repo test suite green (`go test ./...`).
- No authz regression: the authz gate (`authorizeDocumentScope` at `handler.go:1111`) is in the
  delivery layer and is **not touched**. The service never received the `userID` for authz — it
  forwarded it to neither the repo nor any scope check. This is recorded as evidence.

## Non-goals

- No change to `ListComments` repo method or its signature.
- No change to `authorizeDocumentScope` or any delivery-layer authz logic.
- No change to any other `FreezeService` method or field.
- No other parameters removed from any other service method (surgical scope: E2 + E3 only).
- No golangci-lint or repo-wide `any` sweep.

## Validation Gate

| # | Criterion | Command / proof |
|---|-----------|-----------------|
| 1 | `NewFreezeService` last param is `FanoutClient`, not `any`; type assertion at :79 deleted | `grep -n 'fanoutClient any\|fc, _' internal/modules/documents/application/freeze_service.go` → 0 matches |
| 2 | `Service.ListDocumentComments` signature has no `userID` | `grep -n 'ListDocumentComments' internal/modules/documents/application/service.go` shows 3-param signature |
| 3 | `DocumentService` interface in `handler.go` updated to match | `grep -n 'ListDocumentComments' internal/modules/documents/delivery/http/handler.go` shows 3-param |
| 4 | All fakes/mocks updated | `go build ./...` clean |
| 5 | No authz path consumed `userID` | Authz-scope check recorded in evidence.md (done pre-spec — `s.repo.ListComments(ctx, tenantID, documentID)` — `userID` never forwarded; authz gate is delivery-layer `authorizeDocumentScope` before service call) |
| 6 | Whole-repo tests green | `go test ./...` PASS |

## Interview record (authz-scope check)

Pre-spec investigation (session 2026-06-16):

| Question | Finding |
|----------|---------|
| Does `ListDocumentComments` body forward `userID` to the repo? | No — `service.go:434`: `return s.repo.ListComments(ctx, tenantID, documentID)`. `userID` never passed. |
| Does any authz scope derive from `userID` inside the method? | No — method body is a single `return` with no branching on `userID`. |
| Is there a delivery-layer authz gate before the call? | Yes — `handler.go:1111`: `authorizeDocumentScope(w, r, docID)` runs first; service is only called on `ok`. The returned `userID` is then passed to the service, but the service ignores it. |
| HS-2 verdict | Does not trip. Safe to remove. |
| Does passing `*fanout.Client` as `any` to `NewFreezeService` satisfy `FanoutClient`? | Yes — `*fanout.Client` has `Fanout(ctx, FanoutRequest) (FanoutResponse, error)` which matches the interface at `freeze_service.go:30–32`. Both production callers already pass the right type. |
