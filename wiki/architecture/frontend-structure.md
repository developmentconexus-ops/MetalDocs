# Frontend Structure

> **Last verified:** 2026-05-21 (post-freeze terminology sync in state/store guidance)
> **Scope:** Canonical folder layout, naming, routing, state, API, design-system rules for `frontend/apps/web`. Comparison baseline for refactor reviews and the `metaldocs-frontend` skill.
> **Out of scope:** Backend module layout (see `system-overview.md`), eigenpal internals (see `modules/editor-ui-eigenpal.md`).
> **Key files:**
> - `frontend/apps/web/src/main.tsx`     mount + provider tree
> - `frontend/apps/web/src/app/AppRouter.tsx:15`     `createBrowserRouter` router export; public vs. protected route tree
> - `frontend/apps/web/src/features/shell/pages/AppRoot.tsx:28`     auth guard + session bootstrap
> - `frontend/apps/web/src/features/shell/components/AppShell.tsx:9`     Rail + AppToolbar + Outlet layout wrapper
> - `frontend/apps/web/src/features/shell/components/Rail.tsx:41`     56px dark nav sidebar
> - `frontend/apps/web/src/features/shell/components/AppToolbar.tsx:7`     52px top bar (search, notifications, new-doc)
> - `frontend/apps/web/src/features/shell/components/SectionPanel.tsx:4`     224px slot panel (Library only)
> - `frontend/apps/web/src/features/`     one folder per domain
> - `frontend/apps/web/src/components/ui/`     design-system primitives only
> - `frontend/apps/web/src/lib/api/`     apiFetch, ApiError, authBus, openapi-fetch wrapper
> - `frontend/apps/web/src/lib/queryKeys.ts:27`     centralized `QK` constants; all `queryKey` / `invalidateQueries` calls import from here; `QK.templates.byProfile` added
> - `frontend/apps/web/src/lib/api-types/`     generated from `api/openapi/v1/openapi.yaml`
> - `frontend/apps/web/src/styles/tokens.css:2`     Wine-palette design tokens (`--brand-*`, `--rail-*`, `--bg`, `--accent`,    )
> - `frontend/apps/web/design-source/`     claude.design HTML screen specs (committed reference)
> - `packages/shared-tokens/`     cross-app design tokens

## Related skills

- `.claude/skills/metaldocs-frontend/SKILL.md`     architecture rulebook (use for any frontend work).
- `.agents/skills/metaldocs-tanstack-query/SKILL.md`     TanStack Query/API workflow for query keys, cache invalidation, optimistic updates, freshness, and query tests.
- `.claude/skills/metaldocs-screen-implementation/SKILL.md`     6-phase workflow for implementing designed screens from `design-source/<slug>/`. Use when the task is "implement screen X".

---

## 1. Mental model

The frontend is a **feature-sliced SPA**. Each business domain owns its UI, hooks, queries, state, types, and routes. The shell (`app/`) only wires them together. Cross-cutting infra lives in `lib/`. Visual primitives live in `components/ui/`.

Domains today: `auth`, `documents`, `templates`, `taxonomy`, `iam`, `approval`, `controlled-documents`, `notifications`, `feature-flags`, `shell`, `shared`.

---

## 2. Canonical folder layout

