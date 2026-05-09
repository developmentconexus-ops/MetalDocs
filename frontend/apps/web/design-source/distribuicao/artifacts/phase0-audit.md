# Phase 0 — Audit

> **Slug:** distribuicao
> **Confirmed:** 2026-05-08 by user
> **Route resolved:** `documents/:documentId/distribution`

## Keep / Cut / Defer Table

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| Hero — Breadcrumb | doc.code, doc.area, literal "Distribuição" | **Keep** | Navigation context, real data |
| Hero — DocRefCard (3D perspective) | doc.code, doc.version, doc.areaShort | **Keep** | Visual identity, real doc data |
| Hero — Title + description | doc.title + publishedAt + totalTargets | **Keep** | Real data |
| Hero — "Vence em N dias" badge | distribution.deadline + daysLeft | **Keep (mock)** | Real concept, no backend yet |
| Hero — "Lembrete em massa" CTA | POST /fanout/reminders — no endpoint | **Defer (disabled)** | No fanout module |
| Hero — "Exportar relatório" CTA | GET export — no endpoint | **Defer (disabled)** | No export API |
| Hero — "Adicionar destinatários" CTA | POST /fanout/recipients — no endpoint | **Defer (disabled)** | No fanout module |
| Hero — "Política de fanout" CTA | settings modal — no design | **Defer (disabled)** | No fanout module |
| KPI strip — Alvos totais | distribution.totalTargets | **Keep (mock)** | Core metric |
| KPI strip — Reconheceram | distribution.acknowledged + ackPct + progress bar | **Keep (mock)** | Core metric |
| KPI strip — Apenas leram | distribution.read - distribution.acknowledged | **Keep (mock)** | Core metric |
| KPI strip — Pendentes | distribution.pending | **Keep (mock)** | Core metric |
| KPI strip — Em atraso | distribution.overdue | **Keep (mock)** | Core metric |
| §01 — DonutCard (SVG animated donut) | ack%, read%, pending% | **Keep (mock)** | Keep with mock data; real chart lib TBD in backlog |
| §01 — DistributionFacts sidebar | publishedAt, deadline, policy, channel, reminders, groups | **Keep (mock)** | All real concepts |
| §01 — "Editar política" btn | no endpoint | **Defer (disabled)** | No fanout module |
| §02 — CoverageByArea (bullet bars) | byArea: area name, total, read, ack | **Keep (mock)** | Core insight |
| §03 — TimelineCard (SVG line chart) | dailyReads cumulative | **Keep (mock)** | Keep with mock data; real chart lib TBD in backlog |
| §04 — RecipientsCard filter tabs | pending/overdue/read/ack/all tabs | **Keep (mock)** | Core interaction |
| §04 — Search input | client-side filter on mock data | **Keep** | useDebouncedValue |
| §04 — Recipient row (checkbox, avatar, name, area, status, activity, when, actions) | mock people data | **Keep (mock)** | Core table |
| §04 — Bulk action bar | send reminder, reassign, export — no endpoints | **Defer (disabled)** | No fanout module |
| §04 — Row action buttons (mail, more) | individual reminder, view profile | **Defer (disabled)** | No fanout module |
| §04 — Pagination footer | "1 / 23" static | **Keep (mock, static)** | Visual completeness |
| Animated counters | CSS-only cosmetic | **Keep** | Purely visual |

## Open Questions — All Resolved

| # | Question | Resolution |
|---|---|---|
| 1 | Route path | `documents/:documentId/distribution` — distribution is a published-doc view, not editor context |
| 2 | SVG charts | Keep with mock data; backlog item for chart library research |
| 3 | CTA buttons | All `aria-disabled` + "Em breve" — no real routes |

## Deferred Items (backlog)

- All 4 hero CTAs (fanout module prerequisite)
- "Editar política" button (fanout module prerequisite)
- Bulk action bar (fanout module prerequisite)
- Row action buttons (fanout module prerequisite)
- DonutCard + TimelineCard real data wiring (chart library research needed)
- RecipientsCard real pagination (no backend endpoint yet)
