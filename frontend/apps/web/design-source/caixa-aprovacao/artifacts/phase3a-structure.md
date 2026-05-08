# Phase 3a — Structure Mirror Artifact

**Screen:** Caixa de Aprovação (`caixa-aprovacao`)
**Date:** 2026-05-08
**Phase:** 3a (Structure Mirror — TSX skeletons only, no logic/state/data)

---

## Files Created / Replaced

| # | Action | Path |
|---|--------|------|
| 1 | REPLACED | `src/features/approval/pages/InboxPage.tsx` |
| 2 | REPLACED | `src/features/approval/pages/InboxPage.module.css` |
| 3 | CREATED | `src/features/approval/components/InboxToolbar.tsx` |
| 4 | CREATED | `src/features/approval/components/InboxToolbar.module.css` |
| 5 | CREATED | `src/features/approval/components/InboxStack.tsx` |
| 6 | CREATED | `src/features/approval/components/InboxStack.module.css` |
| 7 | CREATED | `src/features/approval/components/InboxApprovalCard.tsx` |
| 8 | CREATED | `src/features/approval/components/InboxApprovalCard.module.css` |
| 9 | CREATED | `src/features/approval/components/InboxTimeline.tsx` |
| 10 | CREATED | `src/features/approval/components/InboxTimeline.module.css` |

---

## Class-Name Mapping Table

### InboxPage

| Design DOM Role | CSS Module Class |
|-----------------|-----------------|
| page root wrapper | `.page` |

### InboxToolbar

| Design DOM Role | CSS Module Class |
|-----------------|-----------------|
| toolbar bar | `.toolbar` |
| "APROVAÇÕES" kicker | `.kicker` |
| breadcrumb separator `›` | `.breadcrumbSep` |
| "Caixa de entrada" label | `.breadcrumbTitle` |
| flex spacer | `.spacer` |
| Filtros button | `.filtersBtn` |
| view switcher pill wrapper | `.viewSwitcher` |
| individual view option button | `.viewSwitcherBtn` |

### InboxStack

| Design DOM Role | CSS Module Class |
|-----------------|-----------------|
| component root | `.root` |
| left queue rail aside | `.queueRail` |
| queue header section | `.queueHeader` |
| count display row | `.queueCount` |
| large count number | `.queueCountNum` |
| "2 vencem hoje" highlight | `.urgentToday` |
| queue item button | `.queueItem` |
| active/selected queue item | `.queueItemActive` |
| numeric position label | `.queueItemNumber` |
| meta column wrapper | `.queueItemMeta` |
| top row (code + dot) | `.queueItemTop` |
| document code span | `.queueItemCode` |
| urgent indicator dot | `.urgentDot` |
| document title line | `.queueItemTitle` |
| deadline + area row | `.queueItemSub` |
| deadline value | `.queueItemDeadline` |
| middle dot separator | `.dot` |
| card main area | `.cardArea` |
| nav bar (counter + prev/next) | `.cardNav` |
| "01 / 04" counter | `.cardCounter` |
| flex spacer | `.spacer` |
| stacked card container | `.cardStack` |
| ghost card base | `.ghostCard` |
| ghost card (front, 96% scale) | `.ghostCard1` |
| ghost card (back, 92% scale) | `.ghostCard2` |
| keyboard shortcut hints bar | `.keyboardHints` |

### InboxApprovalCard

| Design DOM Role | CSS Module Class |
|-----------------|-----------------|
| card article | `.card` |
| header band | `.cardHeader` |
| header inner flex | `.cardHeaderInner` |
| kind badge pill (POP/IT/etc) | `.kindBadge` |
| meta block next to badge | `.cardHeaderMeta` |
| code + version line | `.codeVersion` |
| document title h2 | `.cardTitle` |
| deadline block (right side) | `.deadlineBlock` |
| "VENCE EM" label | `.deadlineLabel` |
| deadline value | `.deadlineValue` |
| card body | `.cardBody` |
| summary paragraph | `.summary` |
| 3-column stats grid | `.statsGrid` |
| individual stat cell | `.statCell` |
| author row (avatar + name) | `.authorRow` |
| author name | `.authorName` |
| changes display row | `.changesRow` |
| large changes number | `.changesNum` |
| stage name value | `.stageName` |
| action buttons row | `.cardActions` |
| "Abrir documento" button | `.btnOpen` |
| "Devolver" button | `.btnReturn` |
| "Aprovar e assinar" button | `.btnApprove` |

