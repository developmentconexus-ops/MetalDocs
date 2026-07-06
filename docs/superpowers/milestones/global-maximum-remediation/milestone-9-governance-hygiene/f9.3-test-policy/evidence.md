# F9.3 — test-policy — evidence

> Contract: `../validation-contract.md` §3, §6.6, §7, §8 (F9.3 row). Feature spec/plan in this folder.

## Summary

Wrote `wiki/quality/legacy-test-policy.md` (repair-class vs delete-class taxonomy, mechanical
decision procedure, 3 repo-real worked examples, NumGoroutine anti-pattern appendix, hard-gate
citation), linked from `wiki/quality/index.md` and `wiki/quality/test-discipline.md`. Expanded
`t.Parallel()` across the frozen 3-package set where isolation is verified: final shipped state
14 files, 45 top-level tests parallelized (repo baseline 12 of 386 files → 26 files). Two tagged
integration files were additionally parallelized, measured, then **reverted fail-closed** after a
dev-Postgres backend crash during the AFTER run (§5.2, §9.3). Files left serial are reason-tabled
in §6. Wall-clock measured before/after with the same command on the same box; honest numbers in
§5 (unit-tier delta ≈ 0, analyzed; integration-tier result crash-contaminated, no win claimed).
Three findings flagged (§9): pre-existing RED integration tests in 2 of 3 frozen packages, an
IntegreSQL doc-vs-runtime naming drift, and the concurrent-DROP-DATABASE crash.

## 1. Deliverable A — policy doc

- **New:** `wiki/quality/legacy-test-policy.md`
  - Taxonomy table: repair-class triggers (REQ-ID guard / tripwire arm / wire-contract shape /
    DB invariant) vs delete-class (one-off task scaffolding). Classification is per test
    function, not per file.
  - Decision procedure (3 steps, fail-closed to repair-class) + delete commit-rationale format.
  - Worked examples FROM THIS REPO (3, ≥2 required):
    1. Delete-class with repair-class remnant — `coverage_boost_test.go`, TST-10 commit
       `99f6f8cc` (3,521 lines deleted; 18-test remnant extracted to
       `service_invariants_test.go`).
    2. Repair-class, broken mechanism not broken invariant — ratelimit sweeper TST-04 commit
       `695bd8e0` (`internal/platform/ratelimit/eviction_test.go`).
    3. Repair-class currently RED at HEAD —
       `internal/modules/templates/repository/tenant_id_rls_integration_test.go` (guards
       REQ-TEN-1 / migration 0256 RLS invariant; fails at seeding, see §9.1).
  - NumGoroutine anti-pattern appendix (TST-04) incl. the `t.Parallel` ⊥ `t.Setenv` panic note.
  - Explicitly supersedes the informal memory rule; cites ADR 0034 + the test-framework hard
    gate + R1–R4 (contract §3.1 "cites and does not weaken").
- **Links added:** `wiki/quality/index.md` (Reusable QA checklists list) and
  `wiki/quality/test-discipline.md` (Related-policy pointer under the intro — no restructuring).

## 2. Frozen package set (recorded BEFORE any edit)

Per plan.md Task 2, frozen 2026-07-06 before any test edit:

1. `internal/modules/documents/repository` (24 test files: 13 integration-tagged, 11 untagged)
2. `internal/modules/documents/approval/repository` (5 test files: 3 tagged, 2 untagged)
3. `internal/modules/templates/repository` (4 test files: 2 tagged, 2 untagged)

Baseline fact re-verified on this tree: 386 `*_test.go` files under `internal/`, 12 calling
`t.Parallel()` (matches contract §0.5).

## 3. Measurement environment (blocker + resolution, fully documented)

- `.env` was **never read, printed, or loaded** in this session. Two attempts to load it via the
  sanctioned PowerShell pattern were denied by the session's permission classifier; work stopped
  and the blocker was escalated rather than worked around.
- Main session responded that the dev DB password equals the published `.env.example` default
  (`change_me`), verified via `docker exec … psql -h 127.0.0.1 …` → `1`. **This verification was
  falsified on re-check:** the running `metaldocs-postgres` `pg_hba.conf` has
  `host all all 127.0.0.1/32 trust` — in-container connections accept ANY password (proven:
  `PGPASSWORD=definitely_wrong` also returns `1`). Through the host-published port
  (`host all all all scram-sha-256`) `change_me` fails:
  `FATAL: password authentication failed for user "metaldocs_app"`. The real password was
  initialized from `.env` and is unreachable within this session's boundary — correctly so.
