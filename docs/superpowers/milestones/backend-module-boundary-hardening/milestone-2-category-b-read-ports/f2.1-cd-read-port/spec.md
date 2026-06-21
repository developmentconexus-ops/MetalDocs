# Feature F2.1 — Spec — CD read-port (profile_code / process_area_code)

> **Milestone:** 2 — Category B: owner-published read-ports  ·  **Folder:** `f2.1-cd-read-port`
> **Status:** Approved (pre-code) — 2026-06-20 / leandrotca (M2 spec-gate approval covers this feature's contract)

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

The consumer contract is **explicit in the three existing call sites** (B1, B5, B6) read in source this
session — the port shape is dictated by what each site already consumes, not invented. The open
questions were *design* decisions (port shape, tx-awareness), resolved against in-repo precedent
(ADR 0029 `UserDisplayNameReader`, the `db.DB` executor abstraction). Recorded below.

| # | Question | Answer |
|---|----------|--------|
| 1 | What fact does each consumer read from `controlled_documents`? | **B1** (`documents/repository/repository.go:1701`): `profile_code` for a cd-id (off-tx, `r.db`). **B5** (`documents/application/document_area.go:37`): `cd.process_area_code` as the 2nd term of `COALESCE(d.process_area_code_snapshot, cd.process_area_code, '')`, via `LEFT JOIN controlled_documents` keyed on `d.controlled_document_id` (in-tx). **B6** (`documents/approval/application/read_service.go:355`): identical `cd.process_area_code` term in `loadInstanceAreaCode`'s COALESCE (in-tx). |
| 2 | One port method or two? | **Two:** `ProfileCode(cdID)` (B1) and `ProcessAreaCode(cdID)` (B5/B6) — distinct facts, distinct call sites; consumer-driven, no speculative union. |
| 3 | How is tx-awareness handled (B5/B6 in-tx, B1 off-tx)? | Each method takes a **`db.DB` executor** argument (`internal/platform/db`). `*sql.DB` (pool) **and** `*sql.Tx` both satisfy `db.DB` structurally (no adapter — see `db/tx.go`). Caller passes its pool off-tx (B1) or its `*sql.Tx` in-tx (B5/B6). One stateless adapter serves both; the read runs in the caller's tx when a tx is passed (HS-PRE-1: plain non-recording `SELECT`, stays non-recording). |
| 4 | B5/B6 reach `cd.process_area_code` via a JOIN on `d.controlled_document_id`. How does the split preserve COALESCE semantics? | The consumer keeps its **own**-table read (`documents.process_area_code_snapshot`, `documents.controlled_document_id`) in its own SQL; only the `controlled_documents` term moves to the port. Go reproduces `COALESCE(snapshot, cd.process_area_code, '')`: a **non-NULL** own snapshot wins even when `''`; only a **NULL** snapshot falls through to the port; a missing cd row (`found=false`) contributes `''`. This NULL-vs-empty ordering is the parity risk and is pinned by the parity tests. |
| 5 | B5's `LoadDocumentAreaCode` is a free function called from ~8 sites; B6's `loadInstanceAreaCode` likewise. How does the port reach them? | A *plan* decision (signature threading vs package wiring). The **contract** is unchanged regardless: `(areaCode string, found bool, err error)` identical to today. Recorded as a plan risk; not a contract question. |

## Consumer contract (FIRST — before any producer)

- **Consumers:**
  - `documents/repository` `Repository.finalizeApprovalPrereqs` (B1) — needs the controlled document's
    `profile_code`.
  - `documents/application.LoadDocumentAreaCode` (B5) and `documents/approval/application.loadInstanceAreaCode`
    (B6) — need the controlled document's `process_area_code` as the JOIN-fallback term in their area
    COALESCE, executed **inside the caller's existing `*sql.Tx`**.
- **Contract** — a CD-owned interface in `controlleddocuments/domain`:
  ```go
  type CDFieldReader interface {
      // ProfileCode returns controlled_documents.profile_code for (tenantID, cdID).
      // ("", nil) when the cd row is absent — matches the current sql.ErrNoRows-tolerant read.
      ProfileCode(ctx context.Context, exec db.DB, tenantID, cdID string) (string, error)
      // ProcessAreaCode returns controlled_documents.process_area_code for (tenantID, cdID).
      // found=false when the cd row is absent (caller coalesces to ""). Runs on exec
      // (pool off-tx, or the caller's *sql.Tx in-tx).
      ProcessAreaCode(ctx context.Context, exec db.DB, tenantID, cdID string) (areaCode string, found bool, err error)
  }
  ```
  (Final method set / whether `ProfileCode` also takes `exec` vs an injected pool is a plan refinement;
  the **facts returned** and the **tx-aware execution** are the binding contract.)
- **Source of truth for the contract:** the three consumer call sites named above (read in source
  2026-06-20). Owner = controlleddocuments (writes `controlled_documents` — F0.2 census owner map).
  Pattern precedent: ADR 0029 `UserDisplayNameReader`, ADR 0039 D3(b).

## What this feature implements

controlleddocuments publishes `CDFieldReader` (interface in `controlleddocuments/domain`, Postgres
adapter in `controlleddocuments/infrastructure`, wired at the composition root). The three consumer
sites stop naming `controlled_documents` in their own SQL and call the port instead; B5/B6 pass their
existing `*sql.Tx` so the read stays in-tx and non-recording. Each raw read is deleted **only after**
its parity test is green.

## Non-goals (mandatory)

- **No change to results, visibility, or authz.** Seam change only; parity tests lock byte-identical output.
- **No migration, no view, no schema change, no denormalized snapshot column.** Reads stay live.
- **No re-porting the `user_process_areas` membership reads** that sit in the same documents/approval
  files — those are **M3** (published view). Touch only the `controlled_documents` reads.
- **No refactor of the COALESCE area-resolver logic or its ~8 call sites** beyond the minimum needed to
  route the `cd.process_area_code` term through the port. `LoadDocumentAreaCode`'s return signature is unchanged.
- **No second/parallel reader** over `controlled_documents` — CD already owns its repository; the new
  port is the single cross-module read surface (ADR 0030 "no parallel reader" rule).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| B1 port path returns the same `profile_code` as the raw read (present / absent-cd cases) | `TestCDFieldReader_ProfileCode_ParityWithRawSQL` (integration, :5433) | real (PG) |
| B5/B6 port path returns the same `process_area_code` term; COALESCE preserved across NULL-snapshot / empty-snapshot / cd-absent cases | `TestCDFieldReader_ProcessAreaCode_ParityWithRawSQL` + resolver-level `TestLoadDocumentAreaCode_ParityPrePostPort` / `TestLoadInstanceAreaCode_ParityPrePostPort` (integration, :5433) | real (PG) |
| `cd.process_area_code` read runs inside the caller's tx (in-tx, non-recording — HS-PRE-1) | reviewer confirms `exec` is the caller's `*sql.Tx` at B5/B6; no `authz_*` write added | real (review) |
| Whole tree builds and tests pass | `go build ./...`; `go test ./...` | real |
| Guard clears B1/B5/B6 | `go run ./tools/cilint ./...` exit 0 with `{documents/repository/repository.go,controlled_documents}`, `{documents/application/document_area.go,controlled_documents}`, `{documents/approval/application/read_service.go,controlled_documents}` removed from `hgPendingRemediation` | real |
| 0 raw `controlled_documents` reads remain under `documents/` | `git grep -n "controlled_documents" internal/modules/documents` shows only the port-call comments / own-table FKs, no `FROM/JOIN controlled_documents` | real |

> TDD: write the failing parity test first, then implement the port to green, then delete the raw read
> (parity must be green **before** deletion — D6). If Docker Postgres :5433 is unavailable, the
> integration parity steps are marked **not-run (HS-3)** in evidence — never false-green.

## ADR needed?

- [x] **No new ADR.** The durable decision (cross-module base-table read → owner-published read-port) is
  already recorded in **ADR-0039** (D1 + D3(b)); this feature is a mechanical application. The
  port's specific shape (two facts, `db.DB` executor for tx-awareness, COALESCE-split parity) is recorded
  in **this spec** + the feature `evidence.md`. No parallel/novel cross-module contract is introduced
  (contrast ADR 0029/0030/0038, which each *established* a pattern; 0039 generalized it).
