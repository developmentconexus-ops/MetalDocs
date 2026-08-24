# T11 — B09-F1 Authority Promotion & Rebaseline Plan

> **For agentic workers:** follow this plan task-by-task with fresh verification after every write.  
> **Status:** EXECUTION PLAN / R2 OPERATOR-RATIFIED / PRODUCT IMPLEMENTATION BLOCKED.  
> **Goal:** promote the operator-ratified B09-F1 Audit investigation decision into bounded durable authority, move the sole API census from 86 to 89, prove the bounded rebaseline, close B09-F1, and resume B09 P7 without starting P8 or Product implementation.  
> **Architecture:** use the repository's established bounded-authority pattern: one durable decision supersedes only conflicting current-tense clauses in larger ratified Product/T8 authorities; the large base authorities are not rewritten wholesale. The numeric census remains separately owned by `docs/decisions/api-operation-census.md`, while roadmap/work ledgers own mutable stage and proof state.  
> **Tech Stack:** Markdown authority/work records + GitHub repository contents API + GitHub Actions Repository Standard verifier.  
> **Spec:** `docs/work/current/t11-b09-audit-capability-decision-candidate-r2.md`.

## Global constraints

```text
Repository current authority > handoff/history.
main must remain cae6ba48df5d611959c0390e0f2b9b8194d62a9d unless fresh revalidation proves otherwise.
Draft PR #162 remains the candidate container; no merge authorization is implied.
Frontend Method v2.3 is OPERATOR-RATIFIED.
No Product code/schema/OpenAPI/runtime/deploy implementation.
No T12 work.
No B09 P8 until resumed P7 exits with no unresolved upstream finding.
No B10-B12 design.
No screen-shaped backend.
No backend-shaped UX.
No generic Audit search/entity/reference-data/deep-link platform.
Preserve Audit evidence/current-state and Audit/Document-History separation.
Preserve B01-B08 LOCK unless new material evidence falsifies a locked decision.
Do not weaken the repository verifier to obtain green CI.
Avoid no-op file updates.
```

---

### Task 1 — Promote one bounded durable Audit investigation authority

**Files:**
- Create: `docs/decisions/audit-investigation-read.md`
- Modify: `docs/decisions/index.md`

**Consumes:**
- Ratified spec: `docs/work/current/t11-b09-audit-capability-decision-candidate-r2.md`
- Existing bounded-authority pattern: `docs/decisions/discussion-notifications-launch.md`

**Produces:**
- One current durable authority for B09-F1.
- Decision-register routing that makes the bounded supersession reachable from Product/T6, T8-C, T8-E, T8-F, T9 and T11.

- [ ] **Step 1: Revalidate candidate and decision-register blobs on the exact PR HEAD.**

Fetch both paths from `arch/t11-implementation-program` and confirm R2 still says `WRITTEN RATIFICATION PENDING` and the register still reports 86 operations.

- [ ] **Step 2: Create `docs/decisions/audit-investigation-read.md` from R2 without weakening semantics.**

The durable header must state:

```text
Status: OPERATOR-RATIFIED / BOUNDED T11 REOPEN
Ratified: 2026-08-24
Implementation: BLOCKED
```

The durable decision must preserve exactly:

```text
op78 listAuditEvents refined, not replaced
op87 listAuditQueryAreas
op88 searchAuditQueryActors
op89 searchAuditQueryResources
operations 86 -> 89
routes 11 unchanged
PermissionCode 16 unchanged
Idempotency-Key creations 11 unchanged
ETag 13/13 unchanged
exact-byte 4 unchanged
owners 4+2 unchanged
new writes 0
```

It must include the R2 structured-query, cursor, evidence/recognition, Query Assist, owner-handoff, YAGNI and supersession laws. It must explicitly say that it supersedes only conflicting current-tense clauses in `docs/product/journeys.md`, `docs/architecture/interfaces.md`, `docs/architecture/wire-contract.md`, `docs/architecture/frontend.md`, and the prior numeric census.

- [ ] **Step 3: Update `docs/decisions/index.md` to route current authority.**

Required register changes:

