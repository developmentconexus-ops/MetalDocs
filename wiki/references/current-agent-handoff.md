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
12. active T4 staging candidate
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
T4 Exact Content/Storage/Restore = ACTIVE / DESIGN
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
roles:
  governance_admin
  area_manager
  author
  approver
  viewer
  governance_viewer

15 Launch permissions
RoleAssignment subject = User | Group
governance_admin = CompanyScope only
area_manager = AreaScope only
author/approver/viewer/governance_viewer = CompanyScope | AreaScope
organization.manage = User/Area/Group identity lifecycle
access.manage = GroupMembership + RoleAssignment mutations
ordinary Author = current responsible owner unless document.owner.manage
Area Manager manages Area work through document.owner.manage
governance.act requires exact active-Step participation + T2 predicates
offboarding atomically disables User + revokes Sessions + removes memberships/direct grants + Audit
re-enable restores no prior authority
security-sensitive actions serialize against offboarding eligibility
Audit = explicit same-local-commit semantic evidence for bounded critical census
AuditEvent = actor/time/operation/resource + Company|Area attribution + bounded PII-minimized facts
audit.read may be Company- or Area-scoped
ordinary read/download/search/autosave/login/logout/preview/deny are not mandatory semantic Audit in Launch
future capabilities never silently broaden existing role bundles
```

The old exact 5×43 catalog remains superseded.

## T4 preserved baseline — do not re-decide

```text
no standalone Artifact semantic owner
exact-content facts live with the semantic record that owns/freezes them
storage/provider identity never becomes semantic identity
WorkingContent = mutable DRAFT authority protected by T2 OCC
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

## Exact next step

Review/adjudicate the active T4 candidate derived only from the official REOPEN set.

After T4 technical adjudication, **do not open T5**. Present the mandatory platform-facing T4 summary and obtain explicit operator ratification first.

No final SQL/table/index design, package layout, async topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.
