# Sequence — Edit + Autosave

> **Last verified:** 2026-06-01
> **Flow:** Eigenpal editor in the browser autosaves every ~1.5s. Bytes go **directly** from browser to MinIO via presigned PUT URLs; the API only signs URLs and records hashes.
> **Why it matters:** This is the scaling pattern. The API never proxies multi-MB docx bytes — 100 concurrent users hitting Save burden MinIO, not the API.
> **Code anchors:**
> - [`packages/editor-ui/src/MetalDocsEditor.tsx`](../../packages/editor-ui/src/MetalDocsEditor.tsx) — the editor wrapper (debounced autosave)
> - [`frontend/apps/web/src/features/documents/hooks/editor/useDocumentAutosave.ts`](../../frontend/apps/web/src/features/documents/hooks/editor/useDocumentAutosave.ts) — autosave hook
> - [`internal/modules/documents/delivery/http/handler.go:117`](../../internal/modules/documents/delivery/http/handler.go) — `presignAutosave` + `commitAutosave` routes
> - [`internal/platform/objectstore/document_presigner.go`](../../internal/platform/objectstore/document_presigner.go) — `PresignRevisionPUT`, scoped to one exact key + TTL + size cap

```mermaid
sequenceDiagram
    autonumber
    actor User as User (author)
    participant Editor as Eigenpal editor (in browser)
    participant API as metaldocs-api
    participant Minio as MinIO

    Note over Editor: User types — autosave debounce 1.5s
    Editor->>Editor: save() → docx bytes + sha256
    Editor->>API: POST /api/v1/documents/{id}/autosave/presign {content_hash}
    activate API
    Note right of API: Server scopes the URL to:<br/>tenants/{tenant}/documents/{doc}/revisions/{hash}.docx<br/>+ TTL + max size + per-user rate limit
    API->>Minio: generate presigned PUT URL (one key, short-lived)
    Minio-->>API: signed URL
    API-->>Editor: {upload_url, pending_upload_id}
    deactivate API

    Editor->>Minio: PUT docx bytes directly (Content-Type: docx)
    Note right of Minio: Bytes never pass through the API.<br/>API is unaware of size or content yet.
    Minio-->>Editor: 200 OK

    Editor->>API: POST /api/v1/documents/{id}/autosave/commit {pending_upload_id}
    activate API
    API->>Minio: HEAD object to confirm upload
    API->>Minio: GET object → re-compute sha256 (server-side hash; client can't lie)
    API->>API: validate content_hash matches the presigned key
    API->>API: write a new revisions row (idempotent on hash)
    API-->>Editor: 200 {revision_id, content_hash}
    deactivate API
```

## Why this is the scalable shape

- **Bytes flow browser ↔ MinIO directly.** The API touches a small JSON envelope only.
- **Hash is server-authoritative.** The server re-derives sha256 from the object on commit — clients can't claim a hash they didn't actually upload.
- **Idempotent.** A second autosave with the same content_hash is a no-op revision (same key).
- **Bounded.** Presigned URL is scoped to *one specific key*, max size, short TTL. A stolen URL can't be reused for anything else.

## Failure modes

| Failure | Outcome |
|---|---|
| MinIO PUT fails | Editor retries on next debounce; nothing is committed; no revision row created |
| Commit before PUT completes | API HEAD returns 404 → `pending_not_found` error → editor retries |
| Hash mismatch (rare; bytes tampered in flight) | API rejects commit with `content_hash_mismatch` |
| API down between presign and commit | The uploaded blob is orphaned; a cleanup job (future) reaps unreferenced presigned keys |

## What this is NOT (deliberate v1 scope)

- **Not Google-Docs-style real-time co-editing.** Single editor per document; last-writer-wins with optimistic concurrency (If-Match on revision_version). Real-time co-edit would require CRDT/OT — different architecture, out of v1.

## Related

- [c4-container-backend.md](c4-container-backend.md) — see the `web ↔ minio` arrow.
- [sequence-create-document.md](sequence-create-document.md) — how the editor gets to this point.
- [`wiki/architecture/attachment-signing.md`](../architecture/attachment-signing.md) — presigning details.
