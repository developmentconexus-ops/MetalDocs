# Discovery Brief: <Mission Name>

> **Mission slug:** `<mission-slug>`  ·  **Type:** remediation | greenfield-build | enhancement | migration
> **Date:** <date>  ·  **Branch:** <branch>
> **Agents / models used:** <e.g. 4× sonnet analysts + 1× sonnet skeptic>
> This is the **evidence base** the mission stands on. Every claim in `mission.md` traces to a finding here.

## Method
What each discovery agent was asked to do, and how findings were verified. State clearly **what was
verified** (re-checked, reproduced, grepped) vs **what was assumed**. Record the skeptic-pass outcome —
which findings survived adversarial review and which were downgraded or dropped.

| Agent / lens | Scope swept | Verified how |
|--------------|-------------|--------------|
| <lens> | <files / modules / docs / web sources> | <grep / reproduced / cross-checked / skeptic> |

## Findings
The cited inventory. For remediation/migration, give `file:line`. For greenfield/enhancement, give the
requirement/capability and its source. Each finding gets a proposed milestone home (or out-of-scope).

| # | Finding (with citation) | Severity / kind | Confidence | Proposed home |
|---|-------------------------|-----------------|------------|---------------|
| 1 | <`path:line` or requirement + source> | <critical/major/minor or feature> | verified / assumed | M<n> / out-of-scope |
| … | | | | |

## Constraints & risks surfaced
House rules, architectural invariants, and blast-radius hotspots the mission must respect or route around
(e.g. shared-API redesign boundaries, advisory-lock hazards, contract-first regen order, consumers that
would break). Flag anything that looks like an HS-2 redesign boundary.

## Open questions for the operator
The decisions Phase 2 must lock — scope, sequencing, proof-of-done, any either/or the evidence can't
settle on its own.

## Coverage statement
What was **not** swept, and why (time-boxed, out-of-scope module, deferred). **No silent caps** — if
discovery bounded itself, say so here so the decomposition isn't read as exhaustive when it isn't.
