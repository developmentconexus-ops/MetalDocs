# R10 Technical Architecture — Active Stage Router

> **Status:** ACTIVE — **PRODUCT CONTRACT REV001 + T1→T5 OPERATOR-RATIFIED; POST-T5 FABLE DELTA APPROVED / DISAGREEMENT EMPTY; CHECKPOINT CLOSURE OPERATOR NEXT; T6 NOT OPEN; IMPLEMENTATION BLOCKED**  
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
4. `wiki/architecture/launch-v1-product-contract.md` — **REV001**
5. `wiki/architecture/whole-product-alignment-review.md`
6. `wiki/architecture/launch-v1-ownership-topology.md`
7. `wiki/architecture/r10-t1-semantic-state-invariants.md`
8. `wiki/architecture/r10-t2-governance-effectivity-transactions.md`
9. `wiki/architecture/r10-t3-authorization-audit-enforcement.md`
10. `wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md`
11. `wiki/architecture/r10-t5-durable-async-search-external-effects.md`
12. `wiki/architecture/rebaseline-decision-registry.md`
13. this router
14. active independent-review/adjudication staging only while the checkpoint is open

Historical R3–R9.5 / old R10 / current implementation/schema/OpenAPI are evidence only unless current authority or the Decision Registry preserves a decision.

Review artifacts are evidence, never target authority by themselves.

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
Product Contract                                        REV001 / OPERATOR-APPROVED
T1 — Semantic State & Invariants                       CLOSED / OPERATOR-RATIFIED
T2 — Governance, Effectivity & Lifecycle Transactions CLOSED / OPERATOR-RATIFIED
T3 — Authorization & Audit Enforcement                CLOSED / OPERATOR-RATIFIED
T4 — Exact Content, Storage Integrity & Restore       CLOSED / OPERATOR-RATIFIED
T5 — Durable Async, Search & External Effects         CLOSED / OPERATOR-RATIFIED
Decision Reconciliation Registry                      CURRENT / OPERATOR-RATIFIED
Post-T5 Fable Round-1 amendments                      OPERATOR-RATIFIED / PROMOTED
Fable delta review                                    APPROVE / DISAGREEMENT EMPTY
Author delta adjudication                             ACCEPTED
Post-T5 checkpoint closure                            OPERATOR ACTION NEXT
T6 — Canonical API / Frontend Journeys                NOT OPEN
T7 — Historical Migration & Cutover                   NOT OPEN
implementation                                         BLOCKED
```

T6 may open only after explicit operator closure of the post-T5 checkpoint.

## 5. Closed authorities and promoted post-T5 amendments

```text
T1 → wiki/architecture/r10-t1-semantic-state-invariants.md
T2 → wiki/architecture/r10-t2-governance-effectivity-transactions.md
T3 → wiki/architecture/r10-t3-authorization-audit-enforcement.md
T4 → wiki/architecture/r10-t4-exact-content-storage-integrity-restore.md
T5 → wiki/architecture/r10-t5-durable-async-search-external-effects.md
```

Operator-ratified post-T5 bounded amendments include:

```text
Product/T1
  title = Revision-governed metadata

T2/T3
  bounded withdrawal of active human-governed obsolescence request
  target remains EFFECTIVE; no fake participant verdict

T3
  provider-disable wording aligned to T5-L

T4
  live admission claim protects in-flight READY content from GC
  all restored ApplicationSessions invalidated before ordinary serving
  post-snapshot security teardown must be reconciled/proven before ordinary authenticated serving
  same-DB durable intent + semantic-fact restore coherence recorded

T5
  Search baseline = canonical PostgreSQL query/view over current canonical facts
  materialized Search + search_refresh + rebuild are conditional on a proven derived/expensive/measured consumer
  if materialized Search exists, projection-write serialization begins before canonical read and spans write/removal
  late rendition for dead candidate = semantic no-op/reclaimable output
```

No formal T-stage reopen occurred. Everything not named above remains frozen.

## 6. Decision Registry

Authority:

`wiki/architecture/rebaseline-decision-registry.md`

The registry has been reconciled to the promoted amendments, including:

```text
Search materialization no longer mandatory
T6 owns the proof/activation question for derived Search facts
source upload/T4 admission UX explicitly in T6 REOPEN set
post-snapshot security-teardown recovery choreography explicitly in T7 REOPEN set
ambiguous SUPERSEDED wording tightened
```

## 7. Post-T5 independent Fable checkpoint — CLOSURE PENDING

Independent Fable review:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-independent-fable-review.md`

Original verdict:

```text
APPROVE T1→T5 WITH MATERIAL FIXES
0 BLOCKER / 3 MAJOR / 5 LOW / 3 NOTE
formal minimal reopen set = NONE
```

Operator-ratified Round-1 adjudication:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-author-adjudication-round1.md`

Independent delta review:

`docs/superpowers/analysis/2026-08-18-t1-t5-integrated-fable-delta-review.md`

Delta verdict:

```text
DELTA VERDICT = APPROVE
M1 = CLOSED
M2 = CLOSED
M3 = CLOSED
L1 = CLOSED
L2 = CLOSED
L3 = CLOSED
L4 = CLOSED
L5 = CLOSED
NEW MATERIAL FINDINGS = 0
DISAGREEMENT SET = EMPTY
T6 READINESS = MAY OPEN
```

Author delta adjudication:

`docs/superpowers/analysis/2026-08-18-t1-t5-fable-delta-adjudication.md`

Author disposition:

```text
ACCEPT DELTA VERDICT IN FULL
no remaining material disagreement
no additional bounded amendment required
no T1→T5 reopen required
```

The Fable non-blocking retitle-concurrency observation remains owned by T6/implementation design and does not reopen T1/T2.

## 8. Current gate

```text
T1→T5 + bounded amendments    OPERATOR-RATIFIED
Decision Registry             CURRENT / RECONCILED
Fable delta verdict           APPROVE
Disagreement set              EMPTY
Author delta adjudication     ACCEPTED
Post-T5 checkpoint            OPERATOR CLOSURE NEXT
T6                            NOT OPEN
implementation                BLOCKED
```

Next:

```text
explicit operator closes post-T5 checkpoint
→ remove/archive completed Fable staging from live tree
→ update router/handoff/PR
→ open T6 Canonical API / Frontend Journeys
```

## 9. T6 — Canonical API / Frontend Journeys — NOT OPEN

When the post-T5 checkpoint closes, T6 consumes only the registry's T6 REOPEN set:

```text
numbering configuration grammar and preview UX
admin journeys for current Organization/AuthZ/config
source upload / T4 admission UX
editor/viewer provider behavior
in-product inspection vs exact-source download
public idempotency/error contracts
search/read/history/audit workspaces
exact Search field/ranking UX + proof whether any derived/expensive fact activates materialized Search seam
DRAFT retitle mutation/concurrency placement under an existing T2 law
EditorSession/UX lease only if a real editor-integration consumer requires it
```

T6 must not reopen T1→T5 absent material evidence and explicit bounded reopen.

## 10. T7 — Historical Migration & Cutover

Consumes registry T7 REOPEN set, including concrete restore/erasure **and post-snapshot security-teardown** reconciliation choreography for the selected recovery model.

## 11. Final gate

After T7:

```text
Integrated Whole-R10 GCR
→ cold independent final review
→ operator final ratification
```

Only then may an implementation spec/plan be authored.

**Implementation remains BLOCKED.**
