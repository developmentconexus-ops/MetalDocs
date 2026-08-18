# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs Whole-Product Alignment / rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 CLOSED / PROMOTED; Decision Reconciliation active; T3 paused.**

The live `docs/superpowers/` tree contains active working material only. Durable accepted truth belongs in `wiki/`; completed/superseded staging is removed and remains recoverable from Git history.

## Current active staging

- `analysis/2026-08-18-rebaseline-decision-reconciliation-candidate.md` — **ACTIVE NON-AUTHORITATIVE reconciliation of prior R3–R9.5 / old R10 decisions; operator review next.**

A T3 candidate written before this reconciliation gate is intentionally removed from the live staging tree. T3 will be rebuilt from the accepted registry rather than repaired from a premature zero-reset.

T1 and T2 were operator-adjudicated, summarized in platform terms, explicitly summary-ratified and promoted into durable `wiki/` authority. Their completed staging is removed from the live tree.

Binding revision convention:

```text
REV000 = initial issuance
REV001 = first revision
REV002 = second revision
...
```

Current durable authority:

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-technical-architecture.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
```

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

The active reconciliation candidate uses:

```text
CURRENT
PRESERVE
REFINED
REOPEN
DEFERRED
SUPERSEDED
```

After operator ratification the candidate should be promoted to a durable decision registry used at the beginning/end of every remaining T-stage.

## Mandatory T-stage closure protocol

For every `Tn`:

```text
candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promotion/closure
→ update Decision Registry
→ only then Tn+1
```

A technical recommendation approval alone never opens the next stage.

## Prior-design evidence retained in the live tree

Prior B1–B6/C artifacts may be consulted as evidence but do not control active technical descent. The reconciliation candidate is the only active place currently deciding their survivorship.

- the 2026-08-14 cohesive redesign ledger is historical inventory/evidence;
- old B1/B2 contain substrate/AuthN/Organization/AuthZ decisions to classify;
- old B3/B4 contain Document/WorkingContent/Submission/governance/Release decisions to classify;
- old B5/B6 preserve future Dossier/Evidence/Records and Audit/migration design evidence;
- `analysis/2026-08-18-r10-c-artifact-physical-integrity-integrated-candidate.md` remains **PAUSED HISTORICAL CANDIDATE / safety evidence only; DO NOT REPAIR OR PROMOTE**.

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Baseline                     ACTIVE / OPERATOR REVIEW NEXT
T3 Authorization & Audit Enforcement                  PAUSED ON RECONCILIATION
T4 Exact Content, Storage Integrity & Restore         NOT OPEN
T5 Durable Async, Search & External Effects           NOT OPEN
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ Integrated Whole-R10 GCR
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

Every material decision must satisfy Launch correctness and the named-future evolution law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

## Hard stop

No product implementation or implementation plan is authorized while the active design gates remain open.