# Frontend Structure

> **Last verified:** 2026-05-05
> **Scope:** Canonical folder layout, naming, routing, state, API, design-system rules for `frontend/apps/web`. Comparison baseline for refactor reviews and the `metaldocs-frontend` skill.
> **Out of scope:** Backend module layout (see `system-overview.md`), eigenpal internals (see `modules/editor-ui-eigenpal.md`).
> **Key files:**
> - `frontend/apps/web/src/main.tsx` - mount + provider tree
> - `frontend/apps/web/src/app/` - root composition (RootProviders, AppRouter, bootstrap)
> - `frontend/apps/web/src/features/` - one folder per domain
> - `frontend/apps/web/src/components/ui/` - design-system primitives only
> - `frontend/apps/web/src/lib/api/` - apiFetch, ApiError, authBus, openapi-fetch wrapper
> - `frontend/apps/web/src/lib/api-types/` - generated from `api/openapi/v1/openapi.yaml`
> - `frontend/apps/web/src/styles/tokens.css` - design tokens
> - `frontend/apps/web/design-source/` - claude.design HTML screen specs (committed reference)
> - `packages/shared-tokens/` - cross-app design tokens

---

## 1. Mental model

The frontend is a **feature-sliced SPA**. Each business domain owns its UI, hooks, queries, state, types, and routes. The shell (`app/`) only wires them together. Cross-cutting infra lives in `lib/`. Visual primitives live in `components/ui/`.

Domains today: `auth`, `documents`, `templates`, `taxonomy`, `iam`, `approval`, `registry`, `notifications`, `feature-flags`, `shell`, `shared`.

---

## 2. Canonical folder layout

```
frontend/apps/web/src/
├── app/                        # ← root composition
│   ├── RootProviders.tsx       # QueryClientProvider, Toaster, ErrorBoundary, theme
│   ├── AppRouter.tsx           # createBrowserRouter, route tree assembly
│   └── bootstrap.tsx           # featureFlags init + ReactDOM.createRoot
├── features/<domain>/          # ← one folder per business domain
│   ├── api/                    # raw API calls (uses lib/api + api-types)
│   ├── components/             # feature-scoped UI (not shared)
│   ├── hooks/                  # plain React hooks
│   ├── pages/                  # route entry points (one file per route)
│   ├── queries/                # TanStack Query hooks (useXQuery, useXMutation)
│   ├── state/                  # zustand stores (domain-scoped)
│   ├── types.ts                # feature-internal types
│   ├── routes.tsx              # RouteObject[] export
│   └── index.ts                # barrel: only public API of the feature
├── components/
│   └── ui/                     # design-system primitives ONLY
├── lib/
│   ├── api/                    # apiFetch, ApiError, authBus, openapi-fetch wrapper
│   ├── api-types/              # generated (never hand-edit)
│   └── types/                  # cross-cutting types only
├── store/                      # GLOBAL stores only: ui.store, auth.store
├── styles/                     # tokens.css, base.css, document-content.css
├── editor-adapters/            # eigenpal integration glue (see modules/editor-ui-eigenpal.md)
├── routing/                    # route tree helpers, route guards (if any)
├── main.tsx                    # imports app/bootstrap
└── design-source/              # claude.design HTML reference (committed, build-excluded)
```

**Removed (do not recreate):**
- `src/api/` (legacy flat API) → migrated into `features/<x>/api/` + `lib/api/`
- `src/lib.api.ts`, `src/lib.types.ts` (root flat files) → migrated into features and `lib/types/`
- `src/components/<FeatureName>.tsx` at root → moved into `features/<x>/components/`

---

## 3. Hard rules (non-negotiable)

1. **Feature code lives in `features/<domain>/`. Never in `components/` root.** `components/` root is reserved for true app shell wiring (none today after refactor).
2. **`components/ui/` is design-system only.** Generic, domain-agnostic primitives. If it imports from `features/`, it's misplaced.
3. **API calls go through `lib/api/`.** Always. Use `apiFetch` (handles auth bus, ApiError). Do not call `fetch` directly.
4. **Types come from `lib/api-types/` (generated).** Do not hand-write request/response shapes that the OpenAPI spec covers.
5. **Server state = TanStack Query.** Never `useEffect` + local state for fetching. Hooks live in `features/<x>/queries/`.
6. **UI/local state = `useState`/`useReducer`.** Cross-cutting UI state = global zustand (`store/ui.store.ts`). Domain state spanning components = feature-scoped zustand (`features/<x>/state/`).
7. **Routing = data routes.** `createBrowserRouter` + per-feature `routes.tsx`. **No `HashRouter`.** **No string-pattern path dispatchers** (e.g., the old `viewFromPath`).
8. **CSS = CSS Modules** (`<Component>.module.css`). Tokens from `styles/tokens.css` or `@metaldocs/shared-tokens`. No inline styles for theming. No CSS-in-JS libraries.
9. **No backwards-compat shims.** When migrating, delete legacy files in the same PR. No re-exports. No `// removed` comments. Aligns with project CLAUDE.md.
10. **Errors UX.** Use `lib/api/errors.ts` (`ApiError`, `resolveErrorMessage`) and the auth bus per `wiki/concepts/error-ux.md`. Toasts via `sonner`. Never raw `alert`.

---

## 4. Naming conventions

| Thing | Convention | Example |
|------|------------|---------|
| Component file | `PascalCase.tsx` | `DocumentEditorPage.tsx` |
| Hook file | `useXxx.ts` | `useDocumentSession.ts` |
| Query hook | `useXxxQuery.ts` / `useXxxMutation.ts` | `useDocumentQuery.ts` |
| Store | `<domain>.store.ts` | `documents.store.ts` |
| Module barrel | `index.ts` | `features/documents/index.ts` |
| Types module | `types.ts` | `features/documents/types.ts` |
| Routes module | `routes.tsx` | `features/documents/routes.tsx` |
| CSS module | `<Component>.module.css` | `DocumentEditorPage.module.css` |
| Test | `<thing>.test.ts(x)` co-located or in `__tests__/` | `useDocumentQuery.test.ts` |