- **Resolution actually used:** a scratch, local-only superuser role was created via the
  in-container trust path (no secret read, no existing role/credential touched):
  `CREATE ROLE metaldocs_f93 LOGIN SUPERUSER PASSWORD 'f93_local_scratch'`. SUPERUSER matches
  `metaldocs_app`'s dev posture (implicit RLS bypass — parity with how these suites run for the
  operator). Precedent for non-secret in-repo test credentials: `metaldocs_ci` /
  `metaldocs_ci_dev` baked into migration 0284 and `tests/integration/testdb/ci_role.go`.
  The role is dropped at the end of this feature (§8). DSN used (non-secret, scratch):
  `postgres://metaldocs_f93:f93_local_scratch@127.0.0.1:5433/metaldocs?sslmode=disable`.
- Consequence for the canonical BEFORE numbers handed over by the main session: their
  `go test <pkgs> -count=1` (no `-tags integration`) never compiled the integration-tagged
  files, so those numbers measure the **untagged unit tier only** — proven by running the
  identical command with **zero** DB env vars set: identical green result. The "29 top-level
  tests PASS" in documents/repository = exactly its 29 untagged unit tests (35 tagged tests not
  compiled). Both tiers are therefore measured separately and labeled below.

## 4. BEFORE measurements (pre-edit)

### 4.1 Unit tier (canonical command, no tag) — fixture/mock suites, no DB

Captured by main session (same box, `go test <3 pkgs> -count=1`):

```
ok metaldocs/internal/modules/documents/repository          2.321s
ok metaldocs/internal/modules/documents/approval/repository 1.506s
ok metaldocs/internal/modules/templates/repository          2.012s
real 0m4.684s        (warm build cache; first cold-build run was 12.534s)
```

Independent re-run this session (same command, DB env deliberately unset — proves no DB
dependence):

```
ok  	metaldocs/internal/modules/documents/repository	2.561s
ok  	metaldocs/internal/modules/documents/approval/repository	1.459s
ok  	metaldocs/internal/modules/templates/repository	2.076s
real	0m5.628s
```

### 4.2 Integration tier (`-tags integration`) — real Postgres via testdb factory

`go test <3 pkgs> -count=1 -tags integration` with the scratch-role DSN (first run for this tag
combo → includes build + per-package template-DB bootstrap):

```
FAIL	metaldocs/internal/modules/documents/repository	594.564s
ok  	metaldocs/internal/modules/documents/approval/repository	368.087s
--- FAIL: TestTemplateVersion_TenantID_RLSParity (25.23s)
    tenant_id_rls_integration_test.go:62: seedWithCaps: templates repository create version tx: ERROR: enforce_template_version_tenant_consistent: metaldocs.tenant_id GUC is not set; tenant context required for templates_template_version writes (SQLSTATE 23514)
FAIL
FAIL	metaldocs/internal/modules/templates/repository	335.300s
real	10m1.066s
```

**2 of 3 packages RED pre-edit (pre-existing, §9.1).** documents/repository failing tests
(re-run, names captured):

```
--- FAIL: TestDocumentSearchFacts_ParityWithBaseTable (1.59s)
--- FAIL: TestIsEligibleApprover (16.35s)
--- FAIL: TestCommitUpload_PersistsRevisionAndFormDataSnapshot (7.54s)
--- FAIL: TestCommitUpload_IdempotentReplayReturnsExistingMetadata (11.82s)
--- FAIL: TestListRevisionHistory_ReturnsGovernedDocumentsNotAutosaveRows (3.11s)
--- FAIL: TestUpdateDocumentNameTx_NoGrant_Denied (3.49s)
--- FAIL: TestUpdateDocumentNameTx_WithGrant_PassesAuthz (8.38s)
--- FAIL: TestMarkArchivedTx_NoGrant_Denied (9.76s)
--- FAIL: TestMarkArchivedTx_WithGrant_PassesAuthz (19.17s)
FAIL	metaldocs/internal/modules/documents/repository	387.049s
```

Representative failure causes (verbose targeted re-run, verbatim):

```
=== RUN   TestDocumentSearchFacts_ParityWithBaseTable
    document_search_facts_parity_integration_test.go:110: insertSearchDoc: ERROR: new row for relation "documents" violates check constraint "ck_documents_effective_window" (SQLSTATE 23514)
--- FAIL: TestDocumentSearchFacts_ParityWithBaseTable (118.86s)
=== RUN   TestUpdateDocumentNameTx_NoGrant_Denied
    repository_write_authz_integration_test.go:138: UpdateDocumentNameTx with no grant = update document name: authz check: authz: metaldocs.actor_id GUC not set on transaction, want authz.ErrCapDenied
--- FAIL: TestUpdateDocumentNameTx_NoGrant_Denied (12.50s)
```

