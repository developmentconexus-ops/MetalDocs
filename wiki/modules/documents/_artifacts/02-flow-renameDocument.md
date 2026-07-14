# Data-flow: renameDocument (PATCH /api/v1/documents/{id})

Last verified: 2026-05-10

---

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | No `operationId` in `api/openapi/v1/openapi.yaml`; route is registered directly in the Go HTTP mux, not via OpenAPI codegen. | (unclear: operationId absent — PATCH /api/v1/documents/{id} is not in the spec) |
| Generated server stub | None — no `api.gen.go` involvement; route bypasses codegen entirely. | (unclear: N/A — no generated stub) |
| Handler | `Handler.renameDocument` | `internal/modules/documents/delivery/http/handler.go:285` |

Route registration: `internal/modules/documents/delivery/http/handler.go:86` and `:115` (duplicate registration — both lines: `mux.HandleFunc("PATCH /api/v1/documents/{id}", h.renameDocument)`).

---

## 2. Call chain

```
HTTP PATCH /api/v1/documents/{id}
  → Handler.renameDocument                      handler.go:285
      → h.authorizeDocumentScope                handler.go:869
          Tier-1 gate: role check (not cap string)
          - requires roleAdmin OR roleDocumentFiller  handler.go:870
          - if roleDocumentFiller: also checks IsDocumentOwner  handler.go:880
          No authz.Require / capability string here (role-based, not cap-based).
      → json.Decode req.Name                    handler.go:296 (400 on failure)
      → h.svc.RenameDocument(ctx, tenantID, userID, docID, req.Name)  handler.go:301
          Service: Service.RenameDocument       application/service.go:564
            - validates name: non-empty, ≤255 chars  service.go:565–567
            - s.repo.GetDocument(ctx, tenantID, docID)  service.go:569
            - guards: doc.Status must == domain.DocStatusDraft  service.go:572 (ErrInvalidStateTransition)
            - No Tier-2 authz.Require inside tx.
            - Transaction boundary: NONE — UpdateDocumentName is a plain ExecContext, no explicit tx.
            - s.repo.UpdateDocumentName(ctx, tenantID, docID, name)  service.go:575
                Repository: Repository.UpdateDocumentName  repository/repository.go:216
                  SQL: UPDATE documents SET name=$2, updated_at=now() WHERE id=$1 AND tenant_id=$3
            - s.audit.Write(ctx, tenantID, userID, "document.renamed", docID, {"name": name})
                  service.go:579  ← OUTSIDE any tx (no tx wraps this call)
      → h.svc.GetDocument(ctx, tenantID, docID)  handler.go:305
      → httpresponse.WriteJSON(w, 200, doc)      handler.go:313
```

**Tier-1 capability check:** no `authz.Require` call — gate is role-based (`roleAdmin` / `roleDocumentFiller` + ownership). Capability string: none.

**Tier-2 authz.Require inside tx:** absent.

**Transaction boundary:** none. `UpdateDocumentName` and the subsequent `audit.Write` run as independent DB calls.

**Audit emission:** `s.audit.Write` at `service.go:579`, outside any tx, sink: `Audit` interface (`application/service.go:81–82`); action string: `"document.renamed"`.

---

## 3. State changes

`documents.status` is unchanged — only `documents.name` and `documents.updated_at` are written.
The service pre-checks that `status == DocStatusDraft` before writing; no status column is mutated.

| Column mutated | Table | Condition |
|---|---|---|
| `name` | `documents` | doc must be in `draft` status |
| `updated_at` | `documents` | always (via `now()` in SQL) |

---

## 4. SQL touched

| File:line | Verb | Tables | Auth-area arg |
|---|---|---|---|
| `repository/repository.go:218` | UPDATE | `documents` | None — no `metaldocs.asserted_caps` SET, no `authz.Require` paired. |

**Tripwire pairing:**

- `enforce_capability_asserted` trigger presence on `documents`: no SQL file in the migrations set contains this trigger on the `documents` table. The `metaldocs.asserted_caps` pattern appears only in test files (`cancel_service_test.go`, `coverage_boost_test.go`, etc.) as mock-query detection strings, not as a live migration.
- **Result: N/A** — no tripwire trigger is attached to `documents`; the UPDATE at `repository.go:218` carries no `SET LOCAL metaldocs.asserted_caps` and no `authz.Require` guard.

Anchors: `repository/repository.go:218` (UPDATE) · no migration anchor found for `enforce_capability_asserted` on `documents`.

---

## 5. Response shape

**2xx (200 OK):** `*domain.Document` serialized as JSON via `httpresponse.WriteJSON`.
The struct is returned by `h.svc.GetDocument` (re-fetch after rename) at `handler.go:305–313`. No dedicated `DocumentResponse` DTO — raw `domain.Document` is written directly.

**Error responses from this handler path:**

| HTTP code | Condition |
|---|---|
| 400 Bad Request (`invalid_body`) | JSON decode failure |
| 400 Bad Request (`invalid_name`) | name empty or > 255 chars (`domain.ErrInvalidName`) |
| 403 Forbidden (`forbidden`) | caller lacks roleAdmin / roleDocumentFiller, or not document owner |
| 404 Not Found (`not_found`) | document not found in repo (`domain.ErrNotFound`) |
| 409 Conflict (`invalid_state_transition`) | document is not in `draft` status (`domain.ErrInvalidStateTransition`) |

---

## 6. Cross-references

- **Idempotency:** No — PATCH, name field is overwritten on each call; repeated calls with the same name are safe but not idempotent in the HTTP sense (each call re-emits an audit event).
- **Pagination:** N/A.
- **Audit log emission:** Yes — `s.audit.Write` at `application/service.go:579`, action `"document.renamed"`, outside transaction, sink: `Audit` interface defined at `application/service.go:81`.
- **Related handler duplicate:** Route `PATCH /api/v1/documents/{id}` registered at both `handler.go:86` and `handler.go:115` — the second registration overwrites the first in the stdlib mux (net/http last-wins semantics). Both point to the same `h.renameDocument` func so there is no behavioral difference, but the duplicate is a latent maintenance hazard.
