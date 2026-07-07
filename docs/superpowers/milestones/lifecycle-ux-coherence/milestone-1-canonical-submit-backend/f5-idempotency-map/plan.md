# Feature F5 — Plan

> **Milestone:** 1 — canonical-submit-backend · **Folder:** `f5-idempotency-map`
> **Status:** Done

## Source
- Milestone row: *Complete the idempotency map — mark-reviewed (16) + templates archive/approval-config
  (17), contract-first: map iff the spec declares the key, else defer with a written trigger.*
- Governing-spec reference: §3 gap register findings 16, 17; `milestone.md` F5 acceptance rule.

## Plan

1. **Parity sweep** — enumerate every mutating op in `openapi.yaml` that declares an `Idempotency-Key`
   header parameter; diff against the Go `idempotentRoutes` maps (`approval/http/router.go` +
   `templates/delivery/http/handler.go`). The invariant: the two sets are equal.
2. **Finding 16** — verify `POST /documents/{id}/review` (mark-reviewed) is already in the approval map;
   no change if present.
3. **Finding 17** — check whether the two template ops declare the header:
   - **If yes** → add to the templates map + a spec↔map parity test.
   - **If no** → **bounded defer** with a written trigger (do NOT add — would break contract-first and
     the header-less approval-config FE caller).

### Files touched
- None (analysis + verification feature; both findings resolve to "already correct" / "documented defer").

### Test strategy
- **Grep/parity verification** — dump both `idempotentRoutes` maps + the spec header decls; assert
  set-equality (25/25, zero orphans).
- Confirm finding 16 present; confirm finding 17 ops absent from BOTH sides (consistent → defer, not gap).
- `go build ./...` + `go test ./...` green.

### Ordering
Sweep spec header decls → diff Go maps → classify each finding → record defer trigger for 17.

## Execution notes
Built inline (spike), retro-formalized. Outcome: **no code change** — finding 16 was already mapped
(router.go:36), finding 17 is a contract-first defer. The §3 "3-line M1 fix" framing for 17 is
therefore a **deviation**, flagged at HS-1 (see spec.md "Deviation flagged for HS-1"). This is the
pre-agreed F5 rule executing correctly, not a scope miss.
