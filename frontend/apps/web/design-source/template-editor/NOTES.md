# template-editor — Design notes

Functional base: `frontend/apps/web/src/features/templates/TemplateAuthorPage.tsx` (legacy template editor, route `/templates/:templateId/versions/:versionNum`).

Visual reference: `template-editor.html` (1194 HTML + 1169 CSS) — Swiss/minimal chrome with topbar + left/right collapsible rails + centered doc canvas. Topbar pattern also informed by `design-source/editor/`.

Tier: **Heavy** — new chrome layout, multi-region, breakpoints, novel rail-collapse interaction.

User intent (paraphrase): keep good legacy chrome, prominent **placeholders left bar** (existing `PlaceholderCatalogPanel`). Steal topbar pattern from `editor/` design.

## Conflict to resolve up-front

**Eigenpal `DocxEditor` ships its own toolbar** (Bold/Italic/Align/Lists/etc.) built into the canvas. The design HTML draws a separate React-driven toolbar above the page. We CANNOT have two toolbars. Decision (defer to user): hide eigenpal toolbar via prop and reimplement controls bound to eigenpal agent commands, OR keep eigenpal toolbar and drop the design toolbar entirely. Eigenpal API exposure unknown without checking — flag as Phase 1 question.
