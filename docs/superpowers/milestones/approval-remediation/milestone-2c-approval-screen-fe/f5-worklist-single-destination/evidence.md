# F5 — Evidence

## Commands + real output

- `npx vitest run src/features/approval src/lib/inbox` → **Test Files 20 passed (20) · Tests 141
  passed (141)** (~9.6s). Includes new `InboxFilters.test.tsx`, `sortByDue.test.ts`, and the 5
  extended `InboxPage.test.tsx` gate cases. Re-run independently by the reviewer from clean state —
  same result.
- `npx tsc --noEmit -p tsconfig.build.json` → **clean, zero output, exit 0** (implementer + reviewer).
- `grep -n "/edit" src/features/approval/pages/InboxPage.tsx` → **ZERO** (main-session + reviewer).
  No author-editor navigation remains in the worklist. Remaining `documents/.*edit` hits elsewhere are
  the ApprovalCockpitPage mock module paths + two explanatory comments — no live bug assertion.
- `grep -rn "useInboxQuery()" src/features/approval` → **ZERO** — no bare (paramless) callers; the
  worklist now always passes mapped filter params.
- `git diff` on `approvalApi.ts`, `approvalTypes.ts`, `useInboxQuery.ts`, `lib/api-types` → **EMPTY**
  (F5 is FE-only; the API layer was F2-complete). No new `interface/type .*Response` — generated DTOs
  only (ADR 0035).

## TDD proof

- Implementer confirmed RED-first: `Test Files 2 failed | 1 passed (3) · Tests 7 failed | 18 passed
  (25)` before implementation — `InboxFilters` module absent, `getByLabelText('Estágio'/'Supervisão')`
  not found, cockpit-nav assertion unmet, teaching-empty copy missing, oversee-403 note missing.
  GREEN after implementing. Tests authored first.

## Runtime proof (observable change) + fixture-vs-real

- **C3 single destination (the core fix):** `InboxPage.openDocument` now navigates
  `/approvals/${item.document_id}` (the cockpit) — no `fetchActiveDocumentInstance`, no
  `/documents/:id/edit`. This closes a real W2-class hole: the worklist previously sent
  reviewers/approvers into the **author editor** (writable session) — the exact vector F3 killed at
  the cockpit. Approve/Reject action buttons (`openDecisionFlow`) still deep-link the cockpit with
  `?decision=…` (one destination, optional intent). Asserted: new test navigates `/approvals/doc-cockpit`
  AND `fetchActiveDocumentInstance` not called on primary open. Labeled **fixture/mock** (vitest +
  mocked `useNavigate`). Real end-to-end open exercised in F8 live QA.
- **Filters → params:** `toInboxParams` maps `stageKind→stage_kind`, `next7→due_before=now+7d ISO`,
  `overdue→due_before=now ISO` + client `overdue` guard, `oversee→scope:'oversee'`. Test asserts the
  exact ISO values and `useInboxQuery` mock call args (non-tautological).
- **Sort:** `sortByDueAsc` copies (`[...items]`) before sort — no in-place mutation of the react-query
  cache; `due_at===null` sorts last; ascending otherwise. Applied once in InboxPage before both
  stack + timeline (single shared path). Non-mutation asserted in `sortByDue.test.ts`.
- **Teaching empty ×2:** driven by `isInboxFilterActive(filters)` → `isFiltered` prop. no-work vs
  filtered-empty render distinct PT-BR copy in both stack + timeline; both asserted distinctly.
- **Overdue chip:** `formatDueRelative(item.due_at)` per item; `overdue` → `--danger`. Null `due_at`
  renders em-dash (no fabricated date). Test asserts `/atrasado há/` for a past due_at.
- **oversee reactive 403 (D2):** no preemptive probe — a `useEffect` reacts to `error instanceof
  ApiError && error.status === 403` while oversee active, shows inline `role="alert"` "Você não tem
  permissão de supervisão.", and reverts `filters.oversee→false`; the flipped guard prevents any
  re-trigger loop. Test confirms toggle → alert → unchecked.

## Key design decisions (verified against runtime truth)

- **Single destination was a behavior gap, not a routing one.** Routing already had one `/approvals`
  list + one `/approvals/:documentId` cockpit (investigator §8); the fix is retargeting the worklist
  item open away from the author editor to the cockpit, which F3/F4 already mode-resolve.
- **oversee gated reactively, not by probe** — a preemptive `scope=oversee` probe on every mount is
  wasted latency; reacting to a real 403 is the lower-cost, behavior-equivalent design (spec D2).
- **doc-type filter deferred** — no `doc_type` param in the contract; adding one is a
  backend/contract change (HS-2 boundary, outside M2c FE scope). Delivered stage-kind + due (spec D1).

## Review / QA disposition

- Independent reviewer subagent (separate from implementer, own tools, no edits): **APPROVE**,
  **0 Critical, 0 Major, 2 Minor**. All 12 adversarial checks PASS, each runtime-proven (executed
  suites + direct code reads). Reviewer re-ran vitest (141/141) + tsc (clean) from clean state.
- **Deleted-test legitimacy confirmed:** the 3 removed InboxPage tests asserted ONLY the old
  `/documents/:id/edit` behavior (obsolete by C3) — no real invariant guard lost. Per the
  legacy-test-deletion rule (task-scaffolding tests deleted when superseded; contract/invariant
  guards repaired). The new single-destination test supersedes them.
- **2 Minor findings (accepted, non-blocking):**
  1. `InboxPage.tsx:47,55-60` — `overseeDenied` note not cleared when the user re-enables the oversee
     toggle after a prior 403 (stale until next resolution). Spec ("revert once, no loop") satisfied;
     small UX edge. **Deferred to F7** (a11y/polish already touches these controls).
  2. `filtersOpen` defaults `true` (panel visible on load) — explicitly plan-permitted ("panel
     visible is fine"), not a deviation.

## Deviations to surface at HS-1

- **D1 — doc-type filter deferred** (contract gap; C3 lists it, no param exists).
- **D2 — oversee gated reactively, not by probe** (behavior-equivalent, lower latency).

## Bounded defers

- Minor #1 (stale oversee-denied note) → F7 polish, no trigger beyond that.
- doc-type filter (D1) → post-M2c contract change if the operator wants it; flagged HS-1.
