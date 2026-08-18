# `docs/superpowers` — Active Design Staging Only

> **Status:** Active staging workspace for the MetalDocs rebaselined R10 technical design.  
> **Reset:** 2026-08-14.  
> **Current gate:** **T1 + T2 + Decision Registry CLOSED / OPERATOR-RATIFIED; T3 ACTIVE.**

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

T3 is `Authorization & Audit Enforcement`.

The active T3 candidate must be rebuilt from the Decision Registry and may design only the registry's T3 `REOPEN` set.

The completed reconciliation candidate is promoted into `wiki/architecture/rebaseline-decision-registry.md` and is removed from live staging. A premature pre-registry T3 candidate was also removed and must not be restored/repaired.

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
T3 Authorization & Audit Enforcement                  ACTIVE
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