# F2 — API layer: reviewVerdict + inbox filters + delegations

> **Milestone:** M2c approval-screen-fe · **Consumer:** the review sidebar (F4) which posts a
> ready/request-changes verdict, the worklist (F5) which filters the inbox by kind/due/scope, and the
> delegation affordance (surfaced, admin UI out of scope). All consume `@metaldocs/editor-ui`-free
> generated DTOs only (ADR 0035).
> **Status:** Approved — 2026-07-07. Approval line below.

## Consumer contract (what downstream requires, defined before producer)

1. **Review verdict mutation.** `approvalApi.reviewVerdict({ instanceId, stageId, etag, body })` POSTs
   `/approval/instances/{instanceId}/stages/{stageId}/review-verdict` with:
   - `If-Match: <etag>` (optimistic concurrency on the instance version) and an auto-generated
     `Idempotency-Key` (uuid) — both **required** by the endpoint.
   - `body: ReviewVerdictRequest = { verdict: 'ready' | 'request_changes'; comment?: string }`
     (comment required by the server when `verdict === 'request_changes'`; the FE surfaces that as a
     form rule in F4, the api layer just forwards).
   - Returns `ReviewVerdictResponse = { verdict_id; was_replay; outcome }`.
   - Types come from `components['schemas']['ReviewVerdictRequest'|'ReviewVerdictResponse']` — **no
     hand-written body types**. Error mapping is the shared `problem+json` path (same as `signoff`).
   - A `useReviewVerdictMutation` hook wraps it and, on success, invalidates the approval subtree so
     the instance + inbox re-fetch (spec C2: no manual staleness banner).
2. **Inbox filter params.** `ListInboxParams` gains `stage_kind?: 'review' | 'approval'`,
   `due_before?: string` (RFC3339 UTC), `scope?: 'oversee'` — all already accepted by the
   `listApprovalInbox` operation. `listInbox` forwards them as query params; `useInboxQuery`'s query
   key already includes the full `params` object, so distinct filters cache distinctly.
3. **Delegations mutations.** `createDelegation(body: CreateApprovalDelegationRequest)` →
   POST `/approval/delegations` (auto `Idempotency-Key`), returns `ApprovalDelegation`;
   `revokeDelegation(id)` → DELETE `/approval/delegations/{id}` (no body, 204). Hooks
   `useCreateDelegationMutation` / `useRevokeDelegationMutation` invalidate the approval subtree.
   `delegator_id` is session-derived server-side — never sent in the body.

## Non-goals

- **Delegation admin screen** (list/manage UI) — appetite rabbit-hole; only the api+hooks layer here.
- **Oversee dashboard** — `scope: 'oversee'` param is wired; the dashboard consuming it is out of M2c.
- **Client-side verdict validation beyond forwarding** — the "comment required on request_changes"
  UX rule lives in F4's form; the api layer forwards whatever it is given (server is the authority).
- **New generated types** — F0 already regenerated `ReviewVerdictRequest/Response`, delegation
  schemas, `stage_kind`, `changes_requested`. F2 consumes them; it does not touch openapi.

## Interview record (B1.5) — api-layer shape

| # | Question | Finding (runtime truth, file:line) | Decision |
|---|----------|-----------------------------------|----------|
| 1 | How do existing mutations set headers + map errors? | `signoff` → `mutate('POST', url, body, {resourceId})` (`approvalApi.ts:74`); `mutate` (`mutationClient.ts:22`) auto-gens `Idempotency-Key` via `crypto.randomUUID()`, pulls `If-Match` from `etagCache.get(resourceId)` or explicit `ifMatch`, maps `problem+json` via `parseProblem` → `ApprovalError`. | reviewVerdict/delegations reuse `mutate` — no new client. |
| 2 | Where does the instance etag live for If-Match? | `etagCache` keyed by `documentId` (set by `getInstance`). reviewVerdict is instance/stage-scoped in its URL but the version is the instance's. | reviewVerdict accepts an explicit `etag` → `ifMatch` (master-plan contract); caller passes the cached instance etag. |
| 3 | Does `mutate` handle a 204 (delegation revoke) + undefined body? | Must verify in `mutationClient.ts`; if it force-parses JSON, revoke uses `requestRaw` (`lib/api`) instead. | Implementer checks 204 handling; pick the path that does not throw on empty body. |
| 4 | Inbox filter params already on the wire? | Yes — `listApprovalInbox` params include `stage_kind`, `due_before`, `scope: 'oversee'` (api-types ~7600). `ListInboxParams` (`approvalTypes.ts:52`) only has `area_code/limit/offset`. | Extend `ListInboxParams` + forward; no openapi change. |
| 5 | Existing delegations FE layer? | None (grep zero). Schemas/operations exist in api-types only. | New `useDelegationsMutations.ts`; nothing to duplicate. |
| 6 | Invalidation convention? | `useSignoffMutation` invalidates `QK.approval.all` (whole `['approval']` subtree) + `documents.all` (`useSignoffMutation.ts:56`). | reviewVerdict/delegations invalidate `QK.approval.all` (covers instance + inbox) + `documents.all`. |

## Validation Gate

- New test `approvalApi.test.ts` cases (house pattern `vi.spyOn(global,'fetch')`):
  - reviewVerdict POSTs the right URL, sets `If-Match: v3` and a uuid `Idempotency-Key`, sends the body.
  - listInbox forwards `stage_kind`/`due_before`/`scope` as query params.
  - createDelegation POSTs `/approval/delegations` with a uuid `Idempotency-Key`; revokeDelegation
    DELETEs `/approval/delegations/{id}` and does not throw on 204.
- `vitest run` (targeted files) PASS; no regression in the existing `approvalApi.test.ts` submit case.
- Typecheck clean (generated DTOs only — grep the new files for hand-written body interfaces → none).

## Approval

- **Contract approved:** 2026-07-07 (main session, per ratified master plan §F2; consumer contract is
  the F4 review sidebar + F5 worklist + delegation affordance; operator holds the HS-1 close gate).
