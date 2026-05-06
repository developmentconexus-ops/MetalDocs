# Module: templates-v2

> **Last verified:** 2026-05-06
> **Status:** Stub. Fill in API surface, schema rules, and lifecycle invariants when refactor stabilizes.
> **Scope:** Template authoring, versioning, approval, publishing.
> **Out of scope:** Document fill-in (see `modules/documents.md`), eigenpal editor wiring (see `modules/editor-ui-eigenpal.md`), toolbar overlay + eigenpal CSS overrides (see `modules/editor-chrome.md`).
> **Key files:**
> - `internal/modules/templates_v2/` — backend module
> - `frontend/apps/web/src/features/templates/TemplatesListPage.tsx` — list
> - `frontend/apps/web/src/features/templates/TemplateCreateDialog.tsx` — new template dialog
> - `frontend/apps/web/src/features/templates/TemplateAuthorPage.tsx` — eigenpal author; consumes `EditorChrome` for toolbar overlay
> - `frontend/apps/web/src/features/templates/VersionActionPanel.tsx` — lifecycle transitions

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
