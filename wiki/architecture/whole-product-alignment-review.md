# Whole-Product Alignment Review — Adjudicated GCR Authority

> **Status:** DURABLE ADJUDICATED GCR — **PRODUCT CONTRACT + A1–A10 + 4+1 OWNERSHIP APPROVED / T1→T7 TECHNICAL REBASELINE APPROVED / T1 ACTIVE**  
> **Date:** 2026-08-18  
> **Implementation:** BLOCKED  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **Active technical authority:** `wiki/architecture/r10-technical-architecture.md`

This page records the operator-adjudicated Whole-Product Global Coherence Review. It is durable decision authority for the A1–A10 dispositions below; **current technical-stage routing is now owned by `wiki/architecture/r10-technical-architecture.md`**.

## Trigger and adjudication

During simplified R10-C, the standalone `Artifact` semantic owner exposed a broader structural problem: technical architecture had matured faster than Launch product scope and older decisions were being carried forward after major scope reductions without re-running Structural Inversion.

The Whole-Product GCR was therefore rerun from the accepted Product Contract outward. On 2026-08-18 the operator accepted recommendations A1–A10 as written.

Method outcome:

```text
RESTRUCTURE NOW at whole-product target-design level
```

The operator then approved the replacement Launch ownership topology:

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

The former R10-A 8+3 Launch topology is superseded.

---

## Accepted whole-product dispositions

### A1 — ACCEPTED

Remove standalone `Artifact` semantic ownership. Exact-content facts belong to the semantic record that freezes/owns them; storage, staging, integrity and physical location remain mechanism.

### A2 — ACCEPTED

Keep:

```text
Document != Revision != Working Content != Submission
system-owned Release = effectivity authority
```

### A3 — ACCEPTED

Keep one sequential governance Step semantic. Do not inherit prior quorum/SoD/fresh-auth/overseer/reassignment richness unless a named Launch journey or invariant proves it.

### A4 — ACCEPTED

Governed obsolescence without replacement is a second explicit Launch governance journey. Reuse only the smallest common governance semantics; do not create a generic arbitrary-subject BPM/workflow engine.

### A5 — ACCEPTED

```text
Distribution / Read & Acknowledge → Launch+
Periodic Review                    → Launch+
```

Neither is part of Launch-Core Release atomicity/topology.

### A6 — ACCEPTED

```text
Dossier
Evidence
Records Governance / Retention / Legal Hold / Disposition
```

are Future for Launch and create no dormant modules, tables, permissions, jobs, transaction branches or backward pressure on core persistence.

### A7 — ACCEPTED

Break generic `Interchange` ownership. Historical Migration remains a cutover capability. Governed Export and generic repository `IMPORT/PUBLISH` remain Future absent a named consumer.

### A8 — ACCEPTED

Regenerate Launch Authorization roles/permissions from accepted Launch journeys. Do not preserve the prior exact 5×43 catalog by subtraction or sunk cost. Include a least-privilege Auditor/Governance Viewer path.

### A9 — ACCEPTED

Keep same-local-commit append-only Audit where required, bounded/PII-minimized facts and `Audit != domain state`. Defer deployment-wide cryptographic `AuditChainHead`/global-lock law unless a concrete Launch assurance requirement later justifies it.

### A10 — ACCEPTED

Defer residual adjuncts without a named Launch consumer, including editable dictionary/system-value machinery, classification taxonomy, structured TemplateSpec platform, DRAFT comment platform, scheduled release, auxiliary semantic Rendition for `SourceOnly`, and advanced Approval policy dimensions.

---

## Product kernel that survived the GCR

```text
single-company identity/access
Document Types + numbering
Controlled Document stable identity
Business Revision
mutable DRAFT Working Content + autosave/recovery/concurrency
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

---

## Future-evolution adjudication

The operator explicitly clarified that deferred capabilities are expected future product evolution and must not be forgotten or made unnecessarily difficult to introduce.

Binding law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

The named horizon in `launch-v1-ownership-topology.md` is evidence for technical choices:

```text
Launch+:
  Distribution / Read & Acknowledge
  Periodic Review

Future:
  Dossier
  Evidence
  Retention / Legal Hold / Disposition
  Governed Export
  generic External Repository IMPORT/PUBLISH
  Training/LMS
  generic/multi-document Change Control
  pooled multi-customer tenancy
  realtime coauthoring / CRDT
```

This authorizes **attachment seams**, not dormant implementation or generic frameworks.

---

## Prior R10 consequences

```text
R10-A old 8+3                    → SUPERSEDED FOR LAUNCH
R10-B1                           → evidence only; surviving laws must be re-proven in T-stages
R10-B2                           → AuthN/Org/AuthZ ownership survives; exact catalog/enforcement reopened
R10-B3                           → core Document/Revision/WorkingContent/Submission evidence survives; Artifact/adjuncts reopened
R10-B4                           → one Step + Release evidence survives; separate Approval ownership/Distribution/workflow richness reopened
R10-B5                           → Dossier/Evidence/Records Future
R10-B6                           → Audit separation + migration truth evidence survives; Interchange/hash-chain scope reopened
R10-C                            → PAUSED HISTORICAL CANDIDATE / DO NOT REPAIR OR PROMOTE
old R10-D/E/F                    → stage labels superseded by T5/T6/T7
```

Old artifacts remain evidence/provenance only. A surviving invariant must be carried into the active T-stage rather than treated as authority by file age.

---

## Approved technical descent

On 2026-08-18 the operator approved the replacement technical sequence:

```text
T1 — Semantic State & Invariants
T2 — Governance, Effectivity & Lifecycle Transactions
T3 — Authorization & Audit Enforcement
T4 — Exact Content, Storage Integrity & Restore
T5 — Durable Async, Search & External Effects
T6 — Canonical API / Frontend Journeys
T7 — Historical Migration & Cutover

→ Integrated Whole-R10 Global Coherence Review
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

This sequence is now owned by `wiki/architecture/r10-technical-architecture.md`.

## Current gate

```text
Product Contract        ✅
Whole-Product GCR       ✅ A1–A10
Ownership topology      ✅ 4+1
T1→T7 decomposition     ✅
T1                      ACTIVE NON-AUTHORITATIVE CANDIDATE
T2→T7                   NOT OPEN
implementation          BLOCKED
```

Active T1 staging packet:

`docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md`

**NEXT = operator adjudication of T1.**