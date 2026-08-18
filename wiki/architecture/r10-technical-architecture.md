# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1 + T2 CLOSED / OPERATOR-RATIFIED; DECISION REGISTRY CLOSED / OPERATOR-RATIFIED; T3 DECISIONS ADJUDICATED / SUMMARY RATIFICATION NEXT; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — design/documentation only**

This file is the **technical-stage router**. Detailed accepted semantics live in dedicated authorities; this page owns current stage status, reading order and next action.

## 1. Binding authority chain

Read in order:

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/rebaseline-decision-registry.md`
10. active T-stage staging candidate, when present
11. active operator-adjudication/summary gate, when present

Historical R3–R9.5 / old R10-A→C / current implementation/schema/OpenAPI are evidence only unless the registry/current authority preserves a decision.

## 2. Binding method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
```

Future-evolution law:

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

Revalidation law:

> **Revalidation does not mean reinvention. Preserve a prior simple/coherent decision unless current authority or a concrete failure mode disproves it; rederive only the composite decision whose justification changed; defer only the capability that actually left Launch.**

## 3. Active semantic ownership baseline

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Not Launch semantic owners:

```text
storage / staging / byte integrity / malware inspection
render/view/editor providers
Search
async/outbox/jobs/retry/lease/DLQ
notifications
Historical Migration execution machinery
backup/restore transport/readiness
```

Do not resurrect `Artifact`, separate `Approval`, Distribution, Documentary Context, Records Governance or generic Interchange ownership by technical convenience.

## 4. Technical descent

```text
T1 — Semantic State & Invariants                              CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions        CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                             CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                       DECISIONS ACCEPTED / SUMMARY RATIFICATION PENDING
T4 — Exact Content, Storage Integrity & Restore              NOT OPEN
T5 — Durable Async, Search & External Effects                NOT OPEN
T6 — Canonical API / Frontend Journeys                       NOT OPEN
T7 — Historical Migration & Cutover                          NOT OPEN

→ Integrated Whole-R10 Global Coherence Review
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

## 5. Mandatory T-stage closure protocol

For every `Tn`:

```text
read Decision Registry
→ consume CURRENT / PRESERVE / REFINED
→ design only Tn REOPEN set
→ keep DEFERRED as future seam/counterexample
→ reject SUPERSEDED inheritance
→ candidate/design
→ material decision adjudication
→ platform-facing summary
→ explicit operator summary ratification
→ promote durable Tn conclusions
→ update Decision Registry
→ remove completed staging
→ only then open Tn+1
```

A technical A/B/C approval alone never opens the next stage.

## 6. Closed T1

Detailed authority:

`wiki/architecture/r10-t1-semantic-state-invariants.md`

Headline:

```text
Authentication → ProviderSubjectBinding + ApplicationSession
Organization   → Company/User/UserProfile/Area/Group/GroupMembership
Authorization  → product Role/Permission vocabulary + RoleAssignment
Controlled Docs→ DocumentType, Document, Revision, WorkingContent, Submission,
                 bounded governance, cancellation, Release, required Rendition,
                 template origin and native/imported provenance seam
