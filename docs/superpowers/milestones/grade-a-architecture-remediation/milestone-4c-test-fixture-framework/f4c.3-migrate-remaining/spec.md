# Feature F4c.3 — Spec

> **Milestone:** 4c — Unified test-fixture framework  ·  **Folder:** `f4c.3-migrate-remaining`
> **Status:** Approved (pre-code) — interview answers locked, fan-out authorized.
> **Approved before code:** 2026-06-15 — operator leandrotca.work ("Agreed", post-explanation).

> The feature's contract, written **before** any code. The milestone-validator judges F4c.3 against
> this file (C1). The "how" lives in `plan.md`; close-out proof lives in `evidence.md`.

## Interview record (fail-closed gate)

Pre-spec ambiguity resolved by census against the F4c.1 factory + F4c.2 commit (`0bc7ef13`). Open
questions for the operator are listed below — answered before approval, not after.

| # | Question | Answer |
|---|----------|--------|
| 1 | Scope of "remaining stateful tests" — strictly the files that still inline-`set_config` / hardcode a tenant UUID / own a local seed helper / use bare unqualified `documents`, **excluding** the 4 F4c.2 files? Or also any pgtest-stateful that grep missed? | **Yes — declarative grep rules** = scope. Self-checking against the F4c.4 guard. |
| 2 | F4c.2 left `SELECT set_config('metaldocs.asserted_caps', …, false)` in `commit_upload` (×2) and `fillin` for raw-guarded-write paths. Treat as intentional (F4c.2-shipped `SeedWithCaps` consumer pattern, **out of F4c.3 scope**) — confirm? | **Yes — out of scope.** `SeedWithCaps`-wrapped calls are sanctioned framework pattern. F4c.4 guard rule must allow them. |
| 3 | `pgtest.OpenAndMigrate` callers (7 platform/job files): scope of F4c.3 is to **classify** each (no-write = leave; stateful-write = migrate to `testdb`) but not necessarily migrate every one. Acceptable? | **Yes — classify all 7, migrate only stateful.** Classification is the framework policy F4c.4 will encode. |
| 4 | Milestone.md says "3 local seed helpers" but the census names 2 (`seedCreateDocumentSnapshotRows`, `seedTemplateVersionStateRows`). Treat `tests/integration/iam/membership_area_scope_test.go` (`seedIdentity` / `seedAreaAdminMembership` / `seedSystemAdminRole`) as the 3rd cluster — or scope only to documents/templates and leave iam membership seeds to a later micro-task? | **Defer.** iam membership needs a factory extension (`WithAreaMembership`) — violates F4c.3's "no factory API change" non-goal. Handle in a micro-task before F4c.4 (or as F4c.4 prep). |
| 5 | Subagent execution model: fan-out by cluster (≤8 concurrent), sonnet for clusters w/ judgment risk (pgtest classify, schema-shadow risk) and haiku for purely mechanical regex-shaped rewrites. Acceptable? | **Yes.** sonnet → C2/C4/C8 (judgment); haiku → C1/C3/C5/C6/C7 (mechanical). |
| 6 | If a cluster trips an HS-2 (Family-A schema defect surfacing on migration), that cluster stops + reports; remaining clusters continue. Confirm — or hard-stop the whole feature on any HS-2? | **Per-cluster stop.** Matches F4c.2 fillin precedent. Other clusters keep landing evidence; HS-2 surfaces to operator in parallel. |

## Consumer contract (FIRST — before any producer)

The **producer** here is **the migrated test files themselves**. The **consumers** are:

- **`tests/integration/testdb/factory.go`** (shipped F4c.1) — its `NewTenant` / `NewUser` /
  `NewTaxonomy` / `NewControlledDoc` / `NewDocument` (`WithStatus` / `WithRevisionVersion` /
  `WithScheduleGen`) builders + `Scenario.PublishedDocument` / `Scenario.ScheduledRevision`
  composites + the exported `SeedWithCaps` helper. The API is **stable** — F4c.3 must **consume**
  it, never extend it (any new builder is a return to F4c.1 spec, not F4c.3 scope).
