# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **T1→T5 CLOSED / OPERATOR-RATIFIED; DECISION REGISTRY CURRENT; POST-T5 FABLE CHECKPOINT ACTIVE; T6 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. active independent-review packet/evidence only when a review checkpoint is open

Historical R3–R9.5 / old R10 / current implementation/schema/OpenAPI are evidence only unless current authority or the Decision Registry preserves a decision.

The active Fable packet is **review evidence/staging**, never target authority.

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
T5 — Durable Async, Search & External Effects                CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                             CURRENT / OPERATOR-RATIFIED
Post-T5 integrated independent Fable checkpoint              ACTIVE / REVIEW REQUEST STAGED
T6 — Canonical API / Frontend Journeys                       NOT OPEN
T7 — Historical Migration & Cutover                          NOT OPEN

→ T6 only after Fable findings are adjudicated and checkpoint explicitly closes
→ T7
→ Integrated Whole-R10 Global Coherence Review
→ cold independent final review
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

An independent checkpoint may deliberately hold the next stage closed **after** the prior stage has been fully promoted; review findings remain evidence until separately adjudicated.

## 6. Closed T-stage authorities

```text
T1 → wiki/architecture/r10-t1-semantic-state-invariants.md
T2 → wiki/architecture/r10-t2-governance-effectivity-transactions.md
T3 → wiki/architecture/r10-t3-authorization-audit-enforcement.md
T4 → wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
T5 → wiki/architecture/r10-t5-durable-async-search-external-effects.md
```

### T4 closed exact-content boundary

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

### T5 closed async/Search boundary

```text
one PostgreSQL-backed durable-job mechanism; River selected/reference mechanism
search_refresh(document_id) = always-required durable projection job
official_rendition_render = conditional only for frozen policy-required OfficialRendition
PDF source direct-view by default
DOCX + SourceOnly direct read-only viewer; no persisted PDF merely for viewing
required job enqueue transaction-coupled to semantic fact
provider/renderer execution outside semantic tx
OfficialRendition finalization revalidates T4/T2/T3 and is idempotent
Search = PostgreSQL rebuildable projection keyed by Document
latest-state refresh makes duplicate/out-of-order jobs converge
Search may lag by omission but never grants stale authority/effectivity
full Search rebuild mandatory; always-on crawler not baseline
GC periodic reconciliation over GC_PENDING with immediate canonical recheck
no mandatory Launch notifications/event bus
no mandatory durable external IdP-disable job
no generic ExternalEffectReceipt
at-least-once/idempotent/revalidating/bounded-retry/terminal-visible jobs with bounded-ID payloads
minimum async operational visibility required
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

T5 is reconciled into the registry. T6's official REOPEN set is now the next design set, but T6 remains held closed by the active independent review checkpoint.

## 8. Post-T5 independent Fable checkpoint — ACTIVE

Review request:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-review-request.md`

Purpose:

> Cold-review the complete ratified T1→T5 architecture for cross-stage contradictions, authority duplication, races, accidental complexity and future-seam defects **before T6 API/frontend journeys encode those decisions into public contracts and UX**.

Required reviewer posture:

```text
repo cold start
Method first
ratified authorities are baseline
current code is evidence only
review packet is not authority
Structural Inversion
minimal reopen only on material counterexample
```

The review explicitly attacks:

```text
DRAFT→SUBMIT→governance→optional Rendition→Release→Search
offboarding vs governance/async work
replacement Release / obsolescence vs stale Search
GC vs WorkingContent/Submission/Rendition/backup
restore vs offboarding/privacy
viewer/preview vs OfficialRendition
Audit + required durable job composition
Search/AuthZ stale-projection boundary
Decision Registry consistency
essential vs accidental complexity
future evolution seams
```

### Current gate

```text
T1→T5 durable authorities          CLOSED / RATIFIED
Decision Registry                  CURRENT
Fable review request               STAGED
Fable independent review           PENDING EXTERNAL REVIEWER
T6                                 NOT OPEN
implementation                     BLOCKED
```

When Fable returns:

```text
independent review evidence
→ adjudicate each material finding against current repo authority + Method
→ bounded correction/reopen only if justified
→ optional Fable delta review if material fixes were required
→ explicit post-T5 checkpoint closure
→ only then open T6
```

The Fable checkpoint does **not** replace the Integrated Whole-R10 GCR or final cold review after T7.

## 9. T6 — Canonical API / Frontend Journeys — NOT OPEN

When the Fable checkpoint closes, T6 consumes only the registry's T6 REOPEN set:

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
EditorSession/UX lease only if a real editor-integration consumer requires it
```

T6 must not reopen T1→T5 absent material evidence and explicit bounded reopen.

## 10. T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set: source evidence, migration modes, imported target-owned shapes, ordinal/content/governance provenance, plan/dry-run/idempotency/reconciliation, semantic-unit atomicity and cutover/readiness/rollback/deletion map.

## 11. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent final review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**
