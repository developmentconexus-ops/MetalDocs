# Feature F0.1 — Evidence

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Feature:** `f0.1-authz-effective-from`  ·  **Closed:** 2026-06-15
> **Contract:** `spec.md` (effective_from-only fix at the shared `Require` predicate; effective_to out of scope).

## What was implemented

- Added `AND upa.effective_from <= now()` (single clause) to the capability-grant query in
  `authz.Require` (`internal/modules/iam/authz/authz.go:123`). Conforms the tier-2 predicate to the
  canonical `ResolveEligibleActors` predicate
  (`internal/modules/documents/approval/repository/postgres_approval_repository.go:1140`).
- Producer ↔ consumer: every tier-2 caller of `authz.Require` now sees the consumer-contract
  behavior — a future-dated `user_process_areas` membership denies; current membership grants.
  No caller code, no schema, no other clause changed.
- Updated two sqlmock-mirrored-SQL expectations that hardcoded the prior predicate
  (`internal/modules/auth/application/service_test.go:980`,
  `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go:73,159`). Both are
  mock harness expectations that mirror the `Require` SQL verbatim — not callers; not a symptom
  patch.
- Added `//go:build integration` proof:
  `internal/modules/iam/authz/authz_effective_from_integration_test.go`.

Commits: pending close-out commit on `main`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `go test -tags integration -run 'TestRequire_FutureDatedMembershipDenied\|TestRequire_CurrentMembershipGranted' ./internal/modules/iam/authz/ -v` | RED on pre-fix code: `authz_effective_from_integration_test.go:69: Require with future-dated membership = <nil>, want ErrCapDenied (premature access must be denied) --- FAIL: TestRequire_FutureDatedMembershipDenied (1.60s)` ; GREEN on post-fix code: `--- PASS: TestRequire_FutureDatedMembershipDenied (1.47s) --- PASS: TestRequire_CurrentMembershipGranted (0.14s)` | **real** (live Postgres via `testdb.Open`, curated bootstrap) |
| Static (build) | `go build ./...` | clean (no output) | — |
| Targeted unit | `go test ./internal/modules/iam/authz/` | `ok  metaldocs/internal/modules/iam/authz  1.186s` | fixture (sqlmock) |
| Targeted mocks updated | `go test ./internal/modules/iam/infrastructure/postgres/ ./internal/modules/auth/application/` | both `ok` | fixture (sqlmock) |
| Targeted integration | `go test -tags integration -count=1 -run 'TestRequire_FutureDated\|TestRequire_Current' ./internal/modules/iam/authz/ -v` | `PASS ok metaldocs/internal/modules/iam/authz 1.903s` | **real** |
| iam integration (regression, excl. pre-existing ProbeA) | `go test -tags integration -run 'TestProbe[B-Z]\|TestUpsert\|TestReplace' ./internal/modules/iam/...` | all `ok` | **real** |
| Whole-repo unit | `go test ./...` | grep `FAIL` returns empty — all packages pass (the iam packages all `ok`; auth/application `ok` after mock update) | mixed |
| Runtime proof of the consumer-contract change | The integration test IS the runtime proof — `Require` invoked against a real Postgres seeded with a future-dated membership returns `ErrCapDenied`; with a current membership returns nil. | see TDD row | **real** |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Future-dated membership → `Require` returns `ErrCapDenied` | yes | TDD row, `TestRequire_FutureDatedMembershipDenied` PASS |
| Current membership → `Require` returns `nil` (grant) | yes | TDD row, `TestRequire_CurrentMembershipGranted` PASS |
| No regression in existing authz behavior (unit) | yes | Targeted unit row |
| Whole-repo green | yes | Whole-repo unit row + targeted integration row |

## Review disposition

- Spec-compliance review (self): producer matches the spec's consumer contract — predicate added,
  `effective_to` clause and bypass/cache/asserted-caps paths untouched, no caller modified, no
  schema change. Non-goals respected (no `effective_to > now()` change). PASS.
- Code-quality review (self): one-line predicate addition at the shared point (ADR-0022 root-cause
  rule respected). Three sqlmock-expectation copies updated because they mirror the SQL verbatim —
  that is a known harness cost of the sqlmock pattern, not a symptom patch in production code. No
  dead code, no widened scope. PASS.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Align `Require`'s `effective_to IS NULL` to `(effective_to IS NULL OR effective_to > now())` (the other half of the `ResolveEligibleActors` predicate) | Out of scope per F0.1 spec interview Q2 — changing it **widens** access (a membership scheduled to end in the future is currently denied; aligning would grant it). Not an F5.1-confirmed finding; needs its own security-direction decision. | Trigger: if a re-audit or operator flags premature-revocation as a defect. Owner: grade-a-completion operator. |
| Pre-existing `TestProbeA_DirectInsertUserProcessAreasBlocked` failure (P0001 vs expected 42501) | Reproduced on clean `main` HEAD before F0.1 changes — not introduced by this feature. Unrelated tripwire-vs-RLS drift in a probe assertion. | Trigger: M0 milestone-validator may flag as a regression-bar question; if it isn't a F0.1 regression, it stays as an existing iam-suite issue for a future fix feature (not in this mission's §5 inventory). |
| **Operational impact note (not a defer — a flag for the operator):** if any live `user_process_areas` row has `effective_from > now()` today, that grant is now correctly denied at runtime (B1 security fix). | This is the intended behavior change. | Trigger: none — this **is** the fix. Operator should be aware before deploy. |
