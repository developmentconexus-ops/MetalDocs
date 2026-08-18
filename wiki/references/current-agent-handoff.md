# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + T3 + DECISION REGISTRY OPERATOR-RATIFIED; T4 EXACT CONTENT / STORAGE / RESTORE ACTIVE**  
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
12. `docs/superpowers/analysis/2026-08-18-r10-t4-exact-content-storage-integrity-restore-candidate.md` — **ACTIVE NON-AUTHORITATIVE T4 CANDIDATE / OPERATOR ADJUDICATION NEXT**
13. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records defer overlay
14. old R3–R9.5 / R10-B1→B6/C only as evidence allowed by the registry

`wiki/architecture/cohesive-platform-redesign.md` and `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` are historical inventory/evidence, not current authority.

## Current checkpoint

```text
Product Contract                 = ACCEPTED / REV000 INITIAL
Whole-Product GCR A1–A10         = ACCEPTED
Launch ownership topology        = CLOSED / APPROVED / 4+1
T1 Semantic State & Invariants   = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx     = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit         = CLOSED / OPERATOR-RATIFIED
Decision Registry                = CURRENT / OPERATOR-RATIFIED
T4 Exact Content/Storage/Restore = ACTIVE / NON-AUTHORITATIVE CANDIDATE
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

Detailed authority:

`wiki/architecture/r10-t3-authorization-audit-enforcement.md`

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

## T4 baseline that MUST NOT be re-decided

```text
no standalone Artifact semantic owner
exact-content facts live with the semantic record that owns/freezes them
storage/provider identity never becomes semantic identity
WorkingContent = mutable DRAFT authority under T2 OCC
Submission = immutable exact governed attempt
OfficialRendition binds exact Submission
provider calls never join local semantic transaction
Object Lock/WORM/provider versioning never owns lifecycle
restore with missing/corrupt required bytes is not healthy
```

## T4 official REOPEN set

```text
exact content descriptor/digest algorithm/canonicalization
provider-neutral managed-content mechanism
provider choice/profile/conformance
staging/confirmation/admission
malware policy/scan ordering
immutable byte/no-overwrite enforcement
mutable WorkingContent recovery
backup/restore completeness + privacy non-resurrection
```

The active candidate currently recommends T4-A→T4-O. It preserves the useful old R10-C staging/no-overwrite/malware/restore findings while removing Artifact ownership.

## Exact next step

Operator adjudication of T4 recommendations `T4-A→T4-O` in the active candidate.

After technical adjudication, **do not open T5**. Present the mandatory platform-facing T4 summary and obtain explicit operator ratification first.

No final SQL/table/index design, package layout, async topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.
