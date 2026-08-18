# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + T3 + T4 + DECISION REGISTRY OPERATOR-RATIFIED; T5 CORRECTED ADJUDICATION NEXT**  
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
13. `docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` — parent T5 analysis
14. `docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md` — **RV-1→RV-6 ACCEPTED**
15. `docs/superpowers/analysis/2026-08-18-r10-t5-corrected-adjudication-packet.md` — **CURRENT T5-A→T5-P ADJUDICATION SURFACE**
16. `wiki/architecture/launch-v1-scope-rebaseline.md`
17. old R3–R9.5 / R10-B1→B6/C and current implementation only as evidence allowed by the registry

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
RV-1→RV-6                        = OPERATOR-ADJUDICATED / ACCEPTED
T5-A→T5-P corrected              = OPERATOR ADJUDICATION NEXT
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
→ design Tn REOPEN set
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
opaque managed-content handle is mechanism only
OPEN→READY server-verified admission
create-once/no-overwrite
WorkingContent = DRAFT recovery baseline
SUBMIT/Rendition semantic tx makes zero provider/scanner calls
backup/restore exact-content fail-closed readiness
post-snapshot UserProfile erasure reconciliation before restored profile serving
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

## Accepted rendition/viewer correction

```text
PDF source
  → direct PDF viewer
  → no duplicate generated PDF by default

DOCX + SourceOnly
  → direct read-only DOCX viewer
  → no persistent governed PDF merely for viewing

DOCX + RequireOfficialRendition(PDF)
  → conditional durable render from exact Submission
  → T4 admission
  → immutable OfficialRendition
  → Release gate
```

Preview/viewing and OfficialRendition are different meanings. Renderer product remains evidence-driven; architecture does not freeze Gotenberg/ONLYOFFICE yet.

Corrected T5 job census:

```text
always-required durable job:
  search_refresh

conditional durable job:
  official_rendition_render
  only when frozen representation policy requires it

periodic reconciliation:
  managed-content GC over GC_PENDING
```

## Exact next step

Operator adjudication of corrected `T5-A→T5-P` in:

`docs/superpowers/analysis/2026-08-18-r10-t5-corrected-adjudication-packet.md`

After technical adjudication, **do not open T6**. Present the mandatory platform-facing T5 summary and obtain explicit operator ratification first.

No final SQL/index/package/process topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.