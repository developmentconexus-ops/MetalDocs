# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 + T3 + T4 + Decision Registry CLOSED / OPERATOR-RATIFIED; T5 ACTIVE / OPERATOR ADJUDICATION NEXT.**

Durable accepted truth belongs in `wiki/`. Active, not-yet-promoted design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/r10-t3-authorization-audit-enforcement.md
→ wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Current active staging

- `analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` — **ACTIVE NON-AUTHORITATIVE T5 candidate; operator adjudication of T5-A→T5-P next.**

Completed T4 candidate/adjudication staging was removed after durable promotion. Git history is the archive.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

For every remaining T-stage:

```text
CURRENT / PRESERVE / REFINED → baseline
REOPEN                       → design in owning T-stage
DEFERRED                     → future seam/counterexample only
SUPERSEDED                   → reject inheritance absent explicit material reopen
```

## T5 preserved baseline

```text
Search = rebuildable/eventually-consistent discovery projection
Search never grants access/effectivity
provider calls never join semantic tx
OfficialRendition is a real optional Release gate
T4 managed-content exactness/admission/GC_PENDING laws are closed
notifications are not domain/lifecycle authority
current River/custom outbox code is evidence only
```

## T5 official REOPEN set

```text
which effects actually require durable intent/outbox
worker/lease/retry/DLQ mechanism
renderer execution
notifications if a Launch consumer remains
Search projection/rebuild/freshness/reconciliation
provider effect receipts where needed
```

## Candidate headline

```text
one Postgres-backed durable-job mechanism; River selected/reference implementation
mandatory durable jobs = official_rendition_render + search_refresh
GC = periodic reconciliation over GC_PENDING
transaction-coupled enqueue for required jobs
renderer outside tx; semantic Rendition/Release finalization inside local tx
PostgreSQL rebuildable Search projection keyed by Document
latest-state search refresh makes duplicate/out-of-order jobs harmless
Search lag may omit but never grant stale authority
full rebuild required; always-on reconciliation crawler not baseline
no Launch notifications/event bus
no mandatory durable IdP-disable sync job
at-least-once/idempotent/bounded-retry/fail-loud jobs
no generic external-effect receipt family
minimal async operational visibility required
```

## Mandatory T-stage closure protocol

```text
read registry
→ candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promotion/closure
→ update Decision Registry
→ remove completed staging
→ only then Tn+1
```

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           ACTIVE / CANDIDATE
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ Integrated Whole-R10 GCR
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

## Hard stop

No product implementation or implementation plan is authorized while active design gates remain open.