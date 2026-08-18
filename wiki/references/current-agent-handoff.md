# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT/GCR/4+1 OWNERSHIP APPROVED; REMAINING TECHNICAL ARCHITECTURE RE-DERIVATION NEXT; R10-C PAUSED**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — **ACCEPTED PRODUCT AUTHORITY**
5. `wiki/architecture/whole-product-alignment-review.md` — **OPERATOR-ADJUDICATED GCR / ACTIVE ROUTING AUTHORITY**
6. `wiki/architecture/launch-v1-ownership-topology.md` — **OPERATOR-APPROVED 4+1 SEMANTIC OWNERSHIP AUTHORITY + FUTURE-EVOLUTION LAW**
7. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records-Governance defer overlay
8. prior cohesive/R9.5/R10 material only as evidence where the current authorities do not supersede it
9. `docs/superpowers/analysis/2026-08-18-r10-c-artifact-physical-integrity-integrated-candidate.md` only as **PAUSED EVIDENCE; DO NOT REPAIR OR PROMOTE**

Git history and current runtime/schema/OpenAPI remain evidence, not automatic target authority.

---

## Current checkpoint

```text
Product Contract             = ACCEPTED / PROMOTED
Whole-Product GCR A1–A10     = OPERATOR-ADJUDICATED / ACCEPTED
Launch ownership topology    = CLOSED / OPERATOR-APPROVED / 4+1
R10-A prior 8+3 topology     = SUPERSEDED FOR LAUNCH
R10-B1/B2/B3/B4/B5/B6        = prior technical evidence only where implicated
R10-C                        = PAUSED / NON-AUTHORITATIVE / DO NOT REPAIR
R10-D–F                      = NOT STARTED

remaining technical architecture re-derivation = NEXT
implementation                              = BLOCKED
```

No implementation plan or product code is authorized until the remaining technical design is re-derived, integrated, reviewed cold and finally ratified by the operator.

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

`Artifact`, `Distribution`, `Documentary Context`, `Records Governance` and generic `Interchange` are not Launch semantic owners.

Controlled Documents owns the complete controlled-document lifecycle, including the smallest sequential governance semantics needed for both exact Submission governance and explicit no-replacement obsolescence. This is one semantic authority, not one giant code unit.

---

## Future-evolution law — operator explicit

The operator approved the 4+1 topology with an explicit constraint:

> **Known future capabilities must not be forgotten or made structurally expensive merely because they are deferred from Launch.**

Controlling rule:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Named horizon to preserve:

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

These are **evidenced future direction**, not Launch implementation scope.

Remaining technical design must make reasonable future attachment additive around stable semantic anchors without duplicating/reinterpreting Document, Revision, Submission, Release, User/Group/Area or Audit history. It must not create dormant modules/tables/permissions/jobs or generic frameworks for those future features.

---

## Core invariants that survive

```text
AuthN provider identity != Organization identity != Authorization grants
Document != Revision != WorkingContent != Submission
Revision numbers never reuse
Submission is immutable exact governed attempt
same-Revision resubmit creates a new Submission
Template is an ordinary governed Document role
one sequential governance Step semantic
NoHumanApproval creates no fake approver
Release is system-owned effectivity authority
one current EFFECTIVE truth
obsolescence without replacement is explicit and governed
Audit is evidence, never current-state authority
Search is projection, never access/effectivity authority
storage/provider identity never becomes semantic identity
historical migration never fabricates native history
governed history is preserved in Launch
restore must fail closed on inconsistent semantic/content truth
```

---

## Exact next step

Re-derive the remaining technical architecture **from zero where prior R10 was implicated**, using old R10 only as evidence.

Sequence:

```text
1. minimal persistent semantic facts + invariants per 4+1 owner
2. transactions / concurrency / lifecycle atomicity
3. regenerated Launch AuthZ catalog + check sites
4. exact-content / storage / integrity / restore mechanism
5. async / Search / effects
6. API + frontend journeys
7. Historical Migration / cutover
8. integrated Whole-R10 Global Coherence Review
9. cold independent review
10. operator final ratification
11. implementation spec/plan
12. code
```

For every material decision run:

```text
Launch correctness
+ essential vs accidental complexity
+ authority ownership
+ proof strategy
+ named-future compatibility
```

Do not restore an old entity/context/table simply because it previously solved a future feature. Do not choose a Launch shortcut that forces future features to rewrite immutable history or dismantle core authority when a small seam can prevent that dead end.

## Gate

**Implementation remains BLOCKED.**