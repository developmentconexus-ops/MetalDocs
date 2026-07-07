# F2 — Plan

Seeded from master plan §F2 + the F2 investigator map (runtime truth). TDD via a fresh implementer
subagent + independent reviewer subagent.

## Tasks

1. **[TDD] Failing tests** — extend `frontend/apps/web/src/features/approval/api/__tests__/approvalApi.test.ts`
   (house pattern: `vi.spyOn(global,'fetch').mockResolvedValue(new Response(...))`, assert
   `fetchMock.mock.calls[0]` `[url, init]`, headers from `init.headers`):
   - reviewVerdict → URL `/api/v1/approval/instances/i1/stages/s1/review-verdict`, `If-Match: v3`,
     `Idempotency-Key` matches `/^[0-9a-f-]{36}$/`, body `{ verdict:'request_changes', comment:'…' }`.
   - listInbox with `{ stage_kind:'review', due_before:'2026-07-01T00:00:00Z', scope:'oversee' }` →
     query string contains all three.
   - createDelegation → POST `/api/v1/approval/delegations`, uuid `Idempotency-Key`, body forwarded.
   - revokeDelegation('d1') → DELETE `/api/v1/approval/delegations/d1`, resolves (no throw) on a 204.
2. **approvalTypes.ts** — extend `ListInboxParams` with `stage_kind?: 'review'|'approval'`,
   `due_before?: string`, `scope?: 'oversee'`. Re-export `ReviewVerdictRequest/Response`,
   `CreateApprovalDelegationRequest`, `ApprovalDelegation` from `components['schemas']` (generated only).
3. **approvalApi.ts** — add:
   - `reviewVerdict({ instanceId, stageId, etag, body }: ReviewVerdictArgs): Promise<ReviewVerdictResponse>`
     → `mutate('POST', \`${BASE}/approval/instances/${instanceId}/stages/${stageId}/review-verdict\`, body, { ifMatch: etag })`.
   - `listInbox` — forward the 3 new params into the query string alongside `area_code/limit/offset`.
   - `createDelegation(body)` → `mutate('POST', \`${BASE}/approval/delegations\`, body, {})`.
   - `revokeDelegation(id)` — **first check `mutationClient.ts` 204 handling**; if `mutate` force-parses
     JSON on empty body, use `requestRaw('DELETE', url, { headers:{ 'Idempotency-Key': crypto.randomUUID() }})`
     (or the no-body path) so a 204 does not throw. Pick the path that matches the client's contract.
4. **queries/useReviewVerdictMutation.ts** (new) — `useMutation` wrapping `reviewVerdict`; `onSuccess`
   invalidates `QK.approval.all` + `QK.documents.all` (matches `useSignoffMutation`).
5. **queries/useInboxQuery.ts** — accept the widened `ListInboxParams` (already passes `params`
   through; query key already includes it — confirm no key change needed).
6. **queries/useDelegationsMutations.ts** (new) — `useCreateDelegationMutation` +
   `useRevokeDelegationMutation`, both invalidate `QK.approval.all`.
7. **Verify** — `vitest run` the approval api + query tests GREEN; package typecheck clean; grep the
   new files for hand-written body interfaces → none.
8. **Review pass** — independent reviewer subagent (spec compliance + generated-DTO-only + error/header
   correctness + 204 safety). Apply accepted findings.

## Files touched

- `frontend/apps/web/src/features/approval/api/approvalApi.ts`
- `frontend/apps/web/src/features/approval/api/approvalTypes.ts`
- `frontend/apps/web/src/features/approval/api/__tests__/approvalApi.test.ts`
- `frontend/apps/web/src/features/approval/queries/useReviewVerdictMutation.ts` (new)
- `frontend/apps/web/src/features/approval/queries/useInboxQuery.ts`
- `frontend/apps/web/src/features/approval/queries/useDelegationsMutations.ts` (new)

## Risks

- **204 on revoke** — mutationClient may assume a JSON body. Resolved by task 3 verification.
- **junction drift** (memory `fe-node-modules-junction-drift`) — if vitest won't run, the real fix is a
  full `pnpm install` in `frontend/apps/web`; do NOT hack around it. Report the exact error if it trips.
