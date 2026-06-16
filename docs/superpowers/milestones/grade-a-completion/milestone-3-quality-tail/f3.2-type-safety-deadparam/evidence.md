# Feature F3.2 — Evidence

> **Milestone:** 3 — Code-quality & dead-code tail  ·  **Feature:** `f3.2-type-safety-deadparam`  ·  **Closed:** 2026-06-16
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

**E2 — `fanoutClient any` → `FanoutClient` in `NewFreezeService`:**
- `internal/modules/documents/application/freeze_service.go`: changed last constructor param from
  `fanoutClient any` to `fanoutClient FanoutClient`; deleted `fc, _ := fanoutClient.(FanoutClient)`
  type assertion; changed `fanout: fc` to `fanout: fanoutClient`; added package-scope compile-time
  type guard `var _ FanoutClient = (*fanout.Client)(nil)`. No caller changes — both production
  callers already pass `*fanout.Client` which satisfies the interface.

**E3 — dead `userID` removed from `ListDocumentComments`:**
- `internal/modules/documents/application/service.go:433`: removed `userID` from signature; body
  unchanged (`s.repo.ListComments(ctx, tenantID, documentID)` already ignored it).
- `internal/modules/documents/delivery/http/handler.go:70`: updated `documentComments` interface to
  3-param signature.
- `internal/modules/documents/delivery/http/handler.go:1111`: changed `tenantID, userID, ok :=` to
  `tenantID, _, ok :=` (authz gate return still runs; `userID` simply discarded).
- `internal/modules/documents/delivery/http/handler.go:1116`: removed `userID` from service call.
- Three test fakes updated (`module_wrapper_test.go:75`, `handler_test.go:206`,
  `handler_comments_test.go:34`): 4-param `(context.Context, string, string, string)` →
  3-param `(context.Context, string, string)`.

Commits:
- `4fe11206 fix(f3.2): type FanoutClient param in NewFreezeService (E2)`
- `8e015af4 fix(f3.2): remove dead userID param from ListDocumentComments (E3)`

## Verification

| Check | Command / action | Result | Real vs fixture |
|-------|------------------|--------|-----------------|
| Gate 1: `fanoutClient any` + type assertion gone | `grep -n 'fanoutClient any\|fc, _' internal/modules/documents/application/freeze_service.go` | 0 matches (grep exits 1 = no matches) | — |
| Gate 2: service.go 3-param signature | `grep -n 'ListDocumentComments' internal/modules/documents/application/service.go` | `433:func (s *Service) ListDocumentComments(ctx context.Context, tenantID, documentID string) ([]domain.Comment, error)` | — |
| Gate 3: interface + callsite updated | `grep -n 'ListDocumentComments' internal/modules/documents/delivery/http/handler.go` | `:70` 3-param interface; `:1116` 3-param call | — |
| Gate 4: whole-repo build | `go build ./...` | `BUILD OK` — clean exit | — |
| Gate 5: authz-scope check | pre-spec read of `service.go:433` body | `return s.repo.ListComments(ctx, tenantID, documentID)` — `userID` never forwarded; authz gate is delivery-layer `authorizeDocumentScope` at `handler.go:1111`, runs before service call; safe to remove | N/A |
| Gate 6: documents packages (force-fresh) | `go test -count=1 ./internal/modules/documents/...` | All 10 packages PASS (16 packages, `[no test files]` skipped) | fixture |
| Gate 6: whole-repo | `go test ./...` | All packages PASS (fully cached — documents packages force-run above) | fixture |

## Acceptance vs spec Validation Gate

| # | Criterion | Met? | Evidence |
|---|-----------|------|----------|
| 1 | `any` param + silent assertion gone | yes | grep returns 0 matches |
| 2 | `service.go` signature 3-param | yes | grep shows `(ctx, tenantID, documentID string)` |
| 3 | `handler.go` interface 3-param | yes | grep :70 shows updated interface |
| 4 | All fakes updated; build clean | yes | `go build ./...` OK |
| 5 | Authz-scope check recorded; HS-2 does not trip | yes | `userID` never forwarded in body; delivery-layer gate unchanged |
| 6 | Whole-repo tests green | yes | `go test -count=1 ./internal/modules/documents/...` all PASS |

## Review disposition

Changes are purely compile-time type corrections with zero runtime behavior change:
- E2: callers already pass the right type; typed param enforces this at compile time.
- E3: `userID` was dead in the service body; delivery authz gate (`authorizeDocumentScope`) is unchanged and still the real enforcement point.
No review findings. Both fixes are at the right seam per milestone quality goal #1.

## Bounded defers

None — all E2 and E3 changes are complete and self-contained.
