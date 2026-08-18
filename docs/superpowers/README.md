# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs Whole-Product Alignment / rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 Semantic State & Invariants candidate open / operator adjudication next.**

The previous accumulation of roadmaps, milestones, plans, reports, specs and analysis packets under `docs/superpowers/` was intentionally removed from the live tree during the architecture reset.

Those files remain recoverable from Git history. They are **not current planning authority** and must not be restored by copying old commits back into the live tree.

Current rule:

- durable accepted truth belongs in `wiki/`;
- active, not-yet-promoted redesign analysis belongs here;
- completed/superseded planning artifacts are deleted from the live tree rather than left to compete with current truth.

## Current active staging

- `analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md` — **ACTIVE NON-AUTHORITATIVE T1 adjudication packet.** It proposes the minimum enduring semantic fact set after the accepted Product Contract, GCR and 4+1 ownership topology.

Current durable authority remains:

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-technical-architecture.md
```

## Prior-design evidence still retained in the live tree

The following may be consulted as evidence but do not control the active technical descent:

- `analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 prior product/domain evidence;
- former R10 B1/B2 reviews/corrected targets — evidence for surviving substrate/AuthN/Organization/AuthZ laws;
- former R10 B3/B4 candidates — evidence for Document/Revision/WorkingContent/Submission, one Step, Release and related counterexamples;
- former R10 B5/B6 candidates — historical evidence for future Dossier/Evidence/Records and for Audit/migration concerns;
- `analysis/2026-08-18-r10-c-artifact-physical-integrity-integrated-candidate.md` — **PAUSED HISTORICAL CANDIDATE / safety evidence only; DO NOT REPAIR OR PROMOTE**.

The former Whole-Product GCR staging packet was promoted into durable adjudicated authority and removed from the live staging tree; Git history is its archive.

## Approved active technical path

```text
T1 Semantic State & Invariants                         ACTIVE
T2 Governance, Effectivity & Lifecycle Transactions   NOT OPEN
T3 Authorization & Audit Enforcement                  NOT OPEN
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

Every material stage decision must satisfy both Launch correctness and the accepted named-future compatibility law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

## Hard stop

No product implementation, SQL/schema/package design or implementation plan is authorized by anything in this directory while the active T-stage gates remain open.