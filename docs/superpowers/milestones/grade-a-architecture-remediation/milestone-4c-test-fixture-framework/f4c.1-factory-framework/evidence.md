# Feature F4c.1 — Evidence

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.1-factory-framework`
> **Status:** Closed (evidence complete) — 2026-06-15
> Judged against `spec.md` (approved pre-code) and `plan.md`. All proof is **real** (template-cloned
> Postgres under the operator DSN), not fixture/mock.

## What shipped

- `tests/integration/testdb/factory.go` (new, build tag `integration`, package `testdb`) — functional-builder
  factories generalizing `fixtures.go`: `NewTenant`, `NewUser` (`WithRole`), `NewTaxonomy`, `NewControlledDoc`,
  `NewDocument` (`WithStatus`/`WithRevisionVersion`/`WithScheduleGen`/`WithRevisionNumber`/`WithEffectiveFrom`),
  `NewApprovalRoute`, `NewApprovalInstance`, and the `Scenario` composite (`PublishedDocument`, `ScheduledRevision(gen)`).
  Shared `Spec`/`Opt` so the generic `WithX` names work across builders; auto-wires missing FK parents; mints
  fresh UUIDs + per-call-unique taxonomy codes; asserts the real tripwire cap via `seedWithCaps` (tx-local).
- `tests/integration/testdb/factory_test.go` (new) — TDD self-test, 9 subtests.
- **Untouched:** `db.go` (empty-diff gate), `fixtures.go`, `pgtest`, all consumer test files. Production source untouched.

## TDD (red → green)

Test written first (`factory_test.go`); builders undefined → compile failure (red). Implementing `factory.go`
surfaced **two real curated-baseline schema gates** the consumers' current shared-DB seeds never hit (the drift
this milestone targets), each fixed at root cause — by seeding what the real schema requires, **not** by weakening
any constraint:

| Red | Root cause | Fix |
|-----|-----------|-----|
| `snapshot columns required for status=scheduled/approved/published (SQLSTATE 23514)` | `enforce_snapshot_on_submit` fires on **INSERT** for any non-draft status on the curated baseline; the shared dev DB the approval tests run on is schema-drifted and lacks it | `NewDocument` stubs the 6 snapshot columns (`placeholder_schema_snapshot`/`_hash`, `composition_config_snapshot`/`_hash`, `body_docx_snapshot_s3_key`/`body_docx_hash`; 32-byte hashes) for non-draft status, mirroring `fixtures.go` `SupersedeActiveDocumentForCD` |
| `null value in column "created_by" of relation "approval_routes" (SQLSTATE 23502)` | `approval_routes.created_by` is NOT NULL on the curated baseline (consumer omits it — drift) | `NewApprovalRoute` mints/accepts an owner and sets `created_by` |

> These two finds are the milestone thesis in miniature: the curated-baseline (good) harness enforces gates the
> drifted shared DB silently skips. The factory satisfies the real schema; the tripwire and the CHECKs are never
> weakened, disabled, or edited.

## Validation Gate — criterion → proof

Env (both set to operator DSN): `$env:METALDOCS_DATABASE_URL`, `$env:DATABASE_URL`.
Command: `go test -tags integration -count=1 -v -run TestFactory ./tests/integration/testdb/...`

| Acceptance criterion (spec) | Proof | Result |
|-----------------------------|-------|--------|
| Each builder seeds its rows + returns asserted IDs/columns | `TestFactory/NewTenant…`, `…/NewUser_with_role…`, `…/NewControlledDoc_autowires_parents`, `…/NewDocument_status_and_numeric_overrides` | PASS (real) |
| FK parents + real tripwire cap satisfied (wrong/absent cap → P0001; missing parent → 23503) | all builder subtests green via `seedWithCaps` | PASS (real) |
| Two factory calls in one DB do not collide | `TestFactory/two_calls_one_db_no_collision` (two CDs + two same-tenant taxonomies, distinct codes, no `document_profiles_pkey` 23505) | PASS (real) |
| Minted taxonomy codes match `^[a-z][a-z0-9_-]{1,63}$` | `TestFactory/NewTaxonomy_codes_match_format_and_seed_rows` | PASS (real) |
| `Scenario.PublishedDocument` / `ScheduledRevision(gen)` produce consumer shapes | `…/scenario_published_document` (status published), `…/scenario_scheduled_revision` (status scheduled, gen 3) | PASS (real) |
| Approval route+instance chain wires | `TestFactory/approval_route_and_instance_chain` | PASS (real) |
| Harness untouched | `git diff --exit-code tests/integration/testdb/db.go` → exit 0 | PASS |

Final run: `ok metaldocs/tests/integration/testdb 182.782s` — all 9 subtests PASS. `go vet -tags integration
./tests/integration/testdb/...` → exit 0. (First subtest absorbs the one-time template-DB build ~133s; subsequent
clones run 2–16s each.)

## Scope / non-goals honored

- No consumer migration (F4c.2/.3). No `db.go` edit (empty diff verified). No `templates_template`/editor-`templates`
  builders (`NewDocument` uses a free `template_version_id`). No CI grep-guard (F4c.4). No wiki/ADR (F4c.5; the
  framework decision ADR lands there). No production-source change. No tripwire/CHECK weakening. `pgtest` untouched.
- `git status` for `tests/integration/testdb/`: only `factory.go` + `factory_test.go` added; `fixtures.go` unchanged.

## Bounded defers (with triggers)

- **D-F4c.1-a** — `NewApprovalInstance` asserts `document.submit` and `NewApprovalRoute` writes `created_by`; whether
  approval_routes/approval_instances need additional caps on the curated baseline beyond what the self-test exercised
  will be confirmed when **F4c.2** migrates the 5 approval/repository + 3 approval/jobs tests onto these builders (their
  assertions exercise the real read paths). Trigger: F4c.2 execution.
- **D-F4c.1-b** — taxonomy-code override options (`WithProfileCode` etc.) and editor-lineage/snapshot-display-name
  builders are **not** built in F4c.1 (no consumer needs them yet). Trigger: a consumer in F4c.2/.3 that asserts on a
  specific taxonomy code or on `created_by_display_name_snapshot` (the F4.1a Gate #5 snapshot test).

## Next

F4c.1 closed. **Stop here** — operator approved starting F4c.1 only. Await operator go-ahead before F4c.2
(migrate the M4-blocker files onto these factories, discarding the abandoned WIP).
