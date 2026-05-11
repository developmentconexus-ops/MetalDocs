# Sync Log Entry Template

Append one block to `wiki/modules/<m>/_artifacts/sync-log.md` per sync run. If the file does not exist, create it with this header at top:

```markdown
# Sync log — <m>

> Append-only log of `metaldocs-module-doc-sync` runs against this module. Newest at top.
```

Then prepend each entry directly below the header.

## Entry format

```markdown
## <YYYY-MM-DD> · <one-line change context>

- **Context:** <plan task description | git range | explicit file list>
- **Anchors moved:** N (list: <file>:L<old>→L<new>, ...)
- **Symbols renamed:** N (list: <old>→<new>, ...)
- **T-NNN closed:** <ids or "none"> · evidence: <one line>
- **R-NNN updated:** <ids → new status or "none"> · PR: <url or sha>
- **§11 counts after:** Critical=<n> Major=<n> Minor=<n>
- **Tally gate:** PASS
- **Patched files:** wiki/modules/<m>.md · wiki/modules/<m>-tech-debt.md · wiki/backlog/<m>-refactor.md
```

## Example

```markdown
## 2026-05-12 · resolved iam T-007 (governance logger wired)

- **Context:** plan task "Plan 1 v2 follow-up: wire MembershipGovernanceLogger" · commits abc1234..def5678
- **Anchors moved:** 1 (apps/api/cmd/metaldocs-api/main.go:L217→L221)
- **Symbols renamed:** 0
- **T-NNN closed:** T-007 · evidence: NewAreaMembershipService now receives postgres MembershipGovernanceLogger impl, no nil-arg
- **R-NNN updated:** R-007 → merged · PR: https://github.com/.../pull/482
- **§11 counts after:** Critical=2 Major=4 Minor=5
- **Tally gate:** PASS
- **Patched files:** wiki/modules/iam.md · wiki/modules/iam-tech-debt.md · wiki/backlog/iam-refactor.md
```

## Why this log exists

The next Phase 1 diff scan reads this file to know what has already been reconciled. Without it, the scanner risks re-flagging anchors that were fixed last week. Treat it as the equivalent of a CHANGELOG specifically for doc-sync runs — small but load-bearing.
