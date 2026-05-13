# TanStack Query v5 Notes For MetalDocs

> Last checked: 2026-05-13
> Sources: official TanStack Query docs and current MetalDocs frontend code.

## Official v5 Behaviors To Preserve

- Cached query data is stale by default. Use `staleTime` globally or per query to reduce excessive refetching.
- Inactive queries are garbage-collected after 5 minutes by default unless `gcTime` changes.
- Failed queries retry 3 times by default in TanStack Query; MetalDocs currently overrides global retry to `1` in `RootProviders.tsx`.
- Query functions must throw or return a rejected promise for TanStack Query to enter the error state.
- Query keys must include all variables the query function depends on. Treat keys as dependency arrays for server state.
- `queryOptions` is the recommended v5 helper when sharing `queryKey`, `queryFn`, and defaults across hooks, prefetching, tests, and `setQueryData`.
- Keep one stable `QueryClient` for the application lifecycle.
- Optimistic updates can be implemented via UI variables or cache writes. Cache optimism must cancel affected queries, snapshot previous cache, write optimistic state, rollback on error, and invalidate/refetch after settle.

## MetalDocs Baseline

- Package: `@tanstack/react-query` v5.100.9.
- Query client: `frontend/apps/web/src/app/RootProviders.tsx`.
- Global defaults: `staleTime: 30_000`, `retry: 1`.
- API client: `frontend/apps/web/src/lib/api/client.ts` and `apiFetch`.
- Error UX: `frontend/apps/web/src/lib/api/resolveQueryError.ts`.
- Query keys: `frontend/apps/web/src/lib/queryKeys.ts`.

## MetalDocs Policy

- Use TanStack Query for all server-derived state.
- Keep Zustand limited to global UI/auth state.
- Use generated OpenAPI types when the route is in the public contract.
- Prefer server truth for regulated workflow transitions. Do not optimistic-update approval, publication, permission, or audit-critical state unless explicitly approved by a design plan.
- Add invalidation decisions to the implementation plan. A mutation is incomplete until affected detail/list/stat/inbox/audit keys are handled.

## Dependency Safety Note

TanStack disclosed an npm supply-chain compromise on 2026-05-11. The official postmortem says confirmed-clean families include `@tanstack/query*`, but future dependency upgrades should still be deliberate:

- do not casually bump `@tanstack/*` packages while wiring product features
- review lockfile diffs
- use official advisories/docs for incident-specific guidance
- treat installs from affected dates/versions as a security task, not a feature task

## Official Sources

- Important defaults: https://tanstack.com/query/latest/docs/framework/react/guides/important-defaults
- Query keys: https://tanstack.com/query/latest/docs/framework/react/guides/query-keys
- Query functions: https://tanstack.com/query/v5/docs/framework/react/guides/query-functions
- Query options: https://tanstack.com/query/v5/docs/framework/react/guides/query-options
- Optimistic updates: https://tanstack.com/query/latest/docs/framework/react/guides/optimistic-updates
- Stable QueryClient ESLint rule: https://tanstack.com/query/latest/docs/eslint/stable-query-client
- Incident postmortem: https://tanstack.com/blog/npm-supply-chain-compromise-postmortem
