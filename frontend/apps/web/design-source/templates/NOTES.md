# Templates List Screen

**Route:** `/templates`
**Feature:** `features/templates/`
**Design:** `design-source/templates/templates.html` (renders `Templates` from `screens-2.jsx`)

## Layout

- FrameRail shell (Rail + Toolbar breadcrumb "Templates")
- Card-grid list (3 columns) with header + tab filter + cards

## Statuses (mapped from design mock)

| Design mock | Real system status | Badge |
|---|---|---|
| `frozen` | Published | success/green |
| `review` | _(cut — doesn't exist)_ | — |
| `draft` | Draft | warning/amber |
| _(not in mock)_ | Archived | neutral/gray |

## Tab bar

Todos · Publicados · Rascunhos · Arquivados (counts from API response)

## Cut list (confirmed 2026-05-07)

- **Bound profile pills** — no API endpoint returns this association; not needed on this screen.
- **"Em revisão" tab** — status doesn't exist in our template versioning. Replaced with "Arquivados".
- **Card description text** — no `description` column in templates table yet. Deferred to backend reformulation (TODO).

## Keep

- Header: kicker + display heading + subtitle + "Novo template" button
- Tab bar with counts (4 real statuses)
- Card grid: mini doc preview, status badge, version pill, title, author avatar + name, relative timestamp
- Empty state, loading state, error state

## TODO (backend, not this PR)

- Add `description` column to templates table
- Return template descriptions in list API response
- Potentially return bound profiles association

## Integration audit summary (2026-05-13)

- Runtime/auth/target-route gate passes for `GET /api/v1/templates`.
- Contract surfaces exist for templates runtime, OpenAPI, generated backend API, generated frontend types, and `api/templatesV2.ts`.
- The list screen is partly real already: fetch, tabs, cards, empty/loading/error states, and wizard handoff are implementable now.
- Two backend/API gaps remain for design parity: real `updated_at` and author display name instead of raw `created_by`.
- The feature still contains a legacy parallel API layer in `api/templates.ts`; keep Plan 12.1 on `templatesV2.ts` and only clean legacy paths if the touched screen flow still depends on them.
- Frontend route wiring needs attention before trusting the screen boundary: `routes.tsx` currently declares duplicate `"templates"` entries.

## Plan 12.1 implementation sync (2026-05-13)

Implemented now (screen-local):
- `templatesV2.listTemplates()` hardened for list envelope/meta tolerance on the real `/api/v1/templates` path
- tab counts behavior preserved on fetched data with explicit per-status counting
- card shell fallback values hardened for missing list fields
- mobile tab clipping fixed via horizontal tab scrolling on narrow widths

Verified unchanged:
- `/templates` route ownership in frontend is already single-source in `features/templates/routes.tsx`
- list -> wizard handoff remains `/templates/new`

Still prerequisite:
- true `updated_at` semantics from backend/API
- author display name (instead of raw `created_by`)

Still deferred:
- card description text
- bound profile pills
- extra `em revisao` tab semantics beyond the current 4-tab behavior
