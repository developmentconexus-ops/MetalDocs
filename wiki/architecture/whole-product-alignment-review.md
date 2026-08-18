# Whole-Product Alignment Review — Adjudicated GCR Authority

> **Status:** ACTIVE ROUTING — PRODUCT CONTRACT + GCR + 4+1 OWNERSHIP TOPOLOGY OPERATOR-APPROVED / TECHNICAL RE-DERIVATION NEXT  
> **Date:** 2026-08-18  
> **Implementation:** BLOCKED  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`

This page records the operator-adjudicated Whole-Product Global Coherence Review and routes the program after approval of the replacement Launch ownership topology. Product semantics remain owned by the Product Contract; semantic ownership is now owned by the Launch V1 Ownership Topology.

## Trigger and adjudication

During simplified R10-C, the standalone `Artifact` semantic owner exposed a broader structural problem: technical architecture had matured faster than Launch product scope and several older decisions were being carried forward after major scope reductions without re-running Structural Inversion.

The resulting Whole-Product GCR was completed from the accepted Product Contract outward. On 2026-08-18 the operator accepted recommendations A1–A10 as written.

Method outcome:

```text
RESTRUCTURE NOW at whole-product target-design level
```

The operator subsequently approved the replacement ownership topology:

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

The prior R10-A 8+3 Launch topology is therefore superseded for the active target.

## Accepted whole-product dispositions

### A1 — ACCEPTED

Remove standalone `Artifact` semantic ownership from the Launch target. Exact-content facts belong to the semantic record that freezes them; storage, staging, integrity and byte-location concerns remain mechanism.

### A2 — ACCEPTED

Keep:

```text
Document != Revision != Working Content != Submission
system-owned Release = effectivity authority
```

### A3 — ACCEPTED

Keep one sequential governance Step semantic. Do not inherit prior advanced policy dimensions merely because R10-B4 already modeled them; participant/quorum/SoD/fresh-auth/overseer/reassignment details require named Launch journeys or invariants.

### A4 — ACCEPTED

Governed obsolescence without replacement is a second explicit Launch governance journey. It may reuse the smallest common governance semantics that survive re-derivation, but Launch must not become a generic arbitrary-subject BPM/workflow engine.

### A5 — ACCEPTED

```text
Distribution / Read & Acknowledge → Launch+
Periodic Review                    → Launch+
```

They are removed from Launch-Core Release/transaction/topology obligations.

### A6 — ACCEPTED

```text
Dossier
Evidence
Records Governance / Retention / Legal Hold / Disposition
```

are Future for Launch and create no dormant modules, tables, permissions, transaction branches or backward pressure on core persistence.

### A7 — ACCEPTED

Break the former generic `Interchange` grouping. Historical Migration remains a cutover capability required when migration exists. Governed Export and generic repository `IMPORT/PUBLISH` remain Future absent a named consumer.

### A8 — ACCEPTED

Regenerate Launch Authorization roles/permissions from accepted Launch journeys. Do not preserve the prior exact 5×43 catalog by subtraction or sunk cost. The resulting model must include a least-privilege Auditor/Governance Viewer path.

### A9 — ACCEPTED

Keep same-local-commit append-only Audit for required governed actions, bounded/PII-minimized facts and the law that Audit never becomes domain state. Defer the deployment-wide cryptographic `AuditChainHead`/global lock law unless a concrete Launch tamper-evidence/non-repudiation requirement later justifies it.

### A10 — ACCEPTED

Defer residual adjuncts without a named Launch consumer, including prior editable dictionary/system-value machinery, classification taxonomy, structured TemplateSpec platform, DRAFT comment platform, scheduled release, auxiliary semantic Rendition for `SourceOnly`, and advanced Approval policy dimensions. Reopen only when a concrete Launch journey or invariant requires the specific capability.

## Product kernel that survives

```text
single-company identity/access
Document Types + numbering
Controlled Document stable identity
Business Revision
mutable DRAFT Working Content + autosave/recovery/OCC
Templates as governed Documents
immutable exact Submission attempt
NoHumanApproval OR one sequential governance route
feedback / ACCEPT / RETURN_FOR_CHANGES
withdraw Submission attempt
cancel open Revision
system-owned Release
EFFECTIVE / SUPERSEDED
optional required official Rendition
explicit governed OBSOLETE without replacement
revision/history
current-effective search/read/download
Audit separated from domain truth
truthful historical migration/cutover
exact-content integrity + backup/restore correctness
```

## Future-evolution adjudication

The operator explicitly clarified that deferred capabilities are expected future product evolution and must not be forgotten or made unnecessarily difficult to introduce.

Accepted law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

The named future horizon in `launch-v1-ownership-topology.md` is therefore evidence for later technical choices. Remaining architecture must keep stable semantic anchors and avoid foreseeable dead ends for Distribution/Acknowledgement, Periodic Review, Dossier, Evidence, Records Governance, governed export, repository interchange, Training/LMS, multi-document Change Control, pooled tenancy and realtime coauthoring.

This does **not** authorize dormant modules, tables, permissions, jobs, feature flags or generic frameworks. A future capability becomes a new semantic owner only when a concrete consumer/lifecycle justifies it.

## Prior R10 consequences

Prior R9.5/R10 material remains evidence and provenance. It no longer controls Launch where the Product Contract, GCR adjudication or approved 4+1 topology conflicts.

```text
R10-A prior 8+3 topology          → SUPERSEDED FOR LAUNCH by launch-v1-ownership-topology.md
R10-B1 substrate                  → evidence; re-derive only surviving technical laws
R10-B2 AuthN/Org/AuthZ boundary   → ownership survives; exact role/permission catalog reopened
R10-B3                            → core semantics survive; Artifact/adjunct technical shape reopened
R10-B4                            → governance/Release semantics survive minimally; Distribution/workflow richness reopened
R10-B5                            → Dossier/Evidence/Records Future for Launch
R10-B6                            → Audit separation + migration truth survive; generic Interchange/hash-chain scope reopened
R10-C                             → PAUSED / evidence only / DO NOT REPAIR
```

## Exact next step

**Re-derive remaining technical architecture from the accepted Product Contract + GCR + 4+1 ownership topology.**

Order:

```text
semantic owner boundaries
→ minimal persistent facts / invariants
→ transactions / concurrency / exact-content mechanics
→ Authorization catalog/check sites
→ storage/integrity/restore mechanism
→ async/projection/effects
→ API/frontend journeys
→ historical migration/cutover
→ integrated Whole-R10 GCR
```

Every material decision must run both:

```text
Launch correctness test
+ named-future compatibility test
```

Do not restore prior R10 technical choices by inertia. Reuse them only when they survive the new authority and future-evolution law.

## Gate

```text
Product Contract                         ✅
Whole-Product GCR A1–A10                ✅
Launch ownership topology 4+1           ✅
re-derived remaining technical design   NEXT
Whole-R10 Global Coherence Review
cold independent review
final operator ratification
implementation spec/plan
code
```

Implementation remains **BLOCKED**. No product code or implementation plan is authorized yet.