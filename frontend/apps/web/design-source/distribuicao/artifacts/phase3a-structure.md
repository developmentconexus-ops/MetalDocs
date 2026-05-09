# Phase 3a — Structure mirror: distribuicao

## Page file

`frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.tsx`
`frontend/apps/web/src/features/documents/pages/DocumentDistributionPage.module.css`

## Component files

| Component            | TSX                                                                                                            | CSS Module                                                                                                        |
|----------------------|----------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| DocRefCard           | `src/features/documents/components/distribution/DocRefCard.tsx`           | `DocRefCard.module.css`           |
| KPIStrip             | `src/features/documents/components/distribution/KPIStrip.tsx`             | `KPIStrip.module.css`             |
| DonutCard            | `src/features/documents/components/distribution/DonutCard.tsx`            | `DonutCard.module.css`            |
| DistributionFacts    | `src/features/documents/components/distribution/DistributionFacts.tsx`    | `DistributionFacts.module.css`    |
| CoverageByArea       | `src/features/documents/components/distribution/CoverageByArea.tsx`       | `CoverageByArea.module.css`       |
| TimelineCard         | `src/features/documents/components/distribution/TimelineCard.tsx`         | `TimelineCard.module.css`         |
| RecipientsCard       | `src/features/documents/components/distribution/RecipientsCard.tsx`       | `RecipientsCard.module.css`       |

Lib (pre-existing, not modified): `src/features/documents/lib/distributionMeta.ts`

## Icon.tsx additions

Added 6 new icon names to `src/components/ui/Icon.tsx`:
`calendar`, `clock`, `shield`, `mail`, `chevron-right`, `chevron-left`

The design source used these icons in DistributionFacts and navigation elements. They were absent from the existing IconName union.

## Class-name mapping table

### DocumentDistributionPage.module.css
`.page` `.hero` `.heroGrid` `.heroOverlay` `.breadcrumb` `.breadcrumbLink` `.breadcrumbSep` `.heroContent` `.heroBadges` `.codeBadge` `.deadlineBadge` `.deadlineDot` `.sectionBadge` `.heroTitle` `.heroSubtitle` `.heroCtas` `.ctaDisabled` `.main` `.section` `.sectionHeader` `.sectionKicker` `.sectionTitle` `.sectionAside` `.twoCol`

### DocRefCard.module.css
`.card` `.headerStrip` `.body` `.code` `.typeShort` `.spacer` `.divider` `.footer` `.version` `.statusDot`

### KPIStrip.module.css
`.strip` `.cell` `.cellBorder` `.kicker` `.value` `.valueDanger` `.progressTrack` `.progressFill` `.sub`

### DonutCard.module.css
`.card` `.kicker` `.body` `.donutWrap` `.donutSvg` `.donutCenter` `.donutPct` `.donutPctSymbol` `.donutLabel` `.legend` `.legendRow` `.legendDot` `.legendDotAck` `.legendDotRead` `.legendDotPending` `.legendName` `.legendCount` `.legendPct` `.footer` `.footerMeta` `.footerWarning` `.warningDot`

### DistributionFacts.module.css
`.card` `.header` `.kicker` `.ctaDisabled` `.list` `.factRow` `.factRowBorder` `.iconBox` `.iconBoxWarning` `.factBody` `.factLabel` `.factValue` `.factValueMono` `.factHint`

### CoverageByArea.module.css
`.card` `.header` `.headerNote` `.legend` `.legendChip` `.legendSwatch` `.legendSwatchAck` `.legendSwatchRead` `.legendSwatchGoal` `.rows` `.areaRow` `.areaLabel` `.areaName` `.areaSub` `.barWrap` `.barTrack` `.barRead` `.barAck` `.goalMarker` `.pctWrap` `.pctValue` `.pctLabel`

### TimelineCard.module.css
`.card` `.header` `.subtitle` `.legend` `.legendChip` `.legendLine` `.legendDash` `.svg`

### RecipientsCard.module.css
`.card` `.filterBar` `.ctaDisabled` `.headerRow` `.colCheck` `.colRecipient` `.colArea` `.colStatus` `.colLastActivity` `.colWhen` `.colActions` `.checkbox` `.recipientRow` `.recipientRowBorder` `.recipientInfo` `.recipientName` `.recipientRole` `.lastActivityDanger` `.statusPill` `.statusAck` `.statusRead` `.statusPending` `.statusOverdue` `.statusDot` `.actionBtn` `.paginationFooter` `.paginationCount` `.paginationNav` `.paginationBtn` `.paginationPages`

## tsc result (new errors only)

**Zero new TypeScript errors.** All errors in the tsc output are pre-existing in:
- `src/features/auth/__tests__/useAuthSession.returnTo.test.tsx`
- `src/features/documents/components/LibrarySidebar.tsx`
- `src/features/documents/pages/NewDocumentWizardPage.tsx`
- `src/features/documents/queries/useAreasQuery.ts`
- `src/features/shell/components/Rail.tsx`

## Semantic conflicts found

1. **TabBar used instead of inline tab buttons** — The design renders recipient filter tabs as custom inline `<button>` elements with badge counts. The instructions specify using the `TabBar` primitive instead. The skeleton uses `TabBar` with `activeKey="pending"` (static placeholder) and `onTabChange={() => undefined}`. No state added per Phase 3a rules. The count values are sourced from `RECIPIENT_TABS` in `distributionMeta.ts`.

2. **No `StatusPill` from components/ui** — Per instructions, recipient row status indicators are implemented as inline `<span>` elements with `.statusPill .statusAck/.statusRead/.statusPending/.statusOverdue` classes from `RecipientsCard.module.css`. Dynamic class lookup uses a `STATUS_CLASS` map (string Record) to avoid bracket notation on CSS module objects.

3. **CTAs rendered as `aria-disabled` buttons** — All hero CTAs and "Editar política" are `<button type="button" aria-disabled="true" title="Em breve">` per instructions, not using the `disabled` attribute.

4. **Icon.tsx extended** — 6 icon names were added (`calendar`, `clock`, `shield`, `mail`, `chevron-right`, `chevron-left`) to support design-specified icons not present in the original union. This is a side-effect of Phase 3a that affects a shared primitive.

5. **distributionMeta.ts was pre-existing** — The file existed from an earlier session stub with the full `MOCK_DISTRIBUTION` and `RECIPIENT_TABS` constants already defined. No modification was needed.
