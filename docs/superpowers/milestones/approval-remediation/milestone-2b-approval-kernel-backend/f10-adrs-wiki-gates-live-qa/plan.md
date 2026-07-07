# F10 — Plan (lightweight, closing feature)

1. Verify ADR structure/content for 0074/0075/0076/0077; fix index gaps in
   `wiki/decisions/index.md`.
2. Dispatch wiki-curator (background agent) to sync `wiki/modules/approval.md`
   for F1-F9 combined; verify its output.
3. Run full cross-feature sweep: `go build ./...`, `go build -tags integration
   ./...`, `go test -count=1 ./...`, `go run ./scripts/api-lint -strict
   api/openapi/v1/openapi.yaml .`, grep-zero checks, `TestCapabilityRegistrySize`.
4. Live QA: rebuild via `.\scripts\start-api.ps1 -Build`, confirm readiness,
   log in as dev-seed `admin`, walk as much of the approval lifecycle as is
   reachable through the current HTTP contract. Fix any real bug found along
   the way (in-scope: bugs in F1-F9's own code, discovered by this feature's
   own verification work — same class as F4-F9 fixing bugs found in their own
   QA).
5. Write spec/plan/evidence; self-review against the system-impact analysis'
   locked constraints; stage and commit only F10-touched paths.

## Deviation from plan (recorded, not hidden)

Two real production bugs were found during live QA (not anticipated in this
plan) and fixed in-line, since they block the very lifecycle walkthrough F10
is required to perform:

- Bug #1: `MapPgError` didn't recognize F2's renamed unique constraint
  (`approval_routes_active_profile_uq`), so a genuine duplicate-route-profile
  collision surfaced as a 500 instead of 409.
- Bug #2: `updateInPlaceOrSupersede` attempted a speculative UPDATE without a
  SAVEPOINT, so Postgres's transaction-abort semantics turned the correctly
  triggered `ErrRouteInUse` fallback path into an opaque 500
  (SQLSTATE 25P02) instead of a working transparent-supersede.

Both are scoped, minimal fixes with unit test coverage added and live HTTP
re-verification. See evidence.md for full detail.
