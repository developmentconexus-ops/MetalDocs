# Phase 2 — Flow trace: GetActiveDocument

`GET /api/v1/controlled-documents/{id}/active-document`

### 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| HTTP mux registration | `mux.HandleFunc("GET /api/v1/controlled-documents/{id}/active-document", generated.GetActiveDocument)` | `internal/modules/registry/delivery/http/handler.go:86` |
| Generated wrapper | `(*ServerInterfaceWrapper).GetActiveDocument` | `internal/modules/registry/api/api.gen.go:992` |
| Wrapper dispatch | `siw.Handler.GetActiveDocument(w, r, id)` | `internal/modules/registry/api/api.gen.go:1007` |
| Concrete handler | `(*Handler).GetActiveDocument` | `internal/modules/registry/delivery/http/routes.go:205` |

### 2. Call chain (6 layers)

1. `ServeMux` matches `GET /api/v1/controlled-documents/{id}/active-document` → `generated.GetActiveDocument` (`handler.go:86`). No idempotency middleware on GET.
2. Generated wrapper enters at `api.gen.go:992`.
3. Path param bind/validate for `{id}` into `openapi_types.UUID` (`api.gen.go:1000`).
4. Wrapper invokes `siw.Handler.GetActiveDocument(w, r, id)` (`api.gen.go:1007`).
5. Concrete handler executes FULL OUTER JOIN SQL using `documents.status` as the active-state truth, then performs a secondary approval-instance lookup only when the active row is `under_review`; maps row into `controlleddocumentsapi.ActiveDocumentResponse` (`routes.go:238`, `:255`, `:309`).
6. Response: `httpresponse.WriteJSON(w, 200, resp)` (`routes.go:331`) or `httpresponse.WriteError(...)` (legacy envelope at `internal/platform/httpresponse/response.go:14-15`).

**Tenant resolution:** request context via `tenant.FromContext` inside `tenantIDFromRequest`; no route-local authz GUC (`routes.go:240-244`, `handler.go:50-53`).

### 3. State changes — none

Read-only path; no INSERT/UPDATE/DELETE.

### 4. SQL touched

**Main FULL OUTER JOIN query** (`internal/modules/registry/delivery/http/routes.go:255-273`) — verbatim:

```sql
SELECT active.id,
       COALESCE(active.content_hash_at_submit,
                (SELECT r.content_hash FROM document_revisions r
                  WHERE r.document_id = active.id
                  ORDER BY r.created_at DESC LIMIT 1)),
       active.revision_version,
       active.status,
       pub.id::text
  FROM (SELECT id, content_hash_at_submit, revision_version, status
          FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status IN ('draft','under_review','approved','rejected','scheduled')
         LIMIT 1) active
  FULL OUTER JOIN
       (SELECT id FROM documents
         WHERE tenant_id = $1::uuid
           AND controlled_document_id = $2::uuid
           AND status = 'published'
         ORDER BY revision_number DESC
         LIMIT 1) pub ON TRUE
```

**Secondary approval-instance query** (`routes.go:309-316`) — executed only when the active row status is `under_review`:

```sql
SELECT id::text
 FROM approval_instances
 WHERE document_id = $1::uuid
   AND tenant_id = $2::uuid
   AND status = 'in_progress'
 ORDER BY submitted_at DESC
 LIMIT 1
```

If this secondary lookup returns `sql.ErrNoRows`, the response simply omits `approvalInstanceId`. Any other query failure is treated as `500 INTERNAL_ERROR` because the route would otherwise misrepresent review state.

**Tripwire pairing audit:** VIOLATION.
- No `metaldocs.assert_caps(...)` reference in this route SQL.
- No `current_setting('metaldocs.tenant_id')` — tenant scoped only via `$1::uuid` query arg sourced from request header.
- No `authz.Require` call in handler.

### 5. Response shape

`controlleddocumentsapi.ActiveDocumentResponse` (generated at `api.gen.go:265-272`) — all fields optional:

```go
type activeDocumentResponse struct {
    DocumentID          *string `json:"documentId,omitempty"`
    ApprovalState       *string `json:"approvalState,omitempty"`
    ContentHash         *string `json:"contentHash,omitempty"`
    RevisionVersion     *int    `json:"revisionVersion,omitempty"`
    PublishedDocumentID *string `json:"publishedDocumentId,omitempty"`
    ApprovalInstanceID  *string `json:"approvalInstanceId,omitempty"`
}
```

200 JSON write at `routes.go:331`. 404 paths:
- both FULL OUTER JOIN sides NULL → `if !docID.Valid && !publishedDocID.Valid { 404 }` (`routes.go:291-294`)
- `sql.ErrNoRows` short-circuit (`routes.go:282-284`)

**Envelope:** legacy `{"code","message"}` (`response.go:14-15`). RFC 9457 not used.

### 6. Cross-references

- **Idempotency:** n/a — GET. Middleware applied only to POST create/revision (`handler.go:79-82`).
- **Pagination:** n/a.
- **Audit emission:** no — handler does DB read + HTTP write only; no governance logger call (`routes.go:205-326`).

