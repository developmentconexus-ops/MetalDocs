# Phase 0 — Audit

**Screen:** Templates List (`/templates`)
**Design source:** `design-source/templates/templates.html`
**User confirmation:** 2026-05-07

## Keep/Cut/Defer Table

| Element | Maps to | Verdict | Reason |
|---|---|---|---|
| Header kicker "Templates" | Static label | Keep | Navigation context |
| Display heading "Layouts reutilizáveis" | Static label | Keep | Page identity |
| Subtitle description | Static label | Keep | UX context |
| "Novo template" button | TemplateCreateDialog (exists) | Keep | Functional, already wired |
| Tab bar: Todos | All templates filter | Keep | Core navigation |
| Tab bar: Publicados | status=Published filter | Keep | Real status |
| Tab bar: Em revisão | _(no matching status)_ | **Cut** | Replaced with "Arquivados" |
| Tab bar: Rascunhos | status=Draft filter | Keep | Real status |
| Card: mini doc preview | Decorative affordance | Keep | Visual recognition |
| Card: Status badge | Template version status | Keep | Draft/Published/Archived |
| Card: Version pill | `latest_version` | Keep | Data-backed |
| Card: Title | `template.name` | Keep | Data-backed |
| Card: Description | _(no DB column)_ | **Defer** | Backend TODO |
| Card: Bound profile pills | _(no API)_ | **Cut** | Not needed |
| Card: "+N" overflow pill | _(depends on profiles)_ | **Cut** | Profiles cut |
| Card: "Não vinculado" fallback | _(depends on profiles)_ | **Cut** | Profiles cut |
| Card: Author avatar + name | `created_by` user | Keep | Data-backed |
| Card: Relative timestamp | `updated_at` | Keep | Data-backed |
| Card: Divider | Visual separator | Keep | Between content and footer |

## Status mapping

- Design `frozen` → Published (green badge)
- Design `draft` → Draft (amber badge)
- Added: Archived (gray badge)
- Cut: `review` (doesn't exist in system)
