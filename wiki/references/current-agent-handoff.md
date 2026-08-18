# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1→T5 + DECISION REGISTRY OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT ACTIVE; T6 NOT OPEN**  
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
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md` — **ACTIVE INDEPENDENT REVIEW REQUEST / NOT AUTHORITY**
15. `wiki/architecture/launch-v1-scope-rebaseline.md`
16. old R3–R9.5 / old R10/current implementation only as evidence allowed by current authority/registry

Completed T5 design/adjudication staging was removed after durable promotion. Git history is the archive.

## Current checkpoint

```text
Product Contract                         = ACTIVE / OPERATOR-APPROVED
Whole-Product GCR A1–A10                 = CLOSED / OPERATOR-APPROVED
Launch ownership topology                = CLOSED / OPERATOR-APPROVED / 4+1
T1 Semantic State & Invariants           = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx             = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit                 = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore         = CLOSED / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects          = CLOSED / OPERATOR-RATIFIED
Decision Registry                        = CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint      = ACTIVE / REVIEW REQUEST STAGED
Fable independent verdict                = PENDING EXTERNAL REVIEWER
T6 Canonical API / Frontend Journeys     = NOT OPEN
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Binding laws

```text
REV000 = initial issuance
REV001 = first revision
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

## Mandatory stage closure protocol

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

An independent checkpoint may keep the next stage closed after the prior stage is fully promoted. Review findings remain evidence until explicitly adjudicated.

## Closed T4 headline

Detailed authority:

`wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`

```text
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
opaque managed-content handle is mechanism only
OPEN→READY server-verified admission
create-once/no-overwrite
WorkingContent = DRAFT recovery baseline
SUBMIT/Rendition semantic tx makes zero provider/scanner calls
backup/restore exact-content fail-closed readiness
post-snapshot UserProfile erasure reconciliation before restored profile serving
```

## Closed T5 headline

Detailed authority:

`wiki/architecture/r10-t5-durable-async-search-external-effects.md`

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
search_refresh(document_id) always-required durable projection job
official_rendition_render conditional only for frozen policy-required OfficialRendition
PDF source direct viewer by default
DOCX + SourceOnly direct read-only viewer; no persisted PDF merely for viewing
required job enqueue transaction-coupled to semantic fact
provider/renderer work outside semantic transaction
OfficialRendition finalization T4/T2/T3-revalidated and idempotent
Search = rebuildable PostgreSQL projection keyed by Document
latest-state refresh makes duplicate/out-of-order jobs converge
Search may lag by omission but never grant stale authority/effectivity
full Search rebuild mandatory; always-on crawler not baseline
GC = periodic reconciliation over GC_PENDING with immediate canonical recheck
no mandatory Launch notifications/event bus
no mandatory durable IdP-disable job
no generic ExternalEffectReceipt
bounded-retry terminal-visible/redrivable jobs with bounded-ID payloads
minimum async operational visibility required
```

## Active post-T5 Fable checkpoint

Review request:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md`

The reviewer must cold-reconstruct the repository and attack T1→T5 as one integrated system, especially:

```text
DRAFT→SUBMIT→governance→optional Rendition→Release→Search
offboarding vs governance/async jobs
replacement Release/obsolescence vs stale Search
GC vs current/governed/backup-protected content
restore vs offboarding/privacy
viewer/preview vs OfficialRendition
Audit + required durable-job composition
Search/AuthZ stale projection
Decision Registry drift
essential vs accidental complexity
future capability seams
```

Review verdict is evidence only. Any material finding must name the smallest exact reopen set and everything that remains frozen.

## Exact next step

```text
external Fable review of staged packet
→ adjudicate review findings against current authority + Method
→ bounded corrections/reopen only if justified
→ optional Fable delta review if material fixes occur
→ explicit post-T5 checkpoint closure
→ only then open T6
```

Do **not** open T6 before this checkpoint closes.

No final SQL/index/package/process topology, public API/frontend contract, Historical Migration execution plan, implementation plan or product code is authorized.
