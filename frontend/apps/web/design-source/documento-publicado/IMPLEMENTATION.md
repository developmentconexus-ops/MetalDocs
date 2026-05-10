# Documento Publicado — Implementation Worksheet

> **Slug:** `documento-publicado`
> **Owning feature:** `features/documents`
> **Target route:** `/documents/:id`
> **Reference:** `./documento-publicado.html` + `./NOTES.md`
> **Skill version:** 1.2
> **Started:** 2026-05-08
> **Completed:** —

---

## Open Questions Log

Append a row whenever you must pause for user input. Phase cannot pass while open rows for that phase exist.

| # | Phase | Question | User answer | Resolved |
|---|---|---|---|---|
| 1 | 0 | `GET /api/v2/documents/:id` exists (confirmed). It does NOT return `values_hash` (the ISO seal). Should we add `values_hash` to the backend SELECT (3-line change) or defer the AuditCard entirely? | Defer AuditCard. No backend change. | ✅ |
| 2 | 0 | Comments backend exists. `content` field is ProseMirror rich JSON (`unknown[]`). Render as extracted plain text (write `extractPlainText` util) or use eigenpal ReadonlyEditor for rich rendering? | Defer CommentsCard entirely — needs brainstorm on storage structure first. Goes to backlog. | ✅ |
| 3 | 3b | Phase 3b screenshot triple-diff: `mcp__Claude_Preview__preview_screenshot` returns inline image data, no file save to disk. Skip artifact PNG files? | Yes — parity verified via numerical computed-style diff (`parity-diff.md`) which is the load-bearing evidence per Skill v1.2. Screenshot files documented as deferred until preview tool gains `save_to_disk`. | ✅ |

---

## Phase 0 — Audit (HARD GATE)

### 0.1 Element-by-element audit

| Element (HTML region) | Maps to (state / role / persona / data) | Keep / Cut / Defer | Reason |
|---|---|---|---|
| **Breadcrumb** (Biblioteca → Area → Type → Code) | Real navigation hierarchy via router | Keep | Maps to actual LibraryPage + document fields |
| **DocCardSmall** (tilted 3D miniature) | Decorative identity anchor (code + area + type) | Keep (simplified) | Visual anchor worth keeping; 3D tilt is pure CSS, no domain behavior |
| **Code badge** (PR-EHS-014) | `controlled_document.code` | Keep | Real field |
| **"vigente" status badge** | `document.status === 'published'` | Keep | Real state; hidden when obsolete |
| **Version badge** (v3.2) | `document.revision_label` or `revision_number` | Keep | Real field |
| **Type label** (Procedimento) | `document.profile_code` → profile label | Keep | Real field |
| **h1 title** | `document.name` | Keep | Real field |
| **Subtitle/description paragraph** | No current field | Defer | No description field exists today; open Q5 |
| **"Visualizar documento" button** | Navigate to `/documents-v2/:id` (editor/viewer) | Keep | Real route exists |
| **"Baixar PDF" button** | PDF download endpoint | Defer | Fanout PDF may exist; backend TBD |
| **"Iniciar revisão" button** | `POST /api/v2/controlled-documents/:id/revisions` | Keep (RBAC-gated) | Real endpoint; disabled for non-authors |
| **"Copiar link" button** | `navigator.clipboard.writeText(window.location.href)` | Keep | Client-side only |
| **KPI — Versão atual** | `document.revision_number` / `revision_label` | Keep | Real field |
| **KPI — Cobertura %** | Fanout read coverage | Defer | Needs fanout data API; no endpoint today |
| **KPI — Próxima revisão** | No review date field | Cut | Field does not exist in current data model |
| **KPI — Páginas + size** | No page count / file size field | Cut | Field does not exist |
| **AboutCard — owner banner** | `document.submitted_by` + `document.published_at` | Keep (partial) | Fields exist; approver role text is mock |
| **AboutCard — "Contatar" button** | No mail flow | Cut | No mailto/contact flow defined |
| **Fact — Tipo** | `profile_code` → profile label | Keep | Real field |
| **Fact — Área** | `area_code` → area label | Keep | Real field |
| **Fact — Vigente desde** | No effectiveAt field | Cut | Substitute with `published_at` if kept, else cut |
| **Fact — Próxima revisão** | No review date field | Cut | Field does not exist |
| **Fact — Tamanho (pages + MB)** | No size fields | Cut | Fields do not exist |
| **Fact — Selo ISO (hash)** | `document.hash` (frozen hash) | Keep | Real field on published/frozen docs |
| **CoverageCard** (84.5% + fanout link) | Fanout module, read coverage | Defer | Needs fanout API |
| **AuditCard** (ISO hash + tooltip) | `document.hash` + full SHA-256 | Keep | Real field; tooltip is client-only |
| **SignoffPipeline** (stepper) | Approval signoffs for this version | Defer | Needs signoff history endpoint (Q2) |
| **VersionTimeline** (interactive pins) | Revision history for the controlled doc | Defer | Needs revision list endpoint (Q3) |
| **Version tag** (major/minor/patch) | No tag field in revision model | Cut | No such classification in data model |
| **Diff stats** (+added ~modified -removed) | No diff tracking | Cut | No diff data available |
| **RelatedGrid** (related documents) | No related-documents feature | Defer | No backend relationship model |
| **CommentsCard** (thread + reply box) | No comments feature | Cut | No comments backend whatsoever |
| **ObsoleteBanner** ("OBSOLETO" stamp) | `document.status === 'obsolete'` | Keep | Real state |
| **ISOSeal** (hash tooltip on hover) | `document.hash` SHA-256 expanded | Keep | Client-only hover UX |

### 0.2 Cut list (awaiting user confirmation)

**CUT (no backend / no domain concept):**
- "Contatar" button on owner banner
- KPI: Próxima revisão (no field)
- KPI: Páginas + size (no field)
- Fact: Vigente desde (no field — substitute `published_at`)
- Fact: Próxima revisão (no field)
- Fact: Tamanho (no field)
- Version tags (major/minor/patch)
- Diff stats (+added ~modified -removed)
- CommentsCard (no backend)

**DEFER (backend TBD, open questions):**
- Subtitle/description paragraph (Q5)
- "Baixar PDF" button (fanout PDF endpoint TBD)
- KPI: Cobertura % (fanout API TBD)
- CoverageCard (fanout API TBD)
- SignoffPipeline (signoff history endpoint TBD — Q2)
- VersionTimeline (revision list endpoint TBD — Q3)
- RelatedGrid (no relationship model)

- [ ] User reviewed cut list
- [ ] Cuts recorded in NOTES.md

---

## Phase 1 — Map (HARD GATE)

_(To be filled after Phase 0 user sign-off)_

---

## Phase 2 — Pre-flight (advisory)

_(Subagent — after Phase 1)_

---

## Phase 3a — Structure mirror (HARD GATE)

_(Subagent — after Phase 2)_

---

## Phase 3b — Style port (HARD GATE)

_(Subagent — after Phase 3a user sign-off)_

---

## Phase 3c — State wiring (advisory)

_(Subagent — after Phase 3b user approval)_

---

## Phase 4 — Verify (HARD GATE)

_(After Phase 3c)_

---

## Phase 5 — Document (advisory)

_(After Phase 4)_
