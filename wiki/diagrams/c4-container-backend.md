# C4 Level 2 — Container View (Backend)

> **Last verified:** 2026-06-01 (async freeze refactor — ADR 0015)
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
        Container(worker, "metaldocs-worker", "Go", "Async job runner. Consumes the outbox; handles PDF conversion, docx materialization (ADR 0015), scheduled publish, review reminders.")
        Container(docgen, "docx-renderer", "Node 20 + Fastify + eigenpal/docx-js-editor (headless)", "Server-side docx render. Routes: /render/fanout (reconstructs frozen docx), /health.")
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
    Rel(worker, pg, "Claim outbox rows, write artifact metadata")
    Rel(worker, minio, "Read docx, write PDF")
    Rel(worker, gotenberg, "POST /forms/libreoffice/convert (docx → PDF)")
    Rel(worker, docgen, "POST /render/fanout (async, via materialize outbox — ADR 0015)")

    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="2")
```

## Why each container exists

| Container | Job | Why separate |
|---|---|---|
| **metaldocs-api** | All business logic + authz | Single trust boundary; horizontally scalable; stateless behind a load balancer |
| **metaldocs-worker** | Async side-effects (PDF, DOCX materialization, scheduled jobs) | Decouples slow/external work from user-facing API. Outbox-driven, at-least-once with retries (ADR 0009, ADR 0015) |
| **metaldocs-web** | UI | Vite SPA. No SSR. Editor is heavy (eigenpal), better as a client app |
| **docx-renderer** | Server-side eigenpal render | Eigenpal's docx engine is JavaScript-only. Wrapped behind a stable HTTP contract so the Go side can swap providers later (SuperDoc, etc.) without touching backend logic — anti-corruption layer |
| **postgres** | All persistent state | Single source of truth for facts; transactional outbox lives here too |
| **minio** | Document bytes | API never proxies multi-MB docx — browser ↔ MinIO direct via presigned URLs. The scaling pattern |
| **redis** | Authz cache + rate limit | Hot-path lookups that would otherwise hammer Postgres |
| **gotenberg** | docx → PDF | LibreOffice in a container; stateless; replaceable |

## Coupling notes

- **Worker → docx-renderer is async + retryable** via `materialize_dispatch_outbox` (ADR 0015). The API no longer calls docx-renderer during the signoff transaction — approval availability is independent of docx-renderer uptime.
- **Worker → Gotenberg is async + retryable** (PDF outbox, ADR 0009). Materialize chains into this: after the frozen docx is written, the `pdf_dispatch_outbox` row is enqueued and processed by the existing PDF path.

For external context, see [c4-context.md](c4-context.md). For end-to-end flows, see the sequence diagrams in this folder.
