# F4 — Delete the finalize chain from Go (keep domain sentinels)

> **Milestone:** M1 · **Findings:** 4, 5 · **Status:** spec approved (ADR 0073-ratified)
> **Approved:** 2026-07-06 — from ADR 0073 §1 ("remove the /finalize wrapper"). No operator
> interview needed: the removal is the ADR's explicit decision; contract is the ADR itself.

## Consumer contract

There is exactly **one** submit entrypoint. Every consumer that previously called
`/documents/{id}/finalize` (or the Go `finalizeDocument` handler / `GetFinalizePrereqs` off-tx
reader) now calls the canonical `/submit`, which owns in-tx resolution (F1). The wrapper's Go
surface is **deleted**, not deprecated-in-place:

- No `/finalize` route in the OpenAPI spec, router, or generated `ServerInterface`.
- No `finalizeDocument` handler, no `GetFinalizePrereqs` production method, no finalize-only
  request/response contract types.
- The **domain sentinels** the wrapper surfaced (`ErrProfileNotConfigured`, `ErrApprovalRouteMissing`,
  `ErrDocumentNotDraft`, `ErrRevisionTitleRequired`) are **retained** — they are now returned by the
  in-tx submit path (F1) and mapped by F3. Deleting them would break submit's error contract.
- Historical **comments** may reference the wrapper for provenance (they cite `repository.go:1801-1817`
  as the origin of the moved query); these are documentation, not live code.

**Source of truth:** ADR 0073 §1; the OpenAPI `/submit` description (which records the removal).

## Non-goals
- No removal of the unrelated `FreezeFinalizer` / render `MarkFailed(finalize=true)` subsystems
  (different concept — content-freeze + outbox dead-letter).
- No removal of the legacy-state rejection (`"finalized"`/`"archived"` parse-boundary guard) — that
  is an invariant guard, not the wrapper.
- No migration/DB change (the `document.finalize` capability row cleanup is out of M1 scope; the
  submit path never checks it).

## Validation Gate
- `grep -ri 'func.*finalize|GetFinalizePrereqs|/finalize|finalizeDocument'` under
  `internal/modules/documents/approval` returns **only comments**, zero live symbols.
- OpenAPI `api/openapi/v1/openapi.yaml` has **no** `/finalize` path (only a deprecation note in the
  `/submit` description).
- `go build ./...` + `go test ./...` green (nothing references the deleted symbols).
- The four retained sentinels still exist and are exercised by F1 integration + F3 mapping tests.
