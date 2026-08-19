# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; T6 MATERIAL DECISIONS OPERATOR-APPROVED; PLATFORM-FACING SUMMARY RATIFICATION NEXT; T7 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
14. T6 material-adjudication record
15. active T6 platform-facing summary
16. T6 candidate/evidence staging only for provenance while T6 remains open
17. current implementation only as claim-specific evidence

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
T6 material decisions                                 OPERATOR-APPROVED
T6 platform-facing summary                            STAGED / OPERATOR RATIFICATION NEXT
T6 durable authority                                  NOT YET
T7 — Historical Migration & Cutover                   NOT OPEN
implementation                                         BLOCKED
```

## 4. T6 operator-approved material decision record

`docs/superpowers/analysis/2026-08-18-r10-t6-operator-material-adjudication.md`

The operator approved the final T6 proposal in this precedence order:

```text
base candidate
→ corrected Global-Maximum adjudication packet
→ final refinements FR-1..FR-4 where named
```

The approved direction includes:

```text
rebuild pre-launch /api/v1 from current product semantics; no compatibility layer
OpenAPI contract-first + generated Go/TS boundary; OAS 3.0.3 Launch baseline
Keycloak Authorization Code → MetalDocs ApplicationSession + session-bound CSRF
semantic-lens frontend with stable route meanings
one DRAFT generation exposed as strong ETag/If-Match; PATCH title/source under same T2 OCC
T4-bound upload_id OPEN→READY→OCC attachment; client never owns ExactContentDescriptor
review exact immutable Submission; case participation never mutates WorkingContent
User eligibility = singleton PUT current resource; DISABLED executes T3 offboarding semantics
Governance Step Decision = singleton immutable PUT resource
semantic exact-byte routes hide provider/storage identity
one fidelity-gated DOCX provider; no EditorSession correctness dependency baseline
closed TYPE | TYPE_AREA numbering; no generic formatting grammar
Search materialization/search_refresh OFF for Launch
Domain history != Audit
RFC9457 errors with one canonical MetalDocs problem code authority
natural HTTP idempotency first; durable Idempotency-Key only for truly non-idempotent POST creation
opaque cursor pagination for unbounded lists
blank/template/revise seeds use exact governed source, never OfficialRendition as editable source
```

Everything outside the approved T6 slate remains frozen unless material evidence triggers an explicit bounded reopen.

## 5. Platform-facing T6 summary — ACTIVE GATE

`docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md`

The summary consolidates the operator-approved material decisions into one implementation-facing system description covering:

```text
platform semantic lenses
public contract law
AuthN/session/CSRF
frontend information architecture
Library/Search
create/numbering/seeding
DRAFT OCC
T4 upload/admission
Submission/governance
lifecycle/idempotency transport
Release/source/OfficialRendition presentation
DOCX provider proof gate
Administration
errors
pagination/read models
History vs Audit
frontend technical organization
explicit subtraction from legacy Launch target
implementation-proof obligations
```

It is **not durable authority until explicitly ratified by the operator**.

## 6. T6 hard boundaries

T6 does not reopen by implication:

```text
Document/Revision/WorkingContent/Submission meaning
Revision-owned title
Release/effectivity
T3 Authorization/Audit authority
T4 exact-content/admission authority
viewer/preview vs OfficialRendition distinction
Search authority boundary + canonical-query baseline
no-notification/event-platform baseline
Historical Migration/Cutover execution
```

## 7. Current gate

```text
operator reviews platform-facing T6 summary
→ explicit operator summary ratification
→ promote durable T6 authority to wiki/
→ reconcile Decision Registry
→ update router/handoff/index/PR
→ remove completed T6 staging from live tree (Git history archive)
→ only then open T7
```

The material decision approval already received does **not** itself promote T6 or open T7.

## 8. Final gate after T7

```text
Integrated Whole-R10 Global Coherence Review
→ cold independent final review
→ operator final ratification
→ implementation spec/plan
→ code
```

**Implementation remains BLOCKED.**