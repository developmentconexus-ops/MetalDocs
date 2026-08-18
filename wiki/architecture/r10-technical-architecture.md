# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1 + T2 + T3 CLOSED / OPERATOR-RATIFIED; DECISION REGISTRY CURRENT; T4 DECISIONS ADJUDICATED / SUMMARY RATIFICATION NEXT; IMPLEMENTATION BLOCKED**  
> **Rebaselined:** 2026-08-18  
> **Revision convention:** `REV000` initial issuance / `REV001` first revision  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation gate:** **CLOSED — design/documentation only**

This file is the technical-stage router. Detailed accepted semantics live in dedicated authorities; this page owns current stage status, reading order and exact next action.

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
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/rebaseline-decision-registry.md`
11. active T-stage staging candidate, when present
12. active operator-adjudication/summary gate, when present

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
T3 — Authorization & Audit Enforcement                       CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                             CURRENT / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore              DECISIONS ACCEPTED / SUMMARY RATIFICATION PENDING
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

A technical decision approval alone never opens the next stage.

## 6. Closed T-stage authorities

```text
T1 → wiki/architecture/r10-t1-semantic-state-invariants.md
T2 → wiki/architecture/r10-t2-governance-effectivity-transactions.md
T3 → wiki/architecture/r10-t3-authorization-audit-enforcement.md
```

Binding revision convention:

```text
REV000 = initial issuance
REV001 = first revision
REV002 = second revision
...
```

T3 established the Launch role/permission surface, User|Group + Company|Area scope matrix, author/Area-manager/governance predicates, atomic offboarding, same-local-commit Audit census and Company|Area Audit visibility. No role is a domain-governance bypass.

## 7. Ratified Decision Registry

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

Later stages may not rediscover CURRENT/PRESERVE/REFINED decisions from zero and may not inherit SUPERSEDED decisions by sunk cost.

## 8. T4 — Exact Content, Storage Integrity & Restore — DECISIONS ACCEPTED

T4 consumes only the registry's official T4 REOPEN set:

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

T4 MUST consume as baseline:

```text
no standalone Artifact semantic owner
exact-content facts belong to the semantic record that owns/freezes them
storage/provider identity never becomes semantic identity
WorkingContent is mutable DRAFT authority under T2 OCC
Submission is immutable exact governed attempt
OfficialRendition binds exact Submission
native/imported truth remains distinguishable
provider calls never join semantic local transaction
Object Lock/WORM/provider versioning never becomes document/records lifecycle authority
```

Operator-adjudicated T4 headline:

```text
ExactContentDescriptor = SHA-256 + exact size + ContentFormat
no mandatory whole-Submission JCS digest in Launch
opaque managed-content UUID handle = retrieval mechanism only
one provider-neutral ManagedContentStore / one active store per deployment
Local dev/test/conformance + AWS S3 reference production profile
OPEN→READY admission with server-derived hash/size/format
opaque admission binding prevents arbitrary/cross-root handle reuse
UNTRUSTED_EXTERNAL requires CLEAN malware proof before governed admission
ClamAV/clamd reference MalwareInspector; no scan on every autosave
create-once/no-overwrite; replacement = new handle
WorkingContent current state is DRAFT recovery baseline; no mandatory WorkingSnapshot business history
SUBMIT/Rendition tx performs no provider/scanner calls and freezes exact READY handle+descriptor
only unreferenced/non-governed DRAFT mechanism objects are reclaimable in Launch
backup = DB recovery point + exact required-content manifest/copy + GC fence
restore remains non-serving until required content matches size/SHA-256/format
older restore must reconcile later lawful UserProfile erasures via minimum independent barrier/journal
future content capabilities reuse descriptor+mechanism without restoring Artifact ownership
```

### T4 current gate

```text
T4 candidate/design          = COMPLETE FOR ADJUDICATION
T4-A→T4-O                    = OPERATOR-ADJUDICATED / ACCEPTED
T4 platform-facing summary   = NEXT
T4 promotion/closure         = PENDING SUMMARY RATIFICATION
T5                           = NOT OPEN
implementation               = BLOCKED
```

Next:

```text
present platform-facing T4 summary
→ explicit operator summary ratification
→ promote/close T4
→ update Decision Registry
→ remove completed T4 staging
→ only then open T5
```

## 9. T5–T7 routing

### T5 — Durable Async, Search & External Effects

Consumes registry T5 REOPEN set: real durable intents, worker/retry/DLQ mechanics, renderer execution, named Launch notifications, Search rebuild/freshness/reconciliation and required provider receipts.

### T6 — Canonical API / Frontend Journeys

Consumes registry T6 REOPEN set: numbering grammar/preview, admin journeys, editor/viewer provider behavior, in-product inspection vs source download, public idempotency/errors, search/read/history/audit workspaces.

### T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set: actual source evidence, migration modes, imported target-owned shapes, ordinal/content/governance provenance, plan/dry-run/idempotency/reconciliation, semantic-unit atomicity and cutover/readiness/rollback/deletion map.

## 10. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**
