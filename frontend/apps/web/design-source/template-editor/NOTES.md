# template-editor - Design notes

Functional base: `frontend/apps/web/src/features/templates/TemplateAuthorPage.tsx` (legacy template editor, route `/templates/:templateId/versions/:versionNum`).

Visual reference: `template-editor.html` (1194 HTML + 1169 CSS) - Swiss/minimal chrome with topbar + left/right collapsible rails + centered doc canvas. Topbar pattern also informed by `design-source/editor/`.

Tier: **Heavy** - new chrome layout, multi-region, breakpoints, novel rail-collapse interaction.

User intent (paraphrase): keep good legacy chrome, prominent **placeholders left bar** (existing `PlaceholderCatalogPanel`). Steal topbar pattern from `editor/` design.

## Conflict to resolve up-front

**Eigenpal `DocxEditor` ships its own toolbar** (Bold/Italic/Align/Lists/etc.) built into the canvas. The design HTML draws a separate React-driven toolbar above the page. We cannot have two toolbars. The current Plan 12.5 integration audit classifies the custom design toolbar as an intentional won't-fix: keep eigenpal's toolbar and do not duplicate it in React.

## Integration audit summary (2026-05-17)

Required gates passed before classification:

- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates`
- `scripts/check-module-contract-sync.ps1 -Module templates`

Audit outcome:

- Ready now: inner rail/back navigation, variables + outline panels, editor chrome metadata/autosave/import/submit surface, eigenpal canvas, non-draft review footer
- Local cleanup required: placeholder catalog still has inline legacy styling and duplicate fetch ownership; editor-facing template wrappers/hooks still rely heavily on direct `fetch` instead of the canonical frontend client path
- Defer honestly: version history needs a template versions list endpoint; comments need a template comment model/endpoint; custom design toolbar remains an intentional won't-fix

Implementation slices proposed from the audit:

1. Screen chrome and panel polish
2. Template-editor API/hook normalization
3. Footer action cleanup and convergence-test rewrite
