# Phase 0 — Audit · template-editor

Date: 2026-05-10
Tier: Heavy
Functional base: `TemplateAuthorPage.tsx` (381 lines, DocxEditor + autosave + schemas + catalog).
Visual ref: `template-editor.html` + chrome inspiration from `design-source/editor/`.

Every UI element classified Keep / Cut / Defer. RBAC + state map per `wiki/concepts/design-workflow-audit.md`.

## Topbar

| Element | Map | Decision | Notes |
|---|---|---|---|
| brand-dot + brand-name | branding | Keep | Mirrors AppShell brand; reuse existing `Rail` brand or inline 24px square |
| hamburger | sidebar toggle | Cut | No collapsible nav rail — Rail is fixed in this app shell |
| breadcrumbs (Templates / `<name>`) | nav | Keep | Already present in `EditorChrome.center` via `draft.template?.name` |
| version-badge `v1` | state | Keep | Already shown via `VersionBadge` (`REV01`); replace with `v1` prefix to match design |
| status-pill `Draft` | state | Keep | `StatusPill` component already wired |
| saved-indicator | autosave | Keep | `AutosaveStatus` component already wired |
| history button | versions | **Defer** | Backlog `template-editor:version-history` — no UI; right-rail "versions" panel covers this |
| share button | sharing | Cut | Not a real concept; templates publish through review flow |
| Submit-for-Review primary CTA | mutation | Keep | Existing `submitForReview` API + `handleSubmitForReview` |
| avatar | user | Keep | `useAuthStore` already returns `user.displayName` for initials |
| **Importar .docx** (legacy) | mutation | Keep | Not in design but real (autosave.importDocx); place next to Submit |

## Left rail icons

| Icon | Map | Decision | Notes |
|---|---|---|---|
| blocks (`B`) | block palette | Cut | No backend; eigenpal owns block insertion |
| layout | layout panel | Cut | "soon" placeholder in legacy — drop, no real feature shipping soon |
| image / media | media library | Cut | No media library backend |
| outline | TOC | **Defer** | Backlog `template-editor:outline` — eigenpal exposes headings; future panel |
| search (⌘F) | find | Cut | Use browser Ctrl+F |
| **variables** (NEW priority) | placeholder catalog | Keep | **Primary panel** — already implemented as `PlaceholderCatalogPanel`, expand to design's panel chrome |

Result: left rail = ONE icon (variables) + future outline. Single-icon rail does not justify the design's collapse-to-expanded chrome — collapse to a fixed expanded variables panel + variables icon for highlight.

## Right rail icons

| Icon | Map | Decision | Notes |
|---|---|---|---|
| inspector | block inspector | Cut | Eigenpal owns selection inspector internally; don't duplicate |
| variables | placeholder | Cut | Already on left (avoid duplicate) |
| comments | review comments | **Defer** | Backlog `template-editor:comments` — comments exist on documents, not templates yet |
| versions | version history | **Defer** | Backlog `template-editor:version-history` — surface via right panel, not right rail icon |

Result: right rail dropped entirely. No real right-rail features for templates today. `VersionActionPanel` (in_review/approved/published) stays as a bottom-of-page block, not a right rail.

## Center / canvas

| Element | Map | Decision | Notes |
|---|---|---|---|
| dochead bar | meta strip | Cut | Information already in topbar (name + version + status); double-strip is redundant |
| **toolbar** (undo/redo, fonts, B/I/U/S, color, align, lists, link, table, image, insert) | formatting | **Cut — KEEP eigenpal toolbar** | See "Conflict resolution" below |
| empty-page CTA | new doc CTA | Cut | Wizard flow handles "new template"; editor never lands empty |
| page (the doc) | DocxEditor | Keep | `<DocxEditor>` from `@eigenpal/docx-js-editor/react` |

### Conflict resolution: toolbar

Eigenpal `DocxEditor` ships its own toolbar (built-in Bold/Italic/Align/Lists). Design draws a separate React toolbar. Two toolbars stacked = bad UX. Options:

- **(A) Keep eigenpal toolbar, drop design toolbar.** Zero new code; visual parity sacrificed. ★ Chosen
- (B) Hide eigenpal toolbar (if prop exists), reimplement design toolbar bound to `editorRef.getAgent().exec(...)`. Heavy — requires reverse-engineering eigenpal command surface, ongoing maintenance against eigenpal upgrades. YAGNI.
- (C) Show both. Reject.

**Decision A.** New design's chrome (topbar + side panels) wraps eigenpal as-is. Document this in NOTES.md so the visual reviewer doesn't flag the design toolbar absence as a Major.

## Right block: VersionActionPanel

| Element | Map | Decision | Notes |
|---|---|---|---|
| in_review/approved/published actions | mutation | Keep | Existing component, render below editor for non-draft states |

## Blocks palette / Inspector panel

Both **Cut**. Both rely on a structured-block model that eigenpal owns internally — duplicating in MetalDocs would compete with eigenpal's selection UX. Eigenpal handles selection, inspector, formatting natively.

## Variables panel (the highlight)

Existing `PlaceholderCatalogPanel` already does:
- Lists catalog placeholders
- Marks detected (used in current doc) vs available
- Saves catalog tokens into schema

Design doesn't show a variables panel detail (design shows blocks + inspector instead). **Keep current PlaceholderCatalogPanel; restyle its chrome** (header bar with title + close, body padding) to match `.rail-panel-head` + `.rail-panel-body` patterns from design CSS. No data shape change.

Future enhancements (Defer): drag-into-doc, search, category grouping. Out of scope for this rebuild.

## Final scope

**Keep:** topbar (brand, breadcrumbs, version-badge, status-pill, saved-indicator, Submit-for-Review, Importar, avatar), left rail (variables icon only), variables panel (PlaceholderCatalogPanel restyled), eigenpal-owned canvas + toolbar, VersionActionPanel.

**Cut:** hamburger, share, history button, blocks palette, inspector panel, right rail entirely, dochead bar, empty-page CTA, design's toolbar, layout/media/search rail icons.

**Defer (with backlog rows):**
- `template-editor:version-history` — versions panel in right-rail or modal
- `template-editor:comments` — review comments thread
- `template-editor:outline` — TOC panel

## Backlog rows to write (Phase 1)

`wiki/backlog/template-editor.md` (new file):
- version-history
- comments
- outline
- design-toolbar-parity (decision A documented)

## User checkpoint

Confirm before Phase 1:

1. **Decision A** on toolbar (keep eigenpal's, drop design's)? Y/N
2. **Drop right rail entirely** (no comments/versions/inspector icons)? Y/N
3. **Variables stays as the only left-rail panel** (no blocks/layout/media/outline shipping with this rebuild)? Y/N
4. Topbar pattern (`design-source/editor/` chrome) acceptable as base, plus dropping hamburger + share + history? Y/N

---

## 2026-05-10 — Layout lock (post-preview audit of live document editor)

User directive: **mirror live document editor layout, narrower scope** — templates have placeholders, no doc-instance metadata.

Verified live `/documents-v2/<id>` shell:
- AppShell `Rail` 56px (global brand+nav+avatar) — already provided by `AppShell`, never re-implement here
- Inner `<aside class="rail">` 48px holding ONLY `railBackBtn` (`DocumentEditorPage.tsx:200`)
- Eigenpal canvas with built-in toolbar (Decision A confirmed — doc editor ships it as-is)
- `EditorMetaSidebar` 300px (Metadados / Revisões / Próximos aprovadores) — DOC-ONLY
- `EditorChrome` top overlay (`code-chip | name | VersionBadge | StatusPill` center; `AutosaveStatus + Submeter` right)

Template-editor layout (locked):
```
AppShell Rail 56px (global)                       [unchanged]
  Inner rail 48px:  ← Voltar + variables-toggle icon   [+ outline icon deferred]
    PlaceholderCatalogPanel ~280px (toggleable)
      Eigenpal canvas (built-in toolbar)
EditorChrome (top overlay):
  center: brand-chip + template name + VersionBadge + StatusPill
  right:  AutosaveStatus + Importar .docx + Submeter para revisão
  left:   unused (back button is in inner rail, not in chrome — same as DocumentEditorPage)
```

Existing `TemplateAuthorPage.tsx:215` already renders `<aside class="rail railLeft">` + `<PlaceholderCatalogPanel>` + `<main class="canvas">` — structure correct, only needs chrome restyle to match doc-editor visual idiom.

NO right sidebar. Drop design HTML's right rail + inspector + comments entirely (templates lack the data shape — see `TemplateResponse` vs `DocumentResponse + ApprovalInstance`).

NO design-source toolbar. Eigenpal's toolbar wins (Decision A) — confirmed by the live doc editor which ships the same eigenpal toolbar without duplication.

Inner-rail icons (final):
| Icon | Purpose | Status |
|---|---|---|
| ← Voltar | Back to templates list | Keep (mirror DocumentEditorPage) |
| variables (toggle) | Show/hide PlaceholderCatalogPanel | Keep |
| outline | TOC | **Defer** (`backlog/template-editor.md:outline`) |

Visual reviewer instruction: do NOT flag absence of design HTML's right rail / blocks panel / inspector / dochead bar / design toolbar — all intentional cuts grounded in document-editor parity.

User checkpoint Q1–Q4 implicitly answered Y by directive "mirror documents editor". Proceed to Phase 1.
