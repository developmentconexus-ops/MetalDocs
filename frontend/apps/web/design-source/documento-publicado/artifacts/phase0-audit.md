# Phase 0 Audit — documento-publicado

**Completed:** 2026-05-08  
**User sign-off:** ✅ 2026-05-08 (via chat — Q1 defer, Q2 defer/backlog)

---

## Keep / Cut / Defer table

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| **Breadcrumb** (Biblioteca → Area → Type → Code) | Router + `area_code`, `profile_code`, `code` | Keep | Real nav hierarchy |
| **DocCardSmall** (tilted 3D card) | Decorative identity anchor | Keep (simplified) | Pure CSS 3D tilt; no domain behavior |
| **Code badge** (e.g. PR-EHS-014) | `controlled_document.code` | Keep | Real field via CD lookup |
| **"vigente" status badge** | `document.status === 'published'` | Keep | Real state |
| **Version badge** (e.g. v3.2) | `document.revision_version` | Keep | Real field |
| **Type label** (e.g. Procedimento) | `document.profile_code_snapshot` → profile label | Keep | Real field |
| **h1 title** | `document.name` | Keep | Real field |
| **Subtitle/description paragraph** | No description field in model | Cut | Field doesn't exist |
| **"Visualizar documento" button** | Navigate to `/documents-v2/:id` | Keep | Real route exists |
| **"Baixar PDF" button** | PDF download endpoint | Defer | No endpoint today |
| **"Iniciar revisão" button** | `POST /api/v2/controlled-documents/:id/revisions` | Keep (RBAC-gated) | Real endpoint; disabled for non-authors |
| **"Copiar link" button** | `navigator.clipboard.writeText(window.location.href)` | Keep | Client-only |
| **KPI — Versão atual** | `document.revision_version` | Keep | Real field |
| **KPI — Cobertura %** | Fanout coverage API | Defer | No endpoint |
| **KPI — Próxima revisão** | No review date field | Cut | Field doesn't exist |
| **KPI — Páginas + size** | No page count / size field | Cut | Field doesn't exist |
| **AboutCard — owner banner** | `document.created_by` + `document.created_at` | Keep (partial) | Fields exist |
| **AboutCard — "Contatar" button** | No contact flow | Cut | No mailto/contact flow defined |
| **Fact — Tipo** | `profile_code_snapshot` → label | Keep | Real field |
| **Fact — Área** | `process_area_code_snapshot` → label | Keep | Real field |
| **Fact — Vigente desde** | No effectiveAt field | Cut | Field doesn't exist; `created_at` is poor substitute |
| **Fact — Próxima revisão** | No review date field | Cut | Field doesn't exist |
| **Fact — Tamanho** | No size fields | Cut | Field doesn't exist |
| **Fact — Selo ISO (hash)** | `values_hash` — NOT in API response | Defer | Deferred with AuditCard (Q1) |
| **CoverageCard** | Fanout coverage API | Defer | No endpoint |
| **AuditCard** (ISO hash + tooltip) | `values_hash` — not returned by API | Defer | No backend change approved; goes to backlog |
| **SignoffPipeline** (stepper) | `GET /api/v2/documents/:id/approval-instance` → `stages[]` + `signoffs[]` | Keep | Real endpoint; display names included |
| **VersionTimeline** (interactive pins) | Revision history | Defer | No revision list endpoint exists |
| **Version tags** (major/minor/patch) | No tag field | Cut | No such classification in data model |
| **Diff stats** (+added ~modified) | No diff tracking | Cut | No diff data available |
| **RelatedGrid** | No related-documents backend | Defer | No relationship model |
| **CommentsCard** (thread + reply) | `GET/POST /api/v2/documents/:id/comments` — content is ProseMirror JSON | Defer | Needs brainstorm on storage structure + rendering approach; goes to backlog |
| **Comments attachment button** | No attachment flow | Cut | No backend support |
| **ObsoleteBanner** ("OBSOLETO" stamp) | `document.status === 'obsolete'` or `cd.status === 'obsolete'` | Keep | Real state |
| **ISOSeal** (hash tooltip on hover) | Tied to AuditCard `values_hash` | Defer | Deferred with AuditCard |

---

## Confirmed CUT list

- Subtitle/description paragraph (no field)
- "Contatar" button (no contact flow)
- KPI: Próxima revisão (no field)
- KPI: Páginas + size (no fields)
- Fact: Vigente desde (no field)
- Fact: Próxima revisão (no field)
- Fact: Tamanho (no field)
- Version tags major/minor/patch (no classification in model)
- Diff stats (no diff tracking)
- Comments attachment button (no backend)

## Confirmed DEFER list

- "Baixar PDF" button (no PDF endpoint)
- KPI: Cobertura % (no fanout API)
- CoverageCard (no fanout API)
- AuditCard + ISOSeal (Q1: `values_hash` not in API; no backend change approved)
- Fact: Selo ISO (tied to AuditCard)
- VersionTimeline (no revision list endpoint)
- RelatedGrid (no relationship model)
- CommentsCard (Q2: needs storage brainstorm; ProseMirror JSON rendering TBD)

## KEEP list (ships in this implementation)

- Full HeroC5: breadcrumb, DocCardSmall, code badge, status badge, version badge, type label, h1 title
- Action buttons: Visualizar documento, Iniciar revisão (RBAC-gated), Copiar link
- KPI: Versão atual
- AboutCard: owner banner (created_by + created_at), Tipo fact, Área fact
- SignoffPipeline (approval stages + signoffs with display names)
- ObsoleteBanner (conditional on status)

---

## Open questions resolved

| # | Question | Answer |
|---|---|---|
| Q1 | Add `values_hash` to backend or defer AuditCard? | Defer AuditCard. No backend change. |
| Q2 | CommentsCard render strategy? | Defer entirely. Backlog: needs storage/architecture brainstorm. |
