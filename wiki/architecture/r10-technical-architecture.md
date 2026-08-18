# R10 Technical Architecture — Rebaselined Active Stage Authority

> **Status:** ACTIVE — **TECHNICAL REBASELINE APPROVED / T1 SEMANTIC STATE & INVARIANTS OPEN / IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Method:** `docs/engineering/standards/root-cause-global-maximum-method.md`  
> **Product authority:** `wiki/architecture/launch-v1-product-contract.md`  
> **GCR authority:** `wiki/architecture/whole-product-alignment-review.md`  
> **Ownership authority:** `wiki/architecture/launch-v1-ownership-topology.md`  
> **Implementation gate:** **CLOSED — design/documentation only**

This page is the active technical-stage routing authority after the operator-approved Whole-Product rebaseline. It **supersedes the former R10-A/B1–B6/C–F stage order as the active technical descent**. Git history preserves the old page; the old B1–B6/C artifacts remain evidence only where current authorities do not supersede them.

The rebaseline exists because the old stage structure encoded the superseded 8+3 ownership model (`Artifact`, separate `Approval`, `Distribution`, `Documentary Context`, `Records Governance`, generic `Interchange`). Continuing that order and subtracting features would be a Local Maximum.

---

## 1. Binding technical objective

Derive the **smallest sustainable Launch V1 technical architecture** from:

```text
accepted Product Contract
+ operator-adjudicated Whole-Product GCR
+ operator-approved 4+1 ownership topology
+ known-future evolution law
```

Every material decision must pass:

```text
Launch correctness
+ essential vs accidental complexity
+ one semantic authority per fact
+ proof strategy before implementation
+ named-future compatibility
```

Binding evolution law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Known future direction is a counterexample set, not permission to prebuild modules/tables/frameworks.

---

## 2. Active semantic ownership baseline

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Not Launch semantic owners:

```text
storage / staging / byte integrity / malware inspection
render/view/editor providers
Search
async/outbox/jobs/retry/lease/DLQ
notifications
Historical Migration execution machinery
backup/restore transport/readiness
```

Historical `Artifact`, `Distribution`, `Documentary Context`, `Records Governance`, generic `Interchange`, and separate `Approval` ownership must not be resurrected by technical convenience.

---

# 3. Rebaselined R10 decomposition

