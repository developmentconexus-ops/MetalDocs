# Documento Publicado — Design Notes

**Slug:** `documento-publicado`
**Design file:** `documento-publicado.html`
**JSX source:** `onda1-v5.jsx` → `PublicadoV5`
**Route:** `/documents/:id` (published state)

## Status
- [x] Phase 0 Audit complete — 2026-05-08

## Keep / Cut / Defer (confirmed 2026-05-08)

### KEEP
- HeroC5: breadcrumb, DocCardSmall (CSS decorative), code badge, status badge, version badge, type label, h1 title
- Action buttons: "Visualizar documento" (→ `/documents-v2/:id`), "Iniciar revisão" (RBAC-gated), "Copiar link" (clipboard)
- KPI: Versão atual (`revision_version`)
- AboutCard: owner banner (`created_by` + `created_at`), Fact: Tipo, Fact: Área
- SignoffPipeline (`GET /api/v1/documents/:id/approval-instance` → stages + signoffs with display names)
- ObsoleteBanner (status === 'obsolete')

### CUT (no backend / no domain concept)
- Subtitle/description paragraph
- "Contatar" button
- KPI: Próxima revisão, KPI: Páginas + size
- Facts: Vigente desde, Próxima revisão, Tamanho
- Version tags (major/minor/patch)
- Diff stats
- Comments attachment button

### DEFER (backend TBD)
- "Baixar PDF" button (no PDF endpoint)
- KPI: Cobertura % (no fanout API)
- CoverageCard (no fanout API)
- AuditCard + ISOSeal + Fact: Selo ISO (`values_hash` not in API response — defer, no backend change)
- VersionTimeline (no revision list endpoint)
- RelatedGrid (no relationship model)
- CommentsCard (ProseMirror JSON storage/rendering needs architectural brainstorm)

## Design decisions
- DocCardSmall tilt: pure CSS `perspective` + `rotateX/Y` — keep as-is, no domain behavior
- SignoffPipeline: `ActorUserID` field in API response actually contains display name snapshot (confirmed in handler code)
- ObsoleteBanner: show when `document.status === 'obsolete'` OR when controlled document status is obsolete/superseded
- "Iniciar revisão": disabled for non-authors via RBAC; endpoint is `POST /api/v1/controlled-documents/:id/revisions`
