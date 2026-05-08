# Phase 3c — State Wiring Artifact

Date: 2026-05-08

## Files changed

| File | Change |
|---|---|
| `src/features/approval/pages/InboxPage.tsx` | Added `view` state (lazy localStorage), `selectedIdx` state, `useInboxQuery` wiring, `handleViewChange`/`handleSelect`/`handleNext`/`handlePrev` handlers, mock fallback logic |
| `src/features/approval/components/InboxToolbar.tsx` | Added `title="Em breve"` to disabled Filtros button |
| `src/features/approval/components/InboxStack.tsx` | Full prop interface wired: queue rail from `items`, active/urgent display, counter, prev/next buttons with bounds, loading/error/empty states, keyboard nav via `useEffect` |
| `src/features/approval/components/InboxStack.module.css` | Added `.loading`, `.error`, `.empty` state classes |
| `src/features/approval/components/InboxApprovalCard.tsx` | Accepts `item: RichInboxItem`; all hardcoded values replaced: `kind`, `code`, `version`, `document_title`, `deadline`, `summary`, `submitted_by`, `area_code`, `changes`, `stage_label`; urgent header class applied |
| `src/features/approval/components/InboxTimeline.tsx` | Accepts `items: RichInboxItem[]`; grouped by deadline bucket via `groupByDeadlineBucket()`; rows rendered from real data; heatmap kept hardcoded (TODO comment added) |
| `src/features/approval/lib/mockInboxData.ts` | Added `deadline_at` field to `RichInboxItem` and `MOCK_EXTRAS`; added `MOCK_INBOX_ITEMS` (4 canonical base objects matching Phase 3a hardcoded items) |

## State design

### InboxPage (owns all server + UI state)
- `view: string` — lazy init from `localStorage.getItem('md.inbox.v')` or `'stack'`
- `selectedIdx: number` — index of selected card, clamped in handlers
- `useInboxQuery()` — TanStack Query server state; enriched via `enrichInboxItem()`
- Fallback: if API returns empty/errors → `MOCK_INBOX_ITEMS`
- `handleViewChange(v)` → sets state + `localStorage.setItem`
- `handleSelect(idx)`, `handleNext()`, `handlePrev()` — passed down as callbacks

### Components (stateless, receive props)
- `InboxToolbar` — receives `view` + `onViewChange`; active button state driven by prop
- `InboxStack` — receives `items`, `selectedIdx`, `onSelect`, `onNext`, `onPrev`, `isLoading`, `isError`; `useEffect` keyboard listener scoped to `items.length > 0`
- `InboxApprovalCard` — receives `item: RichInboxItem`; fully data-driven
- `InboxTimeline` — receives `items: RichInboxItem[]`; groups into 4 buckets internally

## Mock data fields and TODO trail

All `RichInboxItem` fields beyond `InboxItem` base are mock-filled:

| Field | Source | TODO |
|---|---|---|
| `code` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.code` from API |
| `kind` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.kind` from API |
| `deadline` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.deadline_at` from API |
| `deadline_at` | Computed ISO string | Needs real timestamp from API |
| `urgent` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.urgent` from API |
| `summary` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.summary` from API |
| `changes` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.changes` from API |
| `version` | `MOCK_EXTRAS[idx]` | Needs `InboxItem.version` from API |

All TODOs tagged `[BACKLOG: caixa-aprovacao.md]`.

## Deferred items

- Approve/return/signoff flows (action buttons and keyboard A/D) — TODO tagged
- Timeline item click → navigate to doc — TODO tagged
- Timeline heatmap real data — TODO tagged
- `title`/`time` display in timeline rows shows `—` (no real time-of-day from API) — TODO tagged implicitly
- `InboxPage` fallback removes mock items when API returns non-empty results (intended behavior)

## TSC result

`pnpm tsc --noEmit -p tsconfig.build.json` exit code: **0 new errors in `features/approval/`**

Pre-existing errors in other features (auth, documents, shell) were not introduced by this phase.
