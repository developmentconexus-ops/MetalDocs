# ADR 0009 — Atomic CD Create + Per-Area Numbering + Idempotency-Key Adoption

> **Status:** Accepted
> **Date:** 2026-05-07
> **Last verified:** 2026-05-07
> **Scope:** How controlled documents are created, numbered, and made idempotent.
> **Out of scope:** System-wide idempotency rollout (future ADR), visibility enforcement (see `backlog/novo-documento.md#visibility`).
> **Key files:**
> - `internal/modules/registry/` — atomic create handler; owns the `POST /api/v2/controlled-documents` route
> - `internal/platform/idempotency/` — generic `Store` + HTTP middleware; consumed by create + revision routes
> - `internal/modules/registry/repository/cd_sequence_counters.go` — per-(tenant, profile_code, process_area_code) counter
> - `internal/modules/documents/` — exposes `CreateDocumentTx(*sql.Tx, ...)` port consumed by registry at create time

## Context

Three separate problems converged:

1. **Spec drift on numbering.** The original `AutoCode` implementation produced 2-segment, 2-digit codes (`DC-01`). The wiki spec and domain model called for 3-segment, 3-digit zero-padded codes (`DC-RH-001`) keyed on (profile, area). The old `profile_sequence_counters` table had no `process_area_code` column, so all areas shared one counter per profile.

2. **Orphan slot risk.** The "Novo Documento Controlado" wizard used two sequential HTTP calls: `POST /api/v2/controlled-documents` to reserve a slot, then `POST /api/v2/documents` to clone the template. A network or server error between the two calls consumed a sequence number and left an orphan CD slot in the registry with no document attached — requiring manual `system_admin` cleanup.

3. **No idempotency on create.** Retrying a failed create (e.g., after a timeout) could produce duplicate CD entries. The approval module already had a hand-rolled idempotency store (`PostgresSignoffIdempStore`); no shared platform existed.

## Decision

Deliver numbering fix, atomic create, and idempotency in a single PR so the three concerns ship together and the legacy two-call path is never left in a half-migrated state.

**Numbering:** Replace `profile_sequence_counters` with `cd_sequence_counters` keyed on `(tenant_id, profile_code, process_area_code)`. Pad sequence to 3 digits. Code format is now `{PROFILE}-{AREA}-{NNN}` (e.g. `DC-RH-001`).

**Atomic create:** `POST /api/v2/controlled-documents` now inserts the CD row, increments the sequence counter, and inserts the first `documents` revision (draft, `storage_key` empty) inside a single `*sql.Tx`. The documents module exposes a `CreateDocumentTx` port that accepts the caller's transaction; registry calls it rather than issuing a second HTTP request. `POST /api/v2/documents` (create from CD) is deleted. `RegistryCreateDialog` and `DocumentCreatePage` are deleted.

**Idempotency:** `internal/platform/idempotency/` provides a generic `Store` (backed by `metaldocs.idempotency_keys`) and HTTP middleware. Wired on two routes for now: `POST /api/v2/controlled-documents` and `POST /api/v2/controlled-documents/{id}/revisions`. The signoff store is refactored to use the shared platform. System-wide rollout deferred to a future ADR.

**Preview endpoint:** `GET /api/v2/controlled-documents/preview-code?profileCode=&areaCode=` returns the next sequence preview read-only (no reservation), closing the `backlog/novo-documento.md#sequence-preview` deferral.

## Consequences

**Positive:**
- Numbering integrity: codes are now spec-compliant and area-isolated; no cross-area counter bleed.
- No orphan slots: CD + first revision either both commit or both roll back.
- Clean REST model: document revisions are a sub-resource of the controlled document (`/controlled-documents/{id}/revisions`), not a sibling resource.
- Shared idempotency platform replaces ad-hoc store in signoff module.

**Negative:**
- `storage_key` on the first revision starts empty; the editor must handle a missing DOCX gracefully (render blank on demand).
- The documents module must expose a `CreateDocumentTx` port that crosses the module boundary — registry now depends on an internal documents port. This is intentional (owned ports, no circular imports) but must be enforced at code review.
- Idempotency coverage is limited to two routes. Other mutating endpoints remain non-idempotent until a follow-up ADR wires the middleware more broadly.

## References

- `wiki/backlog/novo-documento.md#slot-rollback` — closed
- `wiki/backlog/novo-documento.md#sequence-preview` — closed
- `wiki/concepts/controlled-documents.md` — updated endpoint docs
- `wiki/modules/documents.md` — atomic create flow, deleted `POST /api/v2/documents`
- `wiki/modules/approval.md` — signoff idempotency store (now uses shared platform)
