# Concept: Controlled Documents

> **Last verified:** 2026-06-09 (std-execution Family 5 query-param snake_case: preview-code `?profile_code=<code>&area_code=<code>`; prior: 2026-06-08 Phase E1: request body and response fields updated to snake_case; prior: 2026-05-21)
> **Scope:** What a Controlled Document (CD) is, code generation, profile + area binding, sequence counters, atomic create endpoint, preview endpoint.
> **Key files:**
> - `internal/modules/controlleddocuments/` — controlled-documents module; hosts the atomic create handler
> - `internal/platform/idempotency/` — generic `Store` + middleware; used by atomic create + revision endpoints
> - `internal/modules/controlleddocuments/infrastructure/repository.go` — per-(tenant, profile_code, process_area_code) counter and CD CRUD queries
> - `frontend/apps/web/src/features/controlled-documents/pages/ControlledDocumentsPage.tsx` — controlled-documents list

## What it is

A **Controlled Document (CD)** is a unique catalog slot — a code-numbered controlled-document identity. It binds:

- A **profile** (Tipo Documental).
- An **area**.
- A **sequence number** scoped to that (profile, area) pair.

The CD itself is a slot. The actual editable content lives in **document versions** that hang off the CD.

## Code format

`{profile-code}-{area-code}-{sequence-padded}`

Examples:

- `DC-RH-001` — first Descrição de Cargo in RH.
- `DC-RH-002` — second Descrição de Cargo in RH.
- `DC-QUA-001` — first Descrição de Cargo in Qualidade.
- `POP-PROD-014` — fourteenth Procedimento Operacional in Produção.

Sequence pads to **3 digits** with leading zeros (e.g. `001`, `014`).

## Sequence rules

- One counter per `(tenant_id, profile_code, process_area_code)` triple, stored in `cd_sequence_counters`.
- Monotonic — never reused even if a CD is archived.
- Resets: never.

## Lifecycle of a CD

1. Created atomically via `POST /api/v2/controlled-documents` — code generated and first document revision cloned from profile template in a single DB transaction.
2. Future revisions created via `POST /api/v2/controlled-documents/{id}/revisions`.

The CD itself doesn't have an approval state — its document revisions do. The CD just owns the code and the revision history.

## Atomic Create Endpoint

`POST /api/v2/controlled-documents`

**Required header:** `Idempotency-Key: <uuid>`

**Request body:**

| Field | Type | Notes |
|-------|------|-------|
| `profile_code` | string | Must match an existing profile |
| `process_area_code` | string | Must match an existing process area |
| `name` | string | Document name |
| `template_version_id` | UUID | Published template version to clone |
| `owner_user_id` | UUID | Assigned owner |

**Behavior:** Creates the `controlled_documents` row, increments the `cd_sequence_counters` counter, inserts the first `documents` revision (draft), all within a single DB transaction. `storage_key` on the first revision starts empty — the editor renders it on demand.

**Response:** `201 Created` with the CD object (including the server-resolved `code` field). Replay of the same `Idempotency-Key` returns the stored 201 response from `metaldocs.idempotency_keys`.

**Legacy deleted:** `RegistryCreateDialog` and `DocumentCreatePage` were removed. The wizard at `/documents/new` now calls this endpoint directly.

## Preview Endpoint

`GET /api/v2/controlled-documents/preview-code?profile_code=<code>&area_code=<code>`

Returns the **next** code that would be assigned for the given (profile, area) pair if a document were created now — read-only, no sequence reservation. The actual code is assigned at create time and may differ if another document is created concurrently.

**Response:** `200 OK` with `{ "code": "DC-RH-003", "profile_code": "DC", "area_code": "RH", "next_seq": 3 }` (see `PreviewCodeResponse` at `internal/modules/controlleddocuments/api/api.gen.go:207`).

## Revision Endpoint

`POST /api/v2/controlled-documents/{id}/revisions`

**Required header:** `Idempotency-Key: <uuid>`

Creates a new document revision on an existing CD. Requires idempotency key (same replay semantics as atomic create).

---

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 5
- [modules/taxonomy.md](../modules/taxonomy.md)
- [modules/documents.md](../modules/documents.md)
