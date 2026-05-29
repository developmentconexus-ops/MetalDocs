# QA Report — Templates List (qa/templates-list)

- Route(s): `/templates`   Page: `frontend/apps/web/src/features/templates/pages/TemplatesListRoutePage.tsx` → `TemplatesListPage.tsx`
- Owning module: `wiki/modules/templates.md`   QA class: screen
- Acting role(s): admin (dev seed)   Date: 2026-05-29
- CI: GH Actions billing-blocked → local-evidence gate per `wiki/quality/qa-operating-system.md`

## Gate results

- **Gate 0 — scope/startup/auth:** PASS. API on :8081 via `.\scripts\start-api.ps1`; frontend preview on :4173; dev login `admin / AdminMetalDocs123!`; `/templates` reachable. Module routes + contract match `wiki/modules/templates.md` Plan 12.4 truth.
- **Gate 1 — implementation truth:**
  - `corepack pnpm tsc --noEmit -p tsconfig.build.json` → exit 0 (zero errors).
  - `corepack pnpm exec vitest run src/features/templates` → 31 passed, 5 skipped (test files: 8 passed, 1 skipped). The skipped suite is `template-author-page-convergence.test.tsx` (pre-existing skip, unrelated).
- **Gate 3 — product QA (Preview-driven):** PASS for happy path, empty state per tab, tab switching, card navigation, "Novo template" CTA, refresh re-entry, history back. Error-state runtime drive deferred (see Defers).
- **Gate 5 — regression:** PASS. Re-ran tsc + vitest after fix. No regressions in touched surface.

## Findings (severity-ordered)

| # | Severity | Family | Finding | Disposition |
|---|---|---|---|---|
| F1 | HIGH | screen-local implementation | Empty-state per-tab copy rendered the raw English `TabKey` (`Nenhum template draft encontrado.` etc.) instead of Portuguese; broke i18n consistency of the whole page. | **Fixed** in `TemplatesListPage.tsx`: added `EMPTY_STATE_LABEL: Record<Exclude<TabKey,"all">,string>` and substituted in the empty-state JSX. Verified live: `Nenhum template publicado encontrado.` / `Nenhum template arquivado encontrado.` |
| F2 | MEDIUM | shared contract prerequisite | `created_by` is rendered raw as the username string (e.g. `admin`). Card UX implies a person display name + avatar but the contract only exposes the actor id. Requires a user-resolution endpoint or DTO enrichment. | **Bounded defer** — out of scope for screen QA; needs cross-module work (`auth/identity` + `templates` DTO). Logged here; no symptom patch on the screen. |
| F3 | LOW | screen-local implementation | Dead cast `(dto as { updated_at?: unknown }).updated_at` — `TemplateDTO` does not expose `updated_at`, so the branch was unreachable. | **Fixed** in `TemplatesListPage.tsx`: replaced with a single ISO-validity guard on `dto.created_at`, falling back to `new Date().toISOString()`. |
| F4 | LOW | wiki-memory drift | List query has no pagination wiring at the screen level (T-011 in `templates-tech-debt.md`). | **Bounded defer** — already tracked in `wiki/modules/templates-tech-debt.md` T-011; not regressed by this QA. |
| F5 | LOW | workflow/tooling gap | Error-state could not be exercised at runtime because TanStack Query `staleTime: 60_000` + module remount blocked refetch after a fetch monkey-patch; client error UI path is code-proven via `resolveQueryError` + `styles.error` block. | **Bounded defer** — record-only; no fix on the screen. Future Preview runs should clear `QueryClient` cache before injecting failures. |

## Evidence

- **Commands run (from `frontend/apps/web` unless noted):**
  - `corepack pnpm tsc --noEmit -p tsconfig.build.json` → exit 0
  - `corepack pnpm exec vitest run src/features/templates` → 31 passed
  - PowerShell `.\scripts\start-api.ps1` (API :8081, healthy)
- **Runtime proof (Preview `2ec1ba63-...`):**
  - Tabs snapshot: `Todos· 2 | Publicados· 0 | Rascunhos· 2 | Arquivados· 0` (admin tenant `ffffffff-...`, 2 seed drafts via `POST /api/v1/templates` returning 201).
  - Empty-state after fix:
    - tab `Publicados` → `Nenhum template publicado encontrado.`
    - tab `Arquivados` → `Nenhum template arquivado encontrado.`
  - Navigation: card click → `/templates/:id/versions/1`; history back returns to list cleanly; "Novo template" → `/templates/new?step=1`.
- **Persisted/API proof:**
  - `GET /api/v1/templates` returns 2 templates with `archived_at: null`, `published_version_id: null`, `created_by: "admin"`, `latest_version: 1`. Status derivation → `draft` matches.
  - `POST /api/v1/templates` (with Idempotency-Key) → 201, tenant scoped to admin tenant.

## Hard-stops / Bounded defers

- **F2 author display name:** requires a user-resolution endpoint or DTO enrichment; cross-module (`auth/identity` ↔ `templates`). Minimum prerequisite plan: add `created_by_display_name` to `TemplateDTO` (or expose `GET /api/v1/users/:id/profile`) before re-rendering `TemplateCard` author.
- **F4 pagination (T-011):** tracked in `wiki/modules/templates-tech-debt.md`. No regression here.
- **F5 error-state runtime drive:** instrumentation gap, not a screen defect. Future preview harness should expose a `QueryClient.clear()` hook for fault-injection.

## Files touched

- `frontend/apps/web/src/features/templates/TemplatesListPage.tsx` — F1 + F3 fix.
- `wiki/modules/templates.md` — `Last verified:` bumped to 2026-05-29.
