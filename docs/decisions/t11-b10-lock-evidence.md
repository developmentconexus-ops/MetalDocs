---
id: t11-b10-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for exact operator-LOCKED B10 P8 and post-LOCK P9/P10 Evidence preserved outside the merge candidate while T11 remains open.
---

# T11 B10 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.  
> **Scope:** B10 — Organization Administration operator-LOCKED P8 + P9/P10 proof only.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Why this exists

MetalDocs merge candidates contain no `docs/work/**`, while the Frontend Product Experience Method requires exact operator-LOCKED P8 Evidence to remain recoverable for later P11 assembly/fidelity proof. B10 also carries post-LOCK P9/P10 proof that must remain inspectable without promoting temporary planning notes into Product authority.

Therefore PR #170 preserves its exact pre-cleanup planning tree under one immutable Evidence ref and records the canonical B10 identities here before temporary work is removed from the candidate branch.

This locator is Evidence routing only. It is not Product authority, a second roadmap, or permission to reopen B10.

## 2. Preserved Git identity

```text
repository   developmentconexus-ops/MetalDocs
source PR    #170
evidence ref evidence/t11-pr170-b10-locks-20260824
exact commit b8c607cbd30d61d6bcf6ec1ea734ed1653d2569e
```

The ref was created from the exact B10 pre-cleanup HEAD and remotely resolves to that exact commit. It MUST NOT be moved while current T11/P11/P13/P14 proof depends on B10.

## 3. Canonical B10 Evidence

| Evidence | Path on exact Evidence commit | Git blob |
|---|---|---|
| P8 functional LOCK artifact | `docs/work/current/t11-b10-organization-administration-p8.html` | `1d1cc7d5cb42e034ab9ee71a21c96918cdcf691d` |
| Operator LOCK record | `docs/work/current/t11-b10-operator-lock.md` | `576996f3078c6ab93bd36f33c29fcfcba11c709c` |
| P9 Screen Contract | `docs/work/current/t11-b10-screen-contract.md` | `24cc9530268be952cf3f6b1768a9cf217425ed93` |
| P10 pattern consolidation | `docs/work/current/t11-b10-pattern-consolidation.md` | `23303b363d1ecf0284fef8375ceee78aac0f9f2b` |

The P8 blob is the canonical operator-LOCKED interactive evidence for B10. Later P11 must consume that exact artifact, not a reconstructed approximation.

The same exact Evidence commit also preserves B10 P7/planning and P8-realization records that existed under `docs/work/current/`; those remain supporting Evidence only.

## 4. Protected B10 meaning

Durable Product/architecture authority remains in current Product/R10 documents. The LOCK protects the frontend structure proved by the exact P8/P9/P10 Evidence, including:

```text
stable /admin/organization route
Company / Users / Areas / Groups local Organization workspace
separate User Profile / Provider Binding / Eligibility truths
User offboarding vs re-enable consequence
Area identity vs lifecycle separation
Group identity only in B10
GroupMembership + RoleAssignment remain B11
no invented User/Area/Group global search/filter
independent ETag/write domains preserved
```

If current durable authority conflicts with a temporary planning note, current durable authority wins unless material Evidence invokes the bounded reopen law.

## 5. P11 retrieval law

When FP2/P11 later opens:

```text
read current accepted Product/architecture authority
→ read current roadmap
→ use this locator only for B10 LOCK/proof identity
→ fetch exact P8 blob from exact Evidence commit
→ assemble a new disposable P11 prototype
→ fidelity-check B10 protected structure
→ reopen only if integration materially falsifies the LOCK
```

Do not edit the locked P8 blob in place. A material B10 correction requires the smallest normal frontend/upstream reopen and a newly operator-LOCKED Evidence identity.

## 6. Reopen / retirement trigger

B10-A1 paginated-browse sufficiency was validated for current Launch P8; it may reopen only on material scale/findability Evidence or changed operating requirements, not preference.

Keep this locator and Evidence ref while any current T11/P11/P13/P14 proof relies on B10. Retirement requires later accepted authority explicitly proving the exact B10 Evidence is no longer required; repository-cleanup preference alone is not a retirement trigger.
