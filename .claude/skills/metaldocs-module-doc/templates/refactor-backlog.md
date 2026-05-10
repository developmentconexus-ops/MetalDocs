# Refactor Backlog — <Module>

> Actionable rows. One row = one PR. Pulled from `wiki/modules/<module>-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** YYYY-MM-DD

## Schema

| Field | Required | Notes |
|---|---|---|
| `id` | yes | `R-001`, `R-002`, … |
| `title` | yes | imperative, scoped to a single PR |
| `debt_id` | yes | either `T-NNN` from tech-debt register **or** `maint:<kind>` for rows that have no debt origin — see allowed kinds below |
| `effort` | yes | XS · S · M · L (≥L = split first) |
| `impact` | yes | Critical · Major · Minor (mirror debt severity for `T-NNN` rows; pick the highest user-visible tier for `maint:` rows) |
| `blocked_by` | optional | other row id or external ticket |
| `owner` | optional | github handle |
| `status` | yes | open · in-progress · merged · cancelled |
| `pr` | optional | PR URL once opened |

### Allowed `maint:<kind>` values

Real backlogs contain rows that are not bugs: retiring a predecessor doc, bumping a dep, fixing a broken anchor, renaming a test file. Forcing those into a synthetic `T-NNN` would fake debt that the register cannot defend. Use one of:

| Kind | Use for |
|---|---|
| `maint:doc-cleanup` | retiring/renaming/repointing wiki docs |
| `maint:dep-bump` | upgrading a dependency with no behavior change |
| `maint:test-only` | adding / refactoring tests with no production change |
| `maint:docs-link` | fixing broken cross-links or stale `path:LL` anchors |
| `maint:migration-cleanup` | retiring a no-op migration or deduping a redundant one |

`scripts/tally_check.sh` validates every backlog row's `debt_id` against this exact set — anything else fails the gate. Add a new kind via SKILL changelog + same-commit edit if a legitimate maintenance category is missing.

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | <imperative one-liner> | T-001 | S | major | — | — | open | — |
| R-002 | <imperative one-liner, no debt origin> | maint:doc-cleanup | XS | minor | — | — | open | — |

## Notes

- Anything `L` or larger: split before opening a PR. Reviewer cost scales worse than linearly with diff size.
- When merged: bump `status` to `merged`, link PR, and remove the linked TD row from the register in the same commit.
