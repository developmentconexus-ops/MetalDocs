# ADR 0008 — Soft-archive documents via archived_at timestamp

> Status: accepted 2026-05-03

## Context

Migration 0142 enforces a strict status-transition trigger
(`enforce_document_transition`). It defines no transition into `archived`.
The original `Service.Archive` attempted `UpdateDocumentStatus(... → archived)`
and was rejected at runtime (audit C1).

QMS regulatory requirement: a controlled document's terminal status
(`published`, `superseded`, `obsolete`) must remain visible in the audit
trail unchanged. Replacing it with `archived` discards evidence of the
document's lifecycle outcome.

## Decision

Archive is a soft-hide via `documents.archived_at` timestamp. Status field
is never changed by archive. Default list/search queries filter
`archived_at IS NULL`. Admin endpoints opt in to include archived.

`Service.Archive(tenantID, docID, actorID)` — no `fromFinalized` parameter.
Symmetric `Unarchive` clears the timestamp.

`finalized_at` is **not** retained as a denormalized column (see C6) —
finalization timestamp derives from `document_state_history` via the
`v_document_finalized` view. Different from `archived_at` because
`archived_at` is a hot-path filter predicate (queried per list) while
`finalized_at` is cold-path audit data.

## Consequences

- No trigger 0142 change required
- Status field remains source of truth for lifecycle outcome
- `archived_at IS NULL` predicate added to default queries
- Two readers of "is archived?" (status vs timestamp) collapsed to one
- Frontend list views unchanged (default filter applied at repo level)

## References

- Spec: `docs/superpowers/specs/2026-05-03-group-c-doc-lifecycle-design.md`
- Audit: `wiki/bugs/audit-2026-05-03.md` (C1)
- Trigger: `migrations/0142_disable_legacy_compat.sql`
