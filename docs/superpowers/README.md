# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **PRODUCT CONTRACT REV001 + T1→T5 + Decision Registry OPERATOR-RATIFIED; POST-T5 FABLE ROUND-1 AMENDMENTS PROMOTED; DELTA REVIEW PENDING; T6 NOT OPEN.**

Durable accepted truth belongs in `wiki/`. Active review/design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md          REV001
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

## Current active review staging

- `analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — original independent cold-review request / evidence.
- `analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md` — original Fable review / evidence only.
- `analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md` — **OPERATOR-RATIFIED ROUND-1 DISPOSITION / PROMOTION MAP.**
- `analysis/2026-08-18-t1-t5-fable-delta-review-request.md` — **ACTIVE DELTA REVIEW REQUEST.**

Expected Fable response:

- `analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

## Ratified post-T5 delta

```text
M1  optional materialized Search serializes per Document before canonical read through rewrite/removal; no FIFO.
M2  restored sessions invalid before serving; required post-snapshot security teardown reconciled/proven; T7 chooses proof mechanism.
M3  Search baseline = canonical PostgreSQL query/view; materialization/search_refresh/rebuild only on proven derived/expensive/measured consumer.
L1  title = Revision-governed metadata.
L2  late rendition for dead candidate = semantic no-op/reclaimable output.
L3  live bounded admission claim protects READY content from GC.
L4  bounded withdrawal of active human-governed obsolescence request.
L5  provider-disable wording aligned to T5-L.
```

Also ratified:

```text
same-DB durable-intent restore coherence guard
registry wording cleanup
T6 source upload/T4 admission UX explicit
T6 Search materialization proof explicit
T7 post-snapshot security-teardown recovery choreography explicit
```

Everything else remains frozen.

## Active technical path

```text
Product Contract                                       REV001 / OPERATOR-APPROVED
T1 Semantic State & Invariants                        CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions  CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                 CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore        CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects          CLOSED / OPERATOR-RATIFIED
Decision Registry                                     CURRENT / RECONCILED
Fable Round-1 amendments                              OPERATOR-RATIFIED / PROMOTED
Fable delta review                                    PENDING
Post-T5 checkpoint                                    OPEN
T6 Canonical API / Frontend Journeys                 NOT OPEN
T7 Historical Migration & Cutover                    NOT OPEN

→ operator dispatches Fable to active delta-review request
→ Fable writes delta verdict through GitHub
→ adjudicate any exact remaining disagreement
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