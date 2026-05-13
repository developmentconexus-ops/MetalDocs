# Phase 1 — Map

**Screen:** Templates List (`/templates`)

## §1.1 Existing primitives

| Element | Primitive | Action |
|---|---|---|
| Status badge | `components/ui/StatusPill` | Reuse (published/draft/archived) |
| Author avatar | `components/ui/Avatar` (sm) | Reuse |
| Page header | `components/ui/WorkspaceHeroHeader` | Extend: add `kicker` + `action` props |
| Icon in button | `components/ui/Icon` | Reuse |

## §1.2 New components

| Component | Placement | Reason |
|---|---|---|
| `TabBar` | `components/ui/TabBar.tsx` | Generic tab-filter, reusable |
| `TemplateCard` | `features/templates/components/TemplateCard.tsx` | Domain card |
| `MiniDocPreview` | `features/templates/components/MiniDocPreview.tsx` | Decorative, only here |

## §1.3 Decomposition

```
TemplatesListPage
├── WorkspaceHeroHeader (extended: kicker + action button)
├── TabBar (Todos / Publicados / Rascunhos / Arquivados)
└── div.cardGrid (CSS grid 3-col)
    └── TemplateCard (×N)
        ├── MiniDocPreview
        ├── StatusPill + CodeChip (version)
        ├── h3 title
        └── footer: Avatar + name + timestamp
```

## §1.4 Status meta

Reuse existing `StatusPill` type union. Template status derived:
- `archived_at` set → 'archived'
- `published_version_id` set → 'published'
- else → 'draft'

## §1.5 State

- **Server:** TanStack Query `useTemplatesQuery()` → `listTemplates()`
- **Local UI:** `activeTab: 'all' | 'published' | 'draft' | 'archived'`
- **Derived:** counts per tab from filtered list

## §1.6 Backend

- `GET /api/v1/templates` — exists, returns `TemplateDTO[]`
- **Gap:** `created_by` is user_id (no display name). Mock: show id. TODO backlog.
- **Gap:** No `updated_at`. Fallback: `created_at`.
- **Migration:** `templatesV2.ts` uses raw `fetch()` → migrate `listTemplates` to `apiFetch`

## §1.7 Decisions

- TabBar → `components/ui/` (shared, user confirmed)
- WorkspaceHeroHeader → extend with `kicker?: string` + `action?: ReactNode` (user confirmed)
- TemplateCard / MiniDocPreview → feature-local