```text
T6       add bounded Audit investigation authority to current canonical journeys
T6-API   current census = 89; operation 90+ requires new lawful basis
T8-C     add structured Audit read + bounded Query Assist composition
T8-E     current wire = 89; op78 refined + op87-op89
T8-F     frontend coverage = 89; Audit is structured investigation, not paging-only
T8-H/T9  prior closure preserved for unchanged areas; B09 bounded proof is current T11 delta
T11-AUDIT add a dedicated current row pointing to audit-investigation-read.md
```

Do not change unrelated decisions.

- [ ] **Step 4: Verify the durable decision is reachable from `docs/index.md` through `docs/decisions/index.md`.**

The new durable file must not create a second mutable roadmap and must not link from durable docs into `docs/work/**`.

- [ ] **Step 5: Commit through the repository contents API and wait for exact-head CI.**

Expected: Repository Standard verifier SUCCESS without verifier changes.

---

### Task 2 — Promote the sole numeric API census from 86 to 89

**Files:**
- Modify: `docs/decisions/api-operation-census.md`

**Consumes:**
- `docs/decisions/audit-investigation-read.md`
- Existing census arithmetic: original 76 + read symmetry 2 + Discussion/Notifications 8 = 86.

**Produces:**
- Sole numeric census authority at 89 operations.

- [ ] **Step 1: Re-fetch the current census blob after Task 1.**

Confirm it still owns `86 / 11 / 13-13 / 4` before replacement.

- [ ] **Step 2: Add exactly three safe reads.**

The census delta must be:

```text
87 GET /api/v1/audit/query-areas
   listAuditQueryAreas

88 GET /api/v1/audit/query-actors
   searchAuditQueryActors

89 GET /api/v1/audit/query-resources
   searchAuditQueryResources
```

`listAuditEvents` remains existing operation 78 and is marked refined by the bounded decision; it is not counted again.

- [ ] **Step 3: Replace count proof with exact arithmetic.**

```text
original journeys census                    76
bounded read-symmetry precision             +2
T11 Discussion/Mention/Notifications reopen +8
T11 Audit investigation bounded reopen      +3
                                            ---
current application census                  89
```

Supporting counts remain:

```text
Idempotency-Key creations     11
ETag read / mutation domains  13 / 13
exact-byte resources          4
```

- [ ] **Step 4: Update the census supersession law.**

Any older `78`, `86`, `operation 87 absent`, or equivalent closure statement in pre-B09 authorities is historical/current-tense superseded only where `audit-investigation-read.md` says so. Operation 90+ requires semantic normalization already authorized or a new bounded Product/T6 reopen.

- [ ] **Step 5: Commit and verify exact-head CI SUCCESS.**

---

### Task 3 — Prove the bounded Product/T8/FP0 rebaseline and close B09-F1

**Files:**
- Create: `docs/work/current/t11-b09-f1-rebaseline-proof.md`
- Modify: `docs/work/current/t11-b09-audit-upstream-replan.md`
- Modify: `docs/work/current/t11-b09-audit-capability-decision-candidate-r2.md`
- Modify: `docs/roadmap.md`

**Consumes:**
- Durable decision from Task 1.
- Census from Task 2.

**Produces:**
- Explicit proof that the smallest owning authority was reopened and no unrelated locked block was invalidated.
- B09-F1 CLOSED / OPERATOR-RATIFIED.
- B09 P7 resumed; P8 still blocked.

- [ ] **Step 1: Create `t11-b09-f1-rebaseline-proof.md`.**

The proof must contain this exact affected-surface matrix:

```text
Product/T6
  Audit job = point investigation + period/authorized-scope review
  /audit route unchanged
  export/full-text/saved-search remain deferred; custom sort/analytics rejected

T8-C
  Audit structured predicates + historical visibility stay Audit-owned
  application composes optional current recognition and Query Assist owner facts
  no generic resolver/query service

T8-E
  op78 refined
  op87-op89 added as SAFE_READ
  evidence/recognition structurally separated
  cursor/filter laws preserved

T8-F / FP0
  operation coverage 86 -> 89
  stable routes 11 unchanged
  /audit now consumes 78 + 87 + 88 + 89
  no frontend AuthZ evaluator
  no client post-filter over incomplete pages

Censuses
  operations 89
  routes 11
  permissions 16
  idempotency 11
  ETag 13/13
  exact-byte 4
  owners 4+2

Locked blocks
  B01-B08 preserved
  no contradiction found
```

