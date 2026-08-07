# A1 handoff — make the verifier one trusted product

Issue: #87. Program: `docs/superpowers/analysis/2026-08-07-engineering-remediation-program.md`.
Method: `docs/engineering/repo-audit-playbook.md`.

## State (verified 2026-08-07)

```
gh run list --limit 100  -> 100 failure, 0 success
gh run list --limit 200  -> zero success
gh pr list               -> last merged PR #73, 2026-06-05
gh api .../rulesets      -> []
```

- 20 workflows in `.github/workflows/`, 16 on `pull_request`.
- No PRs for 2 months. All work went direct to `main`. PR gates never fire.
- Red is the normal state, so no check carries signal.
- Nothing is `required`. Every check annotates, including ones whose own comments
  claim BLOCKING (`openapi-breaking.yml:9`).
- Repo went PUBLIC 2026-08-07. This is why rulesets are now available at zero cost;
  private repos get `403 Upgrade to GitHub Pro`. Public is the enforcement floor.

Known inert controls:
- `golangci-lint.yml:26` — `only-new-issues: true` (new lines only).
- `invariants.yml:29` — `continue-on-error: true`.
- `test-full.yml` — `push` only; full suite runs after merge, never before.
- 8 of 15 `scripts/{check,verify,run}-*` referenced by zero workflow.

## Rule that governs this axis

A control that does not fire is absent. Firing hierarchy, always climb highest reachable:
unrepresentable > boot-fatal > red build > runtime assertion > discipline. Discipline is not
a control.

Every guard ships with a fixture that makes it fail, in the same change. A guard nobody has
seen fail is a guard nobody knows works.

## Scope

1. Get the 20 workflows genuinely green. Fix causes, do not delete checks to reach green.
2. Kill inert flags: `only-new-issues`, `continue-on-error`, defaulted-off scanners.
3. Wire the 8 orphan scripts, or delete them. An unwired script is a lie.
4. Move the full suite to pre-merge; keep a fast tier under ~5 min.
5. One `verify` entry point with profiles (`fast`, `changed`, `pr`, `full`). CI calls it and
   nothing else, so "green locally" and "green in CI" are the same claim.
6. Promote deterministic checks to `required` in a `main` ruleset. Only checks that are
   deterministic, locally reproducible, and carry a failing fixture earn `required`.
7. Negative fixture per guard.

## Not in scope

- Fixing the defects the newly-green checks reveal — those belong to axes A2–A9.
- AI PR review (CodeRabbit-class). Annotation layer, add after the ruleset exists, never before.
- Any product code change beyond what is needed to make a check pass honestly.

## Acceptance

`gh run list` shows green on `main`. `gh api repos/.../rulesets` returns a ruleset with
required checks. Each required check has a committed fixture proving it fails.

## Open operator decision — blocks step 6

~125 local commits not pushed. Turning on the ruleset protects `main` and forces those through
a PR.

- (a) ruleset first, adopt PR flow now
- (b) green CI first, lock `main` after

## Constraints

Zero spend, free/OSS only. Never push without explicit permission. Never `--no-verify`.
