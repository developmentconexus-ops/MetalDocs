---
name: metaldocs-tanstack-query
description: "Use when wiring or reviewing MetalDocs frontend server state with TanStack Query: feature API wrappers, generated OpenAPI types, query keys, useQuery/useMutation hooks, cache invalidation, optimistic updates, polling, prefetching, staleTime/gcTime choices, query tests, or any frontend work touching `frontend/apps/web/src/features/*/{api,queries}/`, `frontend/apps/web/src/lib/queryKeys.ts`, `frontend/apps/web/src/app/RootProviders.tsx`, or API-driven UI state."
---

# MetalDocs TanStack Query

Use this with `metaldocs-frontend` for frontend work and with `metaldocs-backend-api` when the OpenAPI contract or generated API types changed.

## Read First

1. `wiki/architecture/frontend-structure.md`
2. `wiki/architecture/api-contract.md` if generated types or OpenAPI changed
3. `wiki/concepts/error-ux.md`
4. `frontend/apps/web/src/app/RootProviders.tsx`
5. `frontend/apps/web/src/lib/queryKeys.ts`
6. Affected feature files under `frontend/apps/web/src/features/<domain>/{api,queries}/`

If library behavior matters, use Context7 or official TanStack docs. Current repo baseline: `@tanstack/react-query` v5.

## Workflow

### 1. Classify the server state

Choose the simplest pattern that matches the data:

| State kind | Pattern |
|---|---|
| Detail by ID | `QK.<domain>.detail(id)` + `useQuery` |
| Filtered list | `QK.<domain>.list(normalizedParams)` + `keepPreviousData` or `placeholderData` when paging |
| Reference data | longer `staleTime`; invalidate only on admin mutations |
| User/session/bootstrap | central auth/session flow, not feature-local cache hacks |
| State transition mutation | `useMutation`; prefer server response + invalidation over optimistic cache mutation |
| High-frequency editor/autosave | dedicated hook; avoid broad invalidation loops |
| Near-real-time inbox/activity | short `staleTime` or explicit `refetchInterval`, documented in the hook |
| Governed transitional detail | detail/history/context queries with query-layer polling only while the governed state is transient (`under_review`, `scheduled`, etc.); stop polling when the state stabilizes |

Do not put server-derived data in Zustand or component `useEffect` fetches.

### 2. Design the API wrapper

- Put pure request functions in `features/<domain>/api/<domain>.ts`.
- The canonical API client surface is `frontend/apps/web/src/lib/api/client.ts` plus the shared helpers it exports or composes, including the repo's `apiFetch` path. Feature wrappers call that surface; never call `fetch` directly.
- Use `lib/api-types/` generated types when OpenAPI covers the route.
- For the module in scope, generated API types and `lib/api/client.ts` behavior are authoritative unless the task is explicitly a prerequisite sync.
- If a feature wrapper behavior diverges from runtime/spec or from the canonical API client surface and affects more than the local task, classify it as a shared contract prerequisite and stop.
- Throw errors from API functions. TanStack Query only treats a query as errored when the query function throws or returns a rejected promise.
- Keep request functions React-free so they can be unit tested and reused by prefetching.

### 3. Design query keys before hooks

- Add key factories to `frontend/apps/web/src/lib/queryKeys.ts`; never inline query key arrays in components.
- Every variable used by the query function must be represented in the key.
- Prefer object params for filters: `['documents', 'list', { status, page }]`.
- Normalize params before passing them to a key: omit empty strings, trim search text, use stable defaults.
- Avoid sentinel values like `"__disabled__"` when `enabled` is enough.
- Include `all` roots for broad invalidation when useful:

```ts
documents: {
  all: ['documents'] as const,
  list: (params = {}) => ['documents', 'list', params] as const,
  detail: (id: string) => ['documents', 'detail', id] as const,
}
```

### 4. Prefer queryOptions for reusable query definitions

For queries used by hooks, prefetching, `setQueryData`, or tests, create an options factory in the feature query file:

```ts
import { queryOptions, useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { getDocument } from '../api/documents';

export function documentQueryOptions(id: string) {
  return queryOptions({
    queryKey: QK.documents.detail(id),
    queryFn: () => getDocument(id),
    enabled: Boolean(id),
    staleTime: 30_000,
  });
}

export function useDocumentQuery(id: string) {
  return useQuery(documentQueryOptions(id));
}
```

Do not introduce this abstraction for a one-off query unless it improves prefetching, testing, or cache writes.

### 5. Choose freshness intentionally

Use repo defaults unless the data has a clear reason to differ. Current global default lives in `RootProviders.tsx`.

Guidelines:

- volatile workflow state: `staleTime` 10-30s
- normal detail/list screens: `staleTime` 30-60s
- reference/catalog data: `staleTime` 5 minutes
- feature flags/permissions that only change on login: `staleTime: Infinity`, not `"static"`, so manual invalidation still works
- never set `staleTime: "static"` unless manual invalidation must be ignored
- add `refetchInterval` only for true live-ish surfaces and document why
- if a governed document or approval state can flip on the server without user interaction, keep that synchronization in the query hook via `refetchInterval` or equivalent query options, not in page-level `useEffect` timers

### 6. Mutations: update exact cache, invalidate dependents

Place mutation hooks in `features/<domain>/queries/useXMutation.ts`.

Default mutation policy:

1. Use a pure API `mutationFn`.
2. On success, write exact detail cache when the response contains the updated entity.
3. Invalidate affected lists, aggregates, inboxes, and audit/activity queries.
4. Prefer targeted invalidation over `queryClient.invalidateQueries()` with no key.
5. Use `resolveQueryError(err, fallback)` for user-facing errors.

Example:

```ts
const queryClient = useQueryClient();

return useMutation({
  mutationFn: renameDocument,
  onSuccess: (document, variables) => {
    queryClient.setQueryData(QK.documents.detail(variables.id), document);
    queryClient.invalidateQueries({ queryKey: QK.documents.list() });
    queryClient.invalidateQueries({ queryKey: QK.documents.stats() });
    queryClient.invalidateQueries({ queryKey: QK.audit.recent() });
  },
});
```

If a key factory requires params, add an `all` or `lists` root before broad invalidation. Do not invent partial keys ad hoc in callsites.

### 7. Optimistic updates are opt-in

Use optimistic cache updates only when all are true:

- the action is easily reversible
- the rollback shape is obvious
- the backend response cannot introduce important workflow state the UI must wait for
- failed optimism will not mislead a regulated QMS workflow

For approvals, publish, finalize, archive, supersede, obsolete, or permission changes: prefer pending UI + server response + invalidation. This is less flashy and more correct.

If optimism is justified, cancel affected queries, snapshot previous cache, write optimistic data, rollback on error, and invalidate on settled.

### 8. Performance checks

- Use `enabled` for dependent queries instead of firing invalid requests.
- Use `keepPreviousData` or `placeholderData` for paginated/filterable lists to avoid flicker.
- Use `select` for component-specific projections; do not duplicate derived server state in local state.
- Prefetch on route transitions or wizard next steps only when the next data need is highly likely.
- Avoid request waterfalls: if a page needs independent queries, run them in parallel with multiple hooks.
- Do not create a `QueryClient` inside render; keep the single app-level client.

### 9. Tests and verification

For new query/API work, add at least one of:

- API wrapper unit test with mocked `fetch`/`apiFetch`
- query hook test with `QueryClientProvider`
- page test proving loading/error/success states
- mutation test proving invalidation or cache update

Run:

```powershell
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

If OpenAPI types changed:

```powershell
cd frontend/apps/web
pnpm gen:api
pnpm.cmd tsc --noEmit -p tsconfig.build.json
```

## Templates

Use templates as scaffolds when adding a new API/query/mutation surface. Do not force them onto tiny edits, and do not create a wrapper framework around TanStack Query.

Templates live in `templates/`:

| Template | Use when |
|---|---|
| `feature-api.ts.template` | Adding pure feature API functions backed by generated OpenAPI types |
| `query-options.ts.template` | Adding a reusable query used by hooks, prefetching, tests, or cache writes |
| `mutation.ts.template` | Adding a mutation that updates exact cache and invalidates dependents |
| `invalidation-map.md.template` | Mutation invalidation is non-obvious or affects multiple domains |
| `query-test.tsx.template` | Adding query/mutation tests with an isolated `QueryClient` |
| `cache-policy.md.template` | Freshness, polling, prefetching, or optimism needs an explicit rationale |

Template rules:

- Replace placeholders with real domain names and generated types before committing code.
- Keep API functions React-free.
- Add or reuse `QK` factories before writing hooks.
- Prefer targeted invalidation over manual cache patching.
- Treat optimistic updates as a design decision, not the default.

## Stop Rules

Stop and report when:

- a needed endpoint is missing from generated OpenAPI types
- frontend code needs to hand-write a type that should come from OpenAPI
- a mutation's correct invalidation graph is unclear
- an optimistic update could misrepresent approval, publication, permissions, or audit state
- a query key would need inline arrays outside `lib/queryKeys.ts`
- tests require disabling TanStack Query behavior instead of modeling it with a test `QueryClient`

## Reference

Read `references/tanstack-query-v5-notes.md` when making policy decisions about current TanStack Query behavior, freshness, query options, optimistic updates, or supply-chain-sensitive dependency updates.