### InboxTimeline

| Design DOM Role | CSS Module Class |
|-----------------|-----------------|
| scroll container | `.timelineContainer` |
| inner padding wrapper | `.inner` |
| header flex row | `.timelineHeader` |
| heading + subtitle block | `.timelineHeadingBlock` |
| h1 title | `.timelineTitle` |
| brand-colored title span | `.titleAccent` |
| heatmap widget card | `.heatmap` |
| bars flex container | `.heatmapBars` |
| individual bar | `.heatmapBar` |
| labels row | `.heatmapLabels` |
| date label span | `.heatmapLabel` |
| timeline body | `.timeline` |
| vertical gradient rail line | `.timelineRail` |
| bucket section | `.bucketSection` |
| bucket dot on rail | `.bucketDot` |
| urgent bucket dot modifier | `.bucketDotUrgent` |
| bucket header row | `.bucketHeader` |
| bucket label h2 | `.bucketLabel` |
| urgent bucket label modifier | `.bucketLabelUrgent` |
| bucket sub label (date) | `.bucketSub` |
| flex spacer | `.spacer` |
| doc count badge | `.bucketCount` |
| urgent doc count modifier | `.bucketCountUrgent` |
| empty count modifier | `.bucketCountEmpty` |
| items container | `.bucketItems` |
| item row div[role=button] | `.itemRow` |
| time + deadline column | `.itemTime` |
| time value (large mono) | `.itemTimeValue` |
| title + code meta column | `.itemMeta` |
| top row of meta | `.itemMetaTop` |
| kind badge pill | `.itemKind` |
| document code | `.itemCode` |
| document title | `.itemTitle` |
| author + area + changes row | `.itemAuthLine` |
| middle dot separator | `.itemDot` |
| spacer between meta and stage | `.itemSpacer` |
| stage progress column | `.stageProgress` |
| stage bars container | `.stageBars` |
| individual stage bar | `.stageBar` |
| filled stage bar modifier | `.stageBarFilled` |
| urgent filled bar modifier | `.stageBarUrgent` |
| "ETAPA X/Y" label | `.stageLabel` |
| "Revisar →" button | `.reviewBtn` |
| empty bucket placeholder | `.emptyBucket` |

---

## Semantic HTML Decisions

- `<article>` used for the active approval card in `InboxApprovalCard` — correct, it is a self-contained unit of content.
- `<aside>` used for the queue rail in `InboxStack` — landmark for the supplementary list of pending items.
- `<main>` used for the card display area in `InboxStack` — primary content region of the stack view.
- `<section>` used for each timeline bucket in `InboxTimeline` — each bucket is a thematically grouped region.
- `<h1>` in `InboxTimeline` header, `<h2>` for bucket labels and card title — maintains heading hierarchy.
- Timeline item rows use `<div role="button" tabIndex={0}>` (not `<button>`) — the design uses `cursor:pointer` on a div, not a button. This is semantically correct for a compound interactive region that is not a simple button. Handlers wired in Phase 3c.
- Queue items in `InboxStack` use `<button type="button">` — they are simple activatable items, appropriate as buttons.
- `aria-hidden="true"` on ghost cards (decorative) and timeline rail (decorative line).
- `Icon name="docs"` used for "Abrir documento" action — `eye` is not in the `IconName` type; `docs` is the closest semantic substitute. A TODO comment is implicit; Phase 3c can update if the icon set is extended.

---

## TSC Result

```
pnpm tsc --noEmit -p tsconfig.build.json
```

**Result: no new errors introduced by Phase 3a files.**

Pre-existing errors (unchanged from Phase 2 baseline):
- `src/features/auth/__tests__/useAuthSession.returnTo.test.tsx` — 2 errors (TS2554)
- `src/features/documents/components/LibrarySidebar.tsx` — 2 errors (TS2339, TS7006)
- `src/features/documents/pages/NewDocumentWizardPage.tsx` — 3 errors (TS2339, TS7006, TS2740)
- `src/features/documents/queries/__tests__/useAreasQuery.test.ts` — 3 errors (TS7053)
- `src/features/documents/queries/useAreasQuery.ts` — 1 error (TS2769)
- `src/features/shell/components/Rail.tsx` — 1 error (TS2322)

All errors are in unrelated features/tests. Zero errors in `features/approval/`.
