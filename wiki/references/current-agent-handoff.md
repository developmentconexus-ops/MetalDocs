# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; T6 MATERIAL DECISIONS OPERATOR-APPROVED; PLATFORM-FACING T6 SUMMARY RATIFICATION NEXT; T7 NOT OPEN**  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** **BLOCKED — architecture/design only**

## Fresh-session route

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. this file
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. `wiki/architecture/r10-technical-architecture.md`
14. `docs/superpowers/analysis/2026-08-18-r10-t6-operator-material-adjudication.md`
15. `docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md` — **OPERATOR SUMMARY RATIFICATION NEXT**
16. T6 candidate/evidence files only when provenance for a material decision is needed
17. current implementation only as claim-specific evidence

Completed Fable staging is removed; Git history is the archive.

## Current checkpoint

```text
Product Contract                         = REV001 / OPERATOR-APPROVED
Whole-Product GCR A1–A10                 = CLOSED / OPERATOR-APPROVED
Launch ownership topology                = CLOSED / OPERATOR-APPROVED / 4+1
T1 Semantic State & Invariants           = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx             = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit                 = CLOSED / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore         = CLOSED / OPERATOR-RATIFIED
T5 Durable Async/Search/Effects          = CLOSED / OPERATOR-RATIFIED
Decision Registry                        = CURRENT / RECONCILED / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint      = CLOSED / OPERATOR-APPROVED
T6 material decisions                    = OPERATOR-APPROVED
T6 platform-facing summary               = STAGED / RATIFICATION NEXT
T6 durable authority                     = NOT YET
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## T6 approved platform direction

```text
API
  rebuild pre-launch /api/v1 from current semantics
  OpenAPI contract-first + generated Go/TS boundaries
  no /api/v2 or compatibility shims

AuthN/browser
  Keycloak Authorization Code
  MetalDocs ApplicationSession
  no local credentials/JIT User
  session-bound CSRF on unsafe requests

Frontend
  Library / My Work / exact Governance case / Document official-work-history / Audit / Admin
  stable semantic route meanings

DRAFT/content
  Revision title + source share one WorkingContent generation
  PATCH + strong If-Match; stale = 412
  T4 upload_id OPEN→READY→OCC attach
  client never owns ExactContentDescriptor

Governance
  exact immutable Submission is review truth
  Step Decision is singleton immutable PUT
  reviewer case access never mutates WorkingContent

Search
  canonical PostgreSQL code/title + filters
  materialized Search/search_refresh OFF for Launch

Representation
  semantic byte URLs hide provider identity
  SourceOnly vs OfficialRendition explicit
  one fidelity-gated DOCX provider; no EditorSession correctness baseline

Admin
  Organization / Access / Document Governance only
  User eligibility singleton PUT executes T3 offboarding/reenable semantics
  strong ETag/If-Match on mutable authority-bearing config

Transport
  RFC9457 + canonical semantic code
  natural HTTP idempotency first
  Idempotency-Key only for truly non-idempotent semantic POST creation
  opaque cursor lists default20/max100

Seeds
  blank = trusted mechanism asset
  Template/revise copy exact governed/released SOURCE, never OfficialRendition
```

Current implementation carries no compatibility entitlement from sunk cost.

## Exact next step

Review and explicitly ratify:

`docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md`

If ratified:

```text
promote T6 durable authority to wiki/
→ reconcile Decision Registry
→ update router/handoff/index/PR
→ remove completed T6 staging
→ open T7 only after those steps complete
```

Do **not** open T7 or write implementation plan/code before summary ratification and T6 promotion.