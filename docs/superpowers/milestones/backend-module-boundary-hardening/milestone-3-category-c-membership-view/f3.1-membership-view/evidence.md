# Feature F3.1 — Evidence

> **Milestone:** 3  ·  **Feature:** `f3.1-membership-view`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer contract: published view `metaldocs.v_active_user_areas` = active-now projection of `public.user_process_areas`, columns `(tenant_id, user_id, area_code, role)`).

## What was implemented

- **NEW** `db/migrations/0242_iam_v_active_user_areas_view.sql` — creates `metaldocs.v_active_user_areas`
  as `SELECT tenant_id, user_id, area_code, role FROM public.user_process_areas WHERE effective_to IS NULL`,
  with a `COMMENT ON VIEW` recording it as the ADR-0039 D3(a) published contract and the mandatory
  `public.schema_migrations` row (`ON CONFLICT DO NOTHING`). No `security_invoker` — matches the existing
  `metaldocs.user_process_areas` exposure view's RLS posture.
- **NEW** `internal/modules/iam/infrastructure/postgres/active_user_areas_view_parity_integration_test.go` —
  the view-vs-base set-equality parity test (the producer proves itself before any consumer repoints).
- **EDIT** `wiki/decisions/0039-cross-module-base-table-read-boundary.md` — "Related code" annotated with the
  concrete migration path (no decision change; the decision is M0/F0.1 ADR-0039 D3(a)/D4).
- **Producer matches consumer contract:** columns are exactly the union the three M3 consumers consume
  (`role` included because approval C3 filters by it); the active predicate is exactly `effective_to IS NULL`
  (ADR 0037 D1 / ADR-0039 D4); temporal columns deliberately not exposed. No consumer repointed (F3.2/F3.3).
- Commit: _(staged with F3.1; sha recorded at commit below)_.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first, then green | `go test -tags integration -run ActiveUserAreasView ./internal/modules/iam/infrastructure/postgres/` | **RED** before migration: `relation "metaldocs.v_active_user_areas" does not exist (SQLSTATE 42P01)`; **GREEN** after migration: `ok metaldocs/internal/modules/iam/infrastructure/postgres 3.324s` | real (PG :5434) |
| Migration applies in full bootstrap | `go test -tags integration ./tests/integration/testdb/...` | `ok metaldocs/tests/integration/testdb 3.572s` (baseline + all `db/migrations` incl. 0242 apply clean) | real (PG :5434) |
| Static build | `go build ./...` | `BUILD-OK` (exit 0) | — |
| Guard unchanged (no ledger edit) | `go run ./tools/cilint ./...` | `cilint-exit=0`; `hgcrossmodule.go` not modified | real |
| Cilint unit suite unaffected | `go test ./tools/cilint/...` | `ok metaldocs/tools/cilint/internal/analyzers 4.935s` (PendingBaseline fixture still valid — C1+C2 not yet drained) | real |

> Set-equality, not count: the test ORDER BYs both `metaldocs.v_active_user_areas` and
> `public.user_process_areas WHERE effective_to IS NULL`, compares row slices, and adds explicit
> authz-drift assertions — active rows (area1/qms_admin, area2/approver) **present**, revoked row (editor)
> **excluded**, wrong-tenant row **not leaked**.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Migration applies cleanly in full bootstrap on test PG | yes | testdb bootstrap row above |
| View returns exactly the base-table active-now set (set-equality over seeded active/revoked/multi-area/wrong-tenant) | yes | `TestActiveUserAreasView_ParityWithBaseActiveNow` GREEN row |
| View encodes active-now (revoked excluded, active present) | yes | same test's drift assertions |
| Build unaffected | yes | `go build ./...` BUILD-OK |
| Guard unchanged (no F3.1 ledger edit) | yes | `cilint-exit=0`, no `hgcrossmodule.go` diff |

## Review disposition

- Spec-compliance review: **PASS** — view shape = consumer-derived contract (cols + `effective_to IS NULL`);
  no consumer touched; no temporal columns exposed; existing passthrough view untouched.
- Code-quality review: **PASS** — migration is forward-only, idempotent (`ON CONFLICT DO NOTHING`),
  transactional, commented with ADR refs; test asserts set-equality + drift exclusions (not a count). PK /
  partial-unique-index constraints on `user_process_areas` were discovered empirically during RED
  (`user_process_areas_pkey`, `ux_user_process_areas_one_active`) and the seed corrected to two areas — a
  test-fixture fix, not a contract change.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| none | — | — |
