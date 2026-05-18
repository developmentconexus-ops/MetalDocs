# Distribuição & Cobertura — Deferred Items

> **Last verified:** 2026-05-17
> **Scope:** Deferred items for the Distribuição & Cobertura de Leitura screen (`/documents/:documentId/distribution`).
> **Out of scope:** Published view deferred items (see `backlog/documento-publicado.md`), fanout render pipeline (see `modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx:32` — page entry point
> - `frontend/apps/web/src/features/documents/lib/distributionMeta.ts:71` — `MOCK_DISTRIBUTION` — replace with real query hooks when endpoints exist

---

## Backend endpoints (no implementation yet)

| Endpoint | Purpose | Priority |
|---|---|---|
| `GET /api/v1/documents/:id/distribution` | KPI metrics (totalTargets, acknowledged, read, pending, overdue) + DistributionFacts sidebar data (publishedAt, deadline, policy, channel, reminders, groups) | High |
| `GET /api/v1/documents/:id/distribution/recipients` | Paginated recipient table with status, last activity, timestamps | High |
| `GET /api/v1/documents/:id/distribution/coverage` | By-area breakdown (area name, total, read, ack counts) | Medium |
| `GET /api/v1/documents/:id/distribution/timeline` | Cumulative daily read events for timeline chart | Low |

Wire via TanStack Query hooks in `features/documents/queries/` when endpoints exist.
Replace mock data from `features/documents/lib/distributionMeta.ts` — remove mock constants, keep type interfaces.

## Fanout module (all CTAs blocked on this)

The 4 hero CTAs and all recipient bulk/row actions require a fanout module:

| CTA | Required endpoint | Status |
|---|---|---|
| Lembrete em massa | `POST /api/v1/documents/:id/fanout/reminders/bulk` | No module |
| Exportar relatório | `GET /api/v1/documents/:id/distribution/export` | No module |
| Adicionar destinatários | `POST /api/v1/documents/:id/fanout/recipients` | No module |
| Política de fanout | Settings modal (no design) | No design |
| Row action — mail | `POST /api/v1/documents/:id/fanout/reminders/:recipientId` | No module |
| Bulk action bar | Bulk send reminder, reassign, export | No module |

## Chart library

DonutCard (SVG animated donut) and TimelineCard (SVG line chart) are currently hand-rolled SVG with mock data.
Research chart libraries for replacement: candidates include `recharts`, `victory`, `visx`.
Criteria: tree-shakeable, no forced Canvas, accessible, compatible with CSS token system.

## Pagination

RecipientsCard shows static "1 / 23" footer.
Real pagination requires `GET /api/v1/documents/:id/distribution/recipients?page=&pageSize=` with `Link`/`X-Total-Count` response headers (matching the library pagination pattern).
Wire Pagination primitive (`features/documents/components/Pagination.tsx`) when endpoint exists.
