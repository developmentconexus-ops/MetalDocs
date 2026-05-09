# Phase 1 — Map

> **Slug:** distribuicao  
> **Route:** `documents/:documentId/distribution`  
> **Completed:** 2026-05-08

---

## 1.1 Reusability scan — backward

| Design element | Existing primitive | Decision |
|---|---|---|
| Recipient filter tabs | `components/ui/TabBar.tsx` | Reuse as-is |
| Recipient search input | `components/ui/SearchBar.tsx` | Reuse as-is |
| Recipient row avatar | `components/ui/Avatar.tsx` | Reuse as-is (size `sm`) |
| Debounced search | `lib/hooks/useDebouncedValue.ts` | Reuse as-is |
| Date formatting | `features/documents/lib/documentDetailMeta.ts` | Reuse `formatPublishedAt`, `formatShortDate` |
| Recipient status badge | `components/ui/StatusPill.tsx` | ❌ Can't reuse — StatusPill is for document lifecycle statuses; distribution statuses (ack/read/pending/overdue) are a different domain |

## 1.2 Reusability scan — forward

All new components are specific to the distribuicao screen — none appear in other design files.
All go in `features/documents/components/distribution/`.

Not promoting to `components/ui/` because:
- DocRefCard, DonutCard, TimelineCard, CoverageByArea are chart/data-viz with no other callers.
- DistributionFacts is domain-specific data layout.
- RecipientsCard is a full section component with tab+search+table logic.

## 1.3 Decomposition

```
DocumentDistributionPage
  features/documents/pages/DocumentDistributionPage.tsx
  features/documents/pages/DocumentDistributionPage.module.css

  ├── Hero (inline in page)
  │   ├── Breadcrumb nav (inline)
  │   ├── DocRefCard
  │   │     features/documents/components/distribution/DocRefCard.tsx
  │   ├── Title + description + deadline badge (inline)
  │   └── 4x aria-disabled CTAs (inline buttons)
  │
  ├── KPIStrip
  │     features/documents/components/distribution/KPIStrip.tsx
  │
  ├── §01 — two-col grid (inline layout)
  │   ├── DonutCard (SVG animated donut)
  │   │     features/documents/components/distribution/DonutCard.tsx
  │   └── DistributionFacts sidebar
  │         features/documents/components/distribution/DistributionFacts.tsx
  │
  ├── §02 — CoverageByArea (bullet bars)
  │     features/documents/components/distribution/CoverageByArea.tsx
  │
  ├── §03 — TimelineCard (SVG line chart)
  │     features/documents/components/distribution/TimelineCard.tsx
  │
  └── §04 — RecipientsCard
        features/documents/components/distribution/RecipientsCard.tsx
          ├── TabBar (components/ui/TabBar) ← reuse
          ├── SearchBar (components/ui/SearchBar) ← reuse
          ├── RecipientRow (inline in RecipientsCard)
          │     └── Avatar (components/ui/Avatar) ← reuse
          └── Static pagination footer (inline)
```

## 1.4 Status/enum meta SSOT

New file: `features/documents/lib/distributionMeta.ts`

Contains:
- `RECIPIENT_STATUS_TONE: Record<'ack' | 'read' | 'pending' | 'overdue', { label: string; bg: string; fg: string }>`
  — maps to `var(--success-bg/success/brand-pale/brand/surface-3/text-muted/danger-bg/danger)`
- Mock data interfaces: `DistributionMock`, `RecipientMock`, `CoverageAreaMock`, `DailyReadMock`
- Mock data constants (from design: `DF5` object, `PEOPLE_F5` array, etc.)
- `RECIPIENT_TABS: TabBarItem[]` constant

## 1.5 State design

| State | Type | Location | Mechanism |
|---|---|---|---|
| Active recipient tab | `'pending'|'overdue'|'read'|'ack'|'all'` | Page | `useState('pending')` |
| Search input | `string` | Page | `useState('')` |
| Debounced search | `string` | Page | `useDebouncedValue(search, 300)` |
| Selected recipient IDs | `Set<string>` | Page | `useState(() => new Set<string>())` |
| SVG animations (donut, timeline) | internal animation frame state | DonutCard, TimelineCard | `useEffect` + `requestAnimationFrame` with intersection observer trigger |
| Loading/error | — | Page | All mock for now; TanStack Query wired when endpoints exist |

## 1.6 Backend contract

**No real API exists.** All data is mock.

Mock data lives in `features/documents/lib/distributionMeta.ts`.

TODO comments block:
```ts
// TODO(distribuicao): wire GET /api/v2/documents/:id/distribution — KPI metrics + facts
// TODO(distribuicao): wire GET /api/v2/documents/:id/distribution/recipients — paginated recipient table
// TODO(distribuicao): wire GET /api/v2/documents/:id/distribution/coverage — by-area breakdown
// TODO(distribuicao): wire GET /api/v2/documents/:id/distribution/timeline — cumulative daily reads
// See wiki/backlog/distribuicao.md
```

Backlog file: `wiki/backlog/distribuicao.md`

## 1.7 New route registration

Add to `features/documents/routes.tsx`:
```ts
{
  path: 'documents/:documentId/distribution',
  handle: { workspaceView: 'library' },
  lazy: () => import('./pages/DocumentDistributionPage').then(m => ({ Component: m.DocumentDistributionPage })),
},
```

Placement: after the `documents/:documentId` route (sibling, not child).

## Open questions — resolved

None. All Phase 0 OQs resolved before Phase 1.
