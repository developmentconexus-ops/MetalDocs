# Milestone 4c — Unified test-fixture framework

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** Spec (drafting) — *authored before any feature in this milestone began.* Awaiting operator agreement (Phase 2 / HS-1-open).
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is, **which features**
> it contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of each feature lives in that feature's `plan.md`. The
> end-of-milestone QA (`qa/milestone-qa.md`) validates the milestone against *this* document.

## Origin

This milestone is a **fork of M4's close gate**, opened by operator decision 2026-06-15 ("Full
framework now"). The M4 close gate is still blocked: after the F4b drop migration (`071931c9`, the
legacy `metaldocs.documents` cluster), M4b's remaining seed-fix path (F4b.3 / F4b.4) was found to be a
**test-architecture defect**, not a seed bug. Two integration-test harnesses coexist with no
governance rule:

- `tests/integration/testdb/db.go` — **template-DB-cloned-per-test** (`CREATE DATABASE … TEMPLATE`
  via `ApplyCuratedBootstrap`, db.go:260: prerequisites → baseline → reference-data → migrations).
  This is the IntegreSQL pattern — fresh DB per test, **zero cross-test state**. The GOOD harness.
- `internal/testsupport/pgtest/pgtest.go` — `OpenAndMigrate(t)` connects to the **SHARED dev DB**
  (`METALDOCS_DATABASE_URL` / `DATABASE_URL`). The shared-DB model is the collision source: ~35
  `*_test.go` hardcode tenant UUIDs, ~28 copy-paste inline `set_config('metaldocs.asserted_caps', …)`
  (one with `is_local=false` → leaks caps across tests on the shared connection), ~11 copy-paste
  taxonomy INSERTs, and 5 local seed helpers re-implementing the same `BEGIN→set_config→INSERT→COMMIT`.
  Because taxonomy code columns are globally primary-keyed but `controlled_documents` FKs are composite
  `(tenant_id, code)`, a fixed code (`'po'`/`'quality'`) can belong to exactly one tenant globally —
  so two tests on the shared DB fight over one code and collide (`document_profiles_pkey` 23505).

A good shared seed layer already exists but is bypassed: `tests/integration/testdb/fixtures.go`
(`seedWithCaps` tx-local `is_local=true` pool-safe, `SeedGovernedTaxonomy`, `SeedSystemAdmin`,
`InsertDraftDocument`, `SupersedeActiveDocumentForCD`). The new factories **generalize** these.

Operator decision 2026-06-15: build the full framework (factories + migrate ALL stateful test files +
CI grep-guards) **as its own milestone**, before closing the program — not a symptom patch. This
milestone **supersedes M4b's remaining F4b.3 (seed-fix) and F4b.4 (suite-green)**; M4b's F4b.1
(census) and F4b.2 (drop migration, committed) stand. There are **uncommitted WIP edits** in the tree
from an abandoned per-tenant-code symptom-patch (`postgres_approval_repository_test.go` +
`scheduled_publish_job_test.go`, currently FAIL with `document_profiles_pkey` 23505; plus minor
asserted-caps edits in `repository_revision_history_integration_test.go` + `authz_bypass_test.go`).
That WIP is **replaced** by the factory migration in F4c.2 — it is not built upon.

## Objective

