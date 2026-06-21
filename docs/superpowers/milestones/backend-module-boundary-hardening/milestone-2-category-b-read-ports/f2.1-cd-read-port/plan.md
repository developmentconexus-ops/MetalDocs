# Feature F2.1 — CD read-port (profile_code / process_area_code)

> **Milestone:** 2 — Category B: owner-published read-ports  ·  **Folder:** `f2.1-cd-read-port`
> **Status:** Planning

## Source

- Milestone spec row F2.1: CD publishes a read-port for the profile/CD point-reads consumed by
  documents/approval (B1, B5, B6); tx-aware for the in-tx callers (B5, B6). Parity per site green
  **before** the raw read is deleted; build/test green; cilint clears B1/B5/B6.
- Governing-spec reference: mission §5 rows 4, 8, 9; §7 M2 F2.1. ADR-0039 D1 + D3(b). Precedent ADR 0029
  (`UserDisplayNameReader`), the `db.DB` executor (`internal/platform/db/tx.go`).

## Plan

### Design (locked in spec.md)

- New CD-owned interface `controlleddocuments/domain.CDFieldReader` with two consumer-driven methods,
  each taking a `db.DB` executor so the **same stateless adapter** serves off-tx (pass `*sql.DB`) and
  in-tx (pass the caller's `*sql.Tx`) — `*sql.Tx` satisfies `db.DB` structurally:
  - `ProfileCode(ctx, exec db.DB, tenantID, cdID) (string, error)` → `("", nil)` when absent.
  - `ProcessAreaCode(ctx, exec db.DB, tenantID, cdID) (areaCode string, found bool, err error)`.
- Postgres adapter `controlleddocuments/infrastructure.CDFieldReaderPG` (stateless; the SQL + table
  name `controlled_documents` live here, the only home).
- `NoopCDFieldReader` null-object in `controlleddocuments/domain` for tests (mirrors
  `NoopUserDisplayNameReader`).
- **No import cycle:** `controlleddocuments/domain` imports only `context` + `platform/db` (verify it
  imports nothing from `documents`). Consumers import the **interface** only.
- **B5/B6 COALESCE split (parity-critical):** consumer keeps its own `documents` read
  (`process_area_code_snapshot`, `controlled_document_id`); Go reproduces
  `COALESCE(snapshot, cd.process_area_code, '')` — non-NULL snapshot wins even when `""`; only a NULL
  snapshot falls through to `ProcessAreaCode`; `found=false` ⇒ `""`. The own read must distinguish
  NULL from `""` (scan into `sql.NullString`).

### Tasks (TDD order — parity test green BEFORE each raw read is deleted, D6)

**T1 — Port + adapter + null-object (no consumer change yet).**
- `controlleddocuments/domain/cd_field_reader_port.go`: `CDFieldReader` interface + `NoopCDFieldReader`.
- `controlleddocuments/infrastructure/cd_field_reader.go`: `CDFieldReaderPG` impl (stateless), SQL:
  `SELECT profile_code FROM controlled_documents WHERE id=$1 AND tenant_id=$2` and
  `SELECT process_area_code FROM controlled_documents WHERE id=$1 AND tenant_id=$2`, both via
  `exec.QueryRowContext`, `sql.ErrNoRows` → `("",nil)` / `("",false,nil)`.
- `go build ./...` green.

**T2 — B1 parity test (failing first) → wire → delete raw.**
- Test `controlleddocuments/infrastructure/cd_field_reader_integration_test.go`:
  `TestCDFieldReader_ProfileCode_ParityWithRawSQL` — seed a CD row (canonical testdb fixture, ADR 0034);
  assert `ProfileCode` == the value the inline raw `SELECT profile_code …` returns, for present + absent-cd.
- `documents/repository.Repository`: add field `cdRead controlleddocumentsdomain.CDFieldReader`; `New`
  gains a 3rd param. Replace the raw read at `repository.go:1701` with
  `r.cdRead.ProfileCode(ctx, r.db, tenantID, cdID.String)`.
- Update all `repository.New(...)` call sites: `module.go:63` (wire real `CDFieldReaderPG`) + the test
  sites (`list_documents_paginated_test.go`, `repository_revision_history_integration_test.go`,
  `repository_create_integration_test.go` ×4, `repository_commit_upload_integration_test.go` ×2) → pass
  `controlleddocumentsdomain.NoopCDFieldReader{}` (none exercise the profile-code path).
- Parity test green → delete the raw `controlled_documents` SQL (done above) → remove
  `{"documents/repository/repository.go","controlled_documents"}` from `hgPendingRemediation`.

**T3 — B5 parity test (failing first) → wire → delete raw.**
- Test `documents/application/document_area_parity_integration_test.go`:
  `TestLoadDocumentAreaCode_ParityPrePostPort` — seed doc with (i) non-NULL snapshot, (ii) `""` snapshot,
  (iii) NULL snapshot + CD area present, (iv) NULL snapshot + no CD; assert post-port result == the
  current JOIN query's result for every case.
- `LoadDocumentAreaCode` signature gains `cdRead controlleddocumentsdomain.CDFieldReader`; drop the
  `LEFT JOIN controlled_documents`; read `process_area_code_snapshot` (NullString) +
  `controlled_document_id` from `documents`; COALESCE in Go via `ProcessAreaCode(ctx, tx, …)` (tx-aware).
- Thread `cdRead` through callers (each gains a `cdRead` field + constructor param, wired at the root):
  `documents/application` `reconstruct_service.go:36`, `fillin_authz.go:31`;
  `documents/approval/application` `submit_service.go:88`, `publish_service.go:80,213`,
  `supersede_service.go:54`, `decision_service.go:209`.
- Parity green → JOIN already gone → remove
  `{"documents/application/document_area.go","controlled_documents"}` from the ledger.

**T4 — B6 parity test (failing first) → wire → delete raw.**
- Test `documents/approval/application/read_service_area_parity_integration_test.go`:
  `TestLoadInstanceAreaCode_ParityPrePostPort` — seed instance/doc/CD across the same NULL/empty/absent
  cases (incl. active-stage snapshot precedence `asi.area_code_snapshot`); assert parity.
- `loadInstanceAreaCode` gains `cdRead` param; drop `LEFT JOIN controlled_documents`; keep the
  approval/documents own-table reads (`approval_instances`, `documents`, `approval_stage_instances` are
  documents-owned — same module, compliant); COALESCE `cd.process_area_code` via the port in-tx.
- Thread `cdRead` into `ReadService` (`newReadService`) and any other `loadInstanceAreaCode` caller.
- Parity green → remove `{"documents/approval/application/read_service.go","controlled_documents"}`.

**T5 — composition-root wiring + close-out.**
- `apps/api/cmd/metaldocs-api/main.go`: construct one `cdReader := cdinfra.NewCDFieldReaderPG(deps.SQLDB)`
  and pass it into `documents.Dependencies` (new field, defaulting to `NoopCDFieldReader` when nil) and
  into the approval services that call the two helpers. Mirror in `apps/jobs/cmd/metaldocs-jobs/main.go`
  if it constructs the affected services.
- `go build ./...`; `go test ./...`; `go run ./tools/cilint ./...` exit 0 (B1/B5/B6 removed from ledger).
- `git grep -n "controlled_documents" internal/modules/documents` → 0 `FROM/JOIN controlled_documents`.

### Test strategy

- Parity tests are **integration** (need PG :5433, canonical testdb fixture per ADR 0034). If :5433 is
  down, mark **not-run (HS-3)** in evidence — never false-green. Unit-level guard: a fake `CDFieldReader`
  asserts the COALESCE-split Go logic (NULL vs `""` vs absent) without a DB, so the parity *logic* is
  covered even if integration can't run.
- Existing suites (`resolution_test.go`, area-resolver-dependent authz tests) must stay green unchanged.

## Execution notes

- Blast radius for B5/B6 is the threading of one collaborator through ~6 services / 8 call sites — DI
  of a read-port (the ADR-0029 displayName precedent), **not** an API redesign → not HS-2. Recorded here
  so the validator sees it was planned, not drift.
- If any caller turns out to need a CD fact beyond `process_area_code` (forcing a contract reshape) →
  HS-2 stop. Not expected.
