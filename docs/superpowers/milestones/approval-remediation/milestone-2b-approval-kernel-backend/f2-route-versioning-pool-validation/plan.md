# Feature F2 — Plan

> Input: `spec.md` (this folder). Engine: inline — contract fully prescriptive from verified
> runtime truth, no design exploration needed.

## Files touched

- Create: `db/migrations/0287_approval_route_versioning.sql`
- Modify: `internal/modules/documents/approval/application/route_admin_service.go` (`updateTx`
  supersede-on-in-use branch)
- Modify: `internal/modules/documents/approval/application/submit_service.go` (empty-pool check,
  reusing existing `domain.ErrEmptyEligiblePool`)
- Modify: `internal/modules/documents/approval/http/errors.go` — add ONE mapping entry
  `domain.ErrEmptyEligiblePool` → 422 (fixes a pre-existing unmapped-error gap for
  `decision_service.go`'s quorum path too — verified via `Grep` that this sentinel had no
  mapping entry before this feature)
- Modify: `api/openapi/v1/openapi.yaml` — PUT `/approval/routes/{id}` description text only (no
  schema change)
- Create test: `internal/modules/documents/approval/application/route_versioning_integration_test.go`
  (testdb factory)
- Modify test: `submit_service_test.go` / its integration counterpart (empty-pool case)
- Create: `wiki/decisions/0074-approval-route-versioning.md` (ADR 1); annotate
  `wiki/decisions/0018-approval-route-lifecycle.md` §1/§3 as superseded by 0074

## Ordering (TDD)

1. Migration `0287` first (additive: new `superseded_at` column, replace unique constraints,
   replace `enforce_route_immutable` function body with column-scoped exception — verify against
   the exact baseline SQL already read this session, do not re-derive from scratch).
2. Failing integration test `route_versioning_integration_test.go` (5 cases from spec.md's
   Validation Gate) against the new migration — confirm current code FAILs the in-use case (old
   code has no supersede branch, so `updateTx` returns `ErrRouteInUse` today for any in-use route).
3. Implement `updateTx` supersede branch: attempt in-place `UPDATE ... RETURNING version` as today;
   on `infrastructure.ErrRouteInUse`, INSERT new row (`version = old.version+1`, `active=true`,
   copy `tenant_id`/`profile_code`, new `name`), insert new stages under new id, `UPDATE` old row
   `SET active=false, superseded_at=now() WHERE id=$1` (now permitted by the new trigger body),
   return `{RouteID: newID, NewVersion: newVersion}`. All inside the existing single tx (`runner.Do`).
4. Implement `domain.ErrEmptyStagePool` + `submit_service.go` check (`len(eligibleIDs)==0` →
   return the sentinel, wrapped so the module's error mapper turns it into 422 `problem+json`).
5. Run full approval module suite + targeted new tests + `go build ./...`.
6. Write ADR 0074; commit `feat(approval): F2 versioned route supersession + empty-pool 422 (W1/W6)`.

## Notes

- `Deactivate()` is untouched — spec.md non-goals confirm this explicitly.
- No `openapi.yaml` schema shape changes (`RouteResponse` already generic); only a clarifying
  description addition on the PUT operation.
