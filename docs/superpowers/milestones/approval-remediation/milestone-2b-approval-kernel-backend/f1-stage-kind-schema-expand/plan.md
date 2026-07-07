# Feature F1 — Plan

> Input: `spec.md` (this folder). Engine: inline (writing-plans-equivalent — contract is fully
> prescriptive, no design exploration needed).

## Files touched

- Create: `db/migrations/0286_approval_stage_kinds_expand.sql`
- Modify: `internal/modules/documents/approval/domain/route.go` (add `StageKind` type + `Kind` field
  on the route-stage struct)
- Modify: `internal/modules/documents/approval/domain/instance.go` (add `Kind` field on the instance
  stage-snapshot struct; add `FrozenContentHash *string`, `CancelReason *string` fields on the
  instance struct)
- Modify: `internal/modules/documents/approval/domain/signoff.go` (add `SignatureMeaning` field)
- Modify: `internal/modules/documents/approval/domain/errors.go` (add `ErrInvalidStageKind`)
- Create test: `internal/modules/documents/approval/domain/stage_kind_test.go`
- Create test: `tests/integration/approval/stage_kind_schema_test.go` (testdb factory)
- Modify (if a Go migration/scan registry exists): repository scan/insert/update SQL touching
  `approval_route_stages`, `approval_stage_instances`, `approval_instances`, `approval_signoffs` in
  `internal/modules/documents/approval/infrastructure/postgres_approval_repository.go` — read the
  new columns into the structs above wherever the existing SELECT/INSERT already lists all columns
  (grep every `SELECT ... FROM approval_route_stages`, `approval_stage_instances`, `approval_signoffs`
  in that file first; add the new columns to the same statements, non-exhaustively is a bug — the Go
  struct will silently zero-value otherwise).

## Test strategy (TDD order)

1. Write `stage_kind_test.go` first (`TestStageKindValues`, `TestStageKindValidate`) — run, confirm
   FAIL (`undefined: StageKind`).
2. Implement `StageKind` type + `Validate()` + `ErrInvalidStageKind` in `domain/route.go` /
   `domain/errors.go`. Run again — confirm PASS.
3. Add `Kind StageKind` to the route-stage struct and instance stage-snapshot struct; add
   `FrozenContentHash`, `CancelReason` to the instance struct; add `SignatureMeaning` to the signoff
   struct. `go build ./...` must stay clean — fix every call site the compiler flags (struct literals
   missing the new field are fine in Go, only mismatched positional literals break; check for those).
4. Write the migration `0286` per spec.md's exact column/CHECK text (verify against baseline: the
   real table is `approval_stage_instances`).
5. Write `stage_kind_schema_test.go` using the `testdb` factory (canonical framework) — seed a route
   + stage via the factory's existing seed helper, assert default `stage_kind='approval'` on a
   pre-existing row, then attempt a direct SQL `UPDATE ... SET stage_kind='signature'` and assert the
   CHECK constraint name in the returned error.
6. Update every repository SELECT/INSERT/UPDATE touching the four affected tables to read/write the
   new columns (needed so F2-F8 don't silently drop them later).
7. Run full approval module test suite + `go build ./...` — must stay green (this is additive-only;
   no existing test should need behavior changes, only possibly struct-literal compile fixes).
8. Commit `feat(approval): F1 stage-kind schema expand + domain enum (spec §2.1/§7)`.

## Ordering rationale

Domain enum first (fastest failing-test loop, no DB needed) → migration → integration test →
repository wiring, so that a compile break in step 3 is caught before the DB migration is written
against structs that don't exist yet.
