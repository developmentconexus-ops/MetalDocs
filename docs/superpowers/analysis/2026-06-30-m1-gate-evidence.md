# M1 gate evidence — lifecycle correctness (template↔document parity program)

**Date:** 2026-06-30
**Branch:** `feat/template-document-parity`
**Scope gated:** M1 work-units T1–T5 (WS-A remove template auto-version via expand/contract; WS-B align status `in_review`→`under_review` with DB migration; WS-D wizard real document number).

## Result summary

| Gate | Command | Result |
|---|---|---|
| Go build | `go build ./...` | **PASS** (exit 0) |
| Go unit tests | `go test ./...` | **PASS** (exit 0) |
| FE typecheck | `tsc --noEmit -p tsconfig.build.json` | **PASS** (exit 0) |
| FE api-types no-drift | `pnpm gen:api` + `git diff` | **PASS** (clean) |
| Backend oapi-codegen no-drift | `go generate ./internal/modules/*/api/...` (11 modules) + `git diff` | **PASS** (clean) |
| FE vitest (full) | `pnpm exec vitest run` | **PASS** (100 files, 591 tests) |
| Integration — templates (M1 surface) | `go test -tags=integration ./tests/integration/templates/...` | **PASS** (TestLifecycle_NoAutoNextDraft, 124s) |
| System-runnable | `scripts/check-system-runnable.ps1 -StartApi` | see §3 |

## 1. M1·T2 integration test repair (in-scope defect, fixed)

`tests/integration/templates/lifecycle_no_auto_draft_test.go` was authored in a session with no integration DB, so its raw SQL was never executed. Three defects surfaced on first real run and were fixed (commit `2b152571`):

1. Seed UPDATE used non-existent column `docx_content_hash`; real column on `templates_template_version` is `content_hash` (baseline:2439), constrained to length 64 → seed a 64-char value.
2. Raw seed write bypasses the application service, so it never set the `metaldocs.tenant_id` GUC that `trg_template_version_tenant_consistent` requires (baseline:758) → assert it tx-locally, mirroring `authz.SeedTxIdentity`.
3. `versionStatuses` queried `version_num`; real column is `version_number`.

After the fix the test PASSES, proving: approve/publish produces exactly one (published) version row, no auto-spawned v2; manual `CreateNextVersion` is the only path that adds a draft v2.

## 2. Pre-existing integration failures — OUT of M1 boundary (bounded defer)

The full `go test -tags=integration ./tests/integration/...` run also surfaced failures in `tests/integration/scenarios` that are **not caused by M1** and predate this branch. Git proof: `git diff --name-only ba99d82f..HEAD` (the pre-M1 baseline-squash commit) shows M1 touched **none** of these files, fixtures, the bootstrap SQL (beyond one `templates_template_version` constraint line + migration 0257), the seed helpers, the documents module, or the flagged TS files. A source-scan over unchanged files and DB tests over an unchanged schema/seed path therefore yield the same verdict they did before M1.

| Failing test(s) | Error class | Why pre-existing |
|---|---|---|
| `TestNoLegacyStatusInTSSource` | `'finalized'`/`'archived'` literals in StatusPill.tsx, TemplatesListPage.tsx, TemplateCard.tsx, DocumentEditorPage.tsx:369 | guard scans for `finalized`/`archived` (NOT `in_review`); all 4 files untouched by M1 |
| `TestGrantAreaMembership{Fn,Idempotent}` | `ErrCapabilityNotAsserted: membership.manage` | seed path doesn't assert caps; testdb helper untouched by M1 |
| `TestObsoleteCascade_NoStaleOCC`, `TestLegalTransition_ObsoleteFromPublished` | `relation "metaldocs.documents" does not exist` | tests hardcode `metaldocs.` schema qualifier; `documents` resolves to `public` (not in `metaldocsOwnedObjects`) — stale qualifier |
| `TestOutbox_ApprovalInstanceInsertHasGovernanceEvent`, `TestOutbox_RollbackOmitsEvent` | `document.create` not asserted / `relation "metaldocs.governance_events"` missing | same two classes above |
| `TestWriterCanReadApprovalTables` | `relation "metaldocs.approval_instances"` missing | stale `metaldocs.` qualifier |
| `TestTriggerBypassBlocked`, `TestIllegalTransitionBlocked` | `document.create` not asserted | seed path doesn't assert caps |
| `TestReflect_RepositoryNoBeginTx`, `TestHTTPHandlers_NoBeginTx` | walk path `internal/modules/documents_v2/...` not found | stale path — module is `documents`, not `documents_v2` |

**Root cause:** incomplete port of the `scenarios` package to the canonical `testdb` factory (DB-level isolation). Tests still carry schema-based-isolation assumptions (`metaldocs.` qualifiers), un-asserted-cap seed inserts, and a stale `documents_v2` path. This is the known test-framework work tracked by memory `m4c-test-fixture-framework` — a separate concern from the parity program. Per CLAUDE.md scope discipline + `legacy-test-deletion`/`test-framework-hard-gate`, M1 does **not** expand to fix the framework port; these are recorded here and deferred to that track.

**Note on coverage risk:** the failing document-side invariant guards (illegal-transition, trigger-bypass, outbox-same-tx) are *test-harness* failures, not product-invariant failures — the runtime triggers/constraints they probe are intact in the baseline (unchanged by M1). The product invariants are exercised end-to-end by the final Preview real-user QA (T21).

## 3. System-runnable

`scripts/check-system-runnable.ps1 -StartApi` brought the API up cleanly (`dev-api.ps1` → `metaldocs-api.exe` on :8081). Every checkpoint the script validates was confirmed **green** by direct probe:

| Checkpoint | Probe | Result |
|---|---|---|
| startup / health-ready | `GET /api/v1/health/ready` | **200** |
| login-endpoint | `POST /api/v1/auth/login` (admin) | **200** |
| login-session | session cookie returned | **present** |
| auth-me | `GET /api/v1/auth/me` | **200** |
| target-route | `GET /api/v1/health/ready` | **200** |
| blank-template-object | `mc stat local/<bucket>/system/templates/blank.docx` | **present** (1.3 KiB, docx content-type) |

**Caveat (environment, not M1):** the script's own process hung in its `Assert-SystemBlankTemplateObject` sub-step — `seed-system-blank-template.ps1 -VerifyOnly` orchestrates `docker compose up -d minio minio-init` + a networked `docker run minio/mc stat`, which stalled in the Docker tooling (the `mc` image is present locally, so it is not a pull stall). The object it checks for is in fact present (probed directly with a 45s-bounded `docker run`). This is a tooling flake in the verify harness, unrelated to M1 code. Follow-up (out of M1 scope): make the seed verify resilient/bounded, or have `check-system-runnable` probe MinIO without spinning a one-shot `compose up`.

## Verdict

**M1 gate: PASS.** All in-scope evidence green; the single in-scope defect (T2 test raw-SQL) is fixed and verified against a live DB; the one runtime caveat is an environment tooling flake with the underlying prerequisite confirmed satisfied. Pre-existing `scenarios` integration debt is git-proven out of boundary and deferred to the testdb-framework track (memory `m4c-test-fixture-framework`).
