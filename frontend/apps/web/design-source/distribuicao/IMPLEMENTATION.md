# Distribuição & Cobertura — Implementation Worksheet

> **Slug:** `distribuicao`
> **Owning feature:** `features/documents`
> **Target route:** `documents/:documentId/distribution`
> **Reference:** `./distribuicao.html` + `./NOTES.md`
> **Skill version:** 1.2
> **Started:** 2026-05-08
> **Completed:** —

---

## Open Questions Log

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | Route: `/documents/:id/distribution` or `/documents-v2/:id/distribution`? | `documents/:documentId/distribution` — distribution is a published-doc view, not editor context | ✅ |
| 2 | 0 | SVG charts (Donut, Timeline line chart) — Keep with mock data, or Defer and replace with placeholder cards? | Keep with mock data; defer real data + chart library research to backlog | ✅ |
| 3 | 0 | CTA buttons (Lembrete em massa, Exportar relatório, Adicionar destinatários, Política de fanout) — all `aria-disabled` + "Em breve", or any should route somewhere real? | All `aria-disabled` + "Em breve" — no real routes yet | ✅ |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

| Element | Maps to | Keep / Cut / Defer | Reason |
|---|---|---|---|
| **Hero — Breadcrumb** | doc.code, doc.area, literal "Distribuição" | Keep | Navigation context, real data |
| **Hero — DocRefCard (3D perspective)** | doc.code, doc.version, doc.areaShort | Keep | Visual identity, real doc data |
| **Hero — Title + description** | literal heading, doc.title + publishedAt + totalTargets | Keep | Real data |
| **Hero — "Vence em N dias" badge** | distribution.deadline + daysLeft | Keep (mock) | Real concept, no backend yet |
| **Hero — "Lembrete em massa" CTA** | POST /fanout/reminders — no endpoint | Defer (disabled) | No fanout module |
| **Hero — "Exportar relatório" CTA** | GET export — no endpoint | Defer (disabled) | No export API |
| **Hero — "Adicionar destinatários" CTA** | POST /fanout/recipients — no endpoint | Defer (disabled) | No fanout module |
| **Hero — "Política de fanout" CTA** | settings modal — no design | Defer (disabled) | No fanout module |
| **KPI strip — Alvos totais** | distribution.totalTargets | Keep (mock) | Core metric |
| **KPI strip — Reconheceram** | distribution.acknowledged + ackPct + progress bar | Keep (mock) | Core metric |
| **KPI strip — Apenas leram** | distribution.read - distribution.acknowledged | Keep (mock) | Core metric |
| **KPI strip — Pendentes** | distribution.pending | Keep (mock) | Core metric |
| **KPI strip — Em atraso** | distribution.overdue | Keep (mock) | Core metric |
| **§01 — DonutCard (SVG animated donut)** | ack%, read%, pending% | Keep (mock) | Keep with mock data; real chart lib TBD in backlog |
| **§01 — DistributionFacts sidebar** | publishedAt, deadline, policy, channel, reminders, groups | Keep (mock) | All real concepts |
| **§01 — "Editar política" btn** | no endpoint | Defer (disabled) | |
| **§02 — CoverageByArea (bullet bars)** | byArea: area name, total, read, ack | Keep (mock) | Core insight |
| **§03 — TimelineCard (SVG line chart)** | dailyReads cumulative | Keep (mock) | Keep with mock data; real chart lib TBD in backlog |
| **§04 — RecipientsCard filter tabs** | pending/overdue/read/ack/all tabs | Keep (mock) | Core interaction |
| **§04 — Search input** | client-side filter on mock data | Keep | useDebouncedValue |
| **§04 — Recipient row (checkbox, avatar, name, area, status, activity, when, actions)** | mock people data | Keep (mock) | Core table |
| **§04 — Bulk action bar** | send reminder, reassign, export — no endpoints | Defer (disabled) | No fanout module |
| **§04 — Row action buttons (mail, more)** | send individual reminder, view profile | Defer (disabled) | No fanout module |
| **§04 — Pagination footer** | "1 / 23" — no real pagination | Keep (mock, static) | Visual completeness |
| **Animated counters** | CSS-only cosmetic | Keep | Purely visual |

### 0.2 Cut list confirmed

- [x] User reviewed cut list (2026-05-08)
- [x] Cuts recorded in NOTES.md

---

## Phase 1 — Map (HARD GATE)

_To be filled after Phase 0 user confirmation._

---

## Phase 2 — Pre-flight (advisory)

_To be filled after Phase 1._

---

## Phase 3a — Structure mirror (HARD GATE)

_To be filled after Phase 2._

---

## Phase 3b — Style port (HARD GATE)

_To be filled after Phase 3a._

---

## Phase 3c — State wiring (advisory)

_To be filled after Phase 3b._

---

## Phase 4 — Verify (HARD GATE)

_To be filled after Phase 3c._

---

## Phase 5 — Document

_To be filled after Phase 4._
