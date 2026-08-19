# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE CHECKPOINT CLOSED; T6 CORRECTED GLOBAL-MAXIMUM ADJUDICATION READY; OPERATOR MATERIAL ADJUDICATION NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — architecture/design only**

This file owns current technical-stage status and exact next action.

## 1. Binding authority chain

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md` — REV001
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. T6 bootstrap
15. T6 material candidate
16. T6 evidence docket
17. T6 corrected adjudication packet
18. current API/frontend/runtime only as claim-specific evidence

Current implementation has **no compatibility entitlement**. Structural Inversion controls T6.

## 2. Binding Method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
Structural Inversion
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed.**

## 3. Technical descent

```text
Product Contract                                        REV001 / OPERATOR-APPROVED
T1 — Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CURRENT / OPERATOR-RATIFIED
Post-T5 integrated Fable checkpoint                   CLOSED / OPERATOR-APPROVED
T6 — Canonical API / Frontend Journeys                ACTIVE / CORRECTED ADJUDICATION READY
T7 — Historical Migration & Cutover                   NOT OPEN
implementation                                         BLOCKED
```

## 4. T6 active staging

```text
docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-bootstrap.md
  → stage scope / hard boundaries

docs/superpowers/analysis/2026-08-18-r10-t6-canonical-api-frontend-journeys-candidate.md
  → material architecture candidate + alternatives / Structural Inversion

docs/superpowers/analysis/2026-08-18-r10-t6-external-evidence-docket.md
  → primary/current evidence + claim boundaries

docs/superpowers/analysis/2026-08-18-r10-t6-global-maximum-adjudication-packet.md
  → corrected, more-specific material dispositions for operator decision
```

When the adjudication packet is more specific than the base candidate, it is the proposed disposition. Neither staging artifact is authority before operator ratification.

## 5. T6 Global-Maximum headline

Candidate architecture deliberately rejects legacy module preservation.

```text
pre-launch /api/v1 is rebuilt from current semantics; no /api/v2 compatibility layer
OpenAPI contract-first + generated Go/TS wire boundaries
Keycloak Authorization Code → MetalDocs ApplicationSession; no local credential API
session-bound CSRF for unsafe browser requests
semantic-lens frontend: Library / My Work / exact Governance case / History / Audit / Admin
one DRAFT ETag/If-Match OCC token covers Revision title + WorkingContent
T4-bound upload_id OPEN→READY→OCC attach; client never owns ExactContentDescriptor
reviewer reads exact immutable Submission; no reviewer WorkingContent editing
semantic byte routes hide provider/storage identity
DOCX provider chosen by representative fidelity gate; no dual provider; no EditorSession baseline
numbering = closed TYPE | TYPE_AREA, fixed '-', minimum 3-digit sequence; no custom grammar
Search materialization/search_refresh = OFF for Launch
Domain history != Audit
RFC9457 errors with one canonical MetalDocs problem code authority
natural HTTP idempotency first; Idempotency-Key only where non-idempotent POST retry could duplicate semantic facts
cursor default20/max100 for unbounded lists
blank/template/revise seeds use exact source semantics; never OfficialRendition as editable source
```

## 6. T6 hard boundaries

T6 may not casually reopen:

```text
Document/Revision/WorkingContent/Submission meaning
Revision-owned title
Release/effectivity
T3 Authorization/Audit authority
T4 exact-content/admission authority
viewer/preview vs OfficialRendition distinction
Search authority boundary + canonical-query baseline
no-notification/event-platform baseline
historical migration/cutover execution
```

## 7. Current gate

```text
operator adjudicates corrected T6 material slate
→ correct only rejected/refined items if needed
→ platform-facing T6 summary
→ explicit operator summary ratification
→ promote durable T6 authority
→ update Decision Registry
→ remove completed T6 staging
→ only then open T7
```

A material-decision approval alone does not open T7.

## 8. Final gate

After T7:

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

**Implementation remains BLOCKED.**
