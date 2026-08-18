# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs Whole-Product Alignment / Cohesive Platform Redesign.
> **Reset:** 2026-08-14.
> **Current gate:** Whole-Product GCR findings complete / operator adjudication pending.

The previous accumulation of roadmaps, milestones, plans, reports, specs and analysis packets under `docs/superpowers/` was intentionally removed from the live tree during the architecture reset.

Those files remain recoverable from Git history. They are **not current planning authority** and must not be restored by copying old commits back into the live tree.

Current rule:

- durable accepted truth belongs in `wiki/`;
- active, not-yet-promoted redesign analysis belongs here;
- completed/superseded planning artifacts are deleted from the live tree rather than left to compete with current truth.

## Current active staging

- `analysis/2026-08-18-whole-product-global-coherence-review.md` — **current non-authoritative GCR adjudication packet**. Product authority remains `wiki/architecture/launch-v1-product-contract.md`; these findings become architecture input only after operator adjudication.

## Historical / prior-design evidence still retained in live tree

- `analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen R3–R9.5 decision ledger / historical product-domain evidence, subordinate to the accepted Product Contract wherever Launch scope or product semantics differ.
- accepted/non-final R10 B3–B6 packets and the paused R10-C packet remain evidence for the current GCR/re-derivation; they are not instructions to continue technical descent.

## Hard stop

No product implementation is authorized by anything in this directory.

Current path:

```text
Whole-Product GCR
→ operator adjudication
→ re-derive ownership/topology
→ re-derive remaining technical architecture
→ Whole-R10 review + cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```
