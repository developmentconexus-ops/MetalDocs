# Module Wiki Maturity Audit - 2026-05-13

Scope: `wiki/modules/*` only. Depth: deep mechanical audit against the L0-L4 maturity model in `.claude/skills/metaldocs-module-doc/references/module-wiki-maturity.md`.

## Maturity Model

| Level | Meaning |
|---|---|
| L0 stub | A page exists but cannot guide implementation. |
| L1 partial | Useful facts exist, but no full trio/artifact set. |
| L2 living doc | Trio plus artifacts exist; main facts are usable. |
| L3 mature | Meets the maturity standard and passes gates. |
| L4 current | L3 plus synced after the latest implementation. |

## Executive Summary

The module wiki is on the right path. The strongest pages already work as LLM/developer memory: they contain runtime flows, route truth tables, persistence notes, authz, dependencies, debt, backlog, and source artifacts.

The maturity gap is consistency, not concept. Ten modules have the full living-doc shape. Five module pages are partial and need promotion before they can participate in cheap post-implementation sync.

Hard blockers before calling the module wiki mature:
- Five full living docs currently fail `tally_check.sh`.
- No module page records the new explicit `Maturity: Lx` field yet.
- `wiki/modules/README.md` is stale relative to current module docs and Plan 8 stamps.
- Some full modules lack sync logs or key-file blocks.
- Five partial module pages lack tech-debt registers, refactor backlogs, and artifacts.

## Current Classification

| Module | Current level | Evidence | Main blocker to next level |
|---|---:|---|---|
| `audit` | L3 candidate | Full trio, artifacts, sync log, route truth table, key files, tally PASS. | Add explicit maturity stamp; sync after latest implementation to reach L4. |
| `documents` | L3 candidate | Full trio, artifacts, sync log, route truth table, key files, tally PASS. | Tally warns T-007/T-010 have no backlog rows; add maturity stamp. |
| `taxonomy` | L3 candidate | Full trio, artifacts, sync log, route truth table, key files, tally PASS. | Add maturity stamp; sync after latest implementation to reach L4. |
| `templates_v2` | L3 candidate | Full trio, artifacts, sync log, route truth table, tally PASS. | Missing key-file block; many spec-missing routes remain documented debt. |
| `approval` | L2 | Full trio and artifacts, route truth table present. | Tally FAIL: doc says Critical=1, register has Critical=2; no sync log; no key-file block. |
| `auth` | L2 | Full trio, artifacts, sync log, route truth table. | Tally FAIL: doc missing-ADR count says 9, register grep count is 8. |
| `editor-chrome` | L2 | Full trio and artifacts; frontend module with route table n/a. | Tally FAIL: doc says Major=3, register has Major=4; no sync log. |
| `editor-ui-eigenpal` | L2 | Full trio and artifacts; frontend module with route table n/a. | Tally FAIL: doc says Critical=0/Major=0, register has Critical=1/Major=2; no key-file block. |
| `iam` | L2 | Full trio, artifacts, sync log, route truth table. | Tally FAIL: doc missing-ADR count says 12, register grep count is 11. |
| `registry` | L2 | Full trio, artifacts, sync log, route truth table. | Tally FAIL: doc missing-ADR count says 10, register grep count is 9. |
| `frontend-primitives` | L1 | Useful page exists with key files. | Missing tech-debt register, refactor backlog, artifacts. |
| `novo-documento-wizard` | L1 | Useful page exists with key files. | Missing tech-debt register, refactor backlog, artifacts. |
| `templates-v2` | L1 | Useful predecessor/frontend page exists. | Missing trio/artifacts; likely predecessor of `templates_v2` and should be retired or explicitly scoped. |
| `render-fanout` | L0 | 31-line TBD/stub page. | Needs full promotion or retirement. |
| `search` | L0 | 30-line stub page. | Needs full promotion or retirement. |

## Tally Gate Results

| Module | Result | Detail |
|---|---|---|
| `approval` | FAIL | Severity mismatch: actual 2/4/6, stated 1/4/6. |
| `audit` | PASS | Actual 2/4/6 matches stated. |
| `auth` | FAIL | Missing-ADR mismatch: register count 8, doc stated 9. |
| `documents` | PASS with warnings | T-007 and T-010 have no backlog rows. |
| `editor-chrome` | FAIL | Severity mismatch: actual 0/4/5, stated 0/3/5. |
| `editor-ui-eigenpal` | FAIL | Severity mismatch: actual 1/2/5, stated 0/0/5. |
| `iam` | FAIL | Missing-ADR mismatch: register count 11, doc stated 12. |
| `registry` | FAIL | Missing-ADR mismatch: register count 9, doc stated 10. |
| `taxonomy` | PASS | Actual 5/5/6 matches stated. |
| `templates_v2` | PASS | Actual 4/6/4 matches stated. |

## API Contract Signals

Spec-missing route rows are well documented, which is good. They are product/architecture debt, not wiki drift by themselves.

| Module | Spec-missing rows |
|---|---:|
| `approval` | 16 |
| `taxonomy` | 16 |
| `templates_v2` | 13 |
| `documents` | 11 |
| `iam` | 3 |
| `audit`, `auth`, `registry` | 0 |

Frontend-only modules correctly do not require API route truth tables unless they own frontend API contract behavior.

## Immediate Fix Order

1. Fix mechanical tally failures in existing L2 modules.
2. Add missing sync logs/key-file blocks and explicit `Maturity:` stamps to the 10 full living docs.
3. Refresh `wiki/modules/README.md` so the index matches current Last verified dates, maturity levels, and predecessor/partial status.
4. Promote or retire partial pages:
   - `search`
   - `render-fanout`
   - `frontend-primitives`
   - `novo-documento-wizard`
   - `templates-v2`
5. After each future implementation, use `metaldocs-module-doc-sync` to move touched L3 modules to L4 current.

## Recommendation

Do not start by rewriting the good docs. Start with the small mechanical blockers:
- `approval`, `auth`, `editor-chrome`, `editor-ui-eigenpal`, `iam`, and `registry` count fixes.
- Add maturity stamps.
- Add missing sync-log headers.
- Update the module index.

Then promote the partial pages one at a time with `metaldocs-module-doc`. This gives the best return: the wiki becomes trustworthy quickly, and the expensive deep sweeps are reserved for pages that truly lack the mature structure.
