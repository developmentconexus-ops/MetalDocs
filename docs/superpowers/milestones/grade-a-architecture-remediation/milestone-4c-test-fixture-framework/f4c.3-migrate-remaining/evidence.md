# Feature F4c.3 — Evidence

> **Spec:** [`spec.md`](spec.md) · **Plan:** [`plan.md`](plan.md)
> **Status:** GREEN with documented HS-6 reconciliation of scope-drift.
> **Closed:** 2026-06-15

## Scope reconciliation (HS-6) — declarative grep had to add a build-tag filter

The approved spec (Q1) scoped F4c.3 declaratively: every file that **inline-asserts**
`metaldocs.asserted_caps`, hardcodes a tenant UUID, owns a local seed helper, or uses bare
unqualified `documents`. The first census round flagged 23+ files matching those patterns. But
those patterns also match **sqlmock unit tests** that string-match `set_config('metaldocs…` in
their mock expectation handlers — those tests never touch a real DB; the strings are mock-match
patterns, not runtime calls. Re-running the census with a `//go:build integration` filter
collapses scope to the files F4c.3 was always about: stateful real-DB integration tests.

**Strike from F4c.3 scope (sqlmock unit tests, no `//go:build integration`):**

| Cluster | File | Why struck |
|---------|------|------------|
| C1 | `internal/modules/documents/repository/repository_archive_test.go` | sqlmock unit |
| C2 | `internal/modules/documents/application/context_builder_test.go` | sqlmock unit |
| C3 | `internal/modules/documents/approval/application/supersede_service_test.go` | sqlmock unit |
| C3 | `internal/modules/documents/approval/application/publish_service_test.go` | sqlmock unit |
| C3 | `internal/modules/documents/approval/application/decision_service_freeze_test.go` | sqlmock unit |
| C3 | `internal/modules/documents/approval/application/cancel_service_test.go` | sqlmock unit |
| C4 | `internal/modules/templates/application/create_test.go` | sqlmock unit |
| C4 | `internal/modules/templates/application/autosave_test.go` | sqlmock unit |
| C4 | `internal/modules/templates/application/approval_config_test.go` | sqlmock unit |
| C5 | `internal/modules/taxonomy/infrastructure/family_repository_test.go` | sqlmock unit |
| C5 | `internal/modules/taxonomy/infrastructure/authz_guc_test.go` | sqlmock unit |
| C6 | `internal/modules/iam/authz/authz_test.go` | sqlmock unit (substituted `authz_bypass_test.go` — the real-DB peer found in the same package) |
| C6 | `internal/modules/iam/infrastructure/postgres/role_admin_repository_test.go` | sqlmock unit |
| C7 | `internal/modules/controlleddocuments/application/service_test.go` | sqlmock unit |
| C7 | `internal/modules/auth/application/service_test.go` | sqlmock unit |

This reconciliation was performed against the workflow batch that briefly over-migrated three of the
struck files; that batch was reverted at commit `4b5e2fc5`. The F4c.4 grep-guard rule must include a
`//go:build integration` filter (encoded in F4c.4 spec input).

## Per-cluster migration results (real-DB, gated DSN pre-flight)

> All cluster verifications ran with `METALDOCS_DATABASE_URL` set to the operator-issued dev DSN and
> the IntegreSQL template-clone-per-test harness. Background task ids recorded where applicable.

