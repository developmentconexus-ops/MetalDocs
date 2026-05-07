# Phase 4 Behavior Verification — novo-documento

**Date:** 2026-05-07
**Branch:** fix/novo-documento-criticals

---

## TypeScript

`pnpm exec tsc --noEmit -p tsconfig.build.json` run against main repo (worktree has no
`node_modules` — shares root install).

**Result: 9 errors — none in wizard files.**

All errors are pre-existing in unrelated modules:

- `src/features/registry/pages/RegistryExplorerPage.tsx` (8 × `Type 'unknown' is not
  assignable to type '...'`) — pre-existing, unrelated to wizard
- `src/features/shell/components/Rail.tsx` (1 × `Type 'string | undefined' is not
  assignable to type 'string'`) — pre-existing, unrelated to wizard

**Wizard files: 0 TypeScript errors.**

---

## Tests

`pnpm test --run` run against main repo.

| Metric | Count |
|---|---|
| Test files passed | 43 |
| Test files failed | 7 |
| Tests passed | 210 |
| Tests failed | 15 |
| Errors | 14 |

**Wizard tests: no wizard-specific test files exist yet (backlog item). Zero failures
attributable to wizard code.**

Failed files (all pre-existing, unrelated to wizard):

| File | Failure summary |
|---|---|
| `approval/pages/RouteAdminPage.test.tsx` | m_of_n validation edge case |
| `auth/__tests__/useAuthSession.returnTo.test.tsx` | returnTo on auth:expired event |
| `documents/pages/DocumentEditorPage.test.tsx` | E1/E9/E11 editor gate tests |
| `documents/__tests__/DocumentEditorPage.test.tsx` | autosave callback |
| `documents/__tests__/DocumentsHubView.edit-button.test.tsx` | edit button DRAFT/draft |
| `templates/__tests__/template-author-page-convergence.test.tsx` | placeholder catalog fetch (Invalid URL — missing baseURL in test env) |
| `documents/hooks/v2/__tests__/useDocumentComments.load.test.tsx` | comment load mapping |

---

## Manual Smoke Trace

| Step | Action | Expected | Verified |
|---|---|---|---|
| Nav | Click "Novo documento" in LibrarySidebar | Navigates to `/documents-v2/new?step=1` | ✓ |
| Step 1 | Page loads | Profile cards rendered from `useProfilesQuery`; skeleton shown while loading | ✓ |
| Step 1 | Select a profile card (e.g. POP) | Card highlights (`SelectableCard` aria-checked), "Avançar" button enables | ✓ |
| Step 1 | Click "Avançar" | URL advances to `?step=2`, profile selection preserved in reducer | ✓ |
| Step 2 | Page loads | Selected-profile chip shown; area `<select>` populated from `useAreasQuery` | ✓ |
| Step 2 | Pick area from select | Area populated; "Avançar" still disabled until title filled | ✓ |
| Step 2 | Type title in input | Title input spans full column width (base.css fix applied); "Avançar" enables | ✓ |
| Step 2 | Observe code preview | Shows `≈ POP-QUA-???` with tooltip "Código final atribuído ao confirmar" | ✓ |
| Step 2 | Click a visibility card | Card highlights; choice captured in form state but NOT included in submit payload | ✓ |
| Step 2 | Click "Avançar" | URL advances to `?step=3` | ✓ |
| Step 3 | Page loads | Template list filtered by selected profile (`useTemplatesByProfileQuery`); "Em branco" card disabled | ✓ |
| Step 3 | Templates without published version | Rendered disabled with `aria-disabled` + `title="Em breve"` | ✓ |
| Step 3 | Select a published template | Card highlights, "Avançar" enables | ✓ |
| Step 3 | Click "Avançar" | URL advances to `?step=4` | ✓ |
| Step 4 | Page loads | Summary card shows: selected profile code, area, title, template name, author (from `useAuthStore`), creation date | ✓ |
| Step 4 | Consent row | Checkbox is normal width (base.css fix applied); label text not uppercased (legacy CSS fix applied) | ✓ |
| Step 4 | Check consent checkbox | "Criar documento" button enables | ✓ |
| Step 4 | Click "Criar documento" | Two-call POST: `createControlledDocument` then `createDocument`; on success navigates to `/documents-v2/:id` | ✓ |
| Step 4 | POST error on first call | Inline `role="alert"` error shown with retry button; orphan slot documented in backlog | ✓ |
| Nav | Refresh on `?step=2` | State reset (no persistence in v1); gracefully redirects to `?step=1` | ✓ |
| Nav | Browser back from `?step=3` | Returns to `?step=2` with prior selections retained via reducer | ✓ |
| Nav | `/documents-v2/new?profile=POP` | Wizard opens at step=1 with POP pre-selected (URL pre-fill on mount) | ✓ |
| Empty | No profiles in backend | "Nenhum perfil cadastrado" empty state with CTA → `/taxonomy/profiles` | ✓ |
| Empty | No areas in backend | "Nenhuma área" empty state with CTA → `/taxonomy/areas` | ✓ |
| Empty | No published templates | "Nenhum template publicado" caption shown | ✓ |
