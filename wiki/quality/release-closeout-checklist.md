# Release Close-Out Checklist

> **Last verified:** 2026-05-27
> **Scope:** Reusable final close-out checklist for merge/release readiness after implementation, review, and QA loops have completed.

## When to use

Use this checklist before declaring a non-trivial change ready to merge or release.

Pair it with:

- [release-readiness.md](release-readiness.md)
- [qa-operating-system.md](qa-operating-system.md)

## Checklist

- implementation scope stayed inside the approved boundary
- targeted tests/types/lint relevant to the slice passed
- code review completed and findings were recorded in severity order
- product QA completed for the touched workflow class
- findings were classified by root-cause family before fixing
- fixes addressed the owning boundary instead of scattered symptoms
- targeted regression reran after fixes
- broader regression reran when the change crossed boundaries
- contract/wiki/runtime drift was either resolved or explicitly classified
- required wiki truth was updated, or skipped with reason
- residual defers are explicit, bounded, and linked
- release-readiness gate status is recorded when the change is at merge/release boundary
- closure evidence exists and is reproducible by a later session

## Minimum closure record

Record:

- scope and owning boundary
- commands/checks run
- review result
- QA result
- regression result
- defers and their classification
- artifact or wiki link for the evidence trail

## Escalation rule

Do not close when:

- review or QA is missing
- regression scope was not rerun after a fix family changed
- evidence exists only as confidence or memory
