# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs Whole-Product Alignment / rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 CLOSED / PROMOTED; T3 Authorization & Audit Enforcement discovery/design active.**

The live `docs/superpowers/` tree contains active working material only. Durable accepted truth belongs in `wiki/`; completed/superseded staging is removed and remains recoverable from Git history.

## Current active staging

No T3 material decision packet has been written yet. T3 is in discovery/design and must first compare credible access/Audit approaches against the accepted Product Contract + T1/T2 journeys.

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

## Mandatory T-stage closure protocol

For every `Tn`:

```text
candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promotion/closure
→ only then Tn+1
```

A technical recommendation approval alone never opens the next stage.

## Prior-design evidence retained in the live tree

Prior B1–B6/C artifacts may be consulted as evidence but do not control active technical descent. In particular:

- the 2026-08-14 cohesive redesign ledger is historical/product evidence;
- old B1/B2 may inform surviving substrate/AuthN/Organization/AuthZ laws;
- old B3/B4 may inform Document/Revision/WorkingContent/Submission, one-Step governance and Release counterexamples;
- old B5/B6 are evidence for future Dossier/Evidence/Records and Audit/migration concerns;
- `analysis/2026-08-18-r10-c-artifact-physical-integrity-integrated-candidate.md` is **PAUSED HISTORICAL CANDIDATE / safety evidence only; DO NOT REPAIR OR PROMOTE**.

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  ACTIVE / DISCOVERY-DESIGN
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

No product implementation or implementation plan is authorized while the active T-stage gates remain open.