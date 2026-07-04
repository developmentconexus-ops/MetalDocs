# F4.4 — drop stray docs/release file (HS-4 fix feature)

> **Origin:** milestone-validator FAIL (`qa/milestone-qa.md`), C3/C6 scope-drift breach.
> **Approved for code: 2026-07-04** (mechanical untrack; no source/test/contract change).

## Consumer contract

- **Consumer = the repo history + the M4 gate.** The M4 diff MUST NOT touch any `docs/release/` path
  (contract §5 / CLAUDE.md "never commit `docs/release/`"). Commit `51950c26` (F4.1 Task C) accidentally
  re-added the tracked file `docs/release/v2-name-inventory.md` (316 lines) while staging FE changes — the
  file was `?? docs/release/` (untracked) at session start.
- **Required outcome:** `docs/release/v2-name-inventory.md` is **untracked** again (removed from the git
  index, kept on disk), in a dedicated commit with recorded rationale. After the fix, `git ls-files
  docs/release/` is empty and no M4 commit going forward touches `docs/release/`.

## Non-goals

- NOT deleting the file from disk (it stays as an untracked working file, its session-start state).
- NOT adding a `.gitignore` entry (surfaced to operator as optional hardening at HS-1; out of this
  minimal fix scope).
- NOT any source, test, contract, or generated-code change.
- NOT rewriting history of `51950c26` (untrack-forward, not a rebase).

## Validation gate

`git rm --cached docs/release/v2-name-inventory.md` committed on its own · `git ls-files docs/release/`
returns empty · the file still exists on disk (untracked, `??`) · `go build ./...` still green (unchanged;
sanity) · no other path in the fix commit.

## Interview record

| Q | Answer |
|---|---|
| Fix scope | Validator-named `f4.4-drop-stray-release-file`: untrack the stray file in a dedicated commit, record it, confirm no `docs/release/` path touched. No source/test/contract change. |
