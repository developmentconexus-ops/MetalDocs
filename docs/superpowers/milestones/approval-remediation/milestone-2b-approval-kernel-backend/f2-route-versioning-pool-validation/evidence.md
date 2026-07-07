# Feature F2 — Evidence

> **Milestone:** 2b — Approval Kernel Backend  ·  **Feature:** `f2-route-versioning-pool-validation`  ·
> **Closed:** 2026-07-07
> **Contract:** `spec.md`

## What was implemented

- Additive migration `db/migrations/0287_approval_route_versioning.sql`: adds
  `approval_routes.superseded_at timestamptz`; drops
  `approval_routes_tenant_profile_key UNIQUE(tenant_id, profile_code)`; adds
  `approval_routes_active_profile_uq` (partial unique `(tenant_id, profile_code) WHERE active`) +
  `approval_routes_profile_version_uq` (unique `(tenant_id, profile_code, version)`); replaces
  `enforce_route_immutable()` with a column-scoped exception permitting an
  active/superseded_at-only UPDATE on an in-use or inactive row while still blocking any change to
  `name`/`profile_code`/`version`/`tenant_id`/`created_at`/`created_by`.
- `route_admin_service.go` `updateTx`/`updateInPlaceOrSupersede`: attempts the cheap in-place UPDATE
  first (unchanged path for not-in-use routes); on `infrastructure.ErrRouteInUse`, falls back to the
  supersede sequence in the same transaction — mark the old row `active=false, superseded_at=now()`,
  insert a new row (new id, `version = old+1`, `active=true`, same `tenant_id`/`profile_code`, new
  name), insert the new stage set under the new id, return `{RouteID: newID, NewVersion: newVersion}`.
  `Deactivate()` untouched (its UPDATE bumps `version` too, so it never qualifies for the trigger's
  narrow exemption and remains blocked while in-use — verified, see Acceptance table).
- `submit_service.go`: after `ResolveEligibleActors` per stage, `len(eligibleIDs) == 0` returns the
  existing sentinel `domain.ErrEmptyEligiblePool` (no new sentinel introduced).
- `http/errors.go`: one new mapping entry, `domain.ErrEmptyEligiblePool` → 422
  (`approvalCodeValidationEmptyEligiblePool`) — also fixes the pre-existing unmapped-error gap on
  `decision_service.go`'s quorum-evaluation path.
- `api/openapi/v1/openapi.yaml`: added `422` response to `submitDocumentForApproval`; added a
  description clarifying `updateApprovalRoute`'s `route_id`-may-differ-from-path-id semantics; no
  schema shape change. Regenerated `api.gen.go` via `go generate ./...`.
- ADR `wiki/decisions/0074-approval-route-versioning.md` (new); `wiki/decisions/0018-approval-route-lifecycle.md`
  §1 annotated as superseded.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Static (build) | `go build ./...` | clean, exit 0 | — |
| Static (vet, integration tags) | `go vet -tags integration ./tests/integration/approval/...` | clean, exit 0 | real |
| Targeted package suite | `go test ./internal/modules/documents/approval/...` | all subpackages PASS | real (fixture-driven fake-conn tests) |
| Runtime proof — migration + trigger + supersede sequence | `go test -tags integration ./tests/integration/approval/... -run TestRouteVersioning -v` against live `metaldocs-postgres` container (port 5433) | `TestRouteVersioning_NotInUse_InPlaceUpdateSucceeds` PASS (158.16s) — in-place update on a not-in-use route leaves `active=true`, `superseded_at` NULL; `TestRouteVersioning_InUse_DefinitionUpdateBlocked_ThenSupersede` PASS (13.37s) — definition-column UPDATE on in-use row raises P0001 `ErrRouteInUse`, active/superseded_at-only UPDATE on the same row succeeds, new row created with `version+1`/`active=true`/distinct id, old row's own stage rows (name unchanged) remain resolvable; `TestRouteVersioning_InUse_VersionColumnUpdateBlocked` PASS (18.00s) — bare version-only UPDATE on in-use row still raises P0001 (proves the exemption is columns-together, not any-single-column); `TestRouteVersioning_InactiveRoute_DefinitionUpdateBlocked` PASS (20.18s) — an already-superseded/inactive row still rejects definition-column updates | real (live DB, not fixture) |