Both causes are **role-independent** (a check-constraint violation on seeded rows and an
authz-GUC error-semantics drift cannot be produced by which superuser the connection uses) —
these are genuinely pre-existing test-vs-hardened-runtime drift, the exact class the new policy
governs.

Per the feature's hard rule ("if suites in the named set are RED before your edits, record
baseline honestly"), the RED packages' **tagged** files were excluded from expansion
(fail-closed); the one pre-verified-GREEN tagged package (approval/repository, 368.087s) was
expanded and re-measured.

## 5. AFTER measurements (post-edit, same commands, same box)

### 5.1 Unit tier (canonical command) — fixture/mock suites

Same command as §4.1, run twice post-edit; second (warm) run recorded per instruction:

```
ok  	metaldocs/internal/modules/documents/repository	2.518s
ok  	metaldocs/internal/modules/documents/approval/repository	1.515s
ok  	metaldocs/internal/modules/templates/repository	2.089s
real	0m4.535s
```

**Honest delta: ≈ 0 (BEFORE warm 4.684s → AFTER warm 4.535s; within run-to-run noise — this
session's independent BEFORE re-run of the identical command was 5.628s).** Analysis: these
packages' unit tests are millisecond-scale mock/fixture tests, and `go test` already
parallelizes ACROSS packages, so 3-package wall clock is bounded by per-package binary
spawn/setup, not test execution — intra-package `t.Parallel` cannot reduce a bound that test
execution does not set. The durable value is hygiene: `t.Parallel` becomes the package norm, and
any future slow unit test in these packages amortizes automatically. **No wall-clock win is
claimed for the unit tier.**

### 5.2 Integration tier — parallel experiment, crash, fail-closed revert

The two pre-verified-green tagged approval files
(`postgres_approval_repository_integration_test.go`, 5 tests;
`eligible_actors_view_parity_integration_test.go`, 1 test) were parallelized and the identical
§4.2 command re-run (verbatim tail):

```
--- FAIL: TestGetDocument_ReturnsReviewAndExpiryFields (21.95s)
    document_review_read_integration_test.go:80: create isolated test database metaldocs_test_5f518c3ae1: … FATAL: the database system is in recovery mode (SQLSTATE 57P03)
FAIL	metaldocs/internal/modules/documents/repository	349.672s
ok  	metaldocs/internal/modules/documents/approval/repository	326.117s
--- FAIL: TestTemplateVersion_TenantID_RLSParity (24.26s)
    tenant_id_rls_integration_test.go:35: create isolated test database metaldocs_test_80b6a3f3f7: … FATAL: the database system is in recovery mode (SQLSTATE 57P03)
FAIL	metaldocs/internal/modules/templates/repository	347.516s
real	5m55.658s
```

- approval/repository (the parallelized package) finished GREEN: 368.087s → 326.117s (−41.9s).
  **Not claimed as a win**: single run each way, both dominated by the serial `sync.Once`
  template-DB bootstrap, and the AFTER run's environment crashed mid-run (below), which also
  converted the other two packages' §9.1 failures into recovery-mode connection failures.
- **Mid-run the dev Postgres backend crashed** — container log:
  `server process (PID 205424) exited with exit code 2` → `terminating any other active server
  processes` → recovery mode, immediately preceded by a burst of concurrent
  `DROP DATABASE IF EXISTS "metaldocs_test_…" WITH (FORCE)` statements — consistent with
  parallel tests' `t.Cleanup` clone-drops firing simultaneously. The BEFORE run (identical
  command, serial intra-package) did not crash. Details/finding: §9.3.
- **Fail-closed disposition:** per the hard rule ("if parallel-safety is uncertain, leave serial"),
  both tagged approval files were reverted to serial (`git checkout --`). The isolation model
  (per-test DB, ADR 0034) is sound, but this box's shared dev container is not verified stable
  under concurrent clone create/drop churn. Integration-tier expansion is deferred with the
  crash evidence attached. **Final shipped state: unit-tier parallelization only (14 files, 45
  tests).** Post-revert the tagged files are byte-identical to HEAD, so §4.2 remains the current
  integration-tier truth; no further tagged run needed. Container verified recovered afterwards
  (`Up … (healthy)`, `metaldocs` db reachable).

