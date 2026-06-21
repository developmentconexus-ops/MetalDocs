# Feature F3.3 — Evidence

> **Milestone:** 3  ·  **Feature:** `f3.3-approval-consume-view` (C3, in-tx, H-PRE-1)  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` — `ResolveEligibleActors` reads iam's published `metaldocs.v_active_user_areas`
> in-tx (ADR-0039 D3a), not `metaldocs.user_process_areas` with an interval predicate.

## What was implemented

- **EDIT** `internal/modules/documents/approval/repository/postgres_approval_repository.go` —
  `ResolveEligibleActors` (`:1133`) in-tx SELECT repointed: `FROM metaldocs.user_process_areas WHERE …
  AND effective_from <= now() AND (effective_to IS NULL OR effective_to > now())` →
  `FROM metaldocs.v_active_user_areas WHERE tenant_id=$1 AND area_code=$2 AND role=$3`. The **entire
  temporal predicate dropped** (the view encodes active-now). `role` filter kept. Method doc rewritten to
  record the view-read, the ADR-0037 Model-A set-equality, and the H-PRE-1 preservation. **Only the
  relation token + dropped predicate changed** — signature, `db.Tx` posture, scan loop, never-nil contract
  all unchanged.
- **NEW** `internal/modules/documents/approval/repository/eligible_actors_view_parity_integration_test.go` —
  in-tx parity gate calling `repo.ResolveEligibleActors` on a real `*sql.Tx` vs a **verbatim inline copy of
  the deleted interval SQL** on the same tx.
- **EDIT** `tools/cilint/internal/analyzers/hgcrossmodule.go` — drained the
  `{documents/approval/repository/postgres_approval_repository.go, user_process_areas}` (C3) ledger entry,
  replaced with a "ported (M3/F3.3)" note. No cilint fixture edit needed (the `PendingBaseline` fixture was
  realigned to the C4d search site in F3.2).

## Verification

| Check | Command | Result (evidence) | Real vs fixture |
|-------|---------|-------------------|-----------------|
| TDD — parity green **before** raw deleted (D6 sanity: raw repo == raw baseline) | `go test -tags integration -run TestResolveEligibleActors_ViewParityWithRaw ./…/approval/repository/` (pre-repoint) | `ok …/approval/repository 3.225s` | real (PG :5434) |
| Repoint correct — parity green **after** raw deleted (view repo == raw baseline, zero authz drift) | same command (post-repoint) | `ok …/approval/repository 3.238s` | real (PG :5434) |
| Build | `go build ./...` | `BUILD-DONE` (exit 0) | — |
| Guard exit 0, C3 drained | `go run ./tools/cilint ./...` | `cilint-exit=0` | real |
| Cilint unit suite green | `go test ./tools/cilint/...` | `ok …/analyzers 2.027s` | real |
| `user_process_areas` gone from approval **production** code | `git grep -n user_process_areas -- internal/modules/documents/approval/` | only matches are test files (seeds/assertions) + `http/contracts/route.go:150` (a **comment**); `postgres_approval_repository.go` has **zero** | real |
| Approval-tree regression | `go test -tags integration ./internal/modules/documents/approval/...` | `application`, `domain`, `http`, `http/contracts`, `infrastructure`, `jobs`, `repository`(parity) **ok**; two FAILs — `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` + `TestPostgresLimiter_Live` — are pre-existing raw-DSN defers (proven below), NOT F3.3 | real (PG :5434) |
| Pre-existing-defer proof | stash F3.3 code (`postgres_approval_repository.go`, `hgcrossmodule.go`) → rerun the two tests on HEAD | both FAIL identically on HEAD: `relation "metaldocs.tenants" does not exist` / `relation "public.auth_failure_counters" does not exist` | real |

> **H-PRE-1 (critical) — preserved.** The diff is a single relation token (`user_process_areas` →
> `v_active_user_areas`) plus the dropped temporal predicate, inside the *same* `tx.QueryContext` call.
> No authz-recording read, no lock, no extra round-trip, no tx-structure change was introduced. The view is
> `SELECT`-only (no `security_invoker`, no function), so it is a structurally identical in-tx read. D5 holds.
>
> **Authz-drift discriminators (the load-bearing proof):** seeded two eligible (active qms_admin), a
> wrong-role (approver), a **revoked** (past `effective_to`), a wrong-area (qms_admin on `safety`), and a
> wrong-tenant member. Post-repoint `ResolveEligibleActors(quality, qms_admin)` = exactly {eligibleA,
> eligibleB}; revoked/wrong-role/wrong-area/wrong-tenant all excluded — identical to the interval-form
> baseline. No eligibility change.

## Predicate-equality (ADR-0037 Model A)

The deleted predicate was an **interval** form; the view is `effective_to IS NULL`. They select the same
set because, under Model A (the only write path): `effective_from` is stamped `now()` at insert (never
future) ⇒ `effective_from <= now()` always true; revoke stamps a **past** `effective_to` and no future
grants exist ⇒ `effective_to > now()` is empty ⇒ the interval reduces to `effective_to IS NULL`. The only
rows on which the two could differ (future-dated `effective_to`/`effective_from`) are **unreachable** via
the write path, so the repoint is a behavior-preserving tightening that also retires the Model-B leak
(ADR-0037 D2). The parity test seeds only write-path-reachable rows and proves the equality empirically.

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| `ResolveEligibleActors` view-form set == raw interval-form set on real `*sql.Tx`, all scopes | yes | `TestResolveEligibleActors_ViewParityWithRaw` GREEN |
| Discriminators: both eligibles present; revoked/wrong-role/wrong-area/wrong-tenant absent | yes | same test's assertions |
| Parity green before raw deleted (D6) | yes | pre-repoint GREEN, repoint, post-repoint GREEN |
| H-PRE-1 preserved (plain non-recording in-tx SELECT, no tx-structure change) | yes | one-token diff; method doc + review note |
| Build + approval tree | yes | `go build` exit 0; approval packages ok (two pre-existing raw-DSN defers proven) |
| Guard exit 0, C3 entry drained | yes | `cilint-exit=0`; ledger note |
| `git grep` approval production → 0 raw reads | yes | only test files + one comment remain |
| Cilint unit suite green | yes | `ok …/analyzers 2.027s` |

## Review disposition

- Spec-compliance review: **PASS** — only the relation + temporal predicate changed; signature, in-tx
  posture, scan loop, never-nil contract untouched; no temporal predicate re-derived in approval SQL.
- Code-quality review: **PASS** — H-PRE-1 preserved by construction (single-token relation swap in the same
  `tx.QueryContext`); parity test asserts set-equality against the verbatim deleted interval SQL with revoked
  + wrong-role + wrong-area + wrong-tenant discriminators; ledger drained in-feature. The Model-A
  set-equality is documented and empirically locked.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` FAIL on raw base DSN | Pre-existing (proven identical FAIL with F3.3 stashed); opens `METALDOCS_DATABASE_URL` directly, needs a hand-provisioned `metaldocs.tenants`; orthogonal to F3.3 | Trigger: test migrated onto the testdb bootstrap harness; owner: mission backlog (not M3 scope) |
| `TestPostgresLimiter_Live` FAIL on raw base DSN | Pre-existing (proven identical FAIL with F3.3 stashed); needs `public.auth_failure_counters` from full bootstrap | Trigger: same harness migration; owner: mission backlog (not M3 scope) |
