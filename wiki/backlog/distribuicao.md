# Distribuição & Cobertura — Deferred Items

> **Last verified:** 2026-05-29
> **Scope:** Deferred items for the Distribuição & Cobertura de Leitura screen (`/documents/:documentId/distribution`).
> **Out of scope:** Published view deferred items (see `backlog/documento-publicado.md`), fanout render pipeline (see `modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` — page entry; wires real document identity via `useDocumentDetailQuery`, renders honest "em breve" empty state for everything backend cannot yet serve
> - `frontend/apps/web/src/features/documents/components/distribution/DocRefCard.tsx` — 3-D doc card primitive (props-driven, no mock identity)
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.test.tsx` — locks honesty invariants (real identity rendered, no fabricated PR-EHS-014/LOTO/numbers, em-breve panel, CTAs disabled)

---

## QA 2026-05-29 — `qa/documents-distribution`

Removed all fabricated distribution UI (KPIStrip, DonutCard, DistributionFacts, CoverageByArea, TimelineCard, RecipientsCard, `lib/distributionMeta.ts` mock constants). The page now:

- Wires real document identity (Code, RevisionNumber, Name) via `useDocumentDetailQuery` — same TanStack pattern as `DocumentPublishedPage`.
- Renders an "Em breve" empty-state panel listing the 4 planned capabilities (Cobertura de leitura, Cobertura por área, Curva de adoção, Lista de destinatários).
- Keeps the 4 hero CTAs (Lembrete em massa / Exportar relatório / Adicionar destinatários / Política de fanout) with `aria-disabled="true" title="Em breve"`.
- Mirrors loading + error states from `DocumentPublishedPage` (role=status / role=alert + retry button).

Gate 3 Preview proof (document `c1bb2112-21ea-46fc-ac1f-719d04994d41` / `PO-RH-002` / `REV00` / `DC_Template_Descricao_Cargo`): identity rendered honestly; no fabricated PR-EHS-014 / LOTO / 248 leaks; bogus uuid renders error state cleanly; only `GET /api/v1/documents/:id` is called — zero calls to non-existent distribution endpoints; no console errors.

---

## Backend endpoints (contract design — no implementation yet)

Design per `wiki/architecture/backend-api-structure.md` + `wiki/architecture/api-contract.md` (contract-first via OpenAPI / oapi-codegen). Implement only when fanout / read-tracking module is scheduled (out of `qa/documents-distribution` boundary — see hard-stop note below).

| Endpoint | Purpose | Priority |
|---|---|---|
| `GET /api/v1/documents/:id/distribution` | KPI rollup (`totalTargets`, `acknowledged`, `read`, `pending`, `overdue`) + sidebar facts (`publishedAt`, `readDeadline`, `policy`, `channel`, `remindersScheduled`, `groups[]`) | High |
| `GET /api/v1/documents/:id/distribution/recipients?page=&pageSize=&status=` | Paginated recipient table (`userId`, `name`, `area`, `status`, `lastEventAt`, `readAt`, `acknowledgedAt`) — `X-Total-Count` + `Link` headers per library pagination pattern | High |
| `GET /api/v1/documents/:id/distribution/coverage` | By-area breakdown (`areaCode`, `areaName`, `total`, `read`, `acknowledged`) | Medium |
| `GET /api/v1/documents/:id/distribution/timeline?granularity=day` | Cumulative daily read/ack series for adoption curve | Low |

Wire via TanStack Query hooks in `features/documents/queries/` (`useDistributionSummaryQuery`, `useDistributionRecipientsQuery`, `useDistributionCoverageQuery`, `useDistributionTimelineQuery`) when endpoints exist. Generated FE types from `lib/api-types/`.

## Fanout module (hard-stop — out of `qa/documents-distribution` scope)

The 4 hero CTAs and all recipient bulk/row actions require a fanout + read-tracking Go module that does not exist. Classification: missing-shared-contract + missing-module. **Redesign-grade** (worker/outbox semantics, recipient resolution from groups/areas/explicit users, idempotent reminder dispatch, ack semantics). Per the QA operating-system hard-stop rule, this is reported, not patched.

| CTA | Required endpoint | Status |
|---|---|---|
| Lembrete em massa | `POST /api/v1/documents/:id/fanout/reminders/bulk` | No module |
| Exportar relatório | `GET /api/v1/documents/:id/distribution/export` | No module |
| Adicionar destinatários | `POST /api/v1/documents/:id/fanout/recipients` | No module |
| Política de fanout | Settings modal (no design) | No design |
| Row action — mail | `POST /api/v1/documents/:id/fanout/reminders/:recipientId` | No module |
| Bulk action bar | Bulk send reminder, reassign, export | No module |

Minimum prerequisite plan:
1. New module `internal/modules/distribution` owning recipient resolution, read events, ack events, reminder dispatch (worker + outbox).
2. New tables: `document_distribution_targets`, `document_read_events`, `document_acknowledgement_events`, `document_reminder_jobs`.
3. Contract-first OpenAPI surface (4 endpoints above + fanout endpoints), oapi-codegen wiring.
4. Worker for reminder dispatch (reuses outbox pattern from `render-fanout` if applicable).
5. Wire TanStack hooks + remove `aria-disabled` from CTAs once endpoints land.

## Chart library

Donut + timeline visuals were previously hand-rolled SVG against mock data. When the timeline + coverage endpoints land, choose a chart library: candidates `recharts`, `victory`, `visx`. Criteria: tree-shakeable, no forced Canvas, accessible, compatible with CSS token system.

## Pagination

Recipient table will reuse the library pagination primitive (`features/documents/components/Pagination.tsx`) once `GET /distribution/recipients` returns `X-Total-Count` + `Link` headers.
