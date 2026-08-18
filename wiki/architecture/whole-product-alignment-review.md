# Whole-Product Alignment Review — Adjudicated GCR Authority

> **Status:** ACTIVE ROUTING — PRODUCT CONTRACT ACCEPTED / WHOLE-PRODUCT GCR OPERATOR-ADJUDICATED / OWNERSHIP RE-DERIVATION NEXT  
> **Date:** 2026-08-18  
> **Implementation:** BLOCKED  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`

This page records the operator-adjudicated Whole-Product Global Coherence Review and controls routing into ownership/topology re-derivation. It does not replace the Product Contract's product semantics and does not itself choose the replacement bounded-context/package topology.

## Trigger and adjudication

During simplified R10-C, the standalone `Artifact` semantic owner exposed a broader structural problem: technical architecture had matured faster than Launch product scope and several older decisions were being carried forward after major scope reductions without re-running Structural Inversion.

The resulting Whole-Product GCR was completed from the accepted Product Contract outward. On 2026-08-18 the operator accepted the GCR recommendations A1–A10 as written.

Method outcome:

```text
RESTRUCTURE NOW at whole-product target-design level
```

This preserves the controlled-document kernel while rejecting automatic carry-forward of the prior R10 ownership topology and downstream technical shape.

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

Regenerate Launch Authorization roles/permissions from accepted Launch journeys after this adjudication. Do not preserve the prior exact 5×43 catalog by subtraction or sunk cost. The resulting model must include a least-privilege Auditor/Governance Viewer path.

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

## Prior R10 consequences

Prior R9.5/R10 material remains evidence and provenance. It is not deleted merely because the GCR restructured the target, but it no longer controls Launch where these adjudicated findings conflict.

```text
R10-A prior 8+3 topology          → REOPENED / replacement must be re-derived
R10-B1 substrate                 → conceptually reusable only where it survives re-derivation
R10-B2 AuthN/Org/AuthZ boundary  → largely reusable; exact role/permission catalog reopened
R10-B3                           → core Document/Revision/WorkingContent/Submission survives; Artifact/adjuncts reopened
R10-B4                           → one Step + Release survive; Distribution/workflow richness reopened
R10-B5                           → Dossier/Evidence/Records Future for Launch
R10-B6                           → Audit separation + migration truth survive; generic Interchange/hash-chain scope reopened
R10-C                            → PAUSED / evidence only / DO NOT REPAIR
```

## Exact next step

Re-derive ownership/topology from zero, in this order:

```text
accepted Launch capabilities
→ end-to-end journeys
→ semantic facts/lifecycles that need an owner
→ authority boundaries
→ smallest coherent ownership topology
```

The re-derivation must compare credible alternatives and run the Method against duplicate authority, unnecessary fragmentation, God contexts and support mechanisms masquerading as semantic owners.

## Gate

```text
operator-adjudicated Whole-Product GCR
→ ownership/topology re-derivation
→ operator approval of topology
→ re-derive remaining technical architecture
→ Whole-R10 Global Coherence Review
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

Until ownership/topology is approved, do not author SQL, table/schema design, storage-provider topology, package layout, an R10-C replacement, implementation plans or product code.