The operator approved the following technical descent on 2026-08-18.

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
→ operator final ratification
→ implementation spec/plan
→ code
```

The sequence is by **failure class and proof dependency**, not by old module/context boundaries.

---

## 3.1 T1 — Semantic State & Invariants — ACTIVE

Question:

> What enduring semantic facts and mutation laws must exist so every Launch journey has one authority and future capabilities can attach without rewriting core history?

T1 owns design of:

```text
minimum semantic fact families per 4+1 owner
stable identities versus mutable current truth versus immutable evidence
Document/Revision/WorkingContent/Submission separation
bounded shared governance semantics for Submission + Obsolescence
Release/effectivity causal facts
native/imported provenance seam
Audit semantic minimum
explicit deletion of old uncontracted semantic families
```

T1 explicitly does **not** decide SQL/table/package shape, locks, exact participant rules, concrete permissions, storage handles, async topology, API/UI or migration execution.

Active staging candidate:

`docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md`

Current gate:

```text
T1 candidate → operator adjudication
```

---

## 3.2 T2 — Governance, Effectivity & Lifecycle Transactions — NOT OPEN

Begins only after T1 semantic facts are accepted.

Must prove all named lifecycle transitions, including:

```text
create / allocate code
DRAFT mutation + autosave concurrency
SUBMIT
ACCEPT
RETURN_FOR_CHANGES
RESUBMIT
WITHDRAW Submission attempt
CANCEL Revision
first Release
replacement Release / supersession
OBSOLETE without replacement
```

T2 owns:

```text
transaction boundaries
serialization roots
OCC/concurrency correctness
state-transition eligibility
smallest participant-selection semantics
quorum only if required
SoD only if required
fresh-auth only if required
attempt termination semantics
one-EFFECTIVE atomicity
obsolescence governance behavior
```

T2 must not create a generic BPM engine.

---

## 3.3 T3 — Authorization & Audit Enforcement — NOT OPEN

Begins after T2 proves the concrete Launch actions/relationships.

Derivation:

```text
personas
→ named operations
→ resources
→ scopes
→ owning-domain relationship predicates
→ permissions
→ role bundles
→ check sites
```

T3 regenerates the Launch catalog from zero; the old exact 5×43 catalog is historical evidence only.

Required persona coverage includes a least-privilege:

```text
Auditor / Governance Viewer
```

T3 also establishes:

```text
which governed/security operations require same-local-commit Audit
minimum Audit facts per operation
current-grant versus historical evidence boundary
offboarding access termination and truthful attribution
```

No role is a domain-governance bypass.

---

## 3.4 T4 — Exact Content, Storage Integrity & Restore — NOT OPEN

T4 answers the physical-content problem **without restoring Artifact semantic ownership**.

Must prove:

```text
semantic exact-content identity != provider/storage identity
WorkingContent mutable bytes remain recoverable
Submission/Rendition/imported exact content remains immutable
provider PUT success != semantic success
hash/size/format validation
safe untrusted-content admission / malware gate where required
no overwrite of governed bytes
temporary DRAFT/staging cleanup
provider outage behavior
backup/restore completeness
restore fail-closed on missing/corrupt required content
```

A shared storage/integrity mechanism may serve multiple semantic owners now or later without acquiring their meaning.

Future-seam counterexamples include Evidence, Records/WORM, repository connectors and CRDT DRAFT editing.

---

## 3.5 T5 — Durable Async, Search & External Effects — NOT OPEN

T5 owns mechanism semantics only where asynchronous/external execution is truly required:

```text
durable intents/outbox when necessary
claim/lease/retry/dead-letter behavior
renderer/provider execution
notifications when required
Search projection + rebuild
projection freshness/reconciliation
external effect receipts where a Launch consumer exists
```

Binding laws:

```text
worker state != business state
Search state != effectivity
Search state != Authorization
provider receipt != domain truth unless owning semantic operation explicitly consumes it
```

Named future seams: Distribution, Periodic Review scheduling and repository connectors must be addable without entering core Release authority.

---

## 3.6 T6 — Canonical API / Frontend Journeys — NOT OPEN

T6 proves the externally observable Launch journeys against the accepted semantic/transactional model:

```text
admin configuration
create blank/template-based
edit/autosave
submit
govern/return/resubmit
withdraw
cancel Revision
release/revise
obsolete
search
read/download
history/audit
offboarding consequences
```

Critical serving law:

```text
Search hit
→ canonical resource/state re-resolution
→ canonical Authorization
→ exact intended content resolution
→ serve
```

A stale projection must never serve SUPERSEDED/OBSOLETE content as current official truth.

---

## 3.7 T7 — Historical Migration & Cutover — NOT OPEN

Migration is designed last because it must write into the settled target rather than reshape the target around legacy convenience.

T7 must prove:

```text
native vs imported distinction
unknown remains unknown
no fabricated native Submission/decision/Release/User action
reliable historical ordinal preservation
exact imported content where available
truthful handling of incomplete old history
idempotency/reconciliation
atomic semantic import unit
cutover/readiness/rollback strategy
legacy mapping/deletion only after proof
```

T7 may introduce the smallest target-owned imported-history fact shapes required by actual source evidence. It must not recreate generic `Interchange`, Governed Export or repository-sync product scope.

---

# 4. Named-future compatibility proof

Every T1–T7 candidate contains an explicit future-horizon attack.

Required examples:

```text
Distribution         → can attach to Release + User/Group without becoming effectivity/AuthZ authority
Periodic Review      → can attach to Document + exact current EFFECTIVE Revision
Dossier              → can reference stable Document without owning content/access
Evidence             → can gain independent lifecycle using shared exact-content mechanism
Records              → can attach policy/hold/disposition to stable governed identities/history
Governed Export      → can consume stable identities/content without becoming source authority
Repository connector → can import/publish through target-owner seams without provider IDs becoming product IDs
Training/LMS         → can consume released/distributed content without becoming effectivity
Change Control       → can orchestrate/reference Document/Revision without taking their authority
pooled tenancy       → can reopen substrate around stable Company identity
CRDT                 → can replace WorkingContent concurrency mechanism without changing Revision/Submission semantics
```

Future compatibility does not require zero migration. It requires avoiding needless **authority demolition, immutable-history rewrite, or duplicate semantics**.

---

# 5. Old R10 evidence classification

```text
former R10-A 8+3                     → SUPERSEDED FOR LAUNCH
former R10-B1 relational substrate    → evidence; re-evaluate only where T1–T7 need it
former R10-B2 AuthN/Org/AuthZ         → semantic boundary largely survives; exact catalog/technical enforcement reopened
former R10-B3                         → core Document/Revision/WorkingContent/Submission evidence; Artifact/adjuncts reopened
former R10-B4                         → one Step + Release evidence; Approval-owner richness/Distribution reopened
former R10-B5                         → historical evidence for Future Dossier/Evidence/Records; not Launch target
former R10-B6                         → Audit/migration evidence; Interchange/hash-chain scope reopened
former R10-C                          → paused safety evidence only; Artifact-owned solution rejected
former R10-D/E/F                      → old stage labels superseded by T5/T6/T7
```

Do not repair old candidate files into the new target. Extract a surviving invariant into the current T-stage candidate or leave it in history/evidence.

---

# 6. Proof and review discipline

Each T-stage must include:

1. authority/evidence boundary;
2. Known / Inferred / Unknown / Deferred;
3. root cause and target invariant;
4. credible alternatives;
5. smallest sustainable decision;
6. named future-horizon attack;
7. proof strategy before implementation;
8. reopen triggers;
9. explicit non-decisions;
10. operator adjudication gate.

A stage is not implementation authority. Accepted stage findings are integrated into this durable authority only after operator adjudication.

After T7:

```text
integrated Whole-R10 GCR
→ cold independent review
→ operator final ratification
```

Only after that may implementation spec/plan be authored.

---

# 7. Current gate

```text
Product Contract           = ACCEPTED
Whole-Product GCR A1–A10   = ACCEPTED
Launch ownership 4+1       = ACCEPTED
T1–T7 decomposition        = OPERATOR-APPROVED
T1                         = ACTIVE CANDIDATE / OPERATOR ADJUDICATION NEXT
T2–T7                      = NOT OPEN
implementation             = BLOCKED
```

Current active packet:

`docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md`
