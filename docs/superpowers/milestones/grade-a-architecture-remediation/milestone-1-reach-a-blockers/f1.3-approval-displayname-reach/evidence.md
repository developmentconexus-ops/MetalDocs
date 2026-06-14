# Feature F1.3 — Evidence

> **Milestone:** 1 (Reach-A Blockers)  ·  **Feature:** `f1.3-approval-displayname-reach`  ·  **Closed:** 2026-06-14
> **Contract:** `spec.md` (AC1–AC5). Closes the H-G-class signoff display-name cross-module reach.

## What was implemented

Contained the approval-signoff actor display-name read (Approach-3 step 1). The raw
`SELECT display_name FROM metaldocs.iam_users` that previously ran **inline inside the
lock-holding signoff transaction** is now a contained off-tx method on the approval
module's **own** `ApprovalRepository`, called pre-flight before the lock runner starts.

- `repository/approval_repository.go` — added `LoadActorDisplayName(ctx, tenantID, userID) (string, error)` to the `ApprovalRepository` interface, with a doc comment pinning the off-tx / H-PRE-1 / RLS rationale.
- `repository/postgres_approval_repository.go` — implemented `LoadActorDisplayName` on the pool (`r.db`), `sql.ErrNoRows → ("", nil)`, `MapPgError` otherwise. Query carries the explicit `tenant_id = $2::uuid` predicate so the GUC-less off-tx read stays tenant-correct under the migration-0237 RLS NULL-permissive policy.
- `application/decision_service.go` — hoisted the call **before** `runner.Do` (off-tx, pre-flight); deleted the inline `tx.QueryRowContext`; threads the result into `domain.NewSignoff(… ActorDisplayNameSnapshot: actorDisplayName …)`. No `iam_users` access remains inside the lock section.
- Producer matches consumer contract: the consumer (`decision_service`) calls a method **owned by the approval module**, not a shared/foreign port — exactly AC2. No generalization to a shared IAM port (that is M4/F4.1; HS-6 scope guard respected).

Commits:
- `7e128ed0` refactor(approval): contain signoff display-name read off-tx (H-PRE-1)
- `fdd304b3` refactor(approval): align LoadActorDisplayName err-handling to house style; pin signoff display-name call args in test
- `fb56f7ca` test(approval): live-DB integration proof for off-tx signoff display-name read (F1.3 AC5)

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Unit — threads display-name from repo, off-tx, correct args | `go test ./internal/modules/documents/approval/application -run TestRecordSignoff_ThreadsActorDisplayNameFromRepo` | PASS — asserts `insertedSignoff.ActorDisplayNameSnapshot()` == repo value AND arg-capture `loadActorDisplayNameCalled==true`, tenant=="tenant-1", user==actorID (call-order: before `runner.Do`) | real (logic) |
| Static (build) | `go build ./...` | exit 0 | — |
| Static (vet) | `go vet ./internal/modules/documents/approval/...` | exit 0 | — |
| Targeted suite | `go test ./internal/modules/documents/approval/...` (at `fdd304b3`) | 260 PASS / 0 FAIL / 8 SKIP | real |
| Runtime proof — live real-Postgres off-tx read-back | `go test -tags integration -run TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema ./internal/modules/documents/approval/repository/` against the running dev Postgres (:5433, migrated) | PASS — `LoadActorDisplayName(f13a…aa, approver-displayname-int-f13) = "Alice Approver"`; missing-user = `""` (nil err). Real RLS NULL-permissive policy, no GUC, `tenant_id::uuid` predicate, read on pool (`r.db`), no deadlock/hang. | **real** (real provider, real schema, real RLS) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| AC1 — read runs off-tx, before `runner.Do` | yes | `decision_service.go` hoist; unit arg-capture pins call-before-runner |
| AC2 — contained on approval's own `ApprovalRepository`, no shared port | yes | interface + impl in approval module; no IAM port introduced |
| AC3 — snapshot value unchanged semantics; empty-on-missing; unit test | yes | unit threading test + integration empty-on-missing `""`/nil |
| AC4 — build / vet / test green | yes | build 0, vet 0, 260 PASS suite |
| AC5 — live runtime proof: real value read back from real Postgres = actor's `iam_users.display_name`; no deadlock; by us not deferred | yes (substance) — see defer | integration read-back PASS against live migrated DB under real RLS; off-tx/no-deadlock proven by construction + green run. **Full live HTTP submit→signoff E2E is bounded-deferred** (test-infra drift, below) |

## Review disposition

- **Spec-compliance review:** SPEC-COMPLIANT. Confirmed off-tx placement, contained ownership (no shared port — HS-6 respected), value-threading unchanged.
- **Code-quality review:** CLEAN after `fdd304b3` (error-handling aligned to two-block house style; test args pinned via capture). No open findings.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Full live **HTTP** submit→signoff E2E persisting `approval_signoffs.actor_display_name_snapshot` end-to-end | The riskiest substance of AC5 (off-tx cross-schema read returns the correct value under real RLS, no deadlock, threaded into the snapshot) is already proven by the real method against the live migrated Postgres + the unit threading/arg-capture test. Only the outer HTTP workflow chrome is unproven. | **Trigger:** when the e2e/test-infra rebaseline is done (see next row), run `scripts/tmp/f13-e2e-proof.ps1` (seed→finalize→signoff) and read back the snapshot. **Owner:** M1 close regression / e2e-harness repair. |
| Post-v1-rebaseline **test-infra drift** repair (NOT F1.3 code) | Discovered 3 independent drift points while attempting the live E2E, all outside F1.3's boundary: (1) `internal/test/e2e_seed.go` `ensureTenant` targets removed legacy `public.tenants`; (2) same seed missing the migration-0231 `metaldocs.asserted_caps` GUC and missing post-rebaseline NOT-NULL columns (e.g. `document_profiles.alias`); (3) `tests/integration/testdb` curated bootstrap `0001_product_reference_data.sql` violates current `templates_template.visibility` NOT NULL. Each is a `metaldocs-database` / e2e-harness boundary fix; chasing them mid-F1.3 is HS-6 scope drift. Working tree left at HEAD (no half-applied seed repair committed). | **Trigger:** before any milestone that needs live HTTP E2E seeding (and before the deferred row above). **Owner:** `metaldocs-database` skill + e2e-harness maintainer. |
| Decision-record fold-in (H-PRE-1 / off-tx signoff read) into wiki | Documentation, not code. | **Trigger:** M1 close. **Owner:** `wiki-curator`. |
