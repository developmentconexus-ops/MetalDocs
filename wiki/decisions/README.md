# Decisions

> **Last verified:** 2026-06-13 (Z-27 — status vocabulary added, new-ADR rule added)
> **Purpose:** Canonical ADR landing page and authoring rules.

The maintained decision list is at [index.md](index.md). This file defines authoring rules and the status vocabulary.

## Status vocabulary

Every ADR **MUST** carry a `> **Status:**` header as the first or second line of its opening blockquote, using one of the canonical values below. No other values are permitted.

| Status value | Meaning |
|---|---|
| `Accepted` | Decision ratified and in force. |
| `Accepted (amended YYYY-MM-DD — <summary>)` | In force; one or more amendments applied after initial ratification. |
| `Accepted (in execution)` | Ratified; a multi-phase implementation plan is partially complete. |
| `Accepted (fully executed YYYY-MM-DD)` | Ratified; all implementation phases complete. |
| `Accepted (closed YYYY-MM-DD)` | Ratified and resolved; the underlying issue is closed. |
| `Historical (<context>)` | No longer actionable; retained as a record. Often preceded by a superseding ADR. |
| `Superseded by ADR XXXX` | This ADR's decision was replaced by the numbered ADR. |
| `Deprecated (<reason>)` | Decision withdrawn without a direct replacement; include the reason. |
| `Proposed` | Draft under review; not yet ratified. |

## Rule: new ADRs MUST carry a Status header from this vocabulary

Every new ADR file added to `wiki/decisions/` MUST include a `> **Status:**` line drawn from the vocabulary above. A file without a status header fails the Wave Z verification gate:

```powershell
foreach ($f in Get-ChildItem wiki/decisions/00*.md) {
  if (-not (Select-String -Path $f -Pattern '^> \*\*Status:\*\*' -Quiet)) {
    "NO STATUS: $f"
  }
}
```

This must produce no output on a clean tree.

## Numbering

ADRs are numbered `XXXX` (four zero-padded digits) in the filename, e.g. `0029-my-decision.md`. Gaps in the sequence (0004, 0005, 0006) are permanent — do not reuse them. The next available number is always `MAX(existing) + 1`. Dated filenames (e.g. `2026-06-03-*.md`) are not valid; rename to the next sequence number before merging.
