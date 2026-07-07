# Feature F1 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f1-stage-kind-schema-expand`  ·
> **Closed:** 2026-07-07
> **Contract:** `spec.md`

## What was implemented

- Additive migration `db/migrations/0286_approval_stage_kinds_expand.sql`: `stage_kind` (text, NOT
  NULL DEFAULT `'approval'`, CHECK `IN ('review','approval')`) + `due_in_days` (int, nullable, CHECK
  `> 0`) on `approval_route_stages`; `stage_kind` (same CHECK/default) + `due_at` (timestamptz,
  nullable) on `approval_stage_instances`; `frozen_content_hash` (text, nullable, CHECK
  `^[0-9a-f]{64}$`) + `cancel_reason` (text, nullable) on `approval_instances`; `signature_meaning`
  (text, NOT NULL DEFAULT `'approval'`, CHECK IN `('approval','rejection')`) on `approval_signoffs`.
  Idempotent, ledger insert, no backfill needed (DEFAULT covers existing rows).
- `domain.StageKind` type + `StageKindReview`/`StageKindApproval` consts + `Validate()` +
  `ErrInvalidStageKind` sentinel (`domain/errors.go`, `domain/route.go`).
- Struct field additions: `Stage.Kind`/`Stage.DueInDays` (route.go); `StageInstance.Kind`/`.DueAt`,
  `Instance.FrozenContentHash`/`.CancelReason` (instance.go); `Signoff.signatureMeaning` + getter +
  `SignoffParams.SignatureMeaning` with default `"approval"` and validation restricting to
  `"approval"`/`"rejection"` (signoff.go).
- Full repository wiring in `postgres_approval_repository.go` across every SELECT/INSERT/UPDATE
  touching the four tables (`InsertInstance`, `InsertStageInstances`, `InsertSignoff`,
  `LoadSignoffByActor`, `loadSignoffByStageActor`, `scanSignoff`, `LoadInstance`,
  `LoadActiveInstanceByDocument`, `listRoutesQuery`/`scanRouteListRows`, `loadStageInstances`,
  `loadSignoffsForInstance`, `LoadInstancesByIDs`, `LoadRoute`, `LoadPriorSignoffs`,
  `LoadStageSignoffs`/`scanSignoffsRows`) — no new column is silently dropped.
- `route_admin_service.go`: `insertRouteStages` extended to 11 cols; `loadRouteStagesTx` scans new
  cols; `stagesEqual` bug found+fixed — normalizes empty `Kind` to `StageKindApproval` before
  DeepEqual compare so the DB's concrete default doesn't trigger a spurious "route changed"
  detection.
- No behavior change: `stage_kind` defaults to `'approval'` everywhere; no service wiring of
  stage-kind semantics (deferred to F3/F4/F5 per spec non-goals).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `go test ./internal/modules/documents/approval/domain/... -run TestStageKind -v` | `TestStageKindValues` PASS, `TestStageKindValidate` PASS | real |
| Static (build) | `go build ./...` | clean, exit 0 | — |
| Static (vet, integration tags) | `go vet -tags integration ./tests/integration/approval/... ./tests/integration/testdb/...` | clean, exit 0 | real |
| Targeted package suite | `go test ./internal/modules/documents/approval/... -v` | all subpackages (application, domain, http, http/contracts, infrastructure, infrastructure/idempotency, infrastructure/signature, jobs) PASS | real |
| Runtime proof — migration + DEFAULT + CHECK | `METALDOCS_DATABASE_URL=... go test -tags integration ./tests/integration/approval/... -run StageKind -v` against live `metaldocs-postgres` container | `TestStageKindSchemaExpand_Default` PASS (234.62s) — migration 0286 applies cleanly on 0285 template, DEFAULT `stage_kind='approval'` confirmed on both tables; `TestStageKindSchemaExpand_RejectsUnknownValue` PASS (69.30s) — CHECK constraint rejects invalid value on INSERT (both tables) and UPDATE | real (live DB, not fixture) |

Pre-existing unrelated `go vet -tags integration ./tests/integration/...` failures in
`controlleddocuments/active_instance_parity_test.go` and
`templates/lifecycle_no_auto_draft_test.go` (missing packages) confirmed via `git status` as
untouched by this feature — out of scope, not fixed here.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Go enum has exactly the two spec values and rejects unknown values | yes | TDD row above |
| DB CHECK rejects an unknown `stage_kind` on both tables | yes | Runtime proof row above |
| Migration applies cleanly on `0285`, existing rows get the DEFAULT | yes | Runtime proof row above |
| No regression | yes | Static (build) + targeted package suite rows above; `stagesEqual` bug caught and fixed as part of this |

## Review disposition

- Spec-compliance review: contract matched exactly (columns, CHECKs, defaults, Go field
  placement, non-goals respected — no service wiring, no route versioning/capability/hash-chain
  changes, no `frozen_content_hash`/`signature_meaning` derivation logic).
- Code-quality review: one real gap found and fixed during implementation —
  `stagesEqual` compared raw `Kind` fields via `DeepEqual`, which would have spuriously flagged a
  route as "changed" once the DB started returning a concrete default value not present in the
  original Go zero-value comparison; fixed by normalizing before compare (root-cause fix, not a
  symptom patch).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| None | — | — |