```
frontend/apps/web/src/
          app/                        #     root composition
                RootProviders.tsx       # QueryClientProvider, Toaster, ErrorBoundary, theme
                AppRouter.tsx           # createBrowserRouter, route tree assembly
                bootstrap.tsx           # featureFlags init + ReactDOM.createRoot
          features/<domain>/          #     one folder per business domain
                api/                    # raw API calls (uses lib/api + api-types)
                components/             # feature-scoped UI (not shared)
                hooks/                  # plain React hooks
                pages/                  # route entry points (one file per route)
                queries/                # TanStack Query hooks (useXQuery, useXMutation)
                types.ts                # feature-internal types
                routes.tsx              # RouteObject[] export
                index.ts                # barrel: only public API of the feature
          features/shell/             #     app shell domain (auth guard + layout)
                pages/AppRoot.tsx       # auth guard + session bootstrap (wraps all protected routes)
                components/
                    AppShell.tsx        # Rail + AppToolbar + Outlet; optional SectionPanel slot
                    Rail.tsx            # 56px dark nav sidebar (--rail-* tokens)
                    AppToolbar.tsx      # 52px top bar: search, notifications, new-doc
                    SectionPanel.tsx    # 224px slot panel (Library only, via route handle flag)
          features/shared/            #     cross-feature UI primitives (used by 2+ features)
                components/
                    editor-chrome/      # toolbar overlay + eigenpal overrides for eigenpal-based pages
                          EditorChrome.tsx          # wrapper with left/center/right/alert slots
                          EditorChrome.module.css   # overlay positioning + eigenpal CSS overrides + button primitives
                          parts/
                                VersionBadge.tsx      # monospace revision chip
                                AutosaveStatus.tsx    # idle/saving/saved/error indicator
                          index.ts                  # barrel: EditorChrome, editorChromeStyles, VersionBadge, AutosaveStatus
                    wizard/             # shared wizard chrome (used by document + template wizards)
                        WizardShell.tsx           # parameterized wizard layout: kicker/title/description/Stepper/children
                        WizardFooter.tsx          # shared footer: stepLabel/primaryDisabled/showBack/onAdvance/onBack/onCancel
          components/
                ui/                     # design-system primitives ONLY
                    Icon.tsx            # unified icon wrapper
                    Avatar.tsx          # user avatar (initials fallback); optional `color` prop for hashed background
                    CodeChip.tsx        # inline code badge
                    Logo.tsx            # product logotype
                    StatusPill.tsx      # document/workflow status badge
                    StatusPill.module.css
                    index.ts            # barrel export
          lib/
                api/                    # apiFetch, ApiError, authBus, openapi-fetch wrapper
                api-types/              # generated (never hand-edit)
                hooks/                  # generic domain-agnostic React hooks (e.g. useDebouncedValue)
                queryKeys.ts            # centralized QK constants (see   8)
                types/                  # cross-cutting types only
          store/                      # GLOBAL stores only: ui.store, auth.store
          styles/                     # tokens.css, base.css, document-content.css
          editor-adapters/            # eigenpal integration glue (see modules/editor-ui-eigenpal.md)
          routing/                    # route tree helpers, route guards (if any)
          main.tsx                    # imports app/bootstrap
          design-source/              # claude.design HTML reference (committed, build-excluded)
```

**Removed (do not recreate):**
- `src/api/` (legacy flat API)     migrated into `features/<x>/api/` + `lib/api/`
- `src/lib.api.ts`, `src/lib.types.ts` (root flat files)     migrated into features and `lib/types/`
- `src/components/<FeatureName>.tsx` at root     moved into `features/<x>/components/`
- `features/<x>/state/<x>.store.ts` (feature-scoped zustand)     server state now exclusively in TanStack Query; only `store/auth.store.ts` + `store/ui.store.ts` remain

---

## 3. Hard rules (non-negotiable)

1. **Feature code lives in `features/<domain>/`. Never in `components/` root.** `components/` root is reserved for true app shell wiring (none today after refactor).
2. **`components/ui/` is design-system only.** Generic, domain-agnostic primitives. If it imports from `features/`, it's misplaced.
3. **API calls go through `lib/api/`.** Always. Use `apiFetch` (handles auth bus, ApiError). Do not call `fetch` directly.
4. **Types come from `lib/api-types/` (generated).** Do not hand-write request/response shapes that the OpenAPI spec covers.
5. **Server state = TanStack Query.** Never `useEffect` + local state for fetching. Hooks live in `features/<x>/queries/`.
6. **UI/local state = `useState`/`useReducer`.** Cross-cutting UI state = global zustand (`store/ui.store.ts` or `store/auth.store.ts`). **No feature-scoped zustand stores**     server state belongs exclusively in TanStack Query.
7. **Routing = data routes.** `createBrowserRouter` + per-feature `routes.tsx`. **No `HashRouter`.** **No string-pattern path dispatchers** (e.g., the old `viewFromPath`).
8. **CSS = CSS Modules** (`<Component>.module.css`). Tokens from `styles/tokens.css` or `@metaldocs/shared-tokens`. No inline styles for theming. No CSS-in-JS libraries.
9. **No backwards-compat shims.** When migrating, delete legacy files in the same PR. No re-exports. No `// removed` comments. Aligns with project CLAUDE.md.
10. **Errors UX.** Use `lib/api/errors.ts` (`ApiError`, `resolveErrorMessage`) and the auth bus per `wiki/concepts/error-ux.md`. For TanStack Query `onError` callbacks use `resolveQueryError(err, fallback)` (`lib/api/resolveQueryError.ts`)     it handles `ApiError`, `Error`, and unknown in one call. Toasts via `sonner`. Never raw `alert`.

---

## 4. Naming conventions

