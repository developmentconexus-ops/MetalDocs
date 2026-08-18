# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T5 CLOSED / OPERATOR-RATIFIED; DECISION REGISTRY CURRENT; POST-T5 FABLE REVIEW RECEIVED / AUTHOR ROUND-1 ADJUDICATION PENDING OPERATOR RATIFICATION; T6 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — design/documentation only**

This file is the technical-stage router. Detailed accepted semantics live in dedicated authorities; this page owns current stage status, reading order and exact next action.

## 1. Binding authority chain

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. active independent-review/adjudication staging only when a review checkpoint is open

Historical R3–R9.5 / old R10 / current implementation/schema/OpenAPI are evidence only unless current authority or the Decision Registry preserves a decision.

Independent Fable review and author adjudication are **review evidence/staging**, never target authority by themselves.

## 2. Binding method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

## 3. Active semantic ownership baseline

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
managed content / storage / malware inspection
render/view/editor providers
Search
async/jobs/retry/leases
notifications
Historical Migration execution
backup/restore transport/readiness
```

Do not resurrect `Artifact`, separate `Approval`, Distribution, Documentary Context, Records Governance or generic Interchange ownership by technical convenience.

## 4. Technical descent

```text
T1 — Semantic State & Invariants                              CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions        CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                       CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore              CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects                CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                             CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                          ACTIVE / ROUND-1 ADJUDICATION
Fable independent review                                     RECEIVED
Author Round-1 adjudication                                  WRITTEN / OPERATOR RATIFICATION PENDING
T6 — Canonical API / Frontend Journeys                       NOT OPEN
T7 — Historical Migration & Cutover                          NOT OPEN
implementation                                                BLOCKED
```

T6 may open only after this checkpoint is explicitly closed.

## 5. Closed T-stage authorities

```text
T1 → wiki/architecture/r10-t1-semantic-state-invariants.md
T2 → wiki/architecture/r10-t2-governance-effectivity-transactions.md
T3 → wiki/architecture/r10-t3-authorization-audit-enforcement.md
T4 → wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
T5 → wiki/architecture/r10-t5-durable-async-search-external-effects.md
```

No durable T1→T5 authority has been changed by the Fable checkpoint yet.

## 6. Decision Registry

Authority:

`wiki/architecture/rebaseline-decision-registry.md`

Registry vocabulary:

```text
CURRENT
PRESERVE
REFINED
REOPEN
DEFERRED
SUPERSEDED
```

T5 is reconciled into the registry. T6's official REOPEN set is the next design set, but T6 remains held closed by the active independent-review checkpoint.

## 7. Post-T5 independent Fable checkpoint — ACTIVE

Review request:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md`

Independent review:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md`

Fable commit:

`bdef5fc3c4004aa3ab4deefc9e8373dd3efcf856`

Fable verdict:

```text
APPROVE T1→T5 WITH MATERIAL FIXES
BLOCKER = 0
MAJOR   = 3
LOW     = 5
NOTE    = 3
formal minimal reopen set = NONE
T6 = ready only after material findings are adjudicated
```

Material Fable findings:

```text
M1 — Search latest-state projection can be overwritten by an older overlapping refresh.
M2 — historical restore can resurrect revoked ApplicationSessions/access teardown.
M3 — materialized Search projection + always-required search_refresh were ratified before a named consumer proves materialization.
```

Author Round-1 adjudication:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md`

Author recommendation, still NON-AUTHORITATIVE:

```text
M1 ACCEPT
  if a materialized projection exists, serialize per-Document projection write before canonical read through write; FIFO remains unnecessary.

M2 ACCEPT ROOT CAUSE / REFINE FIX
  invalidate all restored ApplicationSessions before serving;
  require fail-closed post-snapshot security-teardown reconciliation before ordinary serving;
  do not freeze a generic security journal before T7 proves the smallest recovery mechanism.

M3 ACCEPT — OPTION (b)
  Search journey remains required;
  canonical PostgreSQL query/view is Launch baseline for current canonical search facts;
  materialized projection + search_refresh + rebuild activate only if T6 proves a derived/expensive searchable fact or measured requirement.
```

Accepted-in-principle LOW recommendations in the staging adjudication include Revision-governed title, late-Rendition no-op, live admission protection from GC, bounded obsolescence withdrawal, and T3/T5 provider-disable wording alignment. They are not durable authority until operator ratification.

## 8. Current gate

```text
T1→T5 durable authorities                         CLOSED / RATIFIED
Decision Registry                                 CURRENT
Fable independent review                          RECEIVED
Author Round-1 adjudication                        WRITTEN
Operator ratification of Round-1 adjudication     NEXT
Durable bounded amendments                         NOT APPLIED
Post-T5 checkpoint                                 OPEN
T6                                                  NOT OPEN
implementation                                      BLOCKED
```

After operator ratification:

```text
apply only the ratified bounded amendments to T1→T5 + Decision Registry
→ update router/handoff
→ let Fable inspect the GitHub adjudication/delta if requested or if material disagreement remains
→ explicitly close post-T5 checkpoint
→ only then open T6
```

## 9. T6 — Canonical API / Frontend Journeys — NOT OPEN

Current registry T6 REOPEN set remains controlling until the post-T5 amendments are ratified. The author adjudication additionally recommends explicitly naming source upload/admission UX and requiring Search journeys to prove whether canonical query/view is sufficient before materialization is introduced.

T6 must not reopen T1→T5 absent material evidence and explicit bounded reopen.

## 10. T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set: source evidence, migration modes, imported target-owned shapes, ordinal/content/governance provenance, plan/dry-run/idempotency/reconciliation, semantic-unit atomicity and cutover/readiness/rollback/deletion map.

The Fable author adjudication proposes adding post-snapshot security-teardown reconciliation to the existing restore/erasure choreography work, subject to operator ratification.

## 11. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent final review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**
