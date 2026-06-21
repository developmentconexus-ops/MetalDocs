# Feature F2.2 — Evidence — documents active-instance read-port

> **Milestone:** M2 (category-b-read-ports) · **Feature:** `f2.2-documents-read-port` · **Closed:** 2026-06-21
> **Contract:** `spec.md` (consumer = controlleddocuments `GetActiveInstance`; ADR-0039 D3(b) owner read-port).
> Census sites closed: **B2** (`controlleddocuments/infrastructure/repository.go` → `document_revisions`),
> **B3** (→ `documents`), **B4** (→ `approval_instances`).

## What was implemented

By outcome:

- **Owner publishes the projection port.** `documents/domain` gains `ActiveInstanceView` +
  `ActiveInstanceReader` interface (+ `NoopActiveInstanceReader` fail-closed default) —
  `internal/modules/documents/domain/active_instance_port.go`. The consumer (controlleddocuments)
  depends only on this interface; it never names the documents base tables.
- **Owner-side adapter.** `documents/repository` gains `ActiveInstanceReaderPG`
  (`internal/modules/documents/repository/active_instance_reader.go`), which runs the **identical**
  FULL OUTER JOIN projection over `documents`/`document_revisions` plus the derived in-progress
  `approval_instances` lookup that `GetActiveInstance` historically ran inline. All three tables are
  documents-owned (the approval sub-context maps to the `documents` top-level owner), so the read is
  intra-module. Status filters are passed as **owner typed-constant params**
  (`documents/domain.DocStatus*`, `approval/domain.InstanceInProgress`) — no bare literals.
- **Consumer rewired, raw SQL deleted.** `controlleddocuments/infrastructure` struct gains
  `activeInstance documentsdomain.ActiveInstanceReader`; `NewPostgresControlledDocumentRepository(db,
  activeInstance)` (nil→Noop guard). `GetActiveInstance` body is now a single port call + 1:1
  view→`ActiveDocumentInstance` map. The inline `documents`/`document_revisions`/`approval_instances`
  reads are **removed**.
- **Composition root.** `controlleddocuments/module.go` constructs
  `docrepo.NewActiveInstanceReaderPG(deps.DB)` and injects it into the repo.
- **Ledger drained.** B2/B3/B4 removed from `hgPendingRemediation` in
  `tools/cilint/internal/analyzers/hgcrossmodule.go`.

> Producer matches the consumer contract in `spec.md`: the consumer needed exactly
> `{DocumentID, ContentHash, RevisionVersion, ApprovalState→Status, PublishedDocumentID,
> ApprovalInstanceID}`; the port's `ActiveInstanceView` carries precisely those fields and the CD
> repo maps them 1:1.
>
> Note vs `plan.md`: the adapter was placed in `documents/repository` (the canonical home for the
> module's other PG reader adapters) rather than a `documents/infrastructure` package — artifact is
> truth; boundary (consumer→owner-domain interface) is unchanged.

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — parity test first, green before raw-read deletion (D6) | `go test -tags integration -run TestActiveInstanceReader_ParityWithRawGetActiveInstance ./internal/modules/controlleddocuments/infrastructure/` | `PASS` — 4/4 subtests (`active_draft_only`, `published_only`, `under_review_with_in_progress_approval`, `none`) | real (PG :5434) |
| Static — build | `go build ./...` | clean (no output, exit 0) | — |
| Static — guard | `go run ./tools/cilint ./...` | `EXIT=0` with B2/B3/B4 removed from `hgPendingRemediation` | real |
| Targeted tests — CD + documents modules (unit) | `go test ./internal/modules/controlleddocuments/... ./internal/modules/documents/...` | all `ok` (every package) | real |
| Targeted tests — CD integration | `go test -tags integration ./internal/modules/controlleddocuments/...` | `application`/`delivery/http`/`infrastructure` `ok`; one pre-existing env FAIL in `domain` (see defers) | real (PG) |
| Runtime proof — port path == raw path | parity test seeds CD+documents in each state, runs the verbatim pre-port SQL baseline (`rawGetActiveInstance`) and the port-backed `repo.GetActiveInstance`, asserts field-by-field equality | identical projection in all 4 states; under_review case resolves the in-progress approval instance id | real (PG) |
| No bare foreign status literals | `grep -nE "'(draft|under_review|…|in_progress)'" active_instance_reader.go` | none; SQL uses `DocStatus*` + `InstanceInProgress` params | real (review) |
| 0 raw foreign reads under `controlleddocuments/` (non-test) | `grep -rniE 'from/join (documents|document_revisions|approval_instances)'` | none | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Port path returns same projection as raw across active-only / published-only / under-review / none | yes | parity test row — 4/4 PASS |
| Whole tree builds; targeted tests pass | yes | `go build ./...` clean; CD+documents module suites `ok` |
| Guard clears B2/B3/B4 | yes | cilint `EXIT=0`, ledger entries removed |
| 0 raw documents/document_revisions/approval_instances reads under `controlleddocuments/` | yes | grep clean (non-test) |
| 0 bare foreign status literals in ported query | yes | adapter uses typed-constant params |

## Review disposition

- **Spec-compliance review:** PASS. Consumer contract (`ActiveInstanceView` fields) drove the port
  shape; off-tx as specified; owner-typed status constants as specified; no parallel cross-module
  contract introduced (mechanical ADR-0039 D3(b)).
- **Code-quality review:** PASS. Port projection is byte-for-byte the historical SQL with literals
  swapped for typed-constant params (parity test locks behavioral equivalence). Nil-guard at the
  ctor + `NoopActiveInstanceReader` keeps unit call sites fail-closed. No import cycle
  (`controlleddocuments/infrastructure → documents/domain`; `controlleddocuments/module →
  documents/repository → controlleddocuments/domain`; verified by `go build`).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `TestSequenceAllocatorNextAndIncrement_Concurrent` (controlleddocuments/domain) FAILs: `relation "metaldocs.document_profiles" does not exist (SQLSTATE 42P01)` | **Pre-existing / env, not F2.2.** Test connects to the **raw base DSN** (not a migrated per-test clone) and inserts into `metaldocs.document_profiles`, absent on the base `:5434` DB. Not in F2.2 diff; uses only the unchanged `NewPostgresSequenceAllocator` (F2.2 changed `NewPostgresControlledDocumentRepository`). HS-3 class — marked not-run, never false-greened. | Repair raw-DSN base-DB schema/migration for the sequence integration test (env owner); out of M2 seam scope |
| Whole-tree `go test ./...` not run green end-to-end | Known pre-existing breaks elsewhere (iam live limiter, docx_v2 arity, etc. — documented in F2.1 evidence), unrelated to this seam | Tracked in F2.1 defers; not introduced by F2.2 |
