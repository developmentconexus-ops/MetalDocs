# F4.4 evidence — drop stray docs/release file

> **Origin:** M4 milestone-validator FAIL (C3/C6 scope-drift). Fix commit: see below.

## What happened
Commit `51950c26` (F4.1 Task C, FE `rejected` removal) accidentally `git add`-ed
`docs/release/v2-name-inventory.md` (316 lines) — the file was `?? docs/release/` (untracked) at session
start. Contract §5 / CLAUDE.md forbid committing `docs/release/`.

## Fix
`git rm --cached docs/release/v2-name-inventory.md` — removed from the index, kept on disk (returns to its
untracked session-start state). Untrack-forward; no history rewrite of `51950c26`.

## Proof (real output)
```
$ git rm --cached docs/release/v2-name-inventory.md
  rm 'docs/release/v2-name-inventory.md'
$ git ls-files docs/release/
  (empty — no longer tracked)
$ test -f docs/release/v2-name-inventory.md && echo "YES on disk (untracked)"
  YES on disk (untracked)
$ go build ./...
  BUILD OK (unchanged — no functional path touched)
```

## Disposition
- Fix commit touches only the index removal + these f4.4 records. No source/test/contract/generated change.
- **Optional hardening surfaced to operator (HS-1, not done here):** add `docs/release/` to `.gitignore`
  so a future `git add -A` cannot re-stage it. Out of this minimal fix's scope (contract §5 baseline had
  it untracked-but-not-ignored); operator decides.
