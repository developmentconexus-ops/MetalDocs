# Feature F7.5 — Plan

> Engine: inline. Spec: `./spec.md` (approved).

## Files touched

| File | Change |
|------|--------|
| `wiki/architecture/api-contract.md` | New "Response-body typing gate (H-D — honest two-part)" section; `Last verified` stamp → 2026-06-20 (M7). |
| `frontend/apps/web/src/lib/api-types/index.d.ts` | Regenerated (`npm run gen:api`) — reflects F7.4's 4 documents 200 schemas. |

## Task order

1. Add the honest-gate section to api-contract.md (Part A + Part B + allowlist + the alias/multi-line
   blindspot it closes).
2. Update the `Last verified` stamp.
3. `npm run gen:api` in `frontend/apps/web`; inspect the diff (should add the 4 documents response types).
4. Final proof, whole-repo (record in evidence.md): Part A, Part B, `go generate` no-diff, `go build`,
   `go test -count=1 ./...`.
5. Evidence; commit.

## Test strategy

- No new Go tests — F7.5 is documentation + FE type regen + whole-repo proof. The "test" is the honest
  two-part gate run from clean state across the whole repo plus the full suite.

## Risk / rollback

- FE regen could surface unrelated drift (if openapi.yaml had other undeclared changes). Mitigate by
  inspecting the `index.d.ts` diff — it must be limited to the 4 documents types.
- Rollback = `git checkout` of the 2 files.