| Thing | Convention | Example |
|------|------------|---------|
| Component file | `PascalCase.tsx` | `DocumentEditorPage.tsx` |
| Hook file | `useXxx.ts` | `useDocumentSession.ts` |
| Query hook | `useXxxQuery.ts` / `useXxxMutation.ts` | `useDocumentQuery.ts` |
| Store | `<domain>.store.ts` | `ui.store.ts`, `auth.store.ts` (global only     see   9) |
| Module barrel | `index.ts` | `features/documents/index.ts` |
| Types module | `types.ts` | `features/documents/types.ts` |
| Routes module | `routes.tsx` | `features/documents/routes.tsx` |
| CSS module | `<Component>.module.css` | `DocumentEditorPage.module.css` |
| Test | `<thing>.test.ts(x)` co-located or in `__tests__/` | `useDocumentQuery.test.ts` |

---

## 5. Decision rules

When adding code, ask in order:

1. **Is it a generic primitive (Button, Input, Modal, Card, Tabs)?**     `components/ui/`
2. **Is it specific to one domain?**     `features/<domain>/<correct subfolder>/`
3. **Is it cross-cutting infra (HTTP, auth bus, error mapping, types from spec)?**     `lib/`. Generic domain-agnostic React hooks (e.g. debounce, window size)     `lib/hooks/`. Promote a feature hook here only when a **second** feature calls it.
4. **Is it routing wiring?**     feature owns `routes.tsx`; `app/AppRouter.tsx` composes them.
5. **Is it global UI state (sidebar open, theme)?**     `store/ui.store.ts`. Otherwise feature-local.

When in doubt: feature-local first. Promote when a **second** caller appears:
- Generic UI primitive (domain-agnostic)     `components/ui/`
- Feature-coupled shared component (has domain context, used by 2+ features)     `features/shared/components/` (e.g. `editor-chrome/`)
- Generic React hook (no domain dep)     `lib/hooks/`

---

## 6. Routing

```ts
// features/documents/routes.tsx (illustrative subset     see actual file for full list)
import type { RouteObject } from "react-router-dom";

export const documentsRoutes: RouteObject[] = [
  { path: "documents", lazy: () => import("./pages/LibraryPage") },
  // fixed-segment sub-routes first (all/area/:x/type/:x/doc/:x/mine/recent)
  { path: "documents/:documentId", lazy: () => import("./pages/DocumentPublishedPage") },
  { path: "documents/new",      lazy: () => import("./pages/NewDocumentWizardPage") },
  { path: "documents/:documentID/edit", lazy: () => import("./pages/DocumentEditorRoutePage") },
];
```

```ts
// app/AppRouter.tsx  (see frontend/apps/web/src/app/AppRouter.tsx:15)
export const router = createBrowserRouter([
  // Public     no Rail, no Toolbar
  {
    path: '/login',
    lazy: () => import('../features/auth/pages/LoginPage').then((m) => ({ Component: m.LoginPage })),
  },
  // Protected     AppRoot (auth guard + bootstrap)     AppShell (Rail + AppToolbar + Outlet)
  {
    element: <AppRoot />,
    children: [
      {
        lazy: () => import('../features/shell/components/AppShell').then((m) => ({ Component: m.AppShell })),
        children: [
          ...documentsRoutes,
          ...templatesRoutes,
          ...approvalRoutes,
          // ...
          { path: '*', element: <Navigate to="/" replace /> },
        ],
      },
    ],
  },
]);
```

Route hierarchy:
- **Public routes** (e.g., `/login`)     declared at the router root, outside `AppRoot`. No Rail, no Toolbar.
- **Protected routes**     children of `AppRoot` (`features/shell/pages/AppRoot.tsx:28`), which handles auth guard + session bootstrap. Inside `AppRoot`, `AppShell` (`features/shell/components/AppShell.tsx:9`) renders Rail + AppToolbar + Outlet.
- **`SectionPanel`**     rendered conditionally inside `AppShell` when a matched route's `handle.sectionPanel === true` (Library screen only).

Rules:
- Always `createBrowserRouter` (clean URLs). Nginx fallback already wired (`frontend/apps/web/nginx.conf:9`).
- Lazy-load pages with `lazy: () => import(...)`. Keeps initial bundle small.
- Auth guard lives in `AppRoot` (component element), not in route `loader`s or individual page components.

---

## 7. API + types

```ts
// lib/api/client.ts
import createClient from "openapi-fetch";
import type { paths } from "../api-types";
import { apiFetch } from "./apiFetch";

export const api = createClient<paths>({ fetch: apiFetch });
```

```ts
// features/documents/api/documents.ts
import { api } from "../../../lib/api/client";

export async function getDocument(id: string) {
  const { data, error } = await api.GET("/api/v1/documents/{id}", {
    params: { path: { id } },
  });
  if (error) throw error;
  return data;
}
```