- **The F4c.4 CI grep-guard rules (declared in this milestone's `milestone.md`)** — F4c.3's
  migrated tree must already satisfy them, so the F4c.4 guard can be added on a clean tree.
  Specifically per migrated file: no inline `SELECT set_config('metaldocs.asserted_caps', …)`, no
  `is_local=false`, no hardcoded tenant-UUID literal, no bare unqualified `documents` (use
  `testdb.Qualified`).
- **`tests/integration/testdb/db.go`** (HEAD) — fix-not-adapt invariant: F4c.3 must not edit it
  (empty diff is an acceptance gate). The harness is fixed.
- **Production source (`internal/...`, `db/migrations/...`)** — F4c.3 must not edit it except via an
  **HS-2-escalated, operator-approved** Family-A schema fix, recorded with rationale and an ADR.

**Source of truth for the API:** `tests/integration/testdb/factory.go` (read from there; do not invent
new builders).

## What this feature implements

Migrate the remaining stateful integration-test files onto the F4c.1 `testdb` factory so the suite
satisfies the F4c.4 discipline rules from a clean baseline. Scope is **declarative, not
extensional** — every file matching the rules below is in. Authoritative target list, partitioned
into independent clusters (one subagent per cluster):

### Cluster C1 — documents/repository (non-F4c.2, real-DB)
- `internal/modules/documents/repository/repository_revision_history_integration_test.go`
- `internal/modules/documents/repository/repository_create_integration_test.go`
- `internal/modules/documents/repository/snapshot_repository_test.go`
- `internal/modules/documents/repository/repository_archive_test.go`

### Cluster C2 — documents/application + the snapshot local helper
- `internal/modules/documents/application/create_document_snapshot_integration_test.go`
  (delete `seedCreateDocumentSnapshotRows`)
- `internal/modules/documents/application/context_builder_test.go`

### Cluster C3 — documents/approval/application (4 files)
- `internal/modules/documents/approval/application/supersede_service_test.go`
- `internal/modules/documents/approval/application/publish_service_test.go`
- `internal/modules/documents/approval/application/decision_service_freeze_test.go`
- `internal/modules/documents/approval/application/cancel_service_test.go`

### Cluster C4 — templates (incl. the second local helper)
- `internal/modules/templates/infrastructure/template_version_reader_integration_test.go`
  (delete `seedTemplateVersionStateRows`)
- `internal/modules/templates/application/create_test.go`
- `internal/modules/templates/application/autosave_test.go`
- `internal/modules/templates/application/approval_config_test.go`

### Cluster C5 — taxonomy
- `internal/modules/taxonomy/infrastructure/family_repository_test.go`
- `internal/modules/taxonomy/infrastructure/authz_guc_test.go`

### Cluster C6 — iam authz + role admin
- `internal/modules/iam/authz/authz_test.go`
- `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go`

### Cluster C7 — adjacent application/service tests (controlleddocuments + auth + search)
- `internal/modules/controlleddocuments/application/service_test.go`
- `internal/modules/auth/application/service_test.go`
- `internal/modules/search/infrastructure/v2documents/reader_visibility_integration_test.go`

### Cluster C8 — pgtest classification + targeted migration
For each of the 7 `pgtest.OpenAndMigrate` callers
(`internal/platform/idempotency/middleware_test.go`, `h11_schema_test.go`,
`two_phase_test.go`, `middleware_streaming_test.go`, `middleware_concurrency_test.go`,
`internal/platform/bootstrap/api_test.go`, `internal/platform/jobs/river/client_test.go`):
- Determine empirically whether the file performs **stateful writes** that the F4c.4 guard would
  flag (write-tx, non-read SQL touching the tenant tables, hardcoded tenant UUID).
- If stateful → migrate onto `testdb` (template-DB-per-test).
- If genuinely no-write → document the file as a no-write `pgtest` caller in this feature's
  `evidence.md` (F4c.4's guard will then permit `pgtest` only for the documented set).

> **Out of F4c.3 — explicit non-goal:** `tests/integration/iam/membership_area_scope_test.go` and
> any other iam membership stateful test. Q4 above defers the iam-membership seed helpers to a
> later micro-task **unless** the operator answers Q4 = "include them"; in that case Cluster C9
> (`tests/integration/iam/membership_area_scope_test.go`, deleting `seedIdentity` /
> `seedAreaAdminMembership` / `seedSystemAdminRole`) is added before approval.

## Non-goals (mandatory)

- **No factory API change.** No new builders, no signature edits. If a migration needs something the
  factory lacks, stop the cluster + surface — that's a return to F4c.1 spec, not F4c.3 scope.
- **No edit to `tests/integration/testdb/db.go`.** Empty diff is an acceptance gate.
- **No production-source change** (`internal/...`, `db/migrations/...`, `db/baseline/...`,
  `wiki/decisions/...`) — except an **HS-2-escalated, operator-approved** Family-A schema fix
  recorded with rationale + ADR (matches F4c.2 / fillin precedent).
- **No tripwire weakening.** The real capability cap is asserted on every write; never weakened,
  never disabled, never edited in the CASE map.
- **No re-introduction of inline seeding** to make a test pass. Inline `set_config` /
  `is_local=false` / hardcoded tenant UUID / bare unqualified `documents` re-appearing is a FAIL.
- **No F4c.2 file edit** (commit_upload, fillin, postgres_approval_repository,
  scheduled_publish_job) — their remaining `SetWithCaps` patterns are intentional (Q2). If a cluster
  needs to touch them → stop + surface (HS-6 scope drift).
- **No CI grep-guard.** That is F4c.4.
- **No docs/ADR.** That is F4c.5 (unless an HS-2 Family-A fix requires an inline ADR).
- **No `pgtest` retirement.** F4c.4 owns the policy decision; F4c.3 only classifies + migrates the
  stateful ones.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| **AC1** — Each migrated file's named integration tests pass from a clean baseline under the operator DSN. | Per cluster: `go test -tags integration -count=1 ./<cluster-path>/...` — captured output in `evidence.md` | real |
| **AC2** — `tests/integration/testdb/db.go` is unchanged from HEAD (fix-not-adapt). | `git diff --exit-code tests/integration/testdb/db.go` (exit 0) | real |
| **AC3** — Discipline rules satisfied on every migrated file (the F4c.4 guard would pass): no inline `set_config('metaldocs.asserted_caps', …)`, no `is_local=false`, no hardcoded tenant-UUID literal, no bare unqualified `documents` (use `testdb.Qualified` or factory-returned IDs). | Per cluster, in `evidence.md`: `git grep -nE "set_config\('metaldocs\.asserted_caps'" -- <migrated files>` → empty; same for `is_local\s*=\s*false`, the UUID literal pattern, and `\bbare-documents\b`. | real |
| **AC4** — Local seed helpers in scope deleted: `seedCreateDocumentSnapshotRows`, `seedTemplateVersionStateRows` (and any third confirmed via Q4). | `git grep -nE '^func seed(CreateDocumentSnapshotRows\|TemplateVersionStateRows)\b' -- '*_test.go'` → empty | real |
| **AC5** — pgtest classification recorded: each of the 7 callers labeled stateful-migrated or no-write-documented. | Table in `evidence.md`; for any stateful-migrated file the migrated test run is GREEN under the operator DSN. | real |
| **AC6** — Full integration suite green from a clean baseline under the operator DSN. | `go test -tags integration -count=1 ./...` — captured output in `evidence.md` | real |
| **AC7** — Regression: M4-blocker tests (F4.1a Gate #5 + F4c.2 surface) still GREEN. | `go test -tags integration -count=1 -run 'TestCreateDocumentTx_PopulatesAllSnapshotColumns|TestScheduledPublishWorker_|TestValidateScheduledSupersedeTarget_RealRows|TestLoadCurrentPublishedHeadForDocument_RealRows|TestLoadActiveInstanceByDocument_LoadsDocumentRevisionVersion|TestLoadInstance_LoadsDocumentRevisionVersion|TestScheduleGenerationIncrementsOnScheduledWritePath' ./...` GREEN | real |
| **AC8** — No production-source change unless HS-2-escalated. | `git diff --name-only origin/main...HEAD -- internal/ db/` returns only authorized HS-2 paths (or empty) | real |

> TDD discipline: per cluster, run the cluster's named tests against the unmigrated tree first to
> capture the failing/inline-`set_config` baseline (commit-log evidence — not a code change); then
> migrate; then re-run GREEN. Fixture-only proof is not acceptable for any cluster.

## ADR needed?

- [x] No durable decision in F4c.3 itself — the framework decision (ADR) is F4c.5; trigger fix (if
  any HS-2 surfaces) gets its own ADR under that HS-2's escalation.

## HS-6 reconciliation (post-approval — declarative scope amended with build-tag filter)

The approved declarative grep rules (Q1) over-counted: they matched sqlmock unit tests that
**string-match** `set_config('metaldocs.asserted_caps`, ...)` in their mock expectation handlers.
Those strings are mock-match patterns, not runtime DB calls — the tests never touch a real DB.
First Workflow batch briefly over-migrated 3 of these unit-test files; reverted at `4b5e2fc5`.
Scope rules amended:

> A file is in F4c.3 scope only if its first line is `//go:build integration` **and** it inline-asserts
> `metaldocs.asserted_caps` / hardcodes a tenant UUID / owns a local seed helper / uses bare
> unqualified `documents` at runtime (not in a mock-string literal).

Net effect on the cluster lists in this spec — 15 listed files struck (per [`evidence.md`](evidence.md)
strike table). This filter is also the F4c.4 grep-guard input.

## Post-close finding (does NOT block F4c.3 close)

`go test -tags integration ./...` (AC6) is RED at suite level, **fully from pre-existing M4b
teardown debt** in `tests/integration/scenarios/` (dropped `documents.tenant_id`, `governance_events`,
`approval_instances`, `documents_v2/...`). Baseline-equality verified on `4b5e2fc5` (pre-F4c.3-batch-1).
F4c.3 introduced zero new failures. Cleanup of `tests/integration/scenarios/` is bounded out as
**M4b post-teardown debt** — see [`evidence.md`](evidence.md) Defers table.
