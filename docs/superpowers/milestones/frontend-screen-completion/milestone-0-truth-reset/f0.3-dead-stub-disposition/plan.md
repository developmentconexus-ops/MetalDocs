# Feature F0.3 — Plan

> **Milestone:** 0 — Truth reset  ·  **Folder:** `f0.3-dead-stub-disposition`
> Engine: inline plan (bounded deletion — `superpowers:writing-plans` not needed). Input: `spec.md`.
> Blast radius verified this session (see spec Consumer contract).

## Plan

Deletion-by-the-root, compiler-driven. Order matters: delete the modules first so `tsc` red-flags every
remaining reference, then clean each reference until `tsc` + vitest are green.

1. **Delete the dead source** (6 files + 2 emptied dirs):
   - `src/features/operations/pages/OperationsPage.tsx`
   - `src/features/operations/routes.tsx`
   - `src/features/audit/pages/AuditPage.tsx`
   - `src/features/audit/routes.tsx`
   - `src/components/OperationsCenter.tsx`
   - `src/components/OperationsCenter.module.css`
   - then the now-empty `src/features/operations/` and `src/features/audit/` dirs.
2. **Unwire the router** (`src/app/AppRouter.tsx`): remove the `operationsRoutes` (line 10) and
   `auditRoutes` (line 4) imports and their two spread entries (lines 31, 39). Leave all other route
   arrays and the `{ path: '*', … }` catch-all untouched.
3. **Re-express the F0.2 guard** (`src/app/AppRouter.test.tsx`): drop the `operationsRoutes` import and
   the assertion that referenced it; keep/strengthen the invariant as "`dashboardRoutes` declares
   exactly one root-level `index:true`, and it is the sole root index carrier." The F0.2 contract
   (`/` → one root index = Dashboard) is preserved.
4. **Verify** against the spec Validation Gate (ls-gone, grep=0 ×3, tsc 0, targeted test green, full
   suite ≤ baseline-36). Capture in `evidence.md`.

## Files touched
- DELETE: the 6 files above (+ 2 dirs)
- EDIT: `src/app/AppRouter.tsx` (remove 2 imports + 2 spreads)
- EDIT: `src/app/AppRouter.test.tsx` (remove `operationsRoutes` import; re-express invariant)

## Test strategy
- Compiler-as-test: `tsc --noEmit` must go red on the dangling imports after step 1, green after steps
  2–3. The F0.2 unit guard (`AppRouter.test.tsx`) is the durable regression assertion and must stay
  green. Full `vitest run` confirms no new failures from the deletion (orphan-import / broken-router
  check named in milestone validation §3).

## Ordering
F0.3 after F0.2 (F0.2 removed the operations *index* ambiguity; F0.3 removes the operations *route +
files*). F0.3 before F0.4 (F0.4 records the CUT list + DoD; it does not depend on the deletion but the
tracker truth is cleaner once dead surface is gone).