- `apiFetch` (`lib/api/apiFetch.ts`) handles 401     auth bus, error normalization, JSON parsing, base URL.
- `lib/api-types/` regenerated via `pnpm gen:api` (script added with codegen). Never hand-edit.
- Feature `api/*.ts` files are thin wrappers     pure functions, no React.

---

## 8. Server state (TanStack Query)

For implementation workflow, cache policy, mutation invalidation, and performance review, use `.agents/skills/metaldocs-tanstack-query/SKILL.md`.

TanStack Query is the canonical server-state layer for the web app. It owns remote data, freshness, loading/error state, mutation state, invalidation, polling, prefetching, and optimistic updates. It is not a local UI state store and should not be wrapped by a project-specific framework unless repetition proves that need.

Durable rules:

- API request functions live in `features/<domain>/api/`, use `lib/api/`, and use generated `lib/api-types/` shapes when OpenAPI covers the route.
- Query and mutation hooks live in `features/<domain>/queries/`.
- Reusable queries should expose `queryOptions(...)` factories when the same query is used by hooks, prefetching, tests, or cache writes.
- Mutations must explicitly update exact detail cache entries when the server response is authoritative and invalidate dependent lists, aggregates, inbox/activity, and audit queries.
- Governed workflow states that can change without a local click, such as `under_review`, `scheduled`, or other server-driven transitions, must keep freshness policy in the query layer with targeted invalidation and selective `refetchInterval`. Do not implement background synchronization loops in page components with `useEffect`.
- Optimistic updates are opt-in. Avoid them for approval, publication, archive, finalization, permission, signature, and audit-sensitive workflow state unless the rollback model is obvious and safe.
- Use `.agents/skills/metaldocs-tanstack-query/templates/` when adding a new API/query/mutation surface. Templates are scaffolds, not mandatory boilerplate for tiny edits.

```ts
// lib/queryKeys.ts     centralized constants (frontend/apps/web/src/lib/queryKeys.ts:27)
export const QK = {
  documents: {
    list: () => ['documents', 'list'] as const,
    detail: (id: string) => ['documents', 'detail', id] as const,
  },
  // ...approval, templates, taxonomy, audit, notifications, controlledDocuments
} as const;
```

```ts
// features/documents/queries/useDocumentQuery.ts
import { QK } from '../../../lib/queryKeys';

export function useDocumentQuery(id: string) {
  return useQuery({
    queryKey: QK.documents.detail(id),
    queryFn: () => getDocument(id),
  });
}
```

```ts
// invalidation     always use QK, never inline arrays
queryClient.invalidateQueries({ queryKey: QK.documents.detail(id) });
```

- **Never inline raw string arrays** as query keys. Import from `lib/queryKeys.ts`. This ensures invalidation consistency across the codebase.
- Mutations co-located with their query in `features/<x>/queries/`.
- `QueryClient` instance lives in `app/RootProviders.tsx`. Single instance.
- Tests that exercise query hooks or mutations must create an isolated `QueryClient` per test.

---

## 9. Local + domain state (zustand)

- **Global** (`store/ui.store.ts`, `store/auth.store.ts`): the only two valid zustand stores. Covers truly cross-domain state: current user, auth status, global UI flags (sidebar open, active theme).
- **Feature-scoped zustand stores have been deleted.** `documents.store.ts`, `notifications.store.ts`, `registry.store.ts` (legacy literal filename from pre-rename), and `templates.store.ts` no longer exist. Do not recreate them.
- **Server data never lives in zustand.** Use TanStack Query (with `QK` keys) exclusively for all server-derived state.
- **Component-local state**: `useState` / `useReducer`. If state must span multiple components within one feature, lift to the closest common ancestor or use a context scoped to that subtree.

---

## 10. Styling + design tokens

- All visual values come from `styles/tokens.css` (`frontend/apps/web/src/styles/tokens.css:2`) or `@metaldocs/shared-tokens`. No magic colors/sizes in component CSS.
- New tokens: add to `tokens.css` first, document the rationale in the commit message.
- **Typography:** `@fontsource/inter-tight` (sans-serif     `--font-sans`) + `@fontsource/jetbrains-mono` (mono     `--font-mono`). `@fontsource/dm-sans` and `@fontsource/dm-mono` are removed.
- **Token prefix:** `--brand-*` (e.g., `--brand`, `--brand-deep`, `--brand-soft`, `--brand-pale`) for brand colors. `--vinho-*` prefix is gone. Additional namespaces: `--rail-*` (dark sidebar), `--bg`, `--surface`, `--border`, `--text`, `--accent`, `--success`, `--warning`, `--danger`, `--info`, `--shadow-*`, `--r-*`, `--sp-*`.
- **Wine palette** (see `tokens.css:2`): deep crimson brand (`#6b1f2a`), dark rail (`#2a1418`), warm background (`#f4eeee`). Components must not hardcode these hex values     always use the token variable.
- Dark mode (future): tokens flip via `[data-theme="dark"]` selector. Components stay theme-agnostic.

