# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + T3 + DECISION REGISTRY OPERATOR-RATIFIED; T4 DECISIONS ACCEPTED / PLATFORM SUMMARY RATIFICATION NEXT**  
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
10. `wiki/architecture/rebaseline-decision-registry.md`
11. `wiki/architecture/r10-technical-architecture.md` — exact current router
12. `docs/superpowers/analysis/2026-08-18-r10-t4-exact-content-storage-integrity-restore-candidate.md` — **T4 ACCEPTED TECHNICAL CANDIDATE / NOT YET PROMOTED**
13. `docs/superpowers/analysis/2026-08-18-r10-t4-operator-adjudication.md` — **T4-A→T4-O ACCEPTED / PLATFORM SUMMARY GATE**
14. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records defer overlay
15. old R3–R9.5 / R10-B1→B6/C only as evidence allowed by the registry

Historical design files are evidence/provenance only and never override the authority chain above.

## Current checkpoint

```text
Product Contract                 = ACCEPTED / REV000 INITIAL
Whole-Product GCR A1–A10         = ACCEPTED
Launch ownership topology        = CLOSED / APPROVED / 4+1
T1 Semantic State & Invariants   = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx     = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit         = CLOSED / OPERATOR-RATIFIED
Decision Registry                = CURRENT / OPERATOR-RATIFIED
T4 decisions A→O                 = OPERATOR-ADJUDICATED / ACCEPTED
T4 platform summary              = OPERATOR RATIFICATION NEXT
T4 promotion/closure             = PENDING SUMMARY RATIFICATION
T5→T7                            = NOT OPEN
implementation                   = BLOCKED
```

## Revision convention

```text
REV000 = initial issuance
REV001 = first revision
REV002 = second revision
...
```

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

Every remaining T-stage begins from the Decision Registry:

```text
CURRENT / PRESERVE / REFINED → baseline
REOPEN                       → design deliberately in owning T-stage
DEFERRED                     → future seam/counterexample only
SUPERSEDED                   → forbidden inheritance absent new material reopen
```

## Mandatory T-stage closure protocol

```text
read registry
→ design Tn REOPEN set
→ operator adjudication
→ platform-facing Tn summary
→ explicit operator summary ratification
→ promote/close Tn
→ update registry
→ remove completed staging
→ only then Tn+1
```

## Closed T3 headline

Detailed authority: `wiki/architecture/r10-t3-authorization-audit-enforcement.md`.

```text
six Launch roles / 15 Launch permissions
RoleAssignment subject = User | Group
accepted Company|Area scope matrix
organization.manage vs access.manage administration split
responsible-owner / document.owner.manage authoring predicate
governance.act + exact active-Step participation
atomic offboarding + no silent access resurrection
security-action/offboarding User-eligibility serialization
same-local-commit Audit census + PII-minimized facts
Company|Area historical Audit visibility
future features never silently broaden existing role bundles
```

## Accepted T4 headline

```text
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
whole-Submission JCS/composite digest deferred absent named consumer
managed_content_id = opaque retrieval mechanism only
one provider-neutral ManagedContentStore / one active store per deployment
Local dev/test/conformance + AWS S3 reference production profile
OPEN→READY admission; server derives hash/size/format from exact bytes
opaque admission binding blocks arbitrary/cross-root handle reuse
production UNTRUSTED_EXTERNAL bytes require CLEAN before governed admission
ClamAV/clamd reference scanner; no malware scan on every autosave
create-once/no-overwrite; DRAFT replacement creates new handle
WorkingContent current state is DRAFT recovery baseline; no mandatory WorkingSnapshot history
SUBMIT/Rendition tx performs zero provider/scanner calls and freezes exact READY handle+descriptor
only unreferenced/non-governed DRAFT mechanism objects are reclaimable in Launch
backup couples DB recovery point + exact required-content manifest/copy + GC fence
restore remains non-serving until every required handle matches size/SHA-256/format
older restore reconciles later lawful UserProfile erasures via minimum independent barrier/journal
future Evidence/Records/Export/Repository reuse descriptor+mechanism without Artifact owner
```

The product rationale accepted by the operator is that T4 is required to prove four Launch promises:

```text
what exact content was approved
what exact content is officially effective
that governed history did not silently change
that backup/restore truly recovers semantic state + required exact content
```

## Exact next step

Present the mandatory **platform-facing T4 summary** and obtain explicit operator ratification.

Only after summary ratification:

```text
promote/close T4
→ update Decision Registry
→ remove completed T4 staging
→ open T5 Durable Async, Search & External Effects
```

Do not open T5 or write final SQL/table/index design, package layout, async topology, public API/frontend contract, migration execution plan, implementation plan or product code.
