# Feature F2.1 — Evidence — CD read-port (profile_code / process_area_code)

> **Milestone:** 2  ·  **Feature:** `f2.1-cd-read-port`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).
> Covers census sites **B1, B5, B6** (the three `controlled_documents` foreign point-reads under `documents/`).

## What was implemented

By outcome — producer built to match the three consumer call sites read in source (not the reverse):

- **Owner port published.** `controlleddocuments/domain.CDFieldReader` interface
  (`ProfileCode(ctx, exec db.DB, tenantID, cdID) (string, error)`;
  `ProcessAreaCode(ctx, exec db.DB, tenantID, cdID) (areaCode string, found bool, err error)`) +
  `NoopCDFieldReader`. Postgres adapter `CDFieldReaderPG` (stateless, `NewCDFieldReaderPG()`) in
  `controlleddocuments/infrastructure`. `exec db.DB` makes both methods tx-aware: `*sql.DB` (off-tx)
  and `*sql.Tx` (in-tx) both satisfy `db.DB` structurally — one adapter serves both. ErrNoRows ⇒
  `("",nil)` / `("",false,nil)`, matching the prior tolerant reads.
- **B1** — `documents/repository/repository.go` `finalizeApprovalPrereqs`: raw `SELECT profile_code
  FROM controlled_documents …` replaced by `r.cdRead.ProfileCode(ctx, r.db, tenantID, cdID)` (off-tx,
  pool). `Repository.New` gained the `cdRead` param; wired through `documents/module.go` (nil ⇒ Noop).
- **B5** — `documents/application.LoadDocumentAreaCode`: `LEFT JOIN controlled_documents` deleted; own
  `documents` read kept (`process_area_code_snapshot`, `controlled_document_id`); the
  `cd.process_area_code` COALESCE term resolved in Go via `cdRead.ProcessAreaCode(ctx, tx, …)` —
  **in-tx**, non-NULL snapshot wins (even `""`), only NULL falls through. Threaded into Reconstruct,
  FillIn, and the four approval services (Submit/Decision/Publish/Supersede) via `cdRead`.
- **B6** — `documents/approval/application.loadInstanceAreaCode`: `LEFT JOIN controlled_documents`
  deleted; own approval/documents read kept; `cd.process_area_code` term resolved via
  `cdRead.ProcessAreaCode(ctx, tx, …)` — **in-tx**; full COALESCE precedence (active-stage snapshot →
  doc snapshot → CD area → "") reproduced in Go. Wired through `ReadService`/`newReadService` and
  `NewServices`.
- **Composition root** — `apps/api/.../main.go` and `apps/jobs/.../main.go` inject
  `cdinfra.NewCDFieldReaderPG()` into documents deps + approval `NewServices`.
- **Guard ledger** — B1, B5, B6 entries removed from `hgPendingRemediation` in
  `tools/cilint/internal/analyzers/hgcrossmodule.go`.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing parity test first, then green, then delete raw SQL (D6) | per-site parity tests authored RED against the new signature, GREEN after impl, raw read deleted only after green | all parity suites PASS (rows below) | real (PG :5434) |
| B1 parity | `go test -tags integration -run TestCDFieldReader_ProfileCode_ParityWithRawSQL ./internal/modules/controlleddocuments/infrastructure/` | `--- PASS: TestCDFieldReader_ProfileCode_ParityWithRawSQL` | real (PG) |
| B5/B6 port-level parity | `…-run TestCDFieldReader_ProcessAreaCode_ParityWithRawSQL …/infrastructure/` | `--- PASS` | real (PG) |
| B5 resolver parity (NULL/empty/cd-absent/no-link + absent doc) | `…-run TestLoadDocumentAreaCode_ParityPrePostPort ./internal/modules/documents/application/` | `--- PASS` (4 subtests) | real (PG) |
| B6 resolver parity (stage/empty-stage/doc/empty-doc/cd-fallback/unlinked + absent instance) | `…-run TestLoadInstanceAreaCode_ParityPrePostPort ./internal/modules/documents/approval/application/` | `--- PASS` (6 subtests) | real (PG) |
| Static — build | `go build ./...` | `===BUILD OK===` (exit 0) | — |
| Guard clears B1/B5/B6 | `go run ./tools/cilint ./...` | `===CILINT EXIT 0===` after removing the three ledger entries | real |
| Regression — documents tree | `go test -tags integration ./internal/modules/documents/...` | all packages `ok` **except** two pre-existing env-schema failures (see Bounded defers) | real (PG) |
| 0 raw `controlled_documents` reads under `documents/` (non-test) | `git grep -nE '(FROM\|JOIN)\s+(public\.\|metaldocs\.)?controlled_documents' -- 'internal/modules/documents/**/*.go' \| grep -v _test.go` | empty (only `_test.go` parity baselines + FK comments remain) | real |
| In-tx, non-recording (HS-PRE-1) | review: B5/B6 pass the caller's `*sql.Tx` as `exec`; method body is a plain `SELECT`, no `authz_*`/recording write | confirmed | real (review) |

