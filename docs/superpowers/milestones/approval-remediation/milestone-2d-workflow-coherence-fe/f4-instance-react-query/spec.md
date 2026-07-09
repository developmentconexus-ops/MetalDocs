# Feature F2d.4 — Spec (consumer-contract-first)

> **Milestone:** 2d · **Feature:** `f4-instance-react-query`
> **Approved (pre-code):** 2026-07-09 (code-grounded interview; no operator ambiguity — the milestone.md
> row + generated types + existing `QK.approval.instance` fully determine the contract).

## Consumer contract (what the consumer requires BEFORE the producer is built)

**Producer:** a new query hook `useApprovalInstanceQuery(documentId, enabled)` (react-query) that owns the
approval-instance read. **Consumers:**

1. **F2d.5 single screen (primary, not yet built):** subscribes to `QK.approval.instance(documentId)`;
   reads `data`/`isLoading`/`isError`; after a signoff/verdict/publish mutation it invalidates
   `QK.approval.instance(documentId)` and the instance refetches automatically → `deriveWorkspaceMode`
   re-derives. No imperative refetch callback threaded through the tree.
2. **`useDocumentApprovalArtifact` (cockpit adapter, doomed — deleted in F2d.7):** migrated to consume the
   new hook so its instance state is react-query-backed. Preserves its public shape EXCEPT the dead
   `isStale` field (removed). `refetchInstance(): Promise<void>` stays (cockpit page + sidebar still call
   it) but is backed by `query.refetch`.

**Invariants the producer must hold:**
- **ETag seeding is preserved for free.** `getInstance` already writes `etagCache.set(documentId, etag)`
  on success (`approvalApi.ts:55-58`); the `queryFn` calls `getInstance`, so every fetch/refetch reseeds
  the cache. The If-Match write path (`mutationClient.ts` reads `etagCache`) is unchanged.
- **404 = "no active instance", not an error.** The `queryFn` catches a 404 and resolves `null`
  (parity with the old imperative branch, `useDocumentApprovalArtifact.ts:131-134`); any other status is a
  real query error. 404 never seeds an etag (the throw precedes the etag read in `getInstance`).
- **Query key:** `QK.approval.instance(documentId)` (already registered, `queryKeys.ts:87-88`) — no new key.
- **`enabled`** gates the fetch on the caller's readiness (the adapter passes `hasActiveContext`, matching
  the old "only fetch once `content_hash` is confirmed" effect).

## Interview record (code-grounded — no operator questions needed)

| Q | Resolution | Grounding |
|---|------------|-----------|
| New key or reuse? | Reuse `QK.approval.instance(documentId)` | Already in `queryKeys.ts:87-88` |
| Where does ETag seeding live post-migration? | Untouched — inside `getInstance`; `queryFn` inherits it | `approvalApi.ts:55-58` decoupled from React state |
| Is 404 an error under react-query? | No — `queryFn` maps 404→`null` (old imperative parity) | `useDocumentApprovalArtifact.ts:131-134` |
| Does F2d.4 delete `refetchInstanceRef` + `onRefetchInstance`? | **No — bounded defer.** Those live in `ApprovalCockpitPage.tsx` (a page) and thread through `ApprovalSidebar`→`DecisionFooter` (components), ALL replaced by F2d.5 / deleted by F2d.7. Rewiring them to `queryClient.invalidateQueries` now = doomed work on components about to die (locks a local maximum). | grep: `refetchInstanceRef` only in `ApprovalCockpitPage.tsx`; `onRefetchInstance` only in sidebar/footer |
| Is `isStale` (artifact field) safe to delete now? | Yes — dead. No consumer reads `artifact.isStale` (grep across `src` shows only its own definition). | grep clean |
| Grep-gate scope? | Milestone row says "gone from the approval/document **adapters**." `setInterval`/`isStale` are in the adapter (`useDocumentApprovalArtifact.ts`) → removed here. `refetchInstanceRef` is in a **page** → out of adapter scope, deferred with the page. | milestone.md F2d.4 acceptance |

## Non-goals (mandatory)

- **No** rewire of `ApprovalCockpitPage`'s `refetchInstanceRef` ordering hack or the
  `ApprovalSidebar`→`DecisionFooter` `onRefetchInstance` prop thread — owned by F2d.5 (screen rebuild) /
  F2d.7 (cockpit deletion). Bounded defer, triggers recorded in evidence.
- **No** new OpenAPI/DTO change (contract-frozen; this is FE state plumbing only).
- **No** behavior change to signoff/verdict/publish mutations or the If-Match path.
- **No** F2d.5 screen work.

## Validation Gate (acceptance + named tests + proof commands)

| Acceptance | Named test | Proof |
|------------|-----------|-------|
| `useApprovalInstanceQuery` keys on `QK.approval.instance(id)` and is `enabled`-gated | `useApprovalInstanceQuery.test.tsx › keys + enabled` | `vitest run …/useApprovalInstanceQuery.test.tsx` |
| `queryFn` seeds `etagCache` on success | `… › seeds etag on success` | same |
| `queryFn` maps 404 → `null` (no error, no etag) | `… › 404 resolves null, no etag` | same |
| `QK.approval.instance` invalidation refetches + mode re-derives | cockpit suite (real `QueryClient`): after `invalidateQueries` the instance refetches | `vitest run …/useDocumentApprovalArtifact.test.tsx` |
| signoff/publish If-Match still resolves from `etagCache` | cockpit 404-no-seed test + etag-seed on success stay green | same |
| Grep gate: `setInterval`/`isStale` gone from `useDocumentApprovalArtifact.ts`; instance state is react-query | `grep -n 'setInterval\|isStale' useDocumentApprovalArtifact.ts` → none | grep |
| build/types clean | `npx tsc --noEmit` | 0 errors |

## ADR?

None. No durable architectural decision — this applies the established react-query + `QK` pattern
(`useDocumentDetailQuery`) to the approval instance. The bounded defer is a sequencing choice, not a
decision of record.
