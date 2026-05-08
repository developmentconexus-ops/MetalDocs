# Phase 1 Map — Caixa de Aprovação

**Date:** 2026-05-08
**Status:** ✅ Complete — no open questions

## Primitives reused (backward scan)

| Primitive | Action |
|---|---|
| `Avatar` (sm) | use; `xs` size gap → fix in Phase 2 |
| `Icon` | use (filter, eye icons) |
| `CodeChip` | use for doc code (mock data) |
| `TabBar` | NOT suitable — design uses custom pill switcher |
| `TimelineRail` | NOT suitable — design needs custom grouped bucket structure |

## New components (forward scan)

| Component | Location |
|---|---|
| `InboxToolbar` | `features/approval/components/InboxToolbar.tsx` |
| `InboxStack` | `features/approval/components/InboxStack.tsx` |
| `InboxApprovalCard` | `features/approval/components/InboxApprovalCard.tsx` |
| `InboxTimeline` | `features/approval/components/InboxTimeline.tsx` |

Queue rail items + ghost cards = inline sub-components within InboxStack.

## New query hook

`features/approval/queries/useInboxQuery.ts` — TanStack Query, replaces `useEffect`+`setState` in existing `InboxPage`.

## New lib file

`features/approval/lib/mockInboxData.ts` — `RichInboxItem` type + `enrichInboxItem()` function. All mock fields flagged with `// TODO [BACKLOG: caixa-aprovacao.md]`.

## State

- `view: 'stack' | 'timeline'` — persisted `localStorage['md.inbox.v']`, lazy `useState`
- `selectedIdx: number` — local state, InboxStack

## QK key

`QK.inbox()` already exists in `lib/queryKeys.ts`.

## Backend gaps → backlog

All tracked in `wiki/backlog/caixa-aprovacao.md`:
1. `InboxItem.code`, `.kind`, `.deadline_at`, `.urgent`, `.summary`, `.changes`, `.version`
2. `GET /api/v2/approval/my-decisions?days=14` (heatmap)
3. Signoff flow (approve/reject path + password challenge)
