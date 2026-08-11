# Distribuição & Cobertura — Deferred Items

> **Last verified:** 2026-06-21 (write-path design extracted to the parked mission; prior: 2026-06-08)
> **Scope:** M2 (frontend-screen-completion) derive-on-read **coverage-scope** follow-ups for the
> Distribuição & Cobertura de Leitura screen (`/documents/:documentId/distribution`) — the *denominator*
> (real obligated recipients + by-area + deadline + pending) rendered live, read/ack shown as an honest
> *"tracking pending"* state.
> **Write-path moved out:** the full evidence (read/ack) + action (reminders/export/fanout) domain — the
> *numerator* — is parked as a designed mission:
> [`document-distribution-mission.md`](document-distribution-mission.md). Build it only after the
> frontend-screen-completion mission finishes. Everything below the screen context now lives there.
> **Out of scope:** Published view deferred items (see `backlog/documento-publicado.md`), fanout render pipeline (see `modules/render-fanout.md`).
> **Key files:**
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx` — page entry; wires real document identity via `useDocumentDetailQuery`, wraps the full mock design as illustrative scaffolding under aria-hidden + watermark blocks with a role=note em-breve banner above
> - `frontend/apps/web/src/features/documents/components/distribution/DocRefCard.tsx` — 3-D doc card primitive (props-driven, no mock identity)
> - `frontend/apps/web/src/features/documents/components/distribution/{KPIStrip,DonutCard,DistributionFacts,CoverageByArea,TimelineCard,RecipientsCard}.tsx` — design scaffolding sourcing `MOCK_DISTRIBUTION` from `lib/distributionMeta.ts`; rendered only inside `IllustrativeBlock`
> - `frontend/apps/web/src/features/documents/lib/distributionMeta.ts` — mock data (PR-EHS-014/LOTO) used exclusively for design preview; isolated to scaffolded blocks
> - `frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.test.tsx` — locks the scaffolding pattern (real identity in live banner, 5 watermarks + aria-hidden, 4 CTAs disabled, banner names doc)

---

## QA 2026-05-29 — `qa/documents-distribution`

**Design-scaffolding pattern.** Mock distribution UI (KPIStrip, DonutCard, DistributionFacts, CoverageByArea, TimelineCard, RecipientsCard, `lib/distributionMeta.ts`) preserved as a *visual blueprint* for the unbuilt fanout/read-tracking feature. Future implementation has a concrete reference for what each block should render.

- Real document identity (Code, RevisionNumber, Name) wired via `useDocumentDetailQuery` at the honest surfaces: hero breadcrumb, hero badges, `DocRefCard`, and the em-breve banner body.
- Every illustrative section wrapped in `IllustrativeBlock`: `aria-hidden="true"` + `Dados ilustrativos · Em breve` watermark + `pointer-events: none` + `user-select: none` + saturate(0.85) + diagonal-stripe overlay. AT users skip the scaffolding; sighted users see it muted.
- `role="note"` banner above the scaffolding explicitly states numbers/areas/people are illustrative and do not reflect real data for `${docName}`.
- 4 hero CTAs kept with `aria-disabled="true" title="Em breve"`.
- Loading + error states mirror `DocumentPublishedPage` (role=status / role=alert + retry).

Gate 3 Preview proof (document `c1bb2112-21ea-46fc-ac1f-719d04994d41` / `PO-RH-002` / `REV00` / `DC_Template_Descricao_Cargo`): identity correct on every honest surface; scaffolding visible and aria-hidden; only `GET /api/v1/documents/:id` is called — zero calls to non-existent distribution endpoints; no console errors. Vitest 6/6 green.

---

## What M2 builds (derive-on-read coverage-scope)

M2 wires the **denominator** only, from data that exists at baseline (visibility grants + area
membership — `controlled_document_area_grants`, `controlled_document_user_grants`,
`document_process_areas`, `user_process_areas`, view `metaldocs.v_active_user_areas`):

- Real obligated recipients (paginated), real by-area breakdown, real read-deadline, real `pending`.
- Read / acknowledge numerator rendered as an explicit **"tracking pending"** state — never a fake
  `0%`, never illustrative data. The honest scaffolding/watermark pattern (below) is replaced by live
  denominator data; the numerator slots stay truthfully empty.

The **endpoint contract** M2 consumes (and the precise M2-vs-mission field split) is specified in
[`document-distribution-mission.md`](document-distribution-mission.md) §6. M2 ships the
denominator-bearing subset of the first three endpoints; the numerator fields, the timeline endpoint,
the 4 CTAs, and all row/bulk actions are the parked mission's scope (§8 there).

> **Write-path / fanout / read-tracking design** (the 4 tables, the publish-path snapshot-vs-derive
> decision, caps, reminder worker, the full endpoint table, minimum prerequisite plan) was **moved**
> to [`document-distribution-mission.md`](document-distribution-mission.md). It is **redesign-grade**
> and reported, not patched, inside this screen-wiring scope.

## Chart library

Donut + timeline visuals were previously hand-rolled SVG against mock data. When the timeline + coverage endpoints land, choose a chart library: candidates `recharts`, `victory`, `visx`. Criteria: tree-shakeable, no forced Canvas, accessible, compatible with CSS token system.

## Pagination

Recipient table will need a pagination primitive once `GET /distribution/recipients` returns `X-Total-Count` + `Link` headers. The library's prior `features/documents/components/Pagination.tsx` (+ `.module.css`) was deleted as dead code (zero importers) in the A2.2 burn-down (PR #119) — it will need to be rebuilt or a shared primitive extracted at implementation time, not reused as-is.