Replace the drifted, copy-pasted integration-test seeding with **ONE unified, professional
test-fixture framework** built on `testdb` (template-DB-per-test as the single isolation model for
stateful tests), migrate **all** stateful tests onto it, and enforce the discipline with CI
grep-guards. The bar this moves: **the M4 close-gate blocker re-greens at root cause** —
`TestCreateDocumentTx_PopulatesAllSnapshotColumns` (F4.1a Gate #5) passes and the approval
repo/jobs/commit_upload/fillin tests go green **from a clean baseline under the operator DSN, with
`tests/integration/testdb/db.go` unchanged from HEAD** (empty diff = structural proof the framework
fixed the real test architecture, not the harness). Terminal effect: the full integration suite is
green from a clean curated baseline, unblocking the M4 validator re-dispatch and, transitively, M5.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4c.1 | `f4c.1-factory-framework` | Add `tests/integration/testdb/factory.go` — functional-builder factories with sane defaults + `WithX` options, built **on** `testdb` (not a new package), generalizing `fixtures.go`. Mint fresh UUIDs + per-tenant-unique taxonomy codes (matching the `^[a-z][a-z0-9_-]{1,63}$` CHECK), assert the correct tripwire cap via `seedWithCaps` internally (tx-local `is_local=true`), and return structs carrying the IDs tests assert on. Builders: `NewTenant`, `NewUser` (`WithRole`), `NewTaxonomy`, `NewControlledDoc`, `NewDocument` (`WithStatus`/`WithRevisionVersion`/`WithScheduleGen`); plus a `Scenario` composite (`PublishedDocument`, `ScheduledRevision(gen)`). TDD: a self-test exercising each builder against a fresh template-cloned DB. | Factory self-test green from clean baseline under the operator DSN: each builder seeds the contracted rows, satisfies every FK and the tripwire (real cap asserted, **never** weakened/disabled), and two factory calls in one suite do **not** collide (per-tenant-unique codes / clone isolation). Generated taxonomy codes match the `profile_code_format` CHECK. `git diff tests/integration/testdb/db.go` empty. |
| F4c.2 | `f4c.2-migrate-blocker-files` | Migrate the M4-blocker files onto the factories, discarding the abandoned WIP: `documents/approval/repository` (`postgres_approval_repository_test.go`, delete `seedGovernedParents`), `documents/approval/jobs` (`scheduled_publish_job_test.go`, delete `seedScheduledDocument`), `documents/repository` (`repository_commit_upload_integration_test.go`, delete `seedTenantRole`; Family-C missing-`tenants` seed), `documents/repository` (`fillin_repository_integration_test.go`). For fillin, first **determine** whether its failure is a real Family-A schema-shadow needing **HS-2** escalation or a seed gap. | §4-blocker tests all green from a clean baseline under the operator DSN: `TestCreateDocumentTx_PopulatesAllSnapshotColumns`; approval/repository (`TestValidateScheduledSupersedeTarget_RealRows`, `TestLoadCurrentPublishedHeadForDocument_RealRows`, `TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion`, `TestLoadInstance_LoadsDocumentRevisionVersion`, `TestScheduleGenerationIncrementsOnScheduledWritePath`); approval/jobs (`TestScheduledPublishWorker_{PublishesWhenTruthMatches,NoOpWhenGenerationIsStale,NoOpWhenDeliveredBeforeEffectiveTime}`); commit_upload + fillin. `git diff tests/integration/testdb/db.go` **empty** (fix-not-adapt). The 5 abandoned local helpers in scope are deleted. |
| F4c.3 | `f4c.3-migrate-remaining` | Migrate the remaining stateful test files (the other 3 local seed helpers — `create_document_snapshot_integration_test.go` `seedCreateDocumentSnapshotRows`, `template_version_reader_integration_test.go` `seedTemplateVersionStateRows` — and the ~35 hardcoded-tenant / ~28 inline `set_config` / ~11 copy-paste-taxonomy sites) onto the factories. Convert shared-DB `pgtest` stateful tests to the template-DB model. | Migrated suites green from clean baseline under the operator DSN. Per migrated file: no inline `set_config`, no `is_local=false`, no hardcoded tenant UUID literal, no bare unqualified `documents` (uses `testdb.Qualified`). The 5 local seed helpers are gone. `db.go` diff still empty. |
| F4c.4 | `f4c.4-ci-grep-guards` | Add a CI grep-guard (and document the rule) enforcing the discipline in `*_test.go`: no bare `documents` (use `Qualified`), no inline `set_config`, no `is_local=false`, no hardcoded tenant-UUID literal (identity via factories or one `testdb.DevTenant` const). Retire/scope `pgtest` to genuinely no-write tests (or migrate them too) and delete any now-dead local helpers. | The grep-guard runs in CI, **fails** on a planted violation (each of the 4 rules), and **passes** on the migrated tree (exit 0). `pgtest` remaining callers are documented as no-write (or zero). Full integration suite green from clean baseline under the operator DSN (captured output). |
| F4c.5 | `f4c.5-docs-adr` | Document the framework: a wiki page (`wiki/quality/` or `wiki/architecture/`) describing the harness choice (IntegreSQL template-DB-per-test), the factory API, the discipline rules, and a "how to write an integration test" guide; an **ADR** under `wiki/decisions/` for the test-fixture-framework decision; dispatch `wiki-curator` after. | The wiki page + ADR exist, are linked from the relevant index, and accurately describe the shipped `factory.go` API and the CI grep-guard rules (no drift). ADR carries a canonical `Status:` header. |

For each feature, "what to validate" is objectively checkable — a named test that passes, an empty
diff, a grep-guard that fails-on-violation/passes-on-clean, a doc that matches the shipped API. No
"works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For this
milestone the gate enforces:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate"; each
   feature's `spec.md` consumer contract (the test files / CI / docs that consume the factory API) is
   honored. The factory API is read **from its consumers** (the migrated tests), not invented.
2. **The M4-blocker bar (root cause)** — re-measured from a **clean baseline under the operator DSN**:
   the §4 blocker tests are green **and `git diff tests/integration/testdb/db.go` is empty** (db.go at
   HEAD). A green achieved by editing `db.go` or by re-introducing per-test inline seeding is a
   **FAIL**. Family A (42703) and Family B (P0001) classes both at 0.
3. **Workflow-class QA** — `wiki/quality/backend-api-qa-checklist.md` (the changed code is test
   infrastructure for the backend) + the database rules in `wiki/database/` for any
   schema/reference-data touch + `release-closeout-checklist.md`. **No production runtime/route/contract
   change** (this is a test-layer milestone — production source untouched except where a genuine
   Family-A schema defect is found and escalated via HS-2).
4. **Discipline enforced, not asserted** — the CI grep-guard exists, fails on a planted violation of
   each rule, and passes on the migrated tree; the 5 local seed helpers and the inline-seed drift are
   actually removed (grep proof), not merely wrapped.
5. **Regression** — M0–M3 gates still pass; M4's built features (F4.1–F4.6) and M4b's drop migration
   are unaffected; the full integration suite (`go test -tags integration -count=1 ./...`) is green
   from a clean baseline under the operator DSN.
6. **No unplanned scope** — only test infrastructure + its docs/ADR + CI guard are changed. Any
   production-source change is a FAIL unless it is an HS-2-escalated, operator-approved Family-A schema
   fix recorded with rationale.

## Dependencies & constraints

- Depends on: M4 (F4.1–F4.6 built) and M4b (F4b.1 census + F4b.2 drop migration `071931c9` committed).
  This milestone unblocks the **M4 close gate**, not M4's features. Supersedes M4b's F4b.3 / F4b.4.
- Routed through **`metaldocs-database`** (any reference-data / curated-baseline / schema reasoning) and
  **`metaldocs-backend-api`** (Go test code). `superpowers:test-driven-development` for the factory
  self-test; `superpowers:systematic-debugging` (Iron Law: root cause before any fix) for the fillin
  investigation and any failure.
- Constraints: **fix-not-adapt** — `tests/integration/testdb/db.go` stays at HEAD (empty diff is an
  acceptance gate). **No production runtime/route/contract change.** The tripwire
  (`enforce_capability_asserted` / `trg_require_cap_asserted`) must **not** be weakened, disabled, or
  its CASE map edited — factories assert the real cap. **H-PRE-1**: never run an
  authz/`iam_users`-recording read inside a lock-holding atomic tx (advisory-lock deadlock). Reads stay
  live. Run under both `METALDOCS_DATABASE_URL` and `DATABASE_URL`; PowerShell on Windows; heavy-write
  template-DB churn may run on D: if C: SSD is slow.
- Blocks: the **M4 close gate** (re-dispatch the M4 `milestone-validator` after 4c PASS + HS-1) and,
  transitively, **M5** (the terminal re-audit).

## Applicable hard-stops

- **HS-1** — 4c close gate: operator review; no M4-validator re-dispatch and no merge without approval.
  Also at this milestone's open gate (operator agreement on this spec before any feature).
- **HS-2** — if the fillin investigation (or anything) turns out to need a **shared-schema /
  `search_path` redesign beyond the test layer**, or any fix requires a production-source schema/auth
  redesign outside the test boundary → **stop**, report the architecture boundary + minimum
  prerequisite plan; do **not** symptom-patch.
- **HS-3** — if a prerequisite boundary (clean baseline bootstrap, operator DSN auth, the template-clone
  path) fails → repair it (`runtime-contract-prereq` / `metaldocs-database`), rerun the failed
  checkpoint, then resume.
- **HS-4** — 4c `milestone-validator` returns FAIL → open the named fix feature, re-run its lifecycle,
  re-dispatch the validator.
- **HS-6** — any scope drift beyond test infrastructure + its docs/ADR/CI guard (e.g. re-introducing
  inline seeding to make a test pass, editing `db.go`, or quietly changing production source) → stop,
  surface, replan before continuing.