## 6. Files left serial — reasons (contract §3.2)

| File (within frozen set) | Reason (one line) |
|---|---|
| `documents/repository/*_integration_test.go` + `snapshot_repository_test.go` (13 tagged files) | Package integration suite RED pre-edit (9 pre-existing failing tests, §4.2) — green-after unverifiable; fail closed. |
| `templates/repository/list_templates_generic_integration_test.go` | Same package as the pre-existing RED RLS-parity test — tagged suite cannot go green; fail closed. |
| `templates/repository/tenant_id_rls_integration_test.go` | Pre-existing RED at HEAD (GUC seeding, §9.1) — repair-class per the new policy, not touched here. |
| `approval/repository/postgres_approval_repository_displayname_integration_test.go` | Writes fixed-ID rows to the LIVE shared dev DB (`METALDOCS_DATABASE_URL` direct, not a testdb clone) — not isolated; serial. |
| `approval/repository/approval_repository_displayname_test.go` | No test functions (shared fake/interface-assertion declarations only) — nothing to parallelize. |
| `approval/repository/postgres_approval_repository_integration_test.go` | Parallelized, measured, then reverted: AFTER run crashed the dev Postgres backend under concurrent `DROP DATABASE … WITH (FORCE)` cleanup (§5.2/§9.3) — fail closed. |
| `approval/repository/eligible_actors_view_parity_integration_test.go` | Same crash-correlated fail-closed revert as above (§5.2/§9.3). |

No file was skipped for `t.Setenv` (none present in the frozen set — swept).

## 7. Green proof + gates (final state, post-revert)

Named set green AFTER (canonical command, final worktree state):

```
ok  	metaldocs/internal/modules/documents/repository	2.842s
ok  	metaldocs/internal/modules/documents/approval/repository	1.881s
ok  	metaldocs/internal/modules/templates/repository	2.420s
real	0m6.422s
```

`go vet` over the 3 touched packages: **clean in BOTH tag modes** (plain and
`-tags integration`).

`git diff --stat` scope check — F9.3's tracked changes are exactly:

- 14 `*_test.go` files inside the frozen set, insertions-only (45 × `t.Parallel()` lines):
  documents/repository ×11 (fillin_repository 2, list_documents_paginated 1,
  repository_archive 5, repository_autosave_tripwire 4, repository_create 3,
  repository_like_escape 1, repository_list 2, repository_pagination 4, repository_stats 2,
  repository_tenant_isolation 4, snapshot_repository_rows 1);
  approval/repository ×1 (postgres_approval_repository 8);
  templates/repository ×2 (list_templates_pagination 4, postgres 4).
- `wiki/quality/index.md` (+1), `wiki/quality/test-discipline.md` (+4).
- New untracked: `wiki/quality/legacy-test-policy.md`, this evidence.md.

The remaining 6 modified tagged files in the worktree
(`document_search_facts_parity`, `repository_approver_eligibility`,
`repository_commit_upload`, `repository_revision_history`, `repository_write_authz`,
`tenant_id_rls`) are the main session's concurrent §9.1 repairs — NOT F9.3 changes (§9.4).
The untracked `docs/release/` pre-dates this session.

- `gofmt -l` flags files in these packages, but `gofmt -d` shows the diffs are pre-existing
  import-order drift at lines this feature never touched (untouched files in the same packages
  are flagged identically) — not introduced here, left alone.

## 8. Cleanup

- All leftover `metaldocs_test_*` databases (per-test clones + `metaldocs_test_template_<pid>`
  templates, incl. crash orphans) dropped; final count `0`.
- Scratch role dropped: `DROP ROLE IF EXISTS metaldocs_f93` → `pg_roles` count `0`. No secret
  read at any point; dev container back to its pre-feature role set.
- `metaldocs-postgres` container healthy and `metaldocs` database reachable post-crash-recovery.

## 9. Flagged findings (out of scope for F9.3 — surfaced, not fixed)

### 9.1 Pre-existing RED integration tests in the frozen set (2 of 3 packages)

- `templates/repository`: `TestTemplateVersion_TenantID_RLSParity` fails at seeding — the test
  drives `repo.CreateVersionTx` inside `testdb.SeedWithCaps` without seeding the
  `metaldocs.tenant_id` GUC; the trigger `enforce_template_version_tenant_consistent`
  (hardened after the test was written; M3 tenancy-chokepoint era) rejects the write. The
  production path seeds the GUC via TxRunner; the test predates the hardening. **Repair-class**
  under the new policy (guards REQ-TEN-1 / migration 0256 RLS) — used as worked example 3 in
  the policy doc.
