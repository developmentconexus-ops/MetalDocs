# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + DECISION REGISTRY OPERATOR-RATIFIED; T3 DECISIONS ACCEPTED / PLATFORM SUMMARY RATIFICATION NEXT**  
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
9. `wiki/architecture/rebaseline-decision-registry.md` — **RATIFIED PRIOR-DECISION DISPOSITION BASELINE**
10. `wiki/architecture/r10-technical-architecture.md` — active technical-stage router
11. `docs/superpowers/analysis/2026-08-18-r10-t3-authorization-audit-enforcement-candidate.md` — **T3 ACCEPTED TECHNICAL CANDIDATE / NOT YET PROMOTED**
12. `docs/superpowers/analysis/2026-08-18-r10-t3-operator-adjudication.md` — **T3-A→T3-P ACCEPTED / SUMMARY RATIFICATION GATE**
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
Decision Registry                = CLOSED / OPERATOR-RATIFIED
T3 decisions A→P                 = OPERATOR-ADJUDICATED / ACCEPTED
T3 platform summary              = NEXT / EXPLICIT OPERATOR RATIFICATION REQUIRED
T3 final promotion/closure       = PENDING SUMMARY RATIFICATION
T4→T7                            = NOT OPEN
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

## T3 preserved baseline — already decided

```text
Group + GroupMembership
RoleAssignment subject = User | Group
scope = Company | Area
static product-owned Role/Permission vocabularies
additive grants + default deny
live direct + Group-mediated grants
provider roles/groups/claims never canonical AuthZ
Session is not durable Role/Permission authority
Controlled Documents owns relationship/lifecycle/governance predicates
no role bypasses domain governance
Area Manager remains a preserved operational role concept
offboarding preserves historical identity and re-enable never silently restores old grants/memberships/sessions
Audit != current state
same-local-commit Audit principle for critical governed/security mutations
Audit append-only + PII-minimized
no global AuditChainHead/hash-chain Launch requirement
```

## T3 accepted headline

```text
roles:
  governance_admin
  area_manager
  author
  approver
  viewer
  governance_viewer

15 Launch permissions
all roles assignable to User | Group
governance_admin = CompanyScope only
area_manager = AreaScope only
author/approver/viewer/governance_viewer = CompanyScope | AreaScope
organization.manage = User/Area/Group identity lifecycle
access.manage = GroupMembership + RoleAssignment changes
Group deletion fails while live access/governance dependencies remain
author can work on Documents where actor is current responsible owner
actor with document.owner.manage can manage Documents in scope
governance.act requires exact active-Step participation + T2 predicates
offboarding atomically disables User + revokes Sessions + removes memberships/direct grants + required Audit
re-enable restores no old access
security-sensitive User actions serialize against offboarding eligibility
Audit = explicit semantic append-only same-local-commit evidence for bounded critical census
AuditEvent = actor + trusted time + operation/resource + Company|Area visibility attribution + bounded PII-minimized facts
audit.read may be Company- or Area-scoped
ordinary autosave/search/read/download/login/logout/notification/preview/deny are not mandatory semantic Audit in Launch
future capabilities never silently broaden existing role bundles
```

The old exact `5×43` catalog remains **SUPERSEDED** and may not be repaired/subtracted into the target.

## Exact next step

**Present the platform-facing T3 summary and obtain explicit operator summary ratification.**

Only after that:

```text
promote/close T3
→ update Decision Registry
→ remove completed T3 staging
→ open T4 Exact Content, Storage Integrity & Restore
```

Until T3 closes, do not write final SQL/table/index design, package layout, storage locator design, async topology, public API/frontend contract, migration execution plan, implementation plan or product code.