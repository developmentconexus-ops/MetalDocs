# C4 Level 2 — Container View (Backend)

> **Last verified:** 2026-06-01 (docgen-v2 → docx-renderer rename)
> **Scope:** All runtime processes + their immediate dependencies.
> **Source of truth for:** [`wiki/architecture/system-overview.md`](../architecture/system-overview.md).
> **Code anchors:**
> - [`deploy/compose/docker-compose.yml`](../../deploy/compose/docker-compose.yml) — service topology
> - [`apps/api/cmd/metaldocs-api/main.go`](../../apps/api/cmd/metaldocs-api/main.go) — API entrypoint
> - [`apps/worker/cmd/metaldocs-worker/main.go`](../../apps/worker/cmd/metaldocs-worker/main.go) — worker entrypoint
> - [`internal/platform/bootstrap/`](../../internal/platform/bootstrap/) — DI wiring

```mermaid
C4Container
    title MetalDocs — Backend Container View

    Person(user, "User", "Author / approver / admin / reader")

    System_Boundary(md, "MetalDocs") {
        Container(web, "metaldocs-web", "React 18 + Vite + TanStack Query", "SPA. Renders editor (eigenpal), inbox, admin. Talks to API via HTTPS + cookie session. Uploads/downloads docx directly to MinIO via presigned URLs.")
        Container(api, "metaldocs-api", "Go 1.22 (net/http)", "Authoritative business logic. REST API on :8081. Modules: auth, iam, templates, documents, approval, controlleddocuments, taxonomy, render/fanout, search, audit.")
        Container(worker, "metaldocs-worker", "Go", "Async job runner. Consumes the outbox; handles PDF conversion, scheduled publish, review reminders.")
        Container(docgen, "docx-renderer", "Node 20 + Fastify + eigenpal/docx-js-editor (headless)", "Server-side docx render. Only live route today: /render/fanout (reconstructs frozen docx from snapshot + values during approval freeze).")
        ContainerDb(pg, "postgres", "Postgres 16", "Primary datastore: tenants, users, roles, templates, documents, revisions, approval_instances, approval_signoffs, governance_events, outbox tables, audit_log.")
        ContainerDb(minio, "minio", "S3-compatible object store", "Bucket holds: template body docx, document revisions, frozen final docx, generated PDFs. Browser uploads/downloads via presigned URLs.")
        Container(redis, "redis", "Redis 7", "Authz cache (TTL) + rate-limit counters.")
        Container(gotenberg, "gotenberg", "Gotenberg 8 / LibreOffice", "Pure docx → PDF converter. No state.")
    }

    Rel(user, web, "HTTPS + cookie session")
    Rel(web, api, "REST /api/v1/*")
    Rel(web, minio, "Presigned PUT/GET (autosave, view, export download)")
    Rel(api, pg, "SQL (database/sql + pq)")
    Rel(api, redis, "Authz cache + rate limit")
    Rel(api, minio, "Presign URLs, head/size objects")
    Rel(api, docgen, "POST /render/fanout (inside signoff tx — see sequence-signoff-freeze)")
    Rel(worker, pg, "Claim outbox rows, write artifact metadata")
    Rel(worker, minio, "Read docx, write PDF")
    Rel(worker, gotenberg, "POST /forms/libreoffice/convert (docx → PDF)")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

## Why each container exists

| Container | Job | Why separate |
|---|---|---|
| **metaldocs-api** | All business logic + authz | Single trust boundary; horizontally scalable; stateless behind a load balancer |
| **metaldocs-worker** | Async side-effects (PDF, scheduled jobs) | Decouples slow/external work from user-facing API. Outbox-driven, at-least-once with retries (ADR 0009) |
| **metaldocs-web** | UI | Vite SPA. No SSR. Editor is heavy (eigenpal), better as a client app |
| **docx-renderer** | Server-side eigenpal render | Eigenpal's docx engine is JavaScript-only. Wrapped behind a stable HTTP contract so the Go side can swap providers later (SuperDoc, etc.) without touching backend logic — anti-corruption layer |
| **postgres** | All persistent state | Single source of truth for facts; transactional outbox lives here too |
| **minio** | Document bytes | API never proxies multi-MB docx — browser ↔ MinIO direct via presigned URLs. The scaling pattern |
| **redis** | Authz cache + rate limit | Hot-path lookups that would otherwise hammer Postgres |
| **gotenberg** | docx → PDF | LibreOffice in a container; stateless; replaceable |

## Known coupling worth knowing

- **API → docx-renderer is synchronous during signoff** (the `/render/fanout` call inside the approval transaction). This is the strongest coupling in the system today. See [sequence-signoff-freeze.md](sequence-signoff-freeze.md) for why this matters and the planned async refactor.
- **Worker → Gotenberg is async + retryable** (PDF outbox, ADR 0009). This is the pattern freeze should eventually adopt.

For external context, see [c4-context.md](c4-context.md). For end-to-end flows, see the sequence diagrams in this folder.