---

## 11. Design-source intake

**Implemented screens (design-source     feature slice):**

| Slug | Route | Feature slice | Status |
|---|---|---|---|
| `caixa-aprovacao` | `/approvals` | `features/approval/` | Implemented (Phase 3c); 7 deferred items |
| `documento-publicado` | `/documents/:documentId` | `features/documents/pages/DocumentPublishedPage` | Implemented (Phases 0   3c); 9 deferred items |

Screens designed in claude.design land in `frontend/apps/web/design-source/<screen-slug>/`:

- `<slug>.html`     original export (read-only reference)
- `<slug>.png`     screenshot
- `NOTES.md`     implementer notes (which feature, which route, which existing primitives to reuse)

These files ship in the repo (committed) but are excluded from the production bundle (Vite ignores).

When implementing:

1. Place screen in correct `features/<x>/pages/` per domain map.
2. Map HTML structure     existing `components/ui/` primitives. Extract a new primitive only if the same composition appears in 2+ design files.
3. CSS     CSS Module on the page. Feature-scoped styles stay with the page.
4. Wire data via `features/<x>/queries/`. No mock data after first iteration.
5. Update `wiki/modules/<feature>.md` (or `workflows/`) with screenshot + page anchor. Run wiki-curator.

---

## 12. Testing

| Layer | Tool | Location |
|------|------|----------|
| Unit (pure logic, hooks) | vitest | co-located `*.test.ts(x)` or `__tests__/` |
| Component | vitest + Testing Library | co-located |
| E2E | Playwright | `frontend/apps/web/e2e/` |

- Run unit: `pnpm test`. Run e2e: `pnpm e2e:smoke`.
- New page     at minimum a render-without-error vitest. Critical flows     Playwright.

---

## 13. Migration policy (legacy code)

- **No new code in legacy paths** (`src/api/`, `src/lib.api.ts`, `src/lib.types.ts`, `src/components/<FeatureName>*`). Feature-scoped zustand stores (`features/<x>/state/<x>.store.ts`) are deleted     do not recreate them.
- **When you touch a file**, migrate it fully to the canonical location in the same PR. Update all importers. Delete legacy file.
- **No re-export shims.** No `// moved to X` comments.
- **Wiki-curator runs after each migration PR** to refresh anchors and `Last verified` stamps.

---

## 14. Performance defaults

- Routes lazy-loaded (`lazy: () => import(...)`).
- Heavy components (PDF viewer, eigenpal editor) lazy-loaded via `React.lazy` + `Suspense`.
- Images: prefer SVG for icons (`react-icons` already installed). Raster: WebP if >50KB.
- Bundle audit: `pnpm build` + check `dist/` size. Page-level chunks should be <200 KB gzipped.

---

## 15. Anti-patterns (do not do)

-     `HashRouter`     clean URLs only.
-     String-match path dispatchers (e.g., `if (path.startsWith("/x")) return "view-x"`).
-     God components (>400 lines). Split.
-     `useEffect` + `setState` for data fetching. Use TanStack Query.
-     Inline `queryKey` string arrays. Import from `lib/queryKeys.ts` (`QK.*`).
-     Feature-scoped zustand stores for server state. Use TanStack Query.
-     Direct `fetch` calls. Use `lib/api/`.
-     Hand-written types that mirror OpenAPI shapes. Codegen.
-     Cross-feature imports from `features/<a>/components/<x>` into `features/<b>/`. If you need it, the component is a primitive     move to `components/ui/` or `features/shared/`.
-     Inline styles for theming. CSS Modules + tokens.
-     Backwards-compat re-exports during migration.

---

## 16. References

- `wiki/concepts/error-ux.md`     apiFetch, ApiError, auth-bus contract
- `wiki/modules/editor-ui-eigenpal.md`     editor integration
- `wiki/architecture/system-overview.md`     services + ports
- `wiki/architecture/api-contract.md`     OpenAPI spec, oapi-codegen backend codegen, frontend `gen:api` script, CI drift guard
- `wiki/decisions/`     architecture decision records
- `frontend/apps/web/design-source/README.md`     screen intake protocol (added with Block 0)


