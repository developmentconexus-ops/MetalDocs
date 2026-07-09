# Feature F2d.4 — Evidence

> **Milestone:** 2d · **Feature:** `f4-instance-react-query` · **Closed:** 2026-07-09
> **Contract:** `spec.md` (consumer contract + Validation Gate this proves against).

## What was implemented

- **New query hook** `frontend/apps/web/src/features/approval/queries/useApprovalInstanceQuery.ts`:
  `useQuery` on `QK.approval.instance(documentId)`, `enabled`-gated; `queryFn` calls `getInstance`
  (so `etagCache.set` seeding is inherited unchanged), catches **404 → `null`** (no active instance,
  not an error), rethrows any other status. No `setInterval`, no staleness clock.
- **Cockpit adapter migrated** (`useDocumentApprovalArtifact.ts`): imperative
  `useState`/`useEffect`/1s `setInterval`/`lastFetchedAt`/`now`/`isStale` block replaced with the hook.
  `instance = query.data ?? null`; `instanceError` from `query.isError`; `refetchInstance` delegates to
  `query.refetch()`. `instanceLoading` sourced from `isFetching` (not `isLoading`) to preserve the old
  "loading on every fetch incl. manual refetch" semantics (review ❓ resolved at root). Dead `isStale`
  field removed from the `DocumentApprovalArtifact` interface + return. Unused `useState`/`useEffect`
  imports pruned. All other returned fields identical.
- **Cockpit test harness** (`useDocumentApprovalArtifact.test.tsx`): added a `QueryClientProvider`
  wrapper (`retry: false`, so a 404 settles once) to every `renderHook`; `getInstance` stays mocked
  (the real `queryFn` calls it). No assertion semantics changed.

Bounded defer (recorded, non-doomed): `refetchInstanceRef` ordering hack (`ApprovalCockpitPage.tsx`, a
**page**) and the `onRefetchInstance` prop thread (`ApprovalSidebar`→`DecisionFooter`, **components**)
are NOT touched — all three are replaced by F2d.5 / deleted by F2d.7. Rewiring them now = doomed work.

## Verification

| Check | Command | Result | Real vs fixture |
|-------|---------|--------|-----------------|
| TDD — failing test first | `useApprovalInstanceQuery.test.tsx` (5 cases) written before the hook | red (module absent) → green | real |
| Hook unit tests | `vitest run …/useApprovalInstanceQuery.test.tsx` | `5 passed` | real (queryKey/enabled/404→null/etag/non-404) |
| Cockpit regression (real `QueryClient`) | `vitest run …/useDocumentApprovalArtifact.test.tsx` | `19 passed` | real |
| Combined | both suites | `Test Files 2 passed · Tests 24 passed` | real |
| Static — types | `npx tsc --noEmit` (full frontend) | clean (exit 0) | — |
| Grep gate — adapter | `grep -n 'setInterval\|isStale' useDocumentApprovalArtifact.ts` | no matches | real |
| Grep gate — migrated | `grep -n 'useApprovalInstanceQuery' useDocumentApprovalArtifact.ts` | present (import + call) | real |

## Acceptance vs spec Validation Gate

| Criterion | Met? | Evidence |
|-----------|------|----------|
| Hook keys on `QK.approval.instance(id)`, `enabled`-gated | yes | `useApprovalInstanceQuery.test.tsx › keys + enabled` |
| `queryFn` seeds `etagCache` via `getInstance` | yes | seed lives in `getInstance` (`approvalApi.ts:55-58`); `queryFn` delegates; cockpit 404-no-seed test green |
| 404 → `null` (no error, no etag) | yes | `… › 404 resolves null, no etag`; cockpit `does NOT seed a "v0" etag on the 404 branch` |
| non-404 propagates as error | yes | `… › non-404 rejects` |
| `QK.approval.instance` invalidation refetches + mode re-derives | yes | cockpit suite runs on a real `QueryClient`; `refetchInstance`→`query.refetch` |
| signoff/publish If-Match still resolves from `etagCache` | yes | seeding path unchanged; 404-no-seed + success-seed behavior preserved |
| Grep gate: `setInterval`/`isStale` gone from adapter | yes | grep no matches |
| build/types clean | yes | `tsc --noEmit` exit 0; 24/24 green |

## Review disposition

- Independent `caveman:cavecrew-reviewer` (sonnet, 4 files): **0🔴 · 1🟡 · 1❓.**
  - 🟡 `enabled: Boolean(documentId) && enabled` is an extra defensive condition beyond the spec's
    "enabled gates on hasActiveContext" line — reviewer confirms **no-op in practice** (documentId is
    always a non-empty route param at the call site). Kept as harmless hygiene (matches the plan snippet).
  - ❓ `instanceLoading` from v5 `isLoading` would not surface the spinner during a post-404 **manual
    refetch** (isLoading is first-fetch-only). **Fixed at root:** sourced from `isFetching` instead
    (fires on every fetch), strictly closer to the old imperative semantics. Re-ran gates → 24/24 green,
    tsc clean.
  - Everything else confirmed against the spec: 404→null before any etag read, ETag side-effect intact,
    `refetchInstance` delegates correctly, only `isStale` removed from the public shape, harness wrapper
    correct (`retry:false`), no v5 API misuse.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `refetchInstanceRef` ordering hack (`ApprovalCockpitPage.tsx:91,153`) | it's a **page** concern (adapter grep gate doesn't cover it); the page is rebuilt in F2d.5 and deleted in F2d.7 | F2d.5 screen rebuild / F2d.7 cockpit retirement |
| `onRefetchInstance` prop thread (`ApprovalSidebar`→`DecisionFooter`) → `queryClient.invalidateQueries` | the sidebar/footer are replaced by the F2d.5 single-screen surface; rewiring doomed components now = local maximum | F2d.5 single-screen shell |
