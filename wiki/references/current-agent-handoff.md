# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + T3 + T4 + DECISION REGISTRY OPERATOR-RATIFIED; T5 DURABLE ASYNC / SEARCH / EXTERNAL EFFECTS ACTIVE**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — design/documentation only**

## Fresh-session route

Read in this order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/rebaseline-decision-registry.md`
12. `wiki/architecture/r10-technical-architecture.md`
13. `docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` — **ACTIVE NON-AUTHORITATIVE T5 CANDIDATE / OPERATOR ADJUDICATION NEXT**
14. `wiki/architecture/launch-v1-scope-rebaseline.md`
15. old R3–R9.5 / R10-B1→B6/C and current implementation only as evidence allowed by the registry

## Current checkpoint

```text
Product Contract                 = ACCEPTED / REV000 INITIAL
Whole-Product GCR A1–A10         = ACCEPTED
Launch ownership topology        = CLOSED / APPROVED / 4+1
T1 Semantic State & Invariants   = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx     = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit         = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore = CLOSED / OPERATOR-RATIFIED
Decision Registry                = CURRENT / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects  = ACTIVE / NON-AUTHORITATIVE CANDIDATE
T6→T7                            = NOT OPEN
implementation                   = BLOCKED
```

## Binding laws

```text
REV000 = initial issuance
REV001 = first revision
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

## Mandatory stage gate

```text
read registry
→ design only Tn REOPEN set
→ operator adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promote/close Tn
→ update registry
→ remove staging
→ only then Tn+1
```

## Closed T4 headline

Detailed authority:

`wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`

```text
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
no mandatory whole-Submission JCS digest
opaque managed-content handle is mechanism only
one provider-neutral ManagedContentStore / one active store
Local dev/test/conformance + AWS S3 reference production
OPEN→READY server-verified admission
opaque admission binding
UNTRUSTED_EXTERNAL CLEAN malware gate at governed boundary
create-once/no-overwrite
WorkingContent = DRAFT recovery baseline
SUBMIT/Rendition semantic tx makes zero provider/scanner calls
only non-governed unreferenced content reclaimable
backup = DB recovery point + exact required-content set + GC exclusion
restore exact-content fail-closed readiness
post-snapshot UserProfile erasure reconciliation before serving restored profile data
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

The active candidate currently recommends T5-A→T5-P. Headline:

```text
one Postgres-backed durable-job runtime; River retained as selected/reference mechanism
mandatory durable jobs = official_rendition_render + search_refresh
GC = periodic reconciliation over durable GC_PENDING, not per-handle outbox
required jobs transactionally enqueue with semantic transition
Rendition render outside tx; final T4 admission + semantic Rendition/Release revalidation inside local tx
Search = PostgreSQL rebuildable projection keyed by Document
search_refresh(document_id) reloads latest canonical state so duplicate/out-of-order jobs are harmless
Search may lag by omission but stale hit never grants access/effectivity
full Search rebuild required; permanent reconciliation crawler not baseline
no Launch notifications/inbox/fanout/event bus
no mandatory durable external IdP-disable job
jobs = at-least-once, idempotent, bounded-retry, fail-loud, terminal-visible
no generic ExternalEffectReceipt family
minimal async health/backlog/retry/failure observability required
```

## Exact next step

Operator adjudication of T5 recommendations `T5-A→T5-P`.

After technical adjudication, **do not open T6**. Present the mandatory platform-facing T5 summary and obtain explicit operator ratification first.

No final SQL/index/package/process topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.