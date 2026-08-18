# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1→T5 + Decision Registry CLOSED / OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT ACTIVE; T6 NOT OPEN.**

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

- `analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — **ACTIVE INDEPENDENT FABLE COLD-REVIEW REQUEST / NOT TARGET AUTHORITY.**

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

## Closed T5 headline

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
search_refresh(document_id) always-required durable projection job
official_rendition_render conditional only for frozen policy-required OfficialRendition
PDF source direct-view by default
DOCX + SourceOnly direct read-only viewer; no persisted PDF merely for viewing
transaction-coupled enqueue for required future work
provider/renderer execution outside semantic tx
OfficialRendition finalization T4/T2/T3-revalidated and idempotent
PostgreSQL rebuildable Search projection keyed by Document
latest-state refresh makes duplicate/out-of-order jobs converge
Search lag may omit but never grant stale authority/effectivity
full Search rebuild mandatory; always-on crawler not baseline
GC periodic reconciliation over GC_PENDING
no mandatory Launch notifications/event bus
no mandatory durable IdP-disable job
no generic ExternalEffectReceipt
bounded-retry terminal-visible/redrivable jobs with bounded-ID payloads
minimum async operational visibility
```

## Active Fable checkpoint

The independent reviewer must reconstruct T1→T5 cold from repo authority and attack integrated flows, authority uniqueness, races, Decision Registry drift, overengineering and future seams.

Review request:

`analysis/2026-08-18-t1-t5-integrated-fable-review-request.md`

Review findings are evidence only. T6 stays closed until material findings are adjudicated and the post-T5 checkpoint explicitly closes.

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  CLOSED / OPERATOR-RATIFIED
T4 Exact Content, Storage Integrity & Restore         CLOSED / OPERATOR-RATIFIED
T5 Durable Async, Search & External Effects           CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                    ACTIVE / REVIEW REQUEST STAGED
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ Fable findings adjudication
→ explicit post-T5 checkpoint closure
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
