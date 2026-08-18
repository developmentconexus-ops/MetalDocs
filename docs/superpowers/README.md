# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 + Decision Registry CLOSED / OPERATOR-RATIFIED; T3 decisions accepted / platform summary ratification next.**

Durable accepted truth belongs in `wiki/`. Active, not-yet-promoted design analysis belongs here. Completed/superseded staging is removed from the live tree and remains recoverable from Git history.

## Current durable authority

```text
wiki/architecture/launch-v1-product-contract.md
→ wiki/architecture/whole-product-alignment-review.md
→ wiki/architecture/launch-v1-ownership-topology.md
→ wiki/architecture/r10-t1-semantic-state-invariants.md
→ wiki/architecture/r10-t2-governance-effectivity-transactions.md
→ wiki/architecture/rebaseline-decision-registry.md
→ wiki/architecture/r10-technical-architecture.md
```

## Current active staging

- `analysis/2026-08-18-r10-t3-authorization-audit-enforcement-candidate.md` — **T3 technical candidate; T3-A→T3-P accepted, not yet promoted.**
- `analysis/2026-08-18-r10-t3-operator-adjudication.md` — **operator adjudication record; platform summary ratification next.**

The candidate was rebuilt from the ratified Decision Registry and designs only the official T3 `REOPEN` set.

The completed reconciliation candidate was promoted into `wiki/architecture/rebaseline-decision-registry.md` and removed from live staging. A premature pre-registry T3 candidate was also removed and must not be restored/repaired.

## Revalidation law

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

For every remaining T-stage:

```text
CURRENT / PRESERVE / REFINED → baseline
REOPEN                       → design in owning T-stage
DEFERRED                     → future seam/counterexample only
SUPERSEDED                   → reject inheritance absent explicit material reopen
```

## T3 preserved baseline examples

Do not re-decide:

```text
Group + GroupMembership
RoleAssignment subject = User | Group
scope = Company | Area
static product Role/Permission vocabularies
additive grants + default deny
live direct + Group-mediated grants
provider roles/groups never canonical AuthZ
no role bypasses Controlled Documents governance/lifecycle predicates
offboarding preserves history; re-enable never silently restores old access
Audit != current state
same-local-commit Audit principle for critical governed/security operations
Audit append-only + PII-minimized
no global AuditChainHead/hash-chain Launch requirement
```

## T3 accepted surface

```text
roles = governance_admin | area_manager | author | approver | viewer | governance_viewer
15 Launch permissions
all roles may target User | Group
governance_admin Company-only
area_manager Area-only
author/approver/viewer/governance_viewer Company|Area
organization.manage protects User/Area/Group identity lifecycle
access.manage protects GroupMembership + RoleAssignment changes
responsible-owner / document.owner.manage authoring predicate
governance.act + exact active-Step participation + T2 predicates
offboarding exact teardown and no silent authority restoration
Audit explicit same-local-commit bounded census
Audit Company|Area visibility attribution
no mandatory semantic Audit for ordinary autosave/search/read/download/login/logout/notification/preview/deny
future permissions never silently broaden role bundles
```

## Mandatory T-stage closure protocol

```text
read registry
→ candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promotion/closure
→ update Decision Registry
→ remove completed staging
→ only then Tn+1
```

## Active technical path

```text
T1 Semantic State & Invariants                         CLOSED / OPERATOR-RATIFIED
T2 Governance, Effectivity & Lifecycle Transactions   CLOSED / OPERATOR-RATIFIED
Decision Registry                                      CLOSED / OPERATOR-RATIFIED
T3 Authorization & Audit Enforcement                  DECISIONS ACCEPTED / SUMMARY RATIFICATION PENDING
T4 Exact Content, Storage Integrity & Restore         NOT OPEN
T5 Durable Async, Search & External Effects           NOT OPEN
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN

→ Integrated Whole-R10 GCR
→ cold independent review
→ final operator ratification
→ implementation spec/plan
→ code
```

## Prior evidence

Old R3–R9.5/B1–B6/C material remains evidence only where `wiki/architecture/rebaseline-decision-registry.md` classifies its decisions as CURRENT/PRESERVE/REFINED/REOPEN/Future evidence. The registry's anti-legacy list controls what must not leak back into target architecture.

## Hard stop

No product implementation or implementation plan is authorized while the active design gates remain open.