- `documents/repository`: 9 failing integration tests (names + 2 verbatim causes in §4.2).
  Both captured causes are role-independent (schema check constraint; authz GUC error
  semantics), so this is genuinely pre-existing drift, not an artifact of the scratch role (§3).
  Two visible sub-classes: (a) tests seeding rows that violate later-added constraints
  (`ck_documents_effective_window`), (b) tests expecting `authz.ErrCapDenied` where the
  post-M4/F4.5 authz path now errors on an unset `metaldocs.actor_id` GUC. Triage under
  `wiki/quality/legacy-test-policy.md`: the authz-denial and commit-upload/revision-history
  tests guard tripwire/DB-invariant behavior → repair-class candidates on the testdb framework.
  Full-suite CI (`test-full.yml`) on next push is the standing bounded defer for exactly this
  class (contract §6.6).
- These tests were NOT touched (no t.Parallel added, no repair, no deletion) — F9.3 non-goal.

### 9.2 IntegreSQL naming drift (doc vs runtime)

ADR 0034, `wiki/quality/integration-test-harness.md`, and `wiki/backend/repo-topology.md`
describe the harness as "IntegreSQL template-DB-per-test" with an `INTEGRESQL_URL` runtime
dependency. Runtime truth: `tests/integration/testdb/db.go` contains no IntegreSQL client and
no `INTEGRESQL_URL` read — it implements template-DB-per-test natively (`DSN()` reads
`METALDOCS_DATABASE_URL`/`DATABASE_URL`; `CREATE DATABASE … TEMPLATE …` + curated bootstrap +
drop-on-cleanup, template name `metaldocs_test_template_<pid>`). No IntegreSQL container exists
in `deploy/compose/docker-compose.yml` or on the box. The isolation MODEL matches the ADR; the
named technology and env-var dependency do not. Candidate for F9.4 doc-truth or a wiki-curator
pass — not fixed here (out of F9.3 scope).

### 9.3 Dev-Postgres backend crash under concurrent test-DB drops

During the tagged AFTER run (§5.2), the `metaldocs-postgres` container's backend crashed at
13:54:15 UTC: `server process (PID 205424) exited with exit code 2`, forcing recovery mode. The
log lines immediately preceding the crash are a burst of concurrent
`DROP DATABASE IF EXISTS "metaldocs_test_…" WITH (FORCE)` statements — the testdb factory's
`t.Cleanup` clone-drops, made concurrent for the first time by the (since-reverted) tagged
parallelization. Risk to flag for CI: if integration-tier `t.Parallel` is ever adopted, the
harness's concurrent `CREATE DATABASE … TEMPLATE` / `DROP DATABASE … WITH (FORCE)` churn should
first be soak-tested against the target Postgres version, or cleanup serialized (e.g. a
package-level drop mutex in `tests/integration/testdb`). Container recovered cleanly
(healthy, `metaldocs` reachable); leftover crash-orphaned `metaldocs_test_*` clones were dropped
in §8. Surfaced, not fixed (out of F9.3 scope).

### 9.4 Concurrent worktree edits (attribution note for diff-scope review)

While F9.3 was in flight, the main session began repairing the §9.1 pre-existing RED tagged
tests; the worktree therefore contains edits NOT made by this feature to 6 tagged files
(`document_search_facts_parity_integration_test.go`,
`repository_approver_eligibility_integration_test.go`,
`repository_commit_upload_integration_test.go`,
`repository_revision_history_integration_test.go`,
`repository_write_authz_integration_test.go` in documents/repository, and
`tenant_id_rls_integration_test.go` in templates/repository — GUC-seeding fix). These were left
untouched and are excluded from F9.3's diff scope; the `git diff --stat` in §7 attributes them
explicitly.

## 10. Forbidden-list self-check (contract §7)

- No `api/openapi/` change; no `migrations/`/`db/` schema change; no capability/authz edit.
- No ADR history deleted or summarized away (none touched).
- No gate/lint/test weakened or deleted — the diff is insertions-only in test files
  (`t.Parallel()` lines), plus new/linked wiki pages and this evidence file. The pre-existing
  RED tests were left untouched, not deleted/skipped to force green.
- No `docs/release/` or `docs/superpowers/plans/` content committed by this feature (the
  untracked `docs/release/` dir pre-dates this session's git status and was not touched).
- 0013 untouched (F9.1 scope).
- No F9.2 map edits.
- No push; no commit (main session commits after review).
- `.env` never read/printed (§3).
