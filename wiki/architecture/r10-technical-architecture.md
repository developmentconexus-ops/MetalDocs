# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1 + T2 + T3 + T4 CLOSED / OPERATOR-RATIFIED; DECISION REGISTRY CURRENT; T5 ACTIVE; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — design/documentation only**

This file is the technical-stage router. Detailed accepted semantics live in dedicated authorities; this page owns current stage status, reading order and exact next action.

## 1. Binding authority chain

1. `AGENTS.md`
2. `docs/engineering/standards/root-cause-global-maximum-method.md`
3. `wiki/references/current-agent-handoff.md`
4. `wiki/architecture/launch-v1-product-contract.md`
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/rebaseline-decision-registry.md`
12. active T-stage staging candidate, when present
13. active operator-adjudication/summary gate, when present

Historical R3–R9.5 / old R10 / current implementation/schema/OpenAPI are evidence only unless current authority or the Decision Registry preserves a decision.

## 2. Binding method laws

```text
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
```

> **Defer the capability; preserve the evolution seam. Prepare the seam, not the dormant implementation.**

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
managed content / storage / malware inspection
render/view/editor providers
Search
async/jobs/retry/leases
notifications
Historical Migration execution
backup/restore transport/readiness
```

Do not resurrect `Artifact`, separate `Approval`, Distribution, Documentary Context, Records Governance or generic Interchange ownership by technical convenience.

## 4. Technical descent

```text
T1 — Semantic State & Invariants                              CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions        CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                       CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore              CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                             CURRENT / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects                ACTIVE / DESIGN
T6 — Canonical API / Frontend Journeys                       NOT OPEN
T7 — Historical Migration & Cutover                          NOT OPEN

→ Integrated Whole-R10 Global Coherence Review
→ cold independent review
→ operator final ratification
→ implementation spec/plan
→ code
```

## 5. Mandatory T-stage closure protocol

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

A technical decision approval alone never opens the next stage.

## 6. Closed T-stage authorities

```text
T1 → wiki/architecture/r10-t1-semantic-state-invariants.md
T2 → wiki/architecture/r10-t2-governance-effectivity-transactions.md
T3 → wiki/architecture/r10-t3-authorization-audit-enforcement.md
T4 → wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
```

T4 closed the exact-content boundary:

```text
ExactContentDescriptor = SHA-256 + size + ContentFormat
opaque managed-content handle = retrieval mechanism only
one provider-neutral ManagedContentStore / one active store
OPEN→READY server-verified admission
untrusted-content malware gate at governed boundary
create-once/no-overwrite
WorkingContent = DRAFT recovery baseline
zero provider/scanner calls inside SUBMIT/Rendition semantic tx
only non-governed unreferenced mechanism content reclaimable
backup = DB recovery point + exact required content
restore fails closed on content mismatch and post-snapshot profile-erasure resurrection
```

## 7. Decision Registry

Authority:

`wiki/architecture/rebaseline-decision-registry.md`

Registry vocabulary:

```text
CURRENT
PRESERVE
REFINED
REOPEN
DEFERRED
SUPERSEDED
```

Later stages may not rediscover settled decisions from zero and may not inherit superseded decisions by sunk cost.

## 8. T5 — Durable Async, Search & External Effects — ACTIVE

Active candidate:

`docs/superpowers/analysis/2026-08-18-r10-t5-durable-async-search-external-effects-candidate.md`

Official T5 REOPEN set:

```text
which effects actually require durable intent/outbox
worker/lease/retry/DLQ mechanism
renderer execution
notifications if a Launch consumer remains
Search projection/rebuild/freshness/reconciliation
provider effect receipts where needed
```

T5 MUST consume:

```text
Search = rebuildable eventual-consistency projection
Search never grants access/effectivity
required external work may need transaction-coupled durable enqueue
provider calls never join semantic tx
OfficialRendition is a real optional Release gate
T4 GC_PENDING is technical eligibility
current River/custom outbox implementations are evidence only
```

### Current gate

```text
T5 candidate
→ operator adjudication of T5-A→T5-P
→ platform-facing T5 summary
→ explicit operator summary ratification
→ promote/close T5
→ update registry
→ remove staging
→ open T6
```

T5 does not own public API/frontend journeys or Historical Migration execution.

## 9. T6–T7 routing

### T6 — Canonical API / Frontend Journeys

Consumes registry T6 REOPEN set: numbering grammar/preview, admin journeys, editor/viewer behavior, in-product inspection vs source download, public idempotency/errors, search/read/history/audit workspaces and any bounded EditorSession UX seam actually required.

### T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set: source evidence, migration modes, imported target-owned shapes, ordinal/content/governance provenance, plan/dry-run/idempotency/reconciliation, semantic-unit atomicity and cutover/readiness/rollback/deletion map.

## 10. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**