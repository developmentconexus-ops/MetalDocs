# Launch V1 Records-Governance Defer Overlay

> **Status:** ACTIVE NARROW OVERLAY — SUBORDINATE TO ACCEPTED PRODUCT CONTRACT  
> **Date:** 2026-08-18  
> **Applies to:** Launch V1 governed-history retention/disposition scope only  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Detailed rationale:** `docs/superpowers/analysis/2026-08-18-launch-v1-records-governance-defer-rebaseline.md`

This page preserves the earlier operator-approved **Records-Governance defer** decision only. It no longer owns the overall Launch topology or capability list.

Where the former version of this page described Launch ownership, Dossier/Evidence, Distribution, Artifact semantics, B5/B6 scope or the next technical stage, the accepted Product Contract supersedes it.

## Launch invariant retained

> **Launch V1 preserves confirmed governed history and exposes no governed physical deletion/disposition. SUPERSEDED, OBSOLETE and CANCELLED are lifecycle facts, never deletion commands. Only temporary/mechanism state is eligible for ordinary GC.**

## Records Governance deferred

The following are **not Launch V1 implementation scope**:

```text
Records Governance bounded context/module
DocumentTypeRetentionRule
EvidenceTypeRetentionRule
RetentionBinding
RetentionExtension
LegalHold
LegalHoldSubject
DispositionFence
DispositionRecord
retention clocks / expiry eligibility
governed physical deletion workflow
Records-driven ObjectLock/WORM
eDiscovery/custodian machinery
```

Do not create dormant tables, permissions, modules, flags or jobs for them.

Method outcome: `DEFER SAFELY`.

Reopen only on concrete regulatory, contractual, customer or operational evidence requiring finite retention, legal preservation hold or governed destruction.

## Product-contract consequences

The accepted Launch V1 Product Contract further establishes:

```text
Dossier / documentary case context → Future
Evidence capture / quality records → Future
Distribution / Read & Acknowledge → Launch+
Periodic Review                    → Launch+
standalone Artifact semantic owner → remove/restructure from Launch target
Governed Subject Export            → Future absent named consumer
Generic repository IMPORT/PUBLISH  → Future absent named consumer
```

Those are Product Contract decisions, not decisions owned by this page.

## What remains valid from the defer decision

- confirmed governed history is preserved in Launch;
- business lifecycle status never implies physical deletion;
- no Launch RetentionBinding/Hold/Disposition state or transaction graph;
- no Records-driven ObjectLock/WORM requirement;
- migration must not materialize deferred Records-Governance state;
- backup/restore must not resurrect inconsistent governed truth.

## Routing

Current program status and next action are owned by `wiki/references/current-agent-handoff.md` and `wiki/architecture/whole-product-alignment-review.md`.

Do not use historical sections of this overlay to restart R10-C or infer Launch ownership/topology.
