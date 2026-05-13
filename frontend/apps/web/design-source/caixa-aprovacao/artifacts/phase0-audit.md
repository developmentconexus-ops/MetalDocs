# Phase 0 Audit — Caixa de Aprovação

**Slug:** `caixa-aprovacao`
**Date:** 2026-05-08
**Confirmed by user:** ✅ 2026-05-08

## Strategy

Implement full visual fidelity to design. Missing API fields use hardcoded mock data flagged inline:
```ts
// TODO [BACKLOG: caixa-aprovacao.md]: needs <fieldName> from GET /api/v1/approval/inbox
```
All mock fields catalogued in `wiki/backlog/caixa-aprovacao.md`.

## Keep / Cut / Defer Table

| Element | Decision | Data source |
|---|---|---|
| InboxToolbar breadcrumb | KEEP | Real |
| InboxToolbar Filtros button | KEEP (disabled placeholder) | Real API param `area_code`; panel deferred |
| InboxToolbar view-switcher (Foco/Linha do tempo) | KEEP | `localStorage['md.inbox.v']` |
| Queue: "SUA FILA · N decisões" | KEEP | `total` from API |
| Queue: "N vencem hoje" | KEEP (mock) | MOCK — needs `deadline_at` |
| Queue: numbered item list | KEEP | Real `items[]` |
| Queue item: code chip | KEEP (mock) | MOCK — needs `code` field |
| Queue item: kind badge | KEEP (mock) | MOCK — needs `kind` field |
| Queue item: urgency blink dot | KEEP (mock) | MOCK — needs `urgent` field |
| Queue item: deadline countdown | KEEP (mock) | MOCK — needs `deadline_at` |
| Queue item: title | KEEP | `document_title` |
| Queue item: area | KEEP | `area_code` |
| Ghost cards z-stack | KEEP | CSS only |
| Card header urgent gradient | KEEP (mock) | MOCK — needs `urgent` field |
| Card header "VENCE EM" | KEEP (mock) | MOCK — needs `deadline_at` |
| Card header kind + code + version | KEEP (mock) | MOCK — needs `code`, `kind`, `version` |
| Card header title | KEEP | `document_title` |
| Card body: summary | KEEP (mock) | MOCK — needs `summary` field |
| Card body: AUTOR | KEEP | `submitted_by` + `area_code` |
| Card body: ALTERAÇÕES | KEEP (mock) | MOCK — needs `changes` field |
| Card body: ESTÁGIO | KEEP | `stage_label` + parse `quorum_progress` |
| Card action: "Abrir documento" | KEEP | `navigate('/registry-v2/:id')` |
| Card action: "Devolver" | KEEP (mock) | Navigate to doc — MOCK until signoff flow designed |
| Card action: "Aprovar e assinar" | KEEP (mock) | Navigate to doc — MOCK until signoff flow designed |
| Keyboard shortcuts footer | DEFER | Non-critical; add post-core |
| Timeline header "X decisões" | KEEP | `total` from API |
| Timeline heatmap sparkline | KEEP (mock) | MOCK static data — needs `GET /approval/my-decisions?days=14` |
| Timeline deadline buckets | KEEP (mock) | MOCK `deadline_at` on items |
| Timeline item: time + "vence em" | KEEP (mock) | MOCK |
| Timeline item: code + kind | KEEP (mock) | MOCK |
| Timeline item: title / author / area | KEEP | Real fields |
| Timeline item: stage bars | KEEP | Parse `quorum_progress` |
| Timeline item: "Revisar →" | KEEP | `navigate('/registry-v2/:id')` |

## Backlog items (all mock → API gaps)

See `wiki/backlog/caixa-aprovacao.md` for full list. Summary:

1. `InboxItem.code` — document code
2. `InboxItem.kind` — document type short code
3. `InboxItem.deadline_at` — absolute deadline
4. `InboxItem.urgent` — urgency flag
5. `InboxItem.summary` — revision summary
6. `InboxItem.changes` — edit count
7. `InboxItem.version` — version transition string
8. `GET /api/v1/approval/my-decisions?days=14` — heatmap history
9. Signoff flow — approve/reject action path (design + implement)
