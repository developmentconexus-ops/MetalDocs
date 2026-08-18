# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT/GCR/4+1/T1→T7 APPROVED; T1 DECISIONS ACCEPTED / PLATFORM-SUMMARY RATIFICATION NEXT; T2 NOT OPEN**  
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
7. `wiki/architecture/r10-technical-architecture.md` — **ACTIVE REBASELINED T1→T7 TECHNICAL-STAGE AUTHORITY + MANDATORY SUMMARY-RATIFICATION GATE**
8. `docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md` — T1 design candidate/evidence
9. `docs/superpowers/analysis/2026-08-18-r10-t1-operator-adjudication.md` — **ACTIVE T1 ADJUDICATION / SUMMARY-RATIFICATION STAGING**
10. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records-Governance defer overlay
11. prior cohesive/R9.5/R10 B1–B6/C material only as evidence where current authorities do not supersede it

`wiki/architecture/cohesive-platform-redesign.md` is a short **prior-design evidence compatibility page**, not active routing. Git history preserves its former full narrative.

Git history and current runtime/schema/OpenAPI remain evidence, not automatic target authority.

---

## Current checkpoint

```text
Product Contract                 = ACCEPTED / PROMOTED
Whole-Product GCR A1–A10         = OPERATOR-ADJUDICATED / ACCEPTED
Launch ownership topology        = CLOSED / OPERATOR-APPROVED / 4+1
T1→T7 technical decomposition    = CLOSED / OPERATOR-APPROVED
T1 decisions A→J                 = OPERATOR-ADJUDICATED / ACCEPTED
T1 platform summary              = NEXT / EXPLICIT OPERATOR RATIFICATION REQUIRED
T1 final promotion/closure       = PENDING SUMMARY RATIFICATION
T2→T7                            = NOT OPEN
old R10-A 8+3                    = SUPERSEDED FOR LAUNCH
old R10-B1→B6                    = EVIDENCE ONLY
old R10-C                        = PAUSED HISTORICAL CANDIDATE / DO NOT REPAIR
implementation                   = BLOCKED
```

No implementation plan or product code is authorized until T1→T7 are integrated, Whole-R10 GCR + cold independent review complete, and the operator gives final ratification.

---

## Mandatory T-stage comprehension gate — operator explicit

For every technical stage `Tn`:

```text
Tn candidate/design
→ material decision adjudication
→ assistant presents platform-facing summary
→ explicit operator approval of that summary
→ promote/close Tn
→ only then open Tn+1
```

The summary must explain what was decided **and how it will work in the MetalDocs platform**, not merely repeat architecture labels. It must also identify deliberate deferrals/future seams and material reopen triggers.

A prior “approve A/B/C” response does not by itself authorize opening the next stage unless the platform summary has also been explicitly ratified.

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
T1 — Semantic State & Invariants                         SUMMARY RATIFICATION PENDING
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

## T1 adjudicated headline

The operator accepted:

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

Accepted T1-J:

> If the Document Type is `NoHumanApproval`, governed obsolescence may complete with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks. It remains a governed operation and creates no fake System approver.

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

T1 is **not yet durably promoted/closed** because the operator-facing summary must be ratified first.

---

## Exact next step

**Present the T1 platform-facing summary to the operator and obtain explicit summary ratification.**

Do not open T2 before that approval.

Until T1 closes, do not write SQL/table/index design, package layout, storage locator design, exact permission catalog, async topology, API routes, migration execution design, implementation plans or product code.