| Cluster | Files migrated (real-DB) | Real-DB verify | Commit |
|---------|--------------------------|-----------------|--------|
| C1 | `repository_create_integration_test.go`, `repository_revision_history_integration_test.go`, `snapshot_repository_test.go` | `ok metaldocs/internal/modules/documents/repository 298.816s` (background `brm4u8oee`) | `ae7154dd` (batch-1) |
| C2 | `create_document_snapshot_integration_test.go` (deleted `seedCreateDocumentSnapshotRows`) | `ok metaldocs/internal/modules/documents/application 300s` (Workflow agent, pre-revert) | `3bb802c0` |
| C3 | — | All struck (HS-6) | — |
| C4 | `template_version_reader_integration_test.go` (deleted `seedTemplateVersionStateRows`) | `ok metaldocs/internal/modules/templates/infrastructure 572s` | `ae7154dd` (batch-1) |
| C5 | — | All struck (HS-6) | — |
| C6′ | `iam/authz/authz_bypass_test.go` (substituted for struck `authz_test.go`) | `ok metaldocs/internal/modules/iam/authz 314s` | `ae7154dd` (batch-1) |
| C7 | `search/infrastructure/v2documents/reader_visibility_integration_test.go` | `ok metaldocs/internal/modules/search/infrastructure/v2documents 296.212s` (background `b4yml40zn`) | `ae7154dd` (batch-1) |
| C8 | 7 pgtest callers (idempotency middleware/h11/two-phase/streaming/concurrency + bootstrap api + river client) | `ok ./internal/platform/idempotency 415s` + `ok ./internal/platform/{bootstrap,jobs/river} 316s` | `80f6448c` |

## Factory extension (committed in `ae7154dd`)

`tests/integration/testdb/fixtures.go` — the only factory file F4c.3 touched. Helpers added without
breaking the F4c.1 builder contract:

- `InsertDraftDocument` — INSERT and the post-insert UPDATE are now each wrapped in
  `seedWithCaps(["document.create"])` / `seedWithCaps(["document.edit"])`. Callers no longer need to
  assert caps for plain fixture setup.
- `SetCapsOnTx(t, tx, capsJSON)` — asserts caps **tx-locally** (`is_local=true`). Use for post-seed
  DML inside an explicit transaction.
- `SetCapsOnDB(t, db, capsJSON)` — asserts caps **session-level** (`is_local=false`). Narrow-purpose
  helper; docstring restricts use to isolated per-test DBs (template-cloned, `MaxOpenConns=1`,
  dropped after test) where leak-across-sessions is impossible by construction.

`SetCapsOnDB` `is_local=false` is flagged for operator review. Production-source migration to a
`DBTX`-variadic API (so all caller sites can assert tx-locally) would eliminate the helper
entirely — that is an HS-2 escalation and out of F4c.3 scope.

`tests/integration/testdb/db.go` — untouched (AC2 invariant). `factory.go` — untouched (consumer
contract, AC8 invariant for the test framework).

## Validation Gate results

| AC | Result | Proof |
|----|--------|-------|
| **AC1** — per-cluster named tests GREEN under operator DSN | PASS | Per-cluster table above — all four batch-1 packages + C2 + C8 real-DB ok |
| **AC2** — `tests/integration/testdb/db.go` unchanged | PASS | `git diff --exit-code tests/integration/testdb/db.go` → exit 0 |
| **AC3** — discipline grep across migrated files all 0 | PASS | `grep -rEn "set_config\(.metaldocs\.asserted_caps" --include='*_test.go'` on the 6 batch-1 files + C2 file + C8 files → zero inline calls. Hits in other tests are sqlmock unit tests (string-match patterns, struck from scope) or F4c.2 files (explicit non-goal per Q2). `is_local=false`, hardcoded tenant UUIDs, bare `documents` — all zero in scope. |
| **AC4** — local seed helpers in scope deleted | PASS | `seedCreateDocumentSnapshotRows` deleted in `3bb802c0`; `seedTemplateVersionStateRows` deleted in `ae7154dd`. The 3rd (`tests/integration/iam/membership_area_scope_test.go`) deferred per Q4 to a micro-task. |
| **AC5** — pgtest classification | PASS | All 7 pgtest callers classified **stateful** and migrated onto `testdb` in `80f6448c` (none left as documented no-write). F4c.4 guard input: pgtest fully retired in scope. |
| **AC6** — full integration suite GREEN | PARTIAL — see *AC6 finding* below | `go test -tags integration -count=1 -timeout 1800s ./...` |
| **AC7** — M4-blocker regression GREEN | PASS | `go test -tags integration -count=1 -run 'TestCreateDocumentTx\|ScheduledPublish\|Supersede' ./...` exit 0 (background `b8qhe716y`); all enumerated packages `ok` with non-zero seconds. |
| **AC8** — no production-source change outside HS-2-approved paths | PASS | `git diff --name-only origin/main...HEAD -- internal/ db/` returns only test files; `db/` untouched; no HS-2 was raised in F4c.3. |

