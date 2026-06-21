# Feature F2.2 — Spec — documents active-instance read-port

> **Milestone:** 2 — Category B: owner-published read-ports  ·  **Folder:** `f2.2-documents-read-port`
> **Status:** Approved (pre-code) — 2026-06-20 / leandrotca (M2 spec-gate approval covers this feature's contract)
> Covers census sites **B2** (`document_revisions`), **B3** (`documents`), **B4** (`approval_instances`)
> — the three documents-owned reads inside CD's `GetActiveInstance`.

## Interview record (fail-closed gate)

The consumer contract is **explicit in one existing call site**: `PostgresControlledDocumentRepository.GetActiveInstance`
(`controlleddocuments/infrastructure/repository.go:521`). The port shape is dictated by exactly what
that method already assembles into `controlleddocuments/domain.ActiveDocumentInstance`.

| # | Question | Answer |
|---|----------|--------|
| 1 | What does CD read from documents-owned tables? | A single **active-instance projection** for `(tenantID, controlledDocumentID)`: the active document row (`documents`, status ∈ draft/under_review/approved/rejected/scheduled, LIMIT 1) → `id, content_hash_at_submit, revision_version, status`; **B2** the latest `document_revisions.content_hash` as the content-hash fallback; **B3** the most-recent `documents` row with status='published' (the published id); **B4** when the active status is `under_review`, the in-progress `approval_instances.id`. |
| 2 | One port or several? | **One** owner method returning the whole projection. `documents` (top-level) owns all three tables — `approval_instances` is the `documents/approval` sub-context, which `hgOwnerByTable` maps to `documents`. So a single adapter in `documents/infrastructure` does all reads **intra-module** (no H-G); CD calls once and maps 1:1. No "god query" beyond the exact projection CD already consumes (appetite: consumer-driven). |
| 3 | Off-tx or in-tx? | **Off-tx.** `GetActiveInstance` runs on CD's pool (`r.db`), no surrounding tx. The port adapter holds the documents pool (wired at composition root); no tx threading, no HS-PRE-1 concern. |
| 4 | Where does the interface live, and does it create an import cycle? | Interface + view type in **`documents/domain`** (owner). CD `infrastructure` imports `documents/domain` (consumer→owner). `documents/domain` imports nothing from controlleddocuments → no cycle. (documents already imports `controlleddocuments/domain` for F2.1's CDFieldReader; the reverse edge here is infrastructure→domain, a different package pair — no package cycle.) |
| 5 | Foreign status literals? | The moved query's status strings (`'draft'`,`'under_review'`,`'approved'`,`'rejected'`,`'scheduled'`,`'published'`,`'in_progress'`) are the **owner's** vocabulary and must be the owner's typed constants — `documents/domain.DocStatus*` (model.go:11-16) and `documents/approval/domain.InstanceInProgress` (instance.go:19) — referenced in Go and passed as query parameters, not bare literals in the adapter SQL. |

## Consumer contract (FIRST — before any producer)

- **Consumer:** `controlleddocuments/infrastructure.PostgresControlledDocumentRepository.GetActiveInstance`
  — needs the active-instance projection it currently maps to `domain.ActiveDocumentInstance`.
- **Contract** — a documents-owned interface in `documents/domain`:
  ```go
  // ActiveInstanceView is the documents-owned projection of a controlled
  // document's active/published documents (+ in-progress approval instance).
  // All pointer fields are nil when the corresponding row/column is absent.
  type ActiveInstanceView struct {
      DocumentID          *string // active (non-published) document id
      ContentHash         *string // active content hash (content_hash_at_submit, else latest revision)
      RevisionVersion     *int    // active revision_version
      Status              *string // active document status (owner vocabulary)
      PublishedDocumentID *string // most-recent published document id
      ApprovalInstanceID  *string // in-progress approval instance id (only when Status == under_review)
  }

  type ActiveInstanceReader interface {
      // ActiveInstanceForControlledDocument returns the projection for
      // (tenantID, controlledDocumentID), or nil when neither an active nor a
      // published document exists. Off-tx (runs on the owner's pool).
      ActiveInstanceForControlledDocument(ctx context.Context, tenantID, controlledDocumentID string) (*ActiveInstanceView, error)
  }
  ```
- **Source of truth:** the `GetActiveInstance` body (read 2026-06-20). Owner = documents (writes
  `documents`/`document_revisions`/`approval_instances` — F0.2 census owner map). Precedent: ADR-0039
  D3(b), ADR 0029/0030 reader-port shape.

## What this feature implements

documents publishes `ActiveInstanceReader` (interface + `ActiveInstanceView` in `documents/domain`,
Postgres adapter in `documents/infrastructure`, wired at the composition root). The adapter runs the
**identical** FULL OUTER JOIN projection + the derived in-progress approval lookup that
`GetActiveInstance` runs today, with status literals replaced by the owner's typed constants passed as
parameters. CD's `GetActiveInstance` drops its three foreign reads and instead calls the port and maps
the view 1:1 onto `ActiveDocumentInstance`. The raw reads are deleted **only after** the parity test is
green.

## Non-goals (mandatory)

- **No behavior/visibility/authz change.** Seam only; parity test locks byte-identical projection.
- **No migration, view, or schema change.** Reads stay live.
- **No re-port of `user_process_areas`** membership reads in the same file (`repository.go:150,492`) — those are **M3**.
- **No "god query"** beyond the exact projection CD consumes; no speculative fields.
- **No refactor of CD's `ActiveDocumentInstance` mapping** beyond replacing the three reads with the port call.
- **No parallel reader** over documents tables — the new port is the single cross-module read surface.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Port path returns the same projection as the raw `GetActiveInstance` reads across **active-only / published-only / under-review (with approval instance) / none** cases | `TestActiveInstanceReader_ParityWithRawGetActiveInstance` (integration, :5434) | real (PG) |
| Whole tree builds and tests pass | `go build ./...`; `go test ./...` | real |
| Guard clears B2/B3/B4 | `go run ./tools/cilint ./...` exit 0 with `{controlleddocuments/infrastructure/repository.go, document_revisions}`, `{…, documents}`, `{…, approval_instances}` removed from `hgPendingRemediation` | real |
| 0 raw documents/document_revisions/approval_instances reads remain under `controlleddocuments/` | `git grep -nE '(FROM\|JOIN)\s+(public.\|metaldocs.)?(documents\|document_revisions\|approval_instances)\b' internal/modules/controlleddocuments` shows none (non-test) | real |
| 0 bare foreign status literals in the ported query | adapter SQL references `documents/domain.DocStatus*` + `approval/domain.InstanceInProgress` (params), not `'draft'`/`'published'`/… | real (review) |

> TDD: write the failing parity test first, then implement the port to green, then delete the raw
> reads (parity green **before** deletion — D6). PG unavailable ⇒ mark integration steps **not-run
> (HS-3)**, never false-green.

## ADR needed?

- [x] **No new ADR.** Mechanical application of **ADR-0039** D1 + D3(b) (cross-module base-table read →
  owner-published read-port). Port shape (single projection, off-tx, owner-typed status constants)
  recorded in this spec + `evidence.md`. No parallel/novel cross-module contract introduced.
