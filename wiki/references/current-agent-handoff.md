# Current Agent Handoff

> **Last verified:** 2026-08-18  
> **Status:** ACTIVE — **T1 + T2 + DECISION REGISTRY OPERATOR-RATIFIED; T3 AUTHORIZATION + AUDIT ACTIVE**  
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
11. `docs/superpowers/analysis/2026-08-18-r10-t3-authorization-audit-enforcement-candidate.md` — **ACTIVE NON-AUTHORITATIVE T3 CANDIDATE / OPERATOR ADJUDICATION NEXT**
12. `wiki/architecture/launch-v1-scope-rebaseline.md` — narrow Records defer overlay
13. old R3–R9.5 / R10-B1→B6/C only as evidence allowed by the registry

`wiki/architecture/cohesive-platform-redesign.md` and `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` are historical inventory/evidence, not current authority.

## Current checkpoint

```text
Product Contract                 = ACCEPTED / REV000 INITIAL
Whole-Product GCR A1–A10         = ACCEPTED
Launch ownership topology        = CLOSED / APPROVED / 4+1
T1 Semantic State & Invariants   = CLOSED / OPERATOR-RATIFIED
T2 Governance/Effectivity/Tx     = CLOSED / OPERATOR-RATIFIED
Decision Registry                = CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit         = ACTIVE / NON-AUTHORITATIVE DESIGN CANDIDATE
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

## T3 baseline that MUST NOT be re-decided

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
Area Manager remains a preserved role concept, exact bundle open
offboarding preserves historical identity and re-enable never silently restores old grants/memberships/sessions
Audit != current state
same-local-commit Audit principle for critical governed/security mutations
Audit append-only + PII-minimized
no global AuditChainHead/hash-chain Launch requirement
```

## T3 official REOPEN set

```text
exact Launch role vocabulary
exact permission vocabulary/bundles
whether/how area_manager survives
whole-company admin role naming/bundle
role↔scope matrix
access administration law for GroupMembership/RoleAssignment
Group administration/deletion exact law
offboarding exact access teardown transaction
least-privilege Governance Viewer/Auditor
canonical check sites
authorization-sensitive in-flight/offboarding races where material
same-local-commit Audit operation census + minimum bounded facts
Audit read visibility/scoping
```

The old exact `5×43` catalog is **SUPERSEDED** and may not be repaired/subtracted into the target.

## Exact next step

Operator adjudication of T3 recommendations `T3-A→T3-P` in the active candidate.

After technical adjudication, **do not open T4**. Present the mandatory platform-facing T3 summary and obtain explicit operator ratification first.

No final SQL/table/index design, package layout, storage locator design, async topology, public API/frontend contract, migration execution plan, implementation plan or product code is authorized.