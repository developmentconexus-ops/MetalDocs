# F2.2 plan — verify + pin + doc-truth restore

> Executes `spec.md` against `../validation-contract.md §2`. TDD (failing test first for the pin).
> Disjoint from F2.1 (touches permissions tests + ADR/wiki docs; no tripwire files).

## Tasks (one sonnet subagent)

- **T1 Verify (evidence, no change).** Capture source excerpts proving alignment:
  `permissions.go` force-release row + `/approval/routes*` rows and their first-match ordering vs the
  generic `/approval/` fallback; `repository.go:798,828` (`membership.manage`); the route-management
  service's `route.manage` assertion. Confirm `permissions_test.go` already pins the `/approval/routes`
  ordering.
- **T2 Regression pin (TDD).** Add a **targeted** test (extend `permissions_test.go` or
  `permissions_authz_scope_test.go`) asserting, for the two reconciled routes only, that the tier-1
  route→capability resolution equals the tier-2 asserted capability
  (`membership.manage` for force-release; `route.manage` for approval-route management). Write it to
  **fail first** against a temporary local re-divergence (flip the tier-1 cap), confirm RED, revert,
  confirm GREEN on HEAD. Capture both.
- **T3 Doc-truth restore.** In `wiki/decisions/0022-authz-capability-coherence.md`, back-annotate the
  stale Phase 7/8 ⚠️-follow-up lines (≈198, 236–237, 250) with a `RESOLVED in Phase 11 F4 (see
  §349–351)` note for each of the two sites. Grep the wiki for any other page describing force-release
  or approval-route management as an **open** tier-1/tier-2 divergence and correct it. Record the
  deliberate coarse/fine `/approval/` difference as an intentional exception where the ADR discusses it.
  (Dispatch `wiki-curator` if broad wiki edits are needed; otherwise edit inline.)

## Verification (before F2.2 commit)
`go test ./apps/api/... -run 'Permission|Authz'` (or the packages holding the permissions tests) GREEN ·
`go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` — **all 5 authz lints green** ·
`go build ./...` green · ADR 0022 diff shows only annotations (no behavior claim change).
