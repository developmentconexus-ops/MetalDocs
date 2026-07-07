# Feature F5 — Evidence

> **Milestone:** 1 · **Feature:** `f5-idempotency-map` · **Closed:** 2026-07-06
> **Contract:** `spec.md` (idempotency map completion; contract-first: map iff spec declares the key).

## What was implemented
- **No code change.** Analysis/verification feature. Finding 16 already satisfied; finding 17 is a
  documented contract-first defer with a written trigger.
- Producer matches consumer contract: the Go `idempotentRoutes` set equals the set of spec ops that
  declare an `Idempotency-Key` header (no orphan requires a header the spec omits).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Finding 16 — mark-reviewed mapped | `grep idempotentRoutes internal/modules/documents/approval/http/router.go` | `POST /api/v1/documents/{id}/review: true` at **router.go:36** | **real** |
| Parity — spec header decls == Go maps | enumerate `Idempotency-Key` ops in `openapi.yaml` vs both `idempotentRoutes` maps | 25/25 covered, **zero orphans** | **real** |
| Finding 17 — archive/approval-config consistent | check both spec header decls AND Go maps | **absent from both sides** → consistent (defer, not gap) | **real** |
| Static | `go build ./...` | `EXIT 0` | — |
| Regression | `go test ./internal/modules/documents/approval/...` | all `ok` | **real** |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Full parity sweep 25/25, zero orphans | yes | parity row |
| Finding 16 mark-reviewed present in map | yes | router.go:36 |
| Finding 17 absent from both spec + map (consistent) | yes | consistency row |
| build + test green | yes | build EXIT 0; approval `ok` |

## Review disposition
- Spec-compliance: PASS — executing the pre-agreed F5 rule (map iff spec declares the key). Finding
  17's §3 "3-line M1 fix" framing is a **deviation**: adding the two ops would violate contract-first
  and break the header-less approval-config FE caller. Recorded rationale in spec.md + flagged at HS-1.
- Code-quality: PASS — the parity invariant (Go set == spec-declared set) is the correct construction;
  a Go entry without a spec header would be the actual defect.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Finding 17 — templates archive + approval-config idempotency | PUT approval-config is HTTP-idempotent (full-replace); POST archive is OCC/status-guarded (double-archive = no-op conflict) | if OpenAPI adds `Idempotency-Key` to either op → add to templates `idempotentRoutes` + spec↔map parity test in the same change; owner templates module |
