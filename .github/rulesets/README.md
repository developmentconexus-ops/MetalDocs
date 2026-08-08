# Branch rulesets

`main.json` is an export of the ruleset applied to the default branch. It is
checked in so the gate configuration is reviewable in a diff instead of living
only in the repository settings UI, where a change leaves no trace.

**This file is documentation, not the source of truth.** GitHub does not read it.
The live ruleset is the one that enforces. Re-export after any settings change:

```bash
gh api repos/leandrotcawork/MetalDocs/rulesets/20560142 > .github/rulesets/main.json
```

## What it enforces

- `main` cannot be deleted or force-pushed.
- All changes arrive through a pull request. Review threads must be resolved;
  approvals are set to 0 because a solo author cannot approve their own PR, and
  a rule that cannot be satisfied is a rule that gets bypassed.
- 22 status checks are required. Every one is a deterministic check that was
  observed green on a real PR before being listed.

## Why some green checks are not required

A required check is matched **by name**, and a check that never reports counts
as pending forever — so a workflow with a `paths:` filter cannot be required
without deadlocking every PR that does not touch those paths. That rules out
`e2e-coverage-gate`, `openapi-breaking`, `perf`, `req-traceability`, and
`supply-chain` until their filters come off. Diff-scoping belongs one level
down, in `verify --profile=changed`, where the job still always reports.

See the A1 verification ledger §1b for the full argument.

## Why some checks are not required yet

These are red for reasons recorded as defers, not for reasons this ruleset
should paper over. They are promoted when their defect closes, not before:

| Check | Blocked on |
|---|---|
| `conformance` | D-13 test-discipline |
| `gate` | D-1 req-traceability |
| `DB schema dictionary coverage` | D-5 |
| `E2E smoke (approval flows)` | needs an application stack CI does not provision |
| `Perf suite (reduced — PR gate)` | same, plus a `PERF_DATABASE_URL` secret that does not exist |
| `full` | required as soon as one complete result confirms it is green |

## If a required check name changes

There are no bypass actors, deliberately. That means renaming a job without
updating this ruleset will deadlock every open PR: the old name never reports
and nothing can merge. Recovery is an owner-level API call:

```bash
gh api -X PUT repos/leandrotcawork/MetalDocs/rulesets/20560142 --input .github/rulesets/main.json
```

Keep the job `name:` and the required context in the same commit.
