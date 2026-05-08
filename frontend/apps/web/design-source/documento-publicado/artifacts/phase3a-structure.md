# Phase 3a Structure — documento-publicado

**Completed:** 2026-05-08

---

## Files produced

| File | Path |
|---|---|
| Page component | `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.tsx` |
| CSS Module | `frontend/apps/web/src/features/documents/pages/DocumentPublishedPage.module.css` |

## Phase 2 prerequisites added in this run

The worktree branch (`worktree-agent-af830ddd771fbcd74`) diverged from main before Phase 2 was committed there. The following Phase 2 artifacts were applied to this worktree as part of Phase 3a:

| Artifact | Path |
|---|---|
| `eye` icon added to IconName | `src/components/ui/Icon.tsx` |
| `getApprovalInstance` + types | `src/features/documents/api/documentsV2.ts` |
| `documentDetailMeta.ts` | `src/features/documents/lib/documentDetailMeta.ts` |
| `useDocumentDetailQuery.ts` | `src/features/documents/queries/useDocumentDetailQuery.ts` |
| `useApprovalInstanceQuery.ts` | `src/features/documents/queries/useApprovalInstanceQuery.ts` |

---

## Class-name mapping table

| Design concept | Module class | Notes |
|---|---|---|
| Page wrapper | `.root` | Top-level div |
| Hero section | `.hero` | `<header>` element |
| Hero grid pattern overlay | `.heroBg` | Decorative div inside hero |
| Hero two-column grid | `.heroGrid` | Card + content columns |
| Card column wrapper | `.heroCardWrap` | Positions the 3D DocCardMini |
| DocCardMini container | `.docCard` | CSS perspective/tilt applied in Phase 3b |
| DocCard top area | `.docCardHeader` | Area name stripe |
| DocCard body | `.docCardBody` | Code + type + spacer + divider + footer |
| DocCard code text | `.docCardCode` | Monospace code display |
| DocCard type text | `.docCardType` | Profile label |
| DocCard flex spacer | `.docCardSpacer` | Pushes footer to bottom |
| DocCard divider line | `.docCardDivider` | Horizontal rule |
| DocCard footer row | `.docCardFooter` | Version + dot |
| DocCard version label | `.docCardVersion` | "v3.2" |
| DocCard status dot | `.docCardDot` | Color dot |
| Hero right column | `.heroContent` | Badges + title + actions |
| Badges row | `.heroBadges` | CodeChip + status badge + type label |
| Code badge wrapper | `.codeChip` | Passed as className to CodeChip |
| Status + version badge | `.vigeenteBadge` | "v3.2 · vigente" pill |
| Status dot inside badge | `.vigeenteDot` | Green pulsing dot |
| Type label text | `.typeLabel` | "Procedimento" plain label |
| Document title | `.heroTitle` | `<h1>` |
| Action buttons row | `.heroActions` | Flex row |
| Primary action button | `.btnPrimary` | "Visualizar documento" |
| Secondary action button | `.btnSecondary` | "Iniciar revisão" |
| Ghost action button | `.btnGhost` | "Copiar link" |
| Breadcrumb nav | `.breadcrumb` | `<nav>` |
| Breadcrumb links | `.breadcrumbLink` | `<a>` ancestors |
| Breadcrumb separator icon | `.breadcrumbSep` | Chevron icon |
| Current breadcrumb item | `.breadcrumbCurrent` | `<span>` non-link |
| Main content area | `.content` | Below hero |
| KPI strip container | `.kpiStrip` | Horizontal strip |
| Single KPI cell | `.kpiCell` | "Versão atual" block |
| KPI label | `.kpiLabel` | "Versão atual" text |
| KPI value | `.kpiValue` | "v3.2" |
| KPI hint/sub-label | `.kpiHint` | "desde 12 mar" |
| Generic section | `.section` | `<section>` |
| Section header row | `.sectionHead` | Kicker + title row |
| Section kicker text | `.sectionKicker` | "01 · Sobre" |
| Section title | `.sectionTitle` | `<h2>` |
| About section layout | `.aboutLayout` | Card + deferred slots |
| About card container | `.aboutCard` | White card |
| Owner banner row | `.ownerBanner` | Avatar + name/meta |
| Owner info column | `.ownerInfo` | Name + meta text |
| Owner name text | `.ownerName` | Display name |
| Owner meta text | `.ownerMeta` | "publicou em ..." |
| Facts grid | `.factsGrid` | Tipo + Área rows |
| Fact row | `.factCell` | Icon + content |
| Fact icon column | `.factIcon` | Icon wrapper |
| Fact content column | `.factContent` | Label + value |
| Fact label | `.factLabel` | "Tipo", "Área" |
| Fact value | `.factValue` | Resolved label |
| Signoff card container | `.signoffCard` | White card |
| Signoff grid | `.signoffGrid` | Horizontal stage columns |
| Connector line | `.signoffConnector` | Horizontal line behind pins |
| Stage column | `.signoffStage` | Per-stage block |
| Stage pin (circle + icon) | `.signoffPin` | Check icon circle |
| Stage name text | `.signoffStageName` | "Revisão técnica" etc. |
| Signoff actor row | `.signoffActor` | Avatar + name |
| Signoff actor name | `.signoffActorName` | Display name |
| Signoff actor role | `.signoffActorRole` | "Coord. SSMA" |
| Signoff timestamp | `.signoffWhen` | "11 mar · 18:04" |
| Obsolete banner overlay | `.obsoleteBanner` | Fixed/absolute overlay |
| OBSOLETO stamp text | `.obsoleteStamp` | Rotated stamp |

---

## Conflicts found and resolved

| Conflict | Resolution |
|---|---|
| Prompt specified `size="xs"` for signoff actor Avatar, but `Avatar` only supports `'sm' \| 'md' \| 'lg'` | Changed to `size="sm"` — nearest valid size. Phase 3b CSS can reduce the rendered size via class override if needed. |
| `documento-publicado` design-source artifacts not present in worktree (worktree predates Phase 1/2 commits on main) | Read artifact content from main repo working tree; created artifacts dir in worktree. All Phase 2 prerequisites applied manually. |
| Page stub file existed (`DocumentPublishedPage.tsx`) | Replaced entirely per spec. |

---

## TSC result

Green — see Phase 3a commit.
