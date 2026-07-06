# F-R2 — ADR status-field CI gate (Dim 10 → CONFIRMED)

> Consumer-contract-first. Approved **before** the feature's commit.
> **Approval:** APPROVED 2026-07-06 (operator "Go" on M10; contract self-reviewed against
> `validation-contract.md` C4). — *filled before the F-R2 commit.*

## Problem (the DEBT this closes)

M9 F9.1 established the ADR status-field rule (status block ≤3 lines / ≤400 chars) and did a one-time
sweep, but `documentation-governance.md:33-34` explicitly left CI wiring "optional." The rule was a
**convention** with no structural gate — the exact meta-defect class ("hand-synced convention without
a generator/gate") this mission exists to eliminate. Nothing stops a future ADR from re-growing a
mega-status block.

## Consumers & the contract each requires

| Consumer | Required shape (the contract) |
|----------|------------------------------|
| **CI (governance-check.yml)** | On every `pull_request`, a **blocking** step fails the build if any ADR under `wiki/decisions/` has a status block > 3 physical lines or > 400 chars. No `continue-on-error`. |
| **ADR author** | A one-command local sweep (`bash scripts/check-adr-status.sh`) that mirrors CI exactly, so they can self-check before pushing. Same script → local and CI cannot diverge. |
| **Governance doc** | States the rule is now CI-enforced (the "optional future extension" language removed), pointing at the single-source script. |

## Contract details

- **Single source:** the sweep logic lives in `scripts/check-adr-status.sh` (the awk budget check
  lifted verbatim from the doc's one-liner). CI invokes the script; the doc's manual sweep invokes
  the same script. There is no second copy of the logic.
- **Interface:** `check-adr-status.sh [dir]` — scans `dir` (default `wiki/decisions`); exit 0 +
  `adr-status: clean` on pass; exit 1 + offender list (`name: N lines, M chars`) on fail; exit 2 on
  usage/dir error. The optional `dir` arg exists so the gate is testable against a synthetic fixture
  without polluting `wiki/decisions/`.
- **Placement:** a step in the existing `check` job of `governance-check.yml` (ubuntu-latest, bash
  available), after the PowerShell governance check.

## Non-goals (mandatory)

- No change to the ADR status-field **rule** itself (budget, vocabulary) — F9.1 owns that; F-R2 only
  enforces it.
- No reformatting of any existing ADR (the tree is already clean; positive proof confirms).
- No new workflow file — extend the existing governance-check job.
- No enforcement of other ADR fields (Context/Decision/Consequences) — status field only.

## Validation Gate

1. **Negative proof (captured):** run the script against a temp dir containing a synthetic ADR whose
   status block exceeds the budget → exit 1, offender named.
2. **Positive proof (captured):** run against the real `wiki/decisions/` → exit 0, `adr-status: clean`.
3. `governance-check.yml` `check` job has the step, blocking (grep: no `continue-on-error` on it).
4. `documentation-governance.md` no longer says the CI wiring is "optional future extension" (grep the
   old sentence is gone) and states the gate is blocking on every PR.

## ADR?

No. This is enforcement plumbing for an existing decision (F9.1's rule), not a new decision.