The proof must explicitly state that the bounded decision, not wholesale rewrites of large ratified authorities, is the current supersession mechanism.

- [ ] **Step 2: Close the upstream finding ledger.**

Update `t11-b09-audit-upstream-replan.md` to:

```text
Status: CLOSED / OPERATOR-RATIFIED / DURABLE AUTHORITY PROMOTED
Authority: ../../decisions/audit-investigation-read.md
Census: 89
P7: RESUMED / NEXT
P8: BLOCKED pending P7
```

Preserve the investigation rationale as evidence; do not delete the history of why the reopen occurred.

- [ ] **Step 3: Mark R2 as ratified/promoted evidence.**

Update only its status/gate language so it reads:

```text
OPERATOR-RATIFIED / PROMOTED
Durable authority: ../../decisions/audit-investigation-read.md
```

R2 remains work evidence and must not compete with the durable decision.

- [ ] **Step 4: Rebaseline `docs/roadmap.md`.**

Required state:

```text
system census                    89 / 11 / 16 / 11 / 13-13 / 4
FP0                               CLOSED / R2 89/11 REBASELINED
B09                               OPEN / ACTIVE
B09-F1                            CLOSED / OPERATOR-RATIFIED
Structured Audit Query           OPERATOR-RATIFIED
Human recognition                OPERATOR-RATIFIED
Audit Query Assist               OPERATOR-RATIFIED
Owner-lens cross-links           OPERATOR-RATIFIED
Exact op78 + op87-op89 package   OPERATOR-RATIFIED / DURABLE
P7                                RESUMED / NEXT
P8                                BLOCKED pending P7
B10-B12                           NOT OPEN
IMPLEMENTATION                    BLOCKED
```

Exact next action becomes B09 P7 only. Do not open P8 or B10.

- [ ] **Step 5: Commit the work-record/rebaseline writes and verify exact-head CI SUCCESS.**

---

### Task 4 — Synchronize PR metadata and run final consistency proof

**Files:**
- PR #162 metadata only via `update_pull_request`; no repository file for this step.

**Consumes:**
- Final exact HEAD after Tasks 1-3.

**Produces:**
- PR body matching repository authority.
- Fresh final verification evidence.

- [ ] **Step 1: Compare the pre-promotion HEAD to final HEAD.**

Expected changed paths for this promotion/rebaseline stage:

```text
docs/decisions/audit-investigation-read.md
docs/decisions/index.md
docs/decisions/api-operation-census.md
docs/work/current/t11-b09-f1-rebaseline-proof.md
docs/work/current/t11-b09-audit-upstream-replan.md
docs/work/current/t11-b09-audit-capability-decision-candidate-r2.md
docs/roadmap.md
docs/work/current/t11-b09-f1-promotion-rebaseline-plan.md
```

The earlier R1/R2 candidate files already present before promotion remain evidence; no Product/runtime implementation paths may appear.

- [ ] **Step 2: Re-fetch `main`, PR #162 and final HEAD.**

Require:

```text
main unchanged unless separately authorized
PR OPEN / DRAFT / mergeable
head branch arch/t11-implementation-program
```

- [ ] **Step 3: Fetch final workflow run and required job.**

Require every required step `SUCCESS` on the exact final HEAD.

- [ ] **Step 4: Update PR body metadata only.**

It must report:

```text
current authoritative census 89
B09-F1 CLOSED / OPERATOR-RATIFIED
bounded Audit authority promoted
FP0 R2 89/11 REBASELINED
B09 P7 RESUMED / NEXT
B09 P8 BLOCKED pending P7
B10-B12 NOT OPEN
Product implementation none
T12 none
```

- [ ] **Step 5: Final contradiction scan.**

Search current decision register, roadmap and census for live mutable claims still asserting `86` as current after promotion. Historical stage snapshots may remain only where clearly superseded/history. Verify no current gate says B09-F1 open or P7 paused.

- [ ] **Step 6: Stop at the P7 gate.**

Do not generate functional HTML, do not open B10, and do not merge. The next Product Experience action is B09 P7 layout-hypothesis work under the now-ratified Audit authority.
