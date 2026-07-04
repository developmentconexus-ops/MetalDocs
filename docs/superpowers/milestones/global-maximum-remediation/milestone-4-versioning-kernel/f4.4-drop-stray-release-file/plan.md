# F4.4 plan

Mechanical untrack. No source/test/contract change.

## Task
1. `git rm --cached docs/release/v2-name-inventory.md` — remove from index, keep on disk.
2. Commit dedicated (this fix + its f4.4 spec/plan/evidence records).
3. Verify `git ls-files docs/release/` empty; file still present untracked; `go build ./...` green.
4. Re-dispatch milestone-validator.

## Gate
`git ls-files docs/release/` empty · file on disk (`??`) · fix commit touches no other functional path ·
build green.
