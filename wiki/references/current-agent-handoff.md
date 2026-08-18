# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 CLOSED / OPERATOR-RATIFIED; T3 AUTHORIZATION + AUDIT ACTIVE / DISCOVERY-DESIGN NEXT**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — **ACCEPTED PRODUCT AUTHORITY; REV000 INITIAL / REV001 FIRST REVISION**
5. `wiki/architecture/whole-product-alignment-review.md` — **OPERATOR-ADJUDICATED GCR A1–A10**
6. `wiki/architecture/launch-v1-ownership-topology.md` — **OPERATOR-APPROVED 4+1 OWNERSHIP + FUTURE-EVOLUTION LAW**
7. `wiki/architecture/r10-technical-architecture.md` — **ACTIVE T1→T7 TECHNICAL AUTHORITY; T1 + T2 CLOSED / T3 ACTIVE**
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md` — **OPERATOR-RATIFIED T2 TECHNICAL AUTHORITY**
9. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records-Governance defer overlay
10. prior cohesive/R9.5/R10 B1–B6/C material only as evidence where current authorities do not supersede it

`wiki/architecture/cohesive-platform-redesign.md` is prior-design evidence/compatibility routing only. Git history and current runtime/schema/OpenAPI remain evidence, not automatic target authority.

---

## Current checkpoint

```text
Product Contract                 = ACCEPTED / PROMOTED / REV000 INITIAL
Whole-Product GCR A1–A10         = ACCEPTED
Launch ownership topology        = CLOSED / APPROVED / 4+1
T1→T7 technical decomposition    = CLOSED / APPROVED
T1 Semantic State & Invariants   = CLOSED / OPERATOR-RATIFIED / PROMOTED
T2 Governance/Effectivity/Tx     = CLOSED / OPERATOR-RATIFIED / PROMOTED
T3 Authorization & Audit         = ACTIVE / DISCOVERY-DESIGN
T4→T7                            = NOT OPEN
old R10-A 8+3                    = SUPERSEDED FOR LAUNCH
old R10-B1→B6                    = EVIDENCE ONLY
old R10-C                        = PAUSED HISTORICAL CANDIDATE / DO NOT REPAIR
implementation                   = BLOCKED
```

Completed T1/T2 staging is removed from the live tree after durable promotion; Git history is the archive.

---

## Mandatory T-stage comprehension gate

For every `Tn`:

```text
Tn candidate/design
→ material decision adjudication
→ assistant presents platform-facing summary
→ explicit operator approval of that summary
→ promote/close Tn
→ only then open Tn+1
```

A technical recommendation approval alone does not open the next stage.

---

## Revision numbering

Binding product convention:

```text
REV000 = initial issuance
REV001 = first revision after initial issuance
REV002 = second revision
...
```

Initial creation creates `REV000 DRAFT`; first Release makes `REV000 EFFECTIVE`; the first subsequent business change cycle is `REV001`.

---

## Accepted T1 headline

```text
Authentication
  ProviderSubjectBinding
  ApplicationSession

Organization
  Company / User / UserProfile / Area / Group / GroupMembership

Authorization
  product Role/Permission vocabulary
  RoleAssignment

Controlled Documents
  DocumentType + numbering semantics
  Document + Template role/origin
  Revision
  WorkingContent
  Submission
  bounded GovernanceAttempt over SUBMISSION|OBSOLESCENCE
  Step/Decision evidence + SubmissionFeedback
  RevisionCancellation
  Release
  OfficialRendition only when required
  Obsolescence result
  native/imported provenance seam

Audit
  AuditEvent
```

`NoHumanApproval` may complete governed obsolescence with zero human Step after authorized initiation, mandatory reason and all eligibility/invariant checks. No raw status toggle and no fake System approver.

Not Launch T1 semantic state: standalone Artifact, taxonomy/dictionary platform, TemplateSpec, DRAFT comment platform, Periodic Review, Distribution, Dossier/Evidence/Records, generic Interchange/export/repository state, global AuditChainHead/hash chain.

---

## Accepted T2 headline

Detailed authority: `wiki/architecture/r10-t2-governance-effectivity-transactions.md`.

```text
one local ACID transaction per native business transition
Document = lifecycle serialization root
WorkingContent = OCC/CAS for DRAFT races
create = code + Document + REV000 DRAFT + initial WorkingContent atomically
first later revision = REV001
SUBMIT freezes exact expected WorkingContent generation + coherent config snapshots
route selector = NAMED_USER | GROUP
Group Step = ANY-one from activation membership snapshot
one active sequential Step
submitter/initiator cannot self-approve the same attempt
no baseline cross-Step same-user prohibition
RETURN / withdraw / cancel preserve immutable Submission history
Release gates = human gate + optional OfficialRendition gate
system Release may occur in same tx when all gates are already satisfied
replacement Release = predecessor SUPERSEDED + successor EFFECTIVE atomically
Distribution remains outside Launch-Core Release atomicity
obsolescence requires current EFFECTIVE + reason + no open replacement + no competing obsolescence
active obsolescence blocks new Revision
same DocumentType route reused for obsolescence
NoHumanApproval obsolescence = zero human Step
route edits never reinterpret an in-flight attempt
READ COMMITTED + narrow explicit serialization/CAS posture
```

Deferred absent named requirement: ALL/N-of-M quorum, ROLE_IN_AREA routing, cross-Step strict SoD, fresh-auth/eSignature, live reassign/overseer, SLA/escalation, separate obsolescence route, scheduled effectivity.

---

## Future-evolution law

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

These are architecture counterexamples/attachment-seam evidence, not Launch implementation scope.

---

## T3 current gate

T3 is **Authorization & Audit Enforcement**.

Derive from the accepted Product Contract + T1/T2 journeys, not from the old role catalog:

```text
personas
→ named operations
→ canonical resource/state predicates
→ Company/Area scope
→ permissions
→ product role bundles
→ RoleAssignment administration law
→ check sites
→ same-local-commit Audit census
→ minimum bounded/PII-minimized Audit facts
```

T3 must include a least-privilege Auditor/Governance Viewer and must prove:

```text
Authorization grant alone never bypasses Controlled Documents lifecycle/governance predicates
active governance action requires current grant + exact active Step participation
provider roles/groups never become canonical product Authorization
current grants are re-evaluated at action time
historical actor attribution survives offboarding/profile erasure
no circular privilege or hidden admin bypass
Audit proves actions but never owns current state
```

The old exact `5×43` catalog is evidence/counterexample only. Do not preserve it by subtraction.

## Exact next step

**Explore and present credible T3 access/Audit design alternatives and the recommended material choices for operator adjudication.**

Do not open T4 or write final SQL/table/index design, package layout, storage locator design, async topology, API/frontend routes, migration execution plan, implementation plan or product code.