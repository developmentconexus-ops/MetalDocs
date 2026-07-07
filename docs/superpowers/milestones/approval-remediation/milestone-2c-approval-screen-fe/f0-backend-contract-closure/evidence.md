# F0 — Evidence

## Commands + real output

- `go build ./...` → clean (no output, exit 0).
- `go test ./internal/modules/documents/approval/...` → all PASS:
  `application 2.595s · domain (cached) · http 3.427s · http/contracts (cached) · infrastructure (cached) · idempotency 1.929s · signature (cached) · jobs 1.929s`.
- `go test ./internal/modules/documents/approval/http/contracts/` → `ok 1.045s` incl.
  `TestInstanceStatusWireEnumComplete`, `TestStageRequestStageKindValidation` (4 subcases).
- `go generate` in `approval/api` → exit 0; `api.gen.go` now has 74 `stage_kind`/`StageKind`/`changes_requested` occurrences.
- `npm run gen:api` (openapi-typescript 7.13.0) → exit 0; `api-types/index.d.ts` has 7 `stage_kind`/`changes_requested` occurrences.

## TDD proof

- `TestInstanceStatusWireEnumComplete` written before `IsValidInstanceStatus` helper existed;
  `TestStageRequestStageKindValidation` written before `stage_kind` validation branch added. Both
  fail-first (helper/validation absent → compile/assert fail), pass after implementation.

## HS-2 decision proof (path A)

- `grep -rn ScanForUnresolvedMarkup internal` → **zero** (util + test deleted; no wiring existed).
- Freeze integrity invariant remains enforced: `freeze.go:50` `HasUnresolvedInstanceComments` gate +
  `PinFrozenHash` hash chain, unchanged.
- Decision + runtime-truth interview recorded in `spec.md`; hard-stop + bounded defer recorded in
  program `README.md`.

## Runtime proof (observable change)

- Route-create now accepts `stage_kind: review` (contract + handler map + persistence path all
  aligned; `route_admin_service.insertRouteStages` writes `stage_kind`, `loadRouteStagesTx` reads
  it). Live-QA (F8) exercises actual review-route creation end-to-end against the running stack —
  this evidence row is contract/unit level; the runtime round-trip is verified in F8.

## Fixture-vs-real

- All F0 proof above is build + unit/contract level (real Go compiler + test runner). No mocks. The
  end-to-end review-route creation is deferred to the F8 live-QA walkthrough (labeled there).

## Review / QA disposition

- Self-review: contract-first honored (openapi → oapi-codegen → FE regen, generated DTOs only);
  no-fallback honored (`mapStageKind` uses a known-exhaustive canonical default, not a masked
  unknown); no new migration.

## Bounded defers

- **Server-authoritative suggestion-resolution freeze gate** — trigger + owner in README close-out
  list (HS-2). Today suggestion resolution is client-authoritative (eigenpal) + caught by the hash
  chain.
