# F2 — Evidence

## Commands + real output

- `npx vitest run src/features/approval/api/__tests__/approvalApi.test.ts` → **Test Files 1 passed ·
  Tests 5 passed** (submit + reviewVerdict + listInbox filters + createDelegation + revokeDelegation).
- `npx tsc --noEmit -p tsconfig.build.json` → 5 errors total, **ALL 5 outside the F2 change set**
  (ApprovalTimelinePanel / InboxStack / InboxPage / SignoffDetailPage / useDocumentApprovalArtifact
  test fixtures — F0-regen fallout, tracked + fixed separately as task #11). **Zero** errors originate
  in any of the 5 F2 files (reviewer-confirmed by attribution).

## TDD proof

- Implementer confirmed RED-first: the 4 new cases fail before implementation
  (`TypeError: X is not a function` / wrong query params), GREEN after. Tests authored before impl.

## Runtime proof (observable change) + fixture-vs-real

- **Unit level (real vitest, mocked `global.fetch`):** asserts the actual wire contract — reviewVerdict
  POSTs `/api/v1/approval/instances/i1/stages/s1/review-verdict` with `If-Match: v3` + a uuid
  `Idempotency-Key` + the body; listInbox forwards `stage_kind`/`due_before`/`scope` as query params;
  createDelegation POSTs `/api/v1/approval/delegations` with a uuid key; revokeDelegation DELETEs and
  **resolves on a 204** (no throw). Labeled **fixture/mock** — fetch is stubbed.
- **Real end-to-end** (a live review verdict moving an instance to `changes_requested`, a real
  delegation create/revoke) is exercised in the **F8 live-QA walkthrough** against the running stack.

## Key design decisions (verified against runtime truth)

- **204 revoke** → uses `mutate('DELETE', …)`; `mutationClient.ts:54-56` short-circuits
  `res.status === 204 → undefined` before any `res.json()`, so an empty body never throws. No
  `requestRaw` detour needed; revoke inherits the same auto-Idempotency-Key + problem+json mapping.
- **Generated DTOs only (ADR 0035):** the only local type is `ReviewVerdictArgs` (the sanctioned
  instanceId/stageId/etag argument wrapper); every body/response aliases `components['schemas'][...]`.
- **Hook placement:** `queries/` — reviewer confirmed this is the house convention (all react-query
  hooks, reads + writes, live in `queries/`; `approval/hooks/useSignoffMutation.ts` is the pre-existing
  outlier). No move.
- **Invalidation:** reviewVerdict → `QK.approval.all` + `QK.documents.all` (matches useSignoffMutation);
  delegations → `QK.approval.all`. Replaces any manual refresh (spec C2: no staleness banner).

## Review / QA disposition

- Independent reviewer subagent (separate from implementer): **APPROVE**, 0 Critical, 0 Major, 2 Minor
  (both no-fix: an explicit-`undefined` DELETE body matching `publish`'s pattern; `scope` typed to the
  single wired literal `'oversee'`). Tests judged non-tautological. Typecheck errors all attributed
  outside F2.

## Process note (recorded honestly)

- The first implementer subagent mis-reported (delegated to a nested agent, returned a meta-summary
  with no landed work); it was re-driven and did the work directly. A duplicate concurrent pass left a
  stray `git stash` — inspected (identical to the applied F2 diff) and dropped. Final state verified by
  the main session (5/5 green) before this evidence was written.

## Bounded defers

- None new. Delegation admin UI + oversee dashboard remain appetite rabbit-holes (out of M2c);
  `scope: 'oversee'` param is wired for the future dashboard.
