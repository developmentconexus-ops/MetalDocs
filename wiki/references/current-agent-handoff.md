# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT CLOSED; T6 CORRECTED GLOBAL-MAXIMUM ADJUDICATION READY / OPERATOR MATERIAL ADJUDICATION NEXT; T7 NOT OPEN**  
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
14. `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md`
15. `docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md`
16. `docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md`
17. `docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md` — **OPERATOR DECISION TARGET**
18. current implementation only when claim-specific evidence is needed

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
T6 evidence/inversion pass               = COMPLETE ENOUGH FOR ADJUDICATION
T6 base candidate                        = STAGED / NON-AUTHORITATIVE
T6 corrected adjudication packet         = READY / NON-AUTHORITATIVE
operator material adjudication           = NEXT
T7 Historical Migration / Cutover        = NOT OPEN
implementation                           = BLOCKED
```

## Structural Inversion result

Current API/frontend is current-state evidence only and carries major superseded concepts. T6 has no obligation to retain routes/modules/DTOs/screens by migration cost or sunk cost.

The target is rederived from Product Contract REV001 + T1→T5:

```text
semantic public API instead of legacy module API
semantic-lens frontend instead of old navigation ontology
exact immutable Submission review instead of reviewer WorkingContent mutation
current-effective Library instead of polymorphic document screen
canonical Search instead of mandatory Search infrastructure
provider-neutral content/editor mechanisms behind T4/OCC
```

## Corrected T6 material slate

The operator should adjudicate the combined candidate + corrected packet. Headline:

```text
T6-A   rebuild pre-launch /api/v1; OpenAPI 3.0.3 remains Launch wire-contract feature set; no compatibility layer
T6-B   semantic frontend lenses / My Work + exact governance case
T6-C   closed semantic public operation census; no generic action/module-shaped API
T6-D   server-derived allowed_actions = UX hint only
T6-E   numbering TYPE|TYPE_AREA, fixed '-', min width 3, no generic grammar
T6-F   one strong ETag/If-Match DRAFT OCC covers title + WorkingContent; stale = 412
T6-G   bound upload_id OPEN→READY→OCC attach; client never owns descriptor
T6-H   reviewer operates only on exact immutable Submission
T6-I   semantic exact-byte routes; provider/storage identity hidden
T6-J   fidelity-gated EigenPal-class first candidate; ONLYOFFICE fallback; no EditorSession baseline
T6-K   Search materialization/search_refresh OFF for Launch
T6-L   domain history and Audit remain separate
T6-M   minimal Admin Center + lost-update ETags + existing-provider identity binding
T6-N   RFC9457; code canonical/type mechanically derived; closed error families incl. dependency/ratelimit
T6-O   natural HTTP idempotency first; exact Idempotency-Key POST set; 24h replay
T6-P   opaque cursor default20/max100; no totals/offset/generic filter DSL
T6-Q   purpose-built semantic read models; no DB/module DTO leakage
T6-R   preserve React/TanStack feature-sliced mechanism pattern, replace legacy taxonomy
T6-S   Keycloak Authorization Code + MetalDocs ApplicationSession + CSRF; no local credentials/JIT user
T6-V   blank seed = mechanism; Template/revise copy exact released SOURCE, never OfficialRendition
```

Everything not named remains frozen.

## Exact next step

```text
operator adjudicates material T6 slate
→ revise only rejected/refined decisions
→ platform-facing T6 summary
→ explicit operator summary ratification
→ promote T6 durable authority
→ reconcile Decision Registry
→ remove staging
→ only then T7
```

No final SQL/index/package/process topology, Historical Migration execution plan, implementation plan or product code is authorized.
