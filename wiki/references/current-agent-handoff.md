# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + T3 + T4 + DECISION REGISTRY OPERATOR-RATIFIED; T5 ACTIVE / RENDITION-VIEWER SUBGATE**  
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
13. `docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md` — **PARENT T5 CANDIDATE / CORRECTED PACKET PAUSED**
14. `docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md` — **ACTIVE MATERIAL SUBGATE / RV-1→RV-6 OPERATOR DECISION NEXT**
15. `wiki/architecture/launch-v1-scope-rebaseline.md`
16. old R3–R9.5 / R10-B1→B6/C and current implementation only as evidence allowed by the registry

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
T5 Durable Async/Search/Effects  = ACTIVE / RENDITION-VIEWER SUBGATE
T5-A→T5-P whole adjudication     = PAUSED UNTIL RV-1→RV-6
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

## Active T5 rendition/viewer subgate

The operator challenged the assumption that DOCX viewing requires a persistent `OfficialRendition` PDF.

Current recommended hybrid:

```text
PDF source
  → direct PDF viewer
  → no duplicate PDF by default

DOCX + SourceOnly
  → direct read-only DOCX viewer
  → no persistent governed PDF merely for viewing

DOCX + RequireOfficialRendition(PDF)
  → conditional durable render from exact Submission
  → T4 admission
  → immutable OfficialRendition
  → Release gate
```

Preview/viewing PDF and OfficialRendition PDF are different meanings. A preview/cache may be rebuildable mechanism; it is not semantic truth or Release gate.

Current corrected T5 job census candidate:

```text
always-required durable job:
  search_refresh

conditional durable job:
  official_rendition_render
  only for frozen RequireOfficialRendition policy

periodic reconciliation:
  managed-content GC over GC_PENDING
```

Renderer product is not frozen. EigenPal is the lowest-cost native DOCX viewer candidate; ONLYOFFICE is a stronger self-hosted viewer/converter candidate; Gotenberg/LibreOffice is a simple server-side PDF converter candidate. Selection must be proven through a representative DOCX fidelity corpus.

## Exact next step

Operator adjudication of `RV-1→RV-6` in:

`docs/superpowers/analysis/2026-08-18-t5-rendition-viewer-strategy-evaluation.md`

Only after that:

```text
refine/confirm T5-D/T5-E + durable-job census
→ adjudicate corrected T5-A→T5-P
→ platform-facing T5 summary
→ explicit operator ratification
```

Do **not** open T6. No final SQL/index/package/process topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.