Audit          → AuditEvent
```

Binding revision convention:

```text
REV000 = initial issuance
REV001 = first revision
REV002 = second revision
...
```

## 7. Closed T2

Detailed authority:

`wiki/architecture/r10-t2-governance-effectivity-transactions.md`

Headline:

```text
one local ACID transaction per native business transition
Document = lifecycle serialization root
WorkingContent OCC/CAS
create = Document + REV000 DRAFT + WorkingContent atomically
SUBMIT freezes exact generation + coherent config snapshots
route selector = NAMED_USER | GROUP
GROUP Step = ANY-one activation snapshot
bounded submitter/initiator self-approval prohibition
RETURN/withdraw/cancel preserve immutable history
Release gates = human gate + optional official-rendition gate
replacement Release atomically SUPERSEDES predecessor and EFFECTIVEs successor
obsolescence mutually exclusive with replacement Revision in Launch
READ COMMITTED + narrow explicit serialization/CAS
```

## 8. Ratified Decision Registry

Authority:

`wiki/architecture/rebaseline-decision-registry.md`

The registry classifies prior decisions as:

```text
CURRENT
PRESERVE
REFINED
REOPEN
DEFERRED
SUPERSEDED
```

Later T-stages may not rediscover `CURRENT/PRESERVE/REFINED` decisions from zero and may not inherit `SUPERSEDED` decisions by sunk cost.

### Binding anti-legacy examples

```text
standalone Artifact semantic owner
old 8+3 ownership topology
separate Approval semantic owner
old exact 5×43 permission matrix
universal tenant partition/RLS substrate
REV001 as initial issuance
ROLE_IN_AREA Launch routing
configurable ANY|ALL baseline
strict cross-Step SoD baseline
fresh-auth/SLA/reassign/overseer as Launch defaults
Periodic Review/Distribution/Dossier/Evidence/Records Launch state
generic Interchange owner
global AuditChainHead/hash chain
scheduled Release
universal PDF
Artifact-rooted retention
```

These reopen only on new material evidence.

## 9. T3 — Authorization & Audit Enforcement — DECISIONS ACCEPTED / SUMMARY PENDING

T3 was rebuilt only from the registry's explicit T3 `REOPEN` set. On 2026-08-18 the operator accepted recommendations `T3-A→T3-P` as written.

Active staging:

- `docs/superpowers/analysis/2026-08-18-r10-t3-authorization-audit-enforcement-candidate.md`
- `docs/superpowers/analysis/2026-08-18-r10-t3-operator-adjudication.md`

Accepted headline:

```text
roles = governance_admin | area_manager | author | approver | viewer | governance_viewer
15 Launch permissions only
RoleAssignment subject = User | Group
scope = Company | Area
governance_admin Company-only
area_manager Area-only
author/approver/viewer/governance_viewer Company|Area
organization.manage governs User/Area/Group identity lifecycle
access.manage governs GroupMembership + RoleAssignment mutation
Group deletion fails while live access/governance dependencies remain
authoring = responsible owner OR document.owner.manage
governance action = governance.act + exact active-Step participation + T2 predicates
offboarding atomically disables User + tears down Sessions/memberships/direct grants + required Audit
re-enable never restores prior access
security-sensitive user actions serialize against offboarding eligibility
Audit = explicit semantic append-only same-local-commit evidence for bounded census
AuditEvent carries actor/time/operation/resource + immutable Company|Area visibility attribution + bounded PII-minimized facts
audit.read may be Company- or Area-scoped
ordinary autosave/search/read/download/login/logout/notification/preview/deny are not mandatory semantic Audit in Launch
future capabilities never silently broaden existing role bundles
```

The old exact `5×43` catalog remains forbidden as target inheritance.

### T3 current gate

```text
T3 material decisions ACCEPTED
→ platform-facing T3 summary
→ explicit operator summary ratification
→ promote/close T3
→ update Decision Registry
→ remove completed T3 staging
→ open T4
```

T4 is **NOT OPEN** while summary ratification is pending.

## 10. T4–T7 routing

### T4 — Exact Content, Storage Integrity & Restore

Consumes registry T4 REOPEN set: exact descriptor/digest/canonicalization, provider-neutral managed-content mechanism, provider/profile/conformance, staging/admission/malware, no-overwrite, DRAFT recovery, backup/restore + privacy non-resurrection.

### T5 — Durable Async, Search & External Effects

Consumes registry T5 REOPEN set: which effects need durable intent, worker/retry/DLQ mechanics, renderer execution, named Launch notifications, Search rebuild/freshness/reconciliation and necessary provider receipts.

### T6 — Canonical API / Frontend Journeys

Consumes registry T6 REOPEN set: numbering grammar/preview, admin journeys, editor/viewer provider behavior, in-product inspection vs source download, public idempotency/errors, search/read/history/audit workspaces.

### T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set: actual source evidence, migration modes, imported target-owned shapes, ordinal/content/governance provenance, plan/dry-run/idempotency/reconciliation, semantic-unit atomicity and cutover/readiness/rollback/deletion map.

## 11. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**