# Module: templates-v2

> **Last verified:** 2026-05-08
> **Status:** Partial. List screen complete (Phase 5). Author/versioning pages TBD.
> **Scope:** Template authoring, versioning, approval, publishing; Templates List screen (`/templates-v2`).
> **Out of scope:** Document fill-in (see `modules/documents.md`), eigenpal editor wiring (see `modules/editor-ui-eigenpal.md`), toolbar overlay + eigenpal CSS overrides (see `modules/editor-chrome.md`).
> **Key files:**
> - `internal/modules/templates_v2/` — backend module
> - `frontend/apps/web/src/features/templates/TemplatesListPage.tsx:1` — list page; real query, tab filter, loading/error/empty states, semantic clickable cards
> - `frontend/apps/web/src/features/templates/queries/useTemplatesQuery.ts:1` — TanStack Query hook wrapping `listTemplates()`; key `QK.templates.list()`
> - `frontend/apps/web/src/features/templates/components/TemplateCard.tsx:1` — card primitive (MiniDocPreview + StatusPill + CodeChip + Avatar); `role="button"` + keyboard handler
> - `frontend/apps/web/src/features/templates/components/MiniDocPreview.tsx:1` — decorative A4 thumbnail (8 placeholder lines)
> - `frontend/apps/web/src/components/ui/WorkspaceHeroHeader.tsx:13` — `tone?: "banner" | "flat"` prop; `.headerFlat` strips card chrome (transparent bg, no border-bottom, padding 0)
> - `frontend/apps/web/src/components/ui/TabBar.tsx:17` — roving tabIndex (`tabIndex={isActive ? 0 : -1}`), Arrow/Home/End keyboard nav per WAI-ARIA tablist spec
> - `frontend/apps/web/src/features/templates/TemplateCreateDialog.tsx` — new template dialog
> - `frontend/apps/web/src/features/templates/TemplateAuthorPage.tsx` — eigenpal author; consumes `EditorChrome` for toolbar overlay
> - `frontend/apps/web/src/features/templates/VersionActionPanel.tsx` — lifecycle transitions
> - `frontend/apps/web/design-source/templates/artifacts/` — phase 0–5 implementation artifacts

## Templates List screen

**Route:** `/templates-v2`
**Page:** `frontend/apps/web/src/features/templates/TemplatesListPage.tsx:29`

### Status derivation rule

```ts
status = dto.archived_at
  ? "archived"
  : dto.published_version_id
  ? "published"
  : "draft"
```

Source of truth: `TemplatesListPage.tsx:37–41`. No separate status field on `TemplateDTO` — derived client-side.

### WorkspaceHeroHeader tone

The list page uses `tone="flat"` on `WorkspaceHeroHeader` — hero typography without card chrome (transparent background, no `border-bottom`, `padding: 0`). Use `tone="banner"` (default) for pages that need the card chrome (Documents Library, future Operations Center).

`tone` prop added at `WorkspaceHeroHeader.tsx:13`. CSS variant `.headerFlat` in `WorkspaceHeroHeader.module.css`.

### TabBar a11y

`TabBar` (`TabBar.tsx:17`) implements WAI-ARIA `tablist` pattern:
- `tabIndex={isActive ? 0 : -1}` (roving tabIndex)
- `ArrowLeft` / `ArrowRight` / `Home` / `End` keyboard navigation
- `aria-selected` on each `role="tab"` button

### Card grid layout

`TemplatesListPage.module.css` — `padding: 24px 28px`, gap `var(--sp-6)`, responsive 3-col grid (3 columns ≥1440px, 2 at ≥1024px, 1 at 375px).

### formatRelative helper

`formatRelative` (inlined at `TemplatesListPage.tsx:17`) formats ISO timestamps to Portuguese relative strings (hoje / ontem / N dias atrás / N meses atrás / N anos atrás). Currently uses `dto.created_at` because `updated_at` is absent from `TemplateDTO` — see backlog. Promote to `lib/utils/` when a second caller appears.

---

## Lifecycle

```
draft → in_review → approved → published
```

- **draft**: editable by author. Can iterate freely.
- **in_review**: locked content. Awaiting approver.
- **approved**: signed off but not yet usable downstream.
- **published**: selectable when creating documents. Immutable.

Re-authoring → create a new version (e.g. `v2 draft`); previous published version remains until replaced.

## Domain rules

- One template can be the default for multiple profiles (taxonomy binding, see `modules/taxonomy.md`).
- Only **published** versions show up in the document creation wizard.
- ISO segregation: author of a version cannot be its approver.

## API surface

TBD — extract from `internal/modules/templates_v2/transport/` (or equivalent).

## Editor chrome

`TemplateAuthorPage` wraps `MetalDocsEditor` inside `<EditorChrome>`, passing:
- `left` — back button + sidebar toggle icons
- `center` — template title + `VersionBadge` + `StatusPill`
- `right` — `AutosaveStatus` + submit action

Eigenpal CSS overrides (wine formatting bar, compact title bar, gradient scrollbar) live in `EditorChrome.module.css`, not in `TemplateAuthorPage.module.css`. The page CSS module covers only rails, side panel, and canvas layout (~155 lines).

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Steps 2–4
- [workflows/template-authoring.md](../workflows/template-authoring.md) (TBD)
- [concepts/placeholders.md](../concepts/placeholders.md)
- [modules/editor-ui-eigenpal.md](editor-ui-eigenpal.md)
- [modules/editor-chrome.md](editor-chrome.md) — toolbar overlay primitive