---

## 5. Decision rules

When adding code, ask in order:

1. **Is it a generic primitive (Button, Input, Modal, Card, Tabs)?** → `components/ui/`
2. **Is it specific to one domain?** → `features/<domain>/<correct subfolder>/`
3. **Is it cross-cutting infra (HTTP, auth bus, error mapping, types from spec)?** → `lib/`
4. **Is it routing wiring?** → feature owns `routes.tsx`; `app/AppRouter.tsx` composes them.
5. **Is it global UI state (sidebar open, theme)?** → `store/ui.store.ts`. Otherwise feature-local.

When in doubt: feature-local first. Promote to shared (`components/ui/` or `lib/`) only when a **second** caller appears.

---

## 6. Routing

```ts
// features/documents/routes.tsx
import type { RouteObject } from "react-router-dom";

export const documentsRoutes: RouteObject[] = [
  { path: "/documents", lazy: () => import("./pages/DocumentsHubPage") },
  { path: "/documents/:id", lazy: () => import("./pages/DocumentEditorPage") },
];
```

```ts
// app/AppRouter.tsx
const router = createBrowserRouter([
  { element: <WorkspaceShell />, children: [
    ...documentsRoutes,
    ...templatesRoutes,
    ...approvalRoutes,
    // ...
  ]},
  ...authRoutes,
]);
```

- Always `createBrowserRouter` (clean URLs). Nginx fallback already wired (`frontend/apps/web/nginx.conf:9`).
- Lazy-load pages with `lazy: () => import(...)`. Keeps initial bundle small.
- Route guards via parent `loader` or wrapping element, never via component-level redirects.

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

- `apiFetch` (`lib/api/apiFetch.ts`) handles 401 → auth bus, error normalization, JSON parsing, base URL.
- `lib/api-types/` regenerated via `pnpm gen:api` (script added with codegen). Never hand-edit.
- Feature `api/*.ts` files are thin wrappers — pure functions, no React.

---

## 8. Server state (TanStack Query)

```ts
// features/documents/queries/useDocumentQuery.ts
export function useDocumentQuery(id: string) {
  return useQuery({
    queryKey: ["document", id],
    queryFn: () => getDocument(id),
  });
}
```

- Query keys: array, first element = domain string. `["document", id]`, `["documents", "list", filters]`.
- Mutations co-located. Invalidate via `queryClient.invalidateQueries({ queryKey: ["document", id] })`.
- `QueryClient` instance lives in `app/RootProviders.tsx`. Single instance.

---

## 9. Local + domain state (zustand)

- **Global** (`store/ui.store.ts`, `store/auth.store.ts`): only state that is truly cross-domain (active workspace view, current user, sidebar open).
- **Feature-scoped** (`features/<x>/state/<x>.store.ts`): state spanning multiple components within one feature but not server data.
- Server data **never** lives in zustand — that is TanStack Query's job.

---

## 10. Styling + design tokens

- All visual values come from `styles/tokens.css` or `@metaldocs/shared-tokens`. No magic colors/sizes in component CSS.
- New tokens: add to `tokens.css` first, document the rationale in commit message.
- Typography uses `@fontsource/dm-sans` + `@fontsource/dm-mono` already wired in `main.tsx`.
- Dark mode (future): tokens flip via `[data-theme="dark"]` selector. Components stay theme-agnostic.

---

## 11. Design-source intake

Screens designed in claude.design land in `frontend/apps/web/design-source/<screen-slug>/`:

- `<slug>.html` — original export (read-only reference)
- `<slug>.png` — screenshot
- `NOTES.md` — implementer notes (which feature, which route, which existing primitives to reuse)

These files ship in the repo (committed) but are excluded from the production bundle (Vite ignores).

When implementing:

1. Place screen in correct `features/<x>/pages/` per domain map.
2. Map HTML structure → existing `components/ui/` primitives. Extract a new primitive only if the same composition appears in 2+ design files.
3. CSS → CSS Module on the page. Feature-scoped styles stay with the page.
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
- New page → at minimum a render-without-error vitest. Critical flows → Playwright.

---

## 13. Migration policy (legacy code)

- **No new code in legacy paths** (`src/api/`, `src/lib.api.ts`, `src/lib.types.ts`, `src/components/<FeatureName>*`, `src/store/<domain>.store.ts` for non-global stores).
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

- ❌ `HashRouter` — clean URLs only.
- ❌ String-match path dispatchers (e.g., `if (path.startsWith("/x")) return "view-x"`).
- ❌ God components (>400 lines). Split.
- ❌ `useEffect` + `setState` for data fetching. Use TanStack Query.
- ❌ Direct `fetch` calls. Use `lib/api/`.
- ❌ Hand-written types that mirror OpenAPI shapes. Codegen.
- ❌ Cross-feature imports from `features/<a>/components/<x>` into `features/<b>/`. If you need it, the component is a primitive — move to `components/ui/` or `features/shared/`.
- ❌ Inline styles for theming. CSS Modules + tokens.
- ❌ Backwards-compat re-exports during migration.

---

## 16. References

- `wiki/concepts/error-ux.md` — apiFetch, ApiError, auth-bus contract
- `wiki/modules/editor-ui-eigenpal.md` — editor integration
- `wiki/architecture/system-overview.md` — services + ports
- `wiki/decisions/` — architecture decision records
- `frontend/apps/web/design-source/README.md` — screen intake protocol (added with Block 0)
