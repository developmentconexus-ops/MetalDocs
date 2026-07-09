# Feature F2d.4 — Plan

> Input: `spec.md` (approved pre-code). TDD: failing test first → green.

## Plan

### Step 0 — failing unit tests (write first)
`frontend/apps/web/src/features/approval/queries/useApprovalInstanceQuery.test.tsx`. Mirror the
lightweight config-mock harness (`useTemplatesByProfileQuery.test.ts` mocks `useQuery` as identity) for
the key/enabled assertions, plus direct `queryFn` invocation for the 404/etag behaviors:
- `mock '@tanstack/react-query'` → `useQuery: vi.fn((o) => o)`; `mock '../api/approvalApi'` →
  `getInstance: vi.fn()`; import real `etagCache`.
- **keys + enabled:** `options.queryKey` toEqual `QK.approval.instance('doc-1')`; `enabled` reflects the
  passed flag (true when `enabled` arg true + id present; false when `enabled` false).
- **seeds etag on success:** `getInstance` resolves an instance (and, since seeding lives inside
  `getInstance`, assert the `queryFn` returns the instance and calls `getInstance('doc-1')`). Etag-seed
  itself is `getInstance`'s contract (already tested at the cockpit level); here assert `queryFn` delegates
  to `getInstance` and returns its value.
- **404 resolves null, no etag:** `getInstance` rejects `{ status: 404 }`; `await options.queryFn()` ===
  `null`; `etagCache.get('doc-1')` undefined.
- **non-404 rejects:** `getInstance` rejects `{ status: 500 }`; `await expect(options.queryFn()).rejects`.

Red first (module absent).

### Step 1 — the hook (implement to green)
`frontend/apps/web/src/features/approval/queries/useApprovalInstanceQuery.ts`:
```ts
import { useQuery } from '@tanstack/react-query';
import { getInstance } from '../api/approvalApi';
import type { ApprovalInstance } from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export function useApprovalInstanceQuery(documentId: string, enabled: boolean) {
  return useQuery<ApprovalInstance | null>({
    queryKey: QK.approval.instance(documentId),
    queryFn: async () => {
      try {
        return await getInstance(documentId);
      } catch (err) {
        if ((err as { status?: number }).status === 404) return null; // no active instance
        throw err;
      }
    },
    enabled: Boolean(documentId) && enabled,
  });
}
```
(No `refetchInterval`; no `setInterval`; no staleness clock — staleness is react-query's concern now.)

### Step 2 — migrate the cockpit adapter (`useDocumentApprovalArtifact.ts`)
- Replace the imperative block (`useState instance/instanceLoading/instanceError/lastFetchedAt/now`, the
  `refetchInstance` useCallback, the `hasActiveContext` fetch `useEffect`, the 1s `setInterval` effect, and
  `isStale`) with:
  ```ts
  const instanceQuery = useApprovalInstanceQuery(documentId, hasActiveContext);
  const instance = instanceQuery.data ?? null;
  const instanceLoading = instanceQuery.isLoading;
  const instanceError = instanceQuery.isError ? 'Erro ao carregar dados de aprovação.' : null;
  const refetchInstance = useCallback(async () => { await instanceQuery.refetch(); }, [instanceQuery]);
  ```
  Note: `hasActiveContext` is computed from `context` BEFORE this call — move the `useApprovalInstanceQuery`
  call after `hasActiveContext` is defined (it already is, line 113). `instanceQuery.isLoading` is true only
  while `enabled` and fetching; when disabled it is false (react-query v5 `isLoading = isPending &&
  isFetching`) — matches the old "no fetch until hasActiveContext" semantics for the `loading` field.
- Remove `isStale` from the `DocumentApprovalArtifact` interface and the return object. Keep all other
  returned fields identical.
- Remove now-unused imports (`useEffect` if no longer used; `useState` if no longer used — verify before
  deleting).
- Keep the `refetchInstance` callback in the return (cockpit + sidebar consumers unchanged).

### Step 3 — update the cockpit test harness (`useDocumentApprovalArtifact.test.tsx`)
- The adapter now calls `useQuery`, so `renderHook` needs a `QueryClientProvider` wrapper. Add:
  ```ts
  import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
  const makeWrapper = () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={qc}>{children}</QueryClientProvider>
    );
  };
  ```
  Pass `{ wrapper: makeWrapper() }` to every `renderHook(...)`. `getInstance` stays mocked (the real
  `queryFn` calls it). The 404 test (`getInstance.mockRejectedValue({ status: 404 })`) still asserts
  `instance` null + no etag — now via the `queryFn` 404→null map (retry:false so it settles once).
- Do NOT change any assertion semantics; only add the wrapper. Every existing test stays green.

### Step 4 — green + gates
- `npx vitest run` both suites — new hook green, cockpit green.
- `npx tsc --noEmit` clean.
- Grep gate: `grep -n 'setInterval\|isStale' useDocumentApprovalArtifact.ts` → none;
  `grep -n 'useApprovalInstanceQuery' useDocumentApprovalArtifact.ts` → present.

## Files touched
| File | Change |
|------|--------|
| `features/approval/queries/useApprovalInstanceQuery.ts` | NEW — react-query instance hook (404→null, etag via getInstance) |
| `features/approval/queries/useApprovalInstanceQuery.test.tsx` | NEW — key/enabled/404/etag unit tests |
| `features/documents/adapters/useDocumentApprovalArtifact.ts` | migrate to the hook; delete imperative fetch/setInterval/isStale |
| `features/documents/adapters/useDocumentApprovalArtifact.test.tsx` | add QueryClientProvider wrapper (no assertion changes) |

## Non-goals carried from spec
`refetchInstanceRef` (page) + `onRefetchInstance` thread (sidebar/footer) removal → F2d.5/F2d.7. No OpenAPI/
DTO change. No mutation-path change.

## Execution notes

Executed TDD per plan; sonnet implementer, zero deviations. Gates: 5 hook + 19 cockpit = 24/24 green;
`tsc --noEmit` exit 0; grep gates clean (`setInterval`/`isStale` gone, `useApprovalInstanceQuery`
present). `useState`/`useEffect` imports confirmed unused → dropped; `useCallback` + `ApprovalInstance`
stayed. One post-review fix: `instanceLoading` sourced from `isFetching` (not `isLoading`) so a post-404
manual refetch surfaces loading (old imperative parity) — reviewer ❓ closed at root. Bounded defers
(`refetchInstanceRef` page hack, `onRefetchInstance` sidebar thread) → F2d.5/F2d.7, see evidence.md.
