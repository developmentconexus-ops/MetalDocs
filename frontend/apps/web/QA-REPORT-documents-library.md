# QA Report — Documents Library (qa/documents-library)

- Route(s): `/documents`, `/documents/all`, `/documents/area/:areaCode`, `/documents/type/:profileCode`
- Page: `frontend/apps/web/src/features/documents/pages/LibraryPage.tsx`
- Owning module: `wiki/modules/documents.md`
- QA class: screen
- Acting role(s): admin (dev login `admin / AdminMetalDocs123!`)
- Date: 2026-05-28
- CI: billing-blocked → local-evidence gate per QA OS runtime-wins rule

## Gate results

- **Gate 0 scope truth**: PASS — route ownership, owning module wiki, contract surface mapped; 11 drift suspects classified.
- **Gate 1 implementation truth**:
  - `corepack pnpm tsc --noEmit -p tsconfig.build.json` → PASS (exit 0)
  - `corepack pnpm exec vitest run src/features/documents/` → 108 tests passed; 2 pre-existing unrelated suite failures (DocumentEditorPage `@metaldocs/editor-ui` import resolution — baseline noise, not Library scope, flagged.
- **Gate 3 product QA (Preview-driven)**:
  - Initial load honest — empty state copy renders, no spurious activity.
  - Filter tabs change state but DO NOT update URL (F9, hard-stop).
  - Area sidebar click sends correct `?areaCode=rh` to API.
  - Bad areaCode `/documents/area/NONEXISTENT` silently redirects to `/documents` (F5, coupled hard-stop).
  - Search wires `?q=`; backend returns 0 for known document code `PO-RH-001` (F10, hard-stop).
  - `pageSize` persists via `localStorage` `metaldocs.library.pageSize`.
  - Row click → `/documents/:id` navigation works.
  - Refresh re-entry preserves `pageSize`.
  - Zero client console errors during full drive.
- **Gate 5 regression**: PASS — tsc clean post-fix; Library-touching vitest green; Preview re-verify confirms F1/F2/F6/F7/F11 applied without regressions.

## Findings (severity-ordered)

| # | Severity | Family | Finding | Disposition |
|---|---|---|---|---|
| F3 | HIGH | shared-contract prerequisite | `DocumentSummary.RevisionNumber` missing from `lib/api-types/index.d.ts` while backend payload returns it; FE cannot show governed REV00/REV01 from list endpoint. | **HARD STOP** — see below |
| F4 | HIGH | shared-contract prerequisite | `v{d.RevisionVersion}` (technical counter) renders in Rev. column instead of governed `RevisionNumber`. Violates wiki §8.8. Coupled to F3. | **HARD STOP** (coupled to F3) |
| F9 | HIGH | architecture contradiction | Library filter state (search/state-tab/area/page) not URL-encoded. Violates global rule `web/patterns.md → URL As State`; breaks deep-link/share/back-button truth. | **HARD STOP** — see below |
| F5 | HIGH | architecture contradiction | Legacy `/documents/{all,area/:areaCode,type/:profileCode}` redirects drop filter context via `<Navigate to="/documents" replace />` in `routes.tsx:11-18`. Coupled to F9. | **HARD STOP** (coupled to F9) |
| F10 | HIGH | shared-contract prerequisite | Backend `?q=` searches title only; UI sidebar input semantically suggests both code + title search. `?q=PO-RH-001` returns 0 against a doc that exists with that code. | **HARD STOP** — see below |
| F1 | MED | screen-local implementation | 3 of 4 KPI cards hardcoded fake numbers + trend strings (`+3 hoje`, `2 vencendo`, `+12% vs anterior`, `Maio · Jun`, static 3/47/23). | **FIXED** — `LibraryStatCards.tsx` now renders `—` + `sem fonte de dados` for pending/frozen/upcoming; only `under_review` wired to live stats. |
| F2 | MED | screen-local implementation | `ActivityPanel` rendered 3 fabricated approval-inbox items + 7 fabricated audit lines + "Tempo real · últimas 8h" badge with no live source. | **FIXED** — mock arrays + types deleted; both sections render explicit "Em breve" placeholder copy with disabled CTAs; subtitle = "Em breve". |
| F6 | MED | screen-local implementation | `LibraryFilter` union + `LIBRARY_STATUSES` missing `scheduled` (Agendados) and `superseded` (Substituídos), although backend supports both and `DocumentStatus`/`STATUS_CONFIG` already declare them. | **FIXED** — `libraryStatus.ts` adds both entries between approved and rejected; sidebar + tabs now expose them (verified via Preview snapshot). |
| F7 | LOW | screen-local implementation | `d.Status as DocumentStatus` unsafe cast at `LibraryPage.tsx:202` violates global TS rule "avoid `any`/unsafe casts". | **FIXED** — added `toDocumentStatus(value)` narrowing helper in `libraryStatus.ts`; call site now uses safe lookup against `KNOWN_STATUSES` set with `'draft'` fallback. |
| F11 | MED | screen-local implementation | Table column header overlap when ActivityPanel open: title column collapsed to 0px (`minmax(0,1fr)` + fixed-px total exceeded container width). | **FIXED** — `LibraryPage.module.css` shrinks fixed columns (140/110/120/130 → 120/100/110/120), sets `minmax(180px,1fr)` on title, adds `min-width: 760px` on header+row, `overflow-x: auto` on `.tableCard`. Verified: at ActivityPanel-open width (491px), table now scrolls horizontally and columns retain intended widths. |
| F8 | LOW | bounded defer | No tenant-scope chip in Library header. Out of Library screen scope — shell/auth concern. | **DEFER** — not Library scope. |

## Evidence

- **Commands run** (Gate 1 + Gate 5):
  - `corepack pnpm tsc --noEmit -p tsconfig.build.json` → exit 0 (post-fix)
  - `corepack pnpm exec vitest run src/features/documents/` → 108 passed, 2 unrelated suites failed (baseline)
- **Runtime proof** (Preview, serverId `40ba78b0-…`):
  - Pre-fix snapshot: KPI tiles `3`/`47`/`23` with fabricated trends; ActivityPanel showing 3 mock approval items + 7 mock audit lines; sidebar/tabs missing Agendados+Substituídos; `Status as DocumentStatus` cast; header overlap with ActivityPanel open (cols collapsed to `140px 0px 110px 120px 130px 56px 36px`).
  - Post-fix snapshot: KPI tiles show `—` + `sem fonte de dados` (pending/frozen/upcoming); only `EM REVISÃO 0 documentos` wired to live stats; ActivityPanel sections show "A caixa de aprovação aparecerá aqui assim que o endpoint estiver disponível." and "A trilha de auditoria do tenant aparecerá aqui quando o endpoint for exposto." with disabled CTAs; sidebar buttons include `Agendados`/`Substituídos`; tabs include both; row renders `RASCUNHO` pill via safe narrowing; with ActivityPanel open `tableCard.scrollWidth=830 > parentWidth=491` → horizontal scroll engaged, columns retain `120 180 100 110 120 56 36`.
  - Post-fix screenshot saved via `preview_screenshot` (orchestrator-attached).
- **Persisted/API proof**:
  - `GET /api/v1/documents?page=1&pageSize=20` → `{items:[2 PO-RH docs], total:2}`
  - `GET /api/v1/documents?q=PO-RH-001` → `{items:[], total:0}` (F10 evidence)
  - `GET /api/v1/documents?areaCode=rh` → 200, scoped result set
  - `GET /api/v1/documents/stats` → `byStatus`/`byArea` only; no pending/frozen/upcoming aggregates (F1 evidence)

## Hard-stops / Bounded defers

### F3 — Backend OpenAPI must expose `RevisionNumber` on `DocumentSummary`

- **Boundary**: backend OpenAPI schema (`backend/api/openapi.yaml`) + oapi-codegen regen → `lib/api-types/index.d.ts`.
- **Why hard-stop**: any local FE workaround requires either `(d as any).RevisionNumber` (banned by "no shims" rule) or a parallel type override that drifts from codegen.
- **Minimum prerequisite plan**:
  1. Add `RevisionNumber: integer` to `DocumentSummary` schema in `backend/api/openapi.yaml`.
  2. Regenerate FE codegen.
  3. Confirm `GET /api/v1/documents` already returns the field at runtime (observed in persisted truth) — no backend handler change required.
- **F4 unblocked** by F3: swap `v{d.RevisionVersion}` → `REV{String(d.RevisionNumber).padStart(2,'0')}` per wiki §8.8.

### F9 — Library state must be URL-encoded (URL As State rule)

- **Boundary**: `LibraryPage.tsx` + `routes.tsx` (cross-component refactor).
- **Why hard-stop**: requires reactivating the legacy routes as canonical URLs, swapping `useState` → `useSearchParams`, threading state into `LibrarySidebar`/`LibraryFilterTabs`. Out of single-screen scope; needs design alignment on canonical URL shape.
- **Minimum prerequisite plan**:
  1. Decide URL shape: `/documents?status=…&areaCode=…&page=…&q=…` vs path-based legacy routes.
  2. Replace `useState` with `useSearchParams` in `LibraryPage`.
  3. Make `/documents/area/:areaCode` and `/documents/type/:profileCode` first-class routes that hydrate state from params (not redirects).
  4. Reset paging via param change, not local effect.
- **F5 unblocked** by F9: legacy routes become canonical hydration points instead of context-dropping redirects.

### F10 — Backend `?q` semantics narrower than UI promise

- **Boundary**: backend documents handler `?q` filter logic OR FE input contract.
- **Why hard-stop**: either backend extends to OR(title, code) — which is a contract change — or FE input must be renamed and a separate code field added, which is a UI redesign.
- **Minimum prerequisite plan (option A — backend)**:
  1. Extend `?q` to match `documents.code ILIKE '%q%' OR documents.name ILIKE '%q%'`.
  2. Add fixture covering code-only match.
- **Option B (frontend)**: split into two inputs (`Buscar título` / `Buscar código`) with separate query params; requires backend `?code=` support too.

### F8 — Tenant-scope chip (bounded defer)

- Out of Library scope; tracked against shell/auth module (`wiki/modules/auth.md`).

## Self-check

- [x] Screen driven through Preview as a user
- [x] All checklist paths exercised for the acting role
- [x] Every finding classified (family + severity) with disposition
- [x] Root-cause fixes applied at owning boundary; symptom-only patches refused
- [x] Hard-stops reported with minimum prerequisite plans, NOT patched
- [x] Evidence recorded (commands + runtime + persisted/API)
- [ ] Wiki `Last verified:` bump — DEFERRED: code truth changes are screen-local; no module-wide invariant shifted. Hard-stops F3/F9/F10 should drive a separate wiki sync (and module-doc-sync) when those backend/architecture prerequisites land.