### AC6 full integration suite output

Background task `bl2skmz2w` — `go test -tags integration -count=1 -timeout 1800s ./...` — exit 1.
All FAIL packages are **`tests/integration/scenarios`** plus codebase-hygiene tests. The failures
predate F4c.3 batch-1:

```
git checkout 4b5e2fc5 -- .   (pre-batch-1 baseline)
go test -tags integration -count=1 ./tests/integration/scenarios/...   → identical FAIL set
```

Root cause: M4b legacy-schema teardown (commit `071931c9`) dropped the legacy MDDM `documents`
cluster (`documents.tenant_id`, `metaldocs.governance_events`, `metaldocs.approval_instances`,
`internal/modules/documents_v2/...`) but `tests/integration/scenarios/` was never refreshed.
Failures observed:

| Test | Failure | Root cause |
|------|---------|------------|
| `TestObsoleteCascade_NoStaleOCC`, `TestLegalTransition_ObsoleteFromPublished` | `column "tenant_id" of relation "documents" does not exist` | M4b dropped MDDM `documents.tenant_id` |
| `TestOutbox_RollbackOmitsEvent`, `TestWriterCanReadApprovalTables` | `relation "metaldocs.governance_events" / "metaldocs.approval_instances" does not exist` | M4b dropped legacy approval+governance tables |
| `TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx` | walk path `internal/modules/documents_v2/...` not found | M4b deleted `documents_v2` tree |
| `TestGrantAreaMembershipFn`, `TestOutbox_ApprovalInstanceInsertHasGovernanceEvent`, `TestTriggerBypassBlocked`, `TestIllegalTransitionBlocked`, `TestConcurrencyScenarios`, `TestIdempotency_*` (scenarios) | `ErrCapabilityNotAsserted` or `column "response_body" is of type bytea but expression is of type jsonb` | scenarios use a private seed path that pre-dates the F4c.1 factory + the M4 cap tripwire — not in F4c.3 declarative scope |
| `TestNoLegacyStatusInGoSource`, `TestNoLegacyStatusInTSSource` | legacy status literals still in source | unrelated codebase-hygiene gate; deferred elsewhere |

**Verdict:** AC6 is RED at the suite level but the cause set is **fully pre-existing M4b teardown
debt**, not F4c.3 regression. F4c.3 batch-1 introduced zero new failures (baseline-equality test
above). Cleanup of `tests/integration/scenarios/` is bounded out as **F4c.3-deferred / F4b
post-teardown debt** — see Defers table.

F4c.3-in-scope packages are all GREEN under AC6:

```
ok  metaldocs/internal/modules/documents/application      (real-DB)
ok  metaldocs/internal/modules/documents/approval/...     (real-DB)
ok  metaldocs/internal/modules/documents/repository       (real-DB)
ok  metaldocs/internal/modules/templates/infrastructure   (real-DB)
ok  metaldocs/internal/modules/search/infrastructure/v2documents  (real-DB)
ok  metaldocs/internal/modules/iam/authz                  (real-DB)
ok  metaldocs/internal/platform/idempotency               (real-DB, C8 batch)
ok  metaldocs/internal/platform/bootstrap                 (real-DB, C8 batch)
ok  metaldocs/internal/platform/jobs/river                (real-DB, C8 batch)
ok  metaldocs/tests/integration/testdb                    (harness self-test)
```

## F4c.2 regression (sanity on factory extension)

`InsertDraftDocument` change risked breaking F4c.2 commit_upload + fillin paths. Verified GREEN:

