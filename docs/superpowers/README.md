# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1→T5 + Decision Registry CLOSED / OPERATOR-RATIFIED; FABLE REVIEW RECEIVED; AUTHOR ROUND-1 ADJUDICATION PENDING OPERATOR RATIFICATION; T6 NOT OPEN.**

Durable accepted truth belongs in `wiki/`. Active review/design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
→ wiki/architecture/r10-t5-durable-async-search-external-effects.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Current active staging

- `analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — independent cold-review request / evidence.
- `analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` — **FABLE REVIEW RECEIVED / EVIDENCE ONLY.**
- `analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — **AUTHOR ROUND-1 RESPONSE / OPERATOR RATIFICATION NEXT.**

Completed T5 candidate/subgate/adjudication staging was removed after durable promotion. Git history is the archive.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

For remaining T-stages:

```text
CURRENT / PRESERVE / REFINED → baseline
REOPEN                       → design in owning T-stage
DEFERRED                     → future seam/counterexample only
SUPERSEDED                   → reject inheritance absent explicit material reopen
```

## Fable checkpoint status

Fable verdict:

```text
APPROVE T1→T5 WITH MATERIAL FIXES
BLOCKER = 0
MAJOR   = 3
LOW     = 5
formal T-stage reopen = NONE
```

Author Round-1 recommendation, still non-authoritative:

```text
M1 accept — conditional materialized Search projection needs per-Document write serialization; FIFO still unnecessary.
M2 accept/refine — restore invalidates all restored ApplicationSessions and fails closed until required post-snapshot security teardown is reconciled/proven; T7 chooses smallest proof mechanism.
M3 accept option (b) — canonical PostgreSQL query/view is Search baseline; materialized projection + search_refresh + rebuild activate only when T6 proves a real derived/expensive consumer or measured need.

L1 title becomes Revision-governed metadata.
L2 late rendition finalization becomes no-op when Submission/Revision is no longer eligible.
L3 live admission claim/binding protects READY handles from GC until bounded release/expiry.
L4 bounded initiator/manager withdrawal closes active human-governed obsolescence deadlock.
L5 T3 provider-disable wording aligns to T5-L.
```

No durable authority has been modified yet by these recommendations.

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
Fable independent review                               RECEIVED
Author Round-1 adjudication                            WRITTEN / OPERATOR RATIFICATION NEXT
Durable bounded amendments                             NOT APPLIED
Post-T5 checkpoint                                     OPEN
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ operator ratification of author adjudication
→ bounded authority/registry amendments only
→ GitHub delta challenge by Fable if dispatched / material disagreement remains
→ explicit checkpoint closure
→ T6
→ T7
→ Integrated Whole-R10 GCR
→ cold independent final review
→ final operator ratification
→ implementation spec/plan
→ code
```

## Hard stop

No product implementation or implementation plan is authorized while active design/review gates remain open.
