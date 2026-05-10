# Phase 1 — Map · template-editor

Date: 2026-05-10
Tier: **Heavy** (new chrome, multi-region; eigenpal canvas; toggleable side-panel)

Existing `TemplateAuthorPage.tsx:213-330` already implements the locked layout. Scope = cuts + chrome polish, not greenfield.

## 1.1 Backward primitive scan (reuse, don't build)

| Primitive | Path | Reuse here |
|---|---|---|
| `EditorChrome` + slots | `features/shared/components/editor-chrome/` | Top overlay (center+right+alert) — already wired |
| `VersionBadge` | same module | `REV01` chip in chrome center |
| `AutosaveStatus` | same module | Right slot |
| `StatusPill` | `components/ui/StatusPill` | Right of VersionBadge |
| `PlaceholderCatalogPanel` | `features/templates/PlaceholderCatalogPanel.tsx` | Left panel (after inner rail) |
| Inner rail `<aside class="rail">` pattern | `features/documents/pages/DocumentEditorPage.tsx:200` | Mirror class names + tokens |
| `useTemplateDraft` / `useTemplateAutosave` / `useTemplateSchemas` | `features/templates/hooks/` | Untouched |
| `submitForReview` / `fetchPlaceholderCatalog` | `features/templates/api/` | Untouched |
| `VersionActionPanel` | `features/templates/VersionActionPanel.tsx` | Below editor for in_review/approved/published |

No new shared primitives needed. No `components/ui/` additions. No `features/shared/` additions.

## 1.2 Forward placement decision

All new code stays inside `features/templates/`:

- `pages/TemplateEditorPage.tsx` — replaces `TemplateAuthorPage.tsx` (rename + restructure; old file deleted same change, no shim)
- `pages/styles/TemplateEditorPage.module.css` — extracted from current `TemplateAuthorPage.module.css`, restyled to mirror `DocumentEditorPage.module.css` rail/canvas tokens

`PlaceholderCatalogPanel` keeps its current path (already domain-scoped).

## 1.3 Component tree (post-rebuild)

```
TemplateEditorPage
├─ <div.page>
│  └─ <div.body>
│     ├─ <aside.rail>                   ← inner rail (48px)
│     │  ├─ railBackBtn (← Voltar)
│     │  ├─ railDivider
│     │  ├─ railBtn[variables] (active by default, toggles panel)
│     │  └─ railBtn[outline]   (toggles TemplateOutlinePanel, reads eigenpal agent)
│     │
│     ├─ {leftActive === 'variables' && <PlaceholderCatalogPanel />}
│     ├─ {leftActive === 'outline'   && <TemplateOutlinePanel headings={readHeadings(editorRef)} />}
│     │
│     └─ <main.canvas>
│        └─ EditorChrome
│           ├─ center: brand-chip · name · VersionBadge · StatusPill
│           ├─ right:  AutosaveStatus + Importar .docx + Submeter para revisão
│           └─ children: <DocxEditor> (eigenpal native toolbar)
│
└─ {nonDraft && <VersionActionPanel>}
```

Cuts vs current `TemplateAuthorPage`:
- Drop `IconLayout` / `IconImage` / `IconSearch` rail items (`soon` flag — never shipping in this rebuild). Removes 3 disabled buttons + 1 divider.
- Drop English "Solicitar Revisão" label → match doc editor "Submeter para revisão".
- Replace ad-hoc inline-styled `role="alert"` colors in `alert` slot with token-driven CSS Module classes.

Defers (rail icons, with backlog rows):
- outline / TOC → `wiki/backlog/template-editor.md:outline`
- versions panel → `wiki/backlog/template-editor.md:version-history`
- comments → `wiki/backlog/template-editor.md:comments`

## 1.4 Status / enum SSOT

`StatusPill` already SSOT for status display. No new status meta module — templates reuse `DocumentStatus` (`'draft' | 'in_review' | 'approved' | 'published' | 'rejected'`). Already imported from `components/ui`.

## 1.5 State design

| State | Owner | Persistence |
|---|---|---|
| `leftActive` (which panel open) | local `useState` | none — defaults `'variables'` open. Future lazy-init from `localStorage` if user wants it persisted (defer). |
| `submitting` / `submitErr` | local `useState` | session only |
| `importing` / `importErr` | local `useState` | session only |
| `liveVersion` (post-submit) | local `useState` | session only |
| `localSchemas` / `catalog` / `detectedVariables` | local | feeds `PlaceholderCatalogPanel` |
| `draft.template` / `draft.docxBytes` | `useTemplateDraft` (TanStack via hook) | server |
| autosave queue | `useTemplateAutosave` | server |

No new query keys, no new server-state surface.

## 1.6 Backend contract

Zero new endpoints. All existing:
- `useTemplateDraft` / `useTemplateAutosave` / `useTemplateSchemas` (existing hooks)
- `submitForReview` (`POST /api/v2/templates/:id/versions/:n/submit-for-review`)
- `fetchPlaceholderCatalog` (existing)
- `autosave.importDocx` (existing autosave hook method)

If any 4xx surfaces during smoke, surface via `resolveErrorMessage` per `concepts/error-ux.md`. Today the page uses raw `setSubmitErr(message)` — upgrade to `ApiError` + `resolveErrorMessage` in Phase 3c.

## 1.7 Tier classification

**Heavy** — locked. Triggers fired:
- New responsive layout (rail + side-panel + canvas regions; eigenpal CSS overrides)
- Multi-region (rail / panel / chrome top / canvas / version-action footer)
- Eigenpal CSS coexistence (token-driven overrides — verify via parity-diff)

Light tier rejected: layout shift not single-column.

## 1.8 User checkpoint (resolved 2026-05-10)

1. ✅ Rename `TemplateAuthorPage.tsx` → `TemplateEditorPage.tsx` in place. No shim, no legacy re-export.
2. ✅ Drop `layout` / `media` / `search` rail icons entirely. No "Em breve" placeholder.
3. ✅ **Keep outline icon — eigenpal native.** Eigenpal exposes document headings via `editorRef.current.getAgent()` (same surface as `getVariables()`). Outline icon toggles a `TemplateOutlinePanel` that reads headings from the eigenpal agent. No backend dependency, no new API. Treat as inner-rail siblings:
   - `back` (← Voltar) — navigation
   - `variables` (active by default) — `PlaceholderCatalogPanel`
   - `outline` — `TemplateOutlinePanel` (Phase 3c implementation; reads `agent.getHeadings()` or equivalent)
   Only one panel open at a time (`leftActive` state, exclusive toggle).
4. ✅ pt-BR everywhere: "Submeter para revisão" / "Importando…" / "Salvo" / "Salvando…" / "Falha ao salvar". Mirrors `DocumentEditorPage` copy. No mixed English.

### Outline panel — minimal contract (locked here, built Phase 3c)

```ts
// features/templates/TemplateOutlinePanel.tsx
type Heading = { id: string; level: 1|2|3|4|5|6; text: string };
function readHeadings(editorRef): Heading[] {
  const agent = editorRef.current?.getAgent?.();
  // probe agent surface; fall back to empty array if absent.
  // Phase 3c spike: confirm exact eigenpal method (getHeadings? getOutline?).
  return agent?.getHeadings?.() ?? [];
}
```

If eigenpal API surface doesn't expose headings → defer outline icon, mark backlog. Verify in Phase 2 pre-flight (single-call probe before committing UI).