> Mock-driver note (honest labeling): unit fakes in `documents/application` + `documents/approval/application`
> were updated to the 2-column area read shape (shared `docAreaRows` driver type; fillin sqlmock gained a
> query-text matcher routing the area read). These are **unit fakes**, not real-provider proof — the
> real-PG parity tests above are the binding acceptance.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| B1 port path == raw `profile_code` (present/absent-cd) | yes | `TestCDFieldReader_ProfileCode_ParityWithRawSQL` PASS |
| B5/B6 port `process_area_code` term; COALESCE preserved (NULL/empty/cd-absent) | yes | `…ProcessAreaCode_ParityWithRawSQL` + `TestLoadDocumentAreaCode_ParityPrePostPort` + `TestLoadInstanceAreaCode_ParityPrePostPort` PASS |
| `cd.process_area_code` read runs in caller's tx (HS-PRE-1) | yes | review row above; `exec` = caller `*sql.Tx`; no recording write |
| Whole tree builds; tests pass | yes (build) / qualified (test) | `go build ./...` exit 0; documents tree green except 2 pre-existing env failures (deferred) |
| Guard clears B1/B5/B6 | yes | cilint exit 0; the three `hgPendingRemediation` entries removed |
| 0 raw `controlled_documents` reads under `documents/` | yes | `git grep` empty (non-test) |

## Review disposition

- **Spec-compliance review:** PASS. Producer shape matches the three consumer sites; no contract
  invented. `LoadDocumentAreaCode` return signature unchanged; only the `cd.process_area_code` term
  moved to the port (non-goals honored — no view/migration/snapshot column, no `user_process_areas`
  touch, no parallel reader).
- **Code-quality review:** PASS by root-cause family. One real cross-cutting issue surfaced and fixed
  at the family level: the B5 area-read column shape changed 1→2 columns, breaking every hand-rolled
  fake driver and the fillin sqlmock. Fixed by a single shared `docAreaRows` driver type (not N
  per-file copies) and a substring query-matcher for the sqlmock — addressing the whole family, not
  each call site ad hoc.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestPostgresLimiter_Live` (`auth_failure_counters` absent) + `TestLoadActorDisplayName_ReadsOffTxAgainstLiveSchema` (`metaldocs.tenants` absent) fail in the throwaway :5434 container | **Pre-existing, not a regression (HS-3):** both fail in test *setup* (table provisioning) before any ported path; neither test file is in this feature's diff; both reference tables unrelated to `controlled_documents`/cdRead. The ephemeral container's template omits these base-schema tables. | Re-run on the fully-migrated dev DB (:5433) where these tables exist; owner = milestone-validator regression step. Marked **not-run (HS-3)**, never false-green. |
| Pre-existing branch build breaks outside scope: `tests/integration/documents/concurrent_revision_test.go` (`internal/testdb` ghost import), `tests/docx_v2` (`application.New` arity drift — no CDFieldReader involved), `tests/integration/iam` (`MembershipGovernanceLogger.LogTx`) | Not in this feature's diff; unrelated to cdRead (vet symbol filter for cd/area returned none of these). | Track under their own owners; out of M2/F2.1 boundary. |
