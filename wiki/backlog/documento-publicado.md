# Backlog: Documento Publicado screen (`/documents/:id`)

> Last updated: 2026-05-08

## Deferred features

### AuditCard / ISO seal — values_hash not in API

`GET /api/v2/documents/:id` does not return `values_hash`. The AuditCard showing the SHA-256 integrity seal requires this field. Backend fix: add `values_hash` to the document SELECT in `internal/modules/documents/delivery/http/handler.go` (`getDocumentByID` query).

Backlog: add `values_hash` to backend response, then implement `AuditCard` + `ISOSeal` components.

---

### CommentsCard — needs architecture brainstorm

Comments backend exists (`GET/POST /api/v2/documents/:id/comments`). Content field is ProseMirror JSON (`unknown[]`). Before implementing the CommentsCard, the team needs to decide:
1. Storage structure: how comments relate to document sections/anchors vs free-form
2. Rendering: `extractPlainText` util vs eigenpal `ReadonlyEditor`
3. Reply threading model: `parent_library_id` is present but UX not designed

---

### PDF download

No PDF generation endpoint exists today. `GET /api/v2/documents/:id/pdf` or similar not implemented.

---

### Coverage card (KPI: Cobertura %)

Requires fanout read coverage API. No endpoint today.

---

### VersionTimeline

Requires revision list endpoint. Only `GET /api/v2/documents/:id/revisions/:rid/url` exists (single revision URL). No list endpoint.

---

### RelatedGrid

No related-documents relationship model in the backend.
