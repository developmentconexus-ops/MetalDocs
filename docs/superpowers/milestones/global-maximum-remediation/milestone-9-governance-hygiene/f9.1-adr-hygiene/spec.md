# F9.1 — adr-hygiene (feature spec)

> **Milestone:** M9 governance-hygiene · **Contract:** `../validation-contract.md` §1 (binding)
> **Approved:** 2026-07-06 — approved against mission.md M9 row + validation-contract §1 (operator-locked
> sources; autonomous session per mission D2). Code (doc edits) may start.

## Consumer contract (first)

**Consumer 1 — any maintainer/reviewer opening an ADR:** reads the `> **Status:**` block and learns,
in ≤3 lines / ≤400 chars, (a) the decision state (Proposed/Accepted/Superseded by NNNN/Deprecated/
Historical, optionally `(amended YYYY-MM-DD by NNNN)`), (b) the key date, (c) where the execution
history lives if it exists. Never a changelog.

**Consumer 2 — docs-governance checks:** `wiki/standards/documentation-governance.md` states the
status-field rule precisely enough that a sweep command can enforce it; the sweep is documented in the
rule (repeatable, no bespoke knowledge).

**Consumer 3 — supersession-chain navigation:** from ADR 0013 a reader reaches whatever decision
governs template revisions today; from the successor a reader reaches 0013 (back-link). `index.md`
Status column agrees with the file.

## Interview record (B1.5 — resolved from normative sources; autonomous session)

| Q | A | Source |
|---|---|--------|
| What is "≤3 lines" when a status is one 2757-char line? | Rule = ≤3 physical lines AND ≤400 chars total for the status block (from `> **Status:**` to next `> **Field:**` marker). | Contract §0.3/§1.1 sweep semantics; review finding 105 intent (legibility) |
| Which ADRs must be split? | All sweep violators on final tree: 0022 (2757c), 0027 (664c), 0070 (467c, 2 lines), 0015 (410c, 2 lines) — plus any others the final sweep finds. | Contract §1.2 "audit all 70"; sweep run 2026-07-06 |
| Where does relocated history go? | One companion doc per split ADR under `wiki/decisions/` (`NNNN-execution-history.md` pattern), content preserved verbatim-or-clearly-restructured, linked from the ≤3-line status. | Contract §1.2 |
| Is 0013 actually superseded? | Research says **not wholly**: `revision_number` persisted + REV{nn} labels are live (`templates/repository/postgres.go:92`); ADR 0052 (manual versioning) explicitly says the 0013 counter is "unchanged" but supersedes the *creation trigger* context 0013 assumed (versions spawn on approve/publish paths). Review 778f494a:105 claims 0052/0053 "moved past it". | Runtime + 0052:56 + review artifact |
| So what stamp does 0013 get? | The **researched true state**, per contract §0.3 escape hatch: expected outcome `Accepted (amended by 0052 — version-creation trigger; labels + persisted revision_number unchanged)` with an explicit note block; if implementer's deeper research proves full supersession, `Superseded by 0052`. Either way index.md row updated and 0052 gets the reciprocal reference. The review-finding disposition (why not a plain "Superseded") is recorded in evidence and surfaced at HS-1. | Contract §0.3, §1.3; HS-6 pre-authorized path |
| Does the rule become a CI gate now? | No — F9.1 delivers the documented rule + repeatable sweep command; wiring into CI is allowed but not required (governance-check.yml extension optional). Mission acceptance asks "rule documented". | mission.md M9 row |

## Non-goals (mandatory)

- No ADR body (Context/Decision/Consequences) content changes — relocation of the status-embedded
  history only. Decision text untouched.
- No re-litigating any decision; no "Current reality" annotations beyond what the status rule needs.
- No renumbering, no index restructuring beyond the touched Status/Superseded-by cells.
- No CI job requirement (optional extension only).
- No edits to ADRs that pass the sweep.

## Validation Gate

1. **Sweep GREEN:** documented sweep command over `wiki/decisions/[0-9]*.md` reports **0 violations**
   (≤3 lines AND ≤400 chars per status block). Command + output captured in evidence.
2. **Split integrity:** for each split ADR, companion doc exists, is linked from the status line, and
   a diff/content check shows the relocated history preserved (validator spot-checks ≥3 phase entries
   of 0022 against the pre-split text).
3. **0013 chain:** 0013 status = researched true state with successor reference; successor ADR
   back-references 0013; `index.md` rows agree; research trail in evidence.
4. **Rule documented:** `wiki/standards/documentation-governance.md` contains the status-field rule +
   vocabulary + the sweep command.
5. **No collateral edits:** `git diff --stat` for the feature commit touches only `wiki/decisions/*`,
   `wiki/standards/documentation-governance.md`, and the feature folder.