Full run: `ok metaldocs/tests/integration/approval 214.735s`.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Route not in use: Update() still mutates in place, same id, version+1 | yes | `TestRouteVersioning_NotInUse_InPlaceUpdateSucceeds` |
| Route in use: Update() creates new row, old row superseded, old row's stages/definition untouched, new row id ≠ path id | yes | `TestRouteVersioning_InUse_DefinitionUpdateBlocked_ThenSupersede` |
| In-flight instance still resolves stages from its original (now-superseded) route row | yes | same test — old route's stage row read back unchanged after supersede |
| Direct SQL UPDATE of `name`/`profile_code`/`version` on an in-use or inactive row → P0001 tripwire | yes | `..._ThenSupersede`, `..._VersionColumnUpdateBlocked`, `..._InactiveRoute_DefinitionUpdateBlocked` |
| Direct SQL UPDATE of only `active`/`superseded_at` on an in-use row → succeeds | yes | `..._ThenSupersede` |
| Empty stage pool at submit → `ErrEmptyEligiblePool` → 422, no instance/stage rows persisted | yes | `submit_service.go` unit tests (fixture, real-shaped SQL against fake driver) + `http/errors.go` mapping to 422 confirmed by `go test ./internal/.../http/...` |
| No regression | yes | `go build ./...` clean; full approval module suite green (12 pre-existing test fixtures updated to match F2's intentional behavior changes — see Review disposition) |

## Review disposition

- Spec-compliance review: contract matched — transparent supersede on in-use Update, column-scoped
  trigger exemption, partial-active + version-history unique constraints, `ErrEmptyEligiblePool` reuse
  (no new sentinel), `Deactivate()` non-goal respected (verified via version-bump interaction with the
  trigger's exemption — recorded as a Consequence in ADR 0074).
- Code-quality review / TDD-discipline gap found and closed: implementation of the migration and
  service changes proceeded ahead of the testdb-factory live-DB integration test (a deviation from
  this feature's own `plan.md` step-2 ordering). This was caught and closed before evidence was
  written: `tests/integration/approval/route_versioning_test.go` was authored and run against the
  live DB in this pass, producing the 4 PASS rows above.
- Two pre-existing test-fixture regression classes were root-caused and fixed (not symptom-patched):
  1. `route_admin_service_test.go`'s `TestRouteAdminUpdate_HappyPath`/`TestRouteAdminUpdate_RouteInUse`
     used a fake `routeAdminConn` whose `lockedRouteVersion` (default 2, used by the pre-existing
     `Deactivate()` lock-based OCC pattern) didn't match the `ExpectedVersion` these two `Update()`
     tests passed (3), because `Update()` previously validated OCC via a bare `WHERE version=$N`
     clause rather than the lock-then-compare pattern `Deactivate()` already used. Since F2's
     `updateInPlaceOrSupersede` now needs the locked row state to decide the not-in-use vs in-use
     branch, `Update()` legitimately adopts the same lock-based OCC as `Deactivate()` — fixed by
     setting `lockedRouteVersion` to match each test's intended `ExpectedVersion`.
     `TestRouteAdminUpdate_RouteInUse` additionally required rewriting its assertion: the old test
     asserted `Update()` on an in-use route returns `ErrRouteInUse` to the caller — that is precisely
     the behavior this feature replaces with transparent supersede, so the test now asserts success
     with the new row's id/version instead.
  2. `submit_service_test.go` (11 tests) / `phase5_integration_test.go` (2 tests): their fake
     `ResolveEligibleActors` mock issues a real-shaped SQL query against `metaldocs.user_process_areas`
     that previously always resolved to zero rows (eligibility was never enforced at submit before this
     feature, so these fixtures never needed to seed it). Fixed by adding one query-match branch in
     each fake conn's `Query` dispatcher returning a single non-empty eligible-actor row — these tests
     exercise submit's other behaviors, not eligibility, so seeding a trivially-eligible pool is the
     correct repair (not a weakening of the new W6 business rule).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|--------------------------|
| None | — | — |