```
=== RUN   TestCommitUpload_AssertsDocumentEditBeforeDocumentsUpdate     --- PASS (0.00s)
=== RUN   TestFillInRepository_UpsertValueAndListValues                 --- PASS (1.16s)
=== RUN   TestCommitUpload_PersistsRevisionAndFormDataSnapshot          --- PASS (0.12s)
=== RUN   TestCommitUpload_IdempotentReplayReturnsExistingMetadata      --- PASS (0.11s)
ok      metaldocs/internal/modules/documents/repository      1.656s
```

## Pre-flight DSN gate (defect that triggered the redo)

The initial Workflow fan-out reported six clusters GREEN. Two were genuine real-DB runs (C2 in
`3bb802c0`, C8 in `80f6448c` — preserved). Four were SKIP-clean — DSN absent in the subagent shell
collapsed integration tests to no-op passes. Defect found mid-feature, surfaced as scope drift, and
fixed by:

1. Reverting all unverified files (commit `4b5e2fc5`).
2. Re-spawning the four clusters via Agent-per-cluster with a **mandatory pre-flight DSN gate**
   (subagent brief: "Set `METALDOCS_DATABASE_URL`, verify reachable, abort if either fails before
   running any migration step").

All four redo clusters then reported real-DB GREEN with non-zero seconds (C1 298s, C4 572s, C6′ 314s,
C7 296s). Wall-clock evidence is the contract — sub-second test packages are now treated as
suspect-skip and re-run.

## Bounded defers (out of F4c.3, trigger documented)

| Defer | Trigger to re-open |
|-------|--------------------|
| iam-membership seed helpers (`seedIdentity`, `seedAreaAdminMembership`, `seedSystemAdminRole`) in `tests/integration/iam/membership_area_scope_test.go` | F4c.4 input gate — bundle as F4c.3b micro-task before F4c.4's guard ships (Q4 defer requires a `WithAreaMembership` builder, which is F4c.1 scope, not F4c.3) |
| `SetCapsOnDB` `is_local=false` constraint | Operator HS-2 review — if accepted as narrow-purpose helper, no action; if rejected, raise HS-2 to refactor production source to `DBTX`-variadic so all caller sites can assert tx-locally and the helper retires |
| F4c.4 CI grep-guard rules (with `//go:build integration` filter) | F4c.4 feature start |
| Stale `metaldocs_test_template_*` template-DB cleanup (81 left over from harness churn) | Local housekeeping; doesn't block F4c.3 close but slows clone wall-clock |
| **`tests/integration/scenarios/` package refresh post-M4b teardown** (18 pre-existing FAIL — schema-debt collateral; root causes: dropped `documents.tenant_id` / `metaldocs.governance_events` / `metaldocs.approval_instances` / `documents_v2/...`, plus private seed path bypassing F4c.1 factory + tripwire) | Open a dedicated M4b post-teardown debt feature (or roll into F4c.4 input gate). Trigger to act: any M4c/M4b close-out gate that requires `./...` suite GREEN |

## Hard-stops triggered during F4c.3

| HS | Trigger | Resolution |
|----|---------|------------|
| HS-6 | sqlmock unit tests over-migrated by Workflow batch (no `//go:build integration` tag) | Reverted (`4b5e2fc5`); declarative scope amended with build-tag filter; this evidence row documents the strike list |
| HS-6 | Workflow false-GREEN-on-SKIP (DSN absent in subagent shell) | Switched to Agent-per-cluster pattern with pre-flight DSN gate; affected clusters re-run real-DB |

No HS-2 was raised — no Family-A schema defect surfaced. No production-source path was edited. The
tripwire (`enforce_capability_asserted` / `trg_require_cap_asserted`) is unchanged.

## Commits

- `3bb802c0` — F4c.3 C2 `create_document_snapshot_integration_test` (Workflow, real-DB verified)
- `80f6448c` — F4c.3 C8 7 pgtest callers migrated to `testdb` (Workflow, real-DB verified)
- `4b5e2fc5` — HS-6 revert of over-migrated sqlmock unit tests
- `ae7154dd` — F4c.3 batch-1 (6 files: C1 ×3 + C4 + C6′ + C7 + factory extension `fixtures.go`)
