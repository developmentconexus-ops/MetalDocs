# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT/GCR/4+1/T1→T7 APPROVED; T1 SEMANTIC STATE CANDIDATE OPEN / OPERATOR ADJUDICATION NEXT**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — **ACCEPTED PRODUCT AUTHORITY**
5. `wiki/architecture/whole-product-alignment-review.md` — **OPERATOR-ADJUDICATED GCR A1–A10**
6. `wiki/architecture/launch-v1-ownership-topology.md` — **OPERATOR-APPROVED 4+1 OWNERSHIP + FUTURE-EVOLUTION LAW**
7. `wiki/architecture/r10-technical-architecture.md` — **ACTIVE REBASELINED T1→T7 TECHNICAL-STAGE AUTHORITY**
8. `docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md` — **ACTIVE NON-AUTHORITATIVE T1 CANDIDATE / OPERATOR ADJUDICATION PACKET**
9. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records-Governance defer overlay
10. prior cohesive/R9.5/R10 B1–B6/C material only as evidence where current authorities do not supersede it

`wiki/architecture/cohesive-platform-redesign.md` is now a short **prior-design evidence compatibility page**, not active routing. Git history preserves its former full narrative.

Git history and current runtime/schema/OpenAPI remain evidence, not automatic target authority.

---

## Current checkpoint

```text
Product Contract                 = ACCEPTED / PROMOTED
Whole-Product GCR A1–A10         = OPERATOR-ADJUDICATED / ACCEPTED
Launch ownership topology        = CLOSED / OPERATOR-APPROVED / 4+1
T1→T7 technical decomposition    = CLOSED / OPERATOR-APPROVED
T1 Semantic State & Invariants   = ACTIVE NON-AUTHORITATIVE CANDIDATE
T2→T7                            = NOT OPEN
old R10-A 8+3                    = SUPERSEDED FOR LAUNCH
old R10-B1→B6                    = EVIDENCE ONLY
old R10-C                        = PAUSED HISTORICAL CANDIDATE / DO NOT REPAIR
implementation                   = BLOCKED
```

No implementation plan or product code is authorized until T1→T7 are integrated, Whole-R10 GCR + cold independent review complete, and the operator gives final ratification.

---

## Accepted Launch ownership

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Not semantic owners in Launch:

```text
storage/staging/integrity → mechanism
render/view/editor         → mechanism
Search                     → rebuildable projection
async/outbox/jobs          → mechanism
Historical Migration      → cutover capability
backup/restore             → operations/readiness
```

`Artifact`, separate `Approval`, `Distribution`, `Documentary Context`, `Records Governance` and generic `Interchange` are not Launch semantic owners.

---

## Future-evolution law — operator explicit

> **Known future capabilities must not be forgotten or made structurally expensive merely because they are deferred from Launch.**

Controlling rule:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Named horizon:

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

These are architecture counterexamples and attachment-seam evidence, not Launch implementation scope.

---

## Approved technical descent

```text
T1 — Semantic State & Invariants                         ACTIVE
T2 — Governance, Effectivity & Lifecycle Transactions   NOT OPEN
T3 — Authorization & Audit Enforcement                  NOT OPEN
T4 — Exact Content, Storage Integrity & Restore         NOT OPEN
T5 — Durable Async, Search & External Effects           NOT OPEN
T6 — Canonical API / Frontend Journeys                  NOT OPEN
T7 — Historical Migration & Cutover                     NOT OPEN

→ Integrated Whole-R10 Global Coherence Review
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

Do not use the former B1→B6→C→D→E→F order as active routing.

---

## T1 current candidate — headline

T1 proposes the minimum enduring semantic families needed for Launch:

```text
Authentication
  ProviderSubjectBinding
  ApplicationSession

Organization
  Company
  User
  UserProfile
  Area
  Group
  GroupMembership

Authorization
  product Role/Permission vocabulary
  RoleAssignment

Controlled Documents
  DocumentType + numbering semantics
  Document + Template role/origin
  Revision
  WorkingContent
  Submission
  GovernanceRoute current config
  bounded GovernanceAttempt over SUBMISSION|OBSOLESCENCE
  Step/Decision evidence + SubmissionFeedback
  RevisionCancellation
  Release
  OfficialRendition only when required
  Obsolescence result
  imported/native provenance seam

Audit
  AuditEvent
```

Explicitly absent from Launch T1 unless a later stage proves a named Launch consumer:

```text
Artifact semantic owner
DocumentTypeCategory
generic Dictionary/System Value platform
TemplateSpec platform
DRAFT EditorialComment platform
Periodic Review state
Distribution state
Dossier/Evidence/Records
Interchange / Governed Export / repository receipts
global AuditChainHead/hash chain
```

T1 is **not yet accepted authority**.

---

## Exact next step

**Operator adjudication of T1 recommendations T1-A→T1-I and bounded product-semantic question T1-J.**

T1-J asks:

> If a Document Type uses `NoHumanApproval`, may governed obsolescence also complete with no human Step after authorized initiation/reason/eligibility checks, or must obsolescence always require at least one human governance Step?

Do not answer this silently in T2.

Until T1 closes, do not write SQL/table/index design, package layout, storage locator design, exact permission catalog, async topology, API routes, migration execution design, implementation plans or product code.