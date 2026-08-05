# Adversarial review ledger — unified approval decision surface

**Artifact:** `docs/superpowers/plans/2026-08-05-approval-shape-unified-decision.md` (gitignored
by design; this ledger is its tracked record).
**Reviewer:** Codex `gpt-5.6-sol` / effort `medium`, read-only, independent context.
**Protocol:** `.claude/skills/adversarial-review/SKILL.md`.
**Rounds:** 5. **Status: LOOP PAUSED — §2 escalation, operator decision required.**

---

## Convergence line (§8)

| Round | Findings | Altitude |
|---|---|---|
| 1 | 13 | design: missing authz, lost evidence, wrong structure |
| 2 | 4 BLOCKER / 4 MAJOR / 1 MINOR | design |
| 3 | 4 BLOCKER / 4 MAJOR / 1 MINOR + 4 PARTIAL | fixes that named a problem without closing it |
| 4 | 6 BLOCKER / 5 MAJOR | **mechanical** — nonexistent APIs, wrong status codes |
| 5 | 5 BLOCKER / 1 MAJOR / 1 MINOR + 9 PARTIAL | **structural again** — identity uniqueness, DB enforcement model, missing route registry |

**Verdict on the loop itself: NOT CONVERGING.** Round 4's altitude drop was not the design
search being exhausted — it was the design search being *interrupted* by a wave of compile-level
defects. Round 5 returns to structure and returns *higher* than round 4. Under §8 that is the
stop condition, not a reason for round 6: same-or-rising altitude across rounds means the
structure is wrong and no number of rounds fixes it.

Reinforcing signal, §1 repeated-finding rule: **signer identity is the third consecutive round
on one construct** (R3-C4 → R4-J2-2 → R5-1). The rule says stop bounding and go to §2.

**Resolution (2026-08-05).** §2 was run on both structures and the operator chose outcome (a)
for each — the signer-identity relation and the persisted-state discriminator are built, not
bounded. The rising altitude is what forced that, which is the rule working rather than the
loop failing. Round 6 resumes against the new structures; if it returns findings at the same
altitude on *those*, the stop is real.

---

## Round 5 — Job 1 dispositions (reviewer, verbatim)

```
1.  R3-C2 / R3-J2-6: PARTIAL — apps/api/cmd/metaldocs-api/main.go:490
2.  R3-C3 / R3-J2-7: PARTIAL — plan:2072
3.  R3-C4 / R3-J2-5: PARTIAL — plan:1390
4.  R3-J2-3:         CLOSED  — plan:1671
5.  R4-J2-1:         CLOSED  — plan:722
6.  R4-J2-2:         PARTIAL — plan:1394
7.  R4-J2-3:         PARTIAL — plan:208
8.  R4-J2-4:         CLOSED  — plan:1068
9.  R4-J2-5:         CLOSED  — plan:1985
10. R4-J2-6:         CLOSED  — plan:993
11. R4-J2-7:         CLOSED  — plan:1687
12. R4-J2-8:         PARTIAL — postgres_approval_repository.go:697
13. R4-J2-9:         PARTIAL — spec:243
14. R4-J2-10:        PARTIAL — main.go:490
15. R4-J2-11:        PARTIAL — errors.go:590
```

6 CLOSED / 9 PARTIAL / 0 OPEN.

---

## Round 5 — Job 2 findings and author disposition

Every finding was verified against source before disposition (§5 symmetric duty). All five
anchors checked came back **confirming the reviewer**.

### 1 — BLOCKER · signer-identity uniqueness (third round on this construct)
> ROOT CAUSE: a signed act stores two identity roles on one row, but uniqueness and domain
> checks compare selected columns rather than the complete identity set
> `{physical actor, effective signer}`.
> IMPOSSIBLE IF: signed decisions materialize participating identities into an
> instance-scoped relation with `UNIQUE (approval_instance_id, identity_id)`.

**Disposition: ESCALATED to §2, not patched.** This is the third bound proposed on one
construct. Per §1 that is a structural signal. The reviewer's `IMPOSSIBLE IF` names a concrete
structure (a signer-identity relation) and it is a different design than the plan's partial
index — it is not a refinement of it. Deciding between them is an operator call, not a patch.

Regulated consequence, stated plainly: until this is resolved, SoD evidence is not trustworthy
for 21 CFR Part 11 purposes. This is the single reason the loop is paused rather than merely
continued.

### 2 — BLOCKER · pre-authz normalization is structurally unimplementable
> ROOT CAUSE: a route-phase-specific privacy rule was assigned to the **global** error mapper,
> while `RecordDecision` exposes no pre-authz projection type and `ErrInstanceNotVisible` is
> also emitted *after* successful authorization by read services.

**Verified, reviewer correct.** `WriteError` (`internal/modules/approval/http/errors.go:555-565`)
logs **only** `status >= 500`, so the plan's "keep the internal cause in the log" has no
mechanism today. And `errors_test.go:43,64` do not test the codes the plan claimed.

**Disposition: APPLIED as a structural requirement, not a mapper tweak.** Tasks 3–4 must add a
distinct pre-authz projection type and a distinct pre-authz error; the handler emits the literal
canonical problem object and logs the cause itself. The global mapper is not touched. This is
the author's error twice over — the plan asserted a log path that does not exist and cited two
test lines it had not read (catalog Class 27).

### 3 — BLOCKER · Task 1 still cannot compile; capability envelope shape is wrong
> `asserted_caps` elements are **objects**, read as `v_element->>'cap'`
> (`db/baseline/0001_current_schema.sql:498`), not the string array the plan's table used.
> Stale `NewFixture` snippets survive later in the task.

**Verified, reviewer correct** — `baseline:498` reads `(v_element->>'cap')`. **Disposition:
APPLIED.** Caps become `[{"cap":"document.signoff"}]` / `[{"cap":"approval.review"}]`; every
remaining `NewFixture` snippet is replaced.

Independently found by the author in the same window and already fixed: `NewApprovalInstance`
(`tests/integration/testdb/factory.go:635-706`) is **document-only** — it always inserts
`document_id`, while `approval_instances` CHECKs require `document_id IS NULL` for
`subject_kind='template'` (`baseline:2048,2052`) and `trg_approval_instances_default_subject`
(`baseline:3933`) derives the kind from that column. Task 1 now also adds
`testdb.WithTemplateSubject`, or the `template.approve` arm of the 2×2 ships untested.

### 4 — BLOCKER · remapping the shared stale sentinel changes unrelated routes
> `updateApprovalRoute` requires `If-Match` (`api/openapi/v1/openapi.yaml:4184`) but declares
> 409/428 and **no 412** (`:4206-4209`).

**Verified, reviewer correct.** **Disposition: ALREADY APPLIED** in the same window, from the
author's own producer inventory: twelve production producers of `ErrStaleRevision` across six
services. The sentinel keeps 409; a distinct `ErrIfMatchPreconditionFailed` →
`approval/precondition.if_match` 412 is added for the decision path only, deliberately not
related by `errors.Is` (the 409 arm at `errors.go:213` would otherwise shadow it).

Reviewer's addition, accepted: `updateApprovalRoute` is a *pre-existing* contract defect —
requires `If-Match`, cannot express 412. Noted as out of scope, recorded, not silently fixed.

### 5 — BLOCKER · neither route-coverage option is executable in this repo
> `apps/api/cmd/metaldocs-api/main.go:490` builds a local `http.NewServeMux`. chi `Walk` is
> inapplicable. OpenAPI cannot cover off-contract mounts (`/api/v1/metrics`, `main.go:852-859`).

**Verified, reviewer correct** — `main.go:490` is `mux := http.NewServeMux()`. There is no chi
router. **Disposition: BOTH OPTIONS WITHDRAWN.** The plan offered two options and both were
built on an unverified premise — the author's own catalog Class 27, committed one commit
earlier. The honest state is: **there is no enumerable source of mounted routes in this
repository**, and route mounting and authorization classification are two independent
hand-maintained structures (catalog Class 2, at the composition root).

That prerequisite is its own work item. It is not created inside this change.

### 6 — MAJOR · the labelled successor would not delete the GUC; a GUC can be deleted now
> ROOT CAUSE: PostgreSQL observes table, operation, asserted capabilities and GUCs — **not Go
> module ownership** — so relocating the caller behind a Go port leaves the DB discriminator
> requirement unchanged.
> Also: the plan's claim that cancel's actor holds `approval.cancel` is **false** — cancel
> requires `CapDocumentEdit` (`internal/modules/approval/application/cancel_service.go:88`).

**Verified, reviewer correct on both.** `cancel_service.go:80-88` requires `document.edit`.
This falsifies the author's two-GUC rationale, and — more seriously — falsifies the §2
successor design the author wrote one commit earlier: **a `documents` write port does not
delete the GUC.** The tripwire sees the asserted capability, and a signoff-only actor still
does not hold `document.edit` no matter which module issues the UPDATE.

**Disposition: ESCALATED to §2 alongside finding 1.** The reviewer names the actual global
maximum: DB enforcement that derives the *lifecycle operation* from authoritative persisted
state (instance status: `changes_requested` vs `cancelled`) plus row binding, rather than from
an ambient caller marker — under which `approval_return_in_progress` is unnecessary and
`cancel_in_progress` may be sufficient for both paths. That is a documents-owned enforcement
design, and it is a materially different milestone from the one the plan labelled.

### 7 — MINOR · delete the speculative `StageActorSlot` wrapper
> One production consumer, which always overrides the wrapper's only default behavior
> (`StageCard.tsx:158`, `routeGovernance.ts:65`).

**Disposition: ACCEPTED, deferred with an owner.** Correct YAGNI call and the subtractive pass
did its job. Out of this change's boundary (frontend route-admin, untouched here); recorded for
the frontend lane rather than folded in.

---

## §2 verdicts — operator decisions required

Two, both outcome **(c) stop and surface**. Neither is the agent's to make.

### (c)-1 — Signer identity
- **Global maximum:** signed decisions materialize participating identities into an
  instance-scoped relation with `UNIQUE (approval_instance_id, identity_id)`; the domain
  predicate becomes a set-intersection over that relation.
- **Local maximum shipped today:** a pairwise loop plus a partial unique index on the physical
  actor. Third round of bounds; still permits cross-axis and same-stage delegated duplicates.
- **Cost of the global maximum:** one table, one migration, a rewrite of the SoD predicate and
  its tests. Contained.
- **Cost of the local maximum later:** SoD evidence that a regulator can defeat with a
  delegation. Not recoverable retrospectively — the historical rows would already be wrong.
- **Recommendation:** build it now. The cost asymmetry is not close.

### (c)-2 — The documents/UPDATE discriminator
- **Global maximum:** documents-owned DB enforcement that derives the lifecycle operation from
  persisted state (instance status + row binding), not from an ambient caller marker. Under it
  `approval_return_in_progress` is never created, and `cancel_in_progress` may be deletable too.
- **Local maximum currently in the plan:** a new GUC with three bounds, labelled transitional
  with a `documents` write port as successor. **That label is now known to be false comfort** —
  a Go-side port does not change what PostgreSQL observes, so the successor as described cannot
  delete the GUC.
- **Cost of the global maximum:** a redesign of the documents transition arm; crosses into the
  `documents` module. Larger than this change.
- **Cost of the local maximum later:** a permanent ambient marker plus a successor milestone
  that cannot discharge it, i.e. exactly the unlabelled-local-maximum failure the doctrine
  exists to prevent, wearing a label.
- **Recommendation:** do not ship the GUC on the current rationale. Either build the persisted-
  state discriminator in this change, or scope the return path so it does not write `documents`
  at all until the discriminator exists.

---

## Author-side defects found in this round (catalog Part II instances)

Four, all the author's, all Class 27 (confident false citation), all found by verifying
anchors rather than by re-reading prose:

1. "six direct-write call sites" — the real count is **ten writes plus one `FOR UPDATE` lock**,
   and three of them write non-status columns, which falsifies the single-method port design.
2. "`ErrStaleRevision` has exactly one producer" — it has **twelve**.
3. "cancel's actor holds `approval.cancel`" — it requires **`document.edit`**
   (`cancel_service.go:88`).
4. "chi `Walk` can enumerate the mounted routes" — the runtime uses **`http.NewServeMux`**
   (`main.go:490`).

Three of the four were written in the same session that committed the catalog entry describing
the class. That is the finding worth keeping: knowing the class does not prevent the class.
Only anchor verification does.

---

## Operator rulings on the two §2 escalations (2026-08-05)

Both answered by the operator, in session. Recorded here because §8 forbids closing a
DO NOT PROCEED without naming who accepted what.

### (c)-1 — Signer identity → **build the global maximum now**

Operator: *"Construir agora"*. Outcome **(a) restructure now**, not (b).

Applied to the plan:
- `public.approval_decision_signers` added to migration 0318 §3b —
  `PRIMARY KEY (approval_instance_id, identity_id)`, FK to `iam_users`, composite FK to
  `approval_decisions(id, approval_instance_id)`, no `ON DELETE CASCADE`, FORCE RLS with the
  same tenant policy shape as its parent. No `role` column: derivable from `decision_id`, one
  consumer, and the control's point is that the two capacities are indistinguishable.
- `record_decision_signers()` AFTER INSERT trigger flattens every signed approval into one row
  per participating identity via `unnest(ARRAY[actor_user_id, on_behalf_of_user_id])`. Unsigned
  decisions and returns write nothing, which is what preserves multi-stage commenting.
- The partial unique index `(approval_instance_id, actor_user_id) WHERE requires_signature_snapshot
  AND outcome='approve'` is **deleted, not supplemented** — strictly weaker on every delegated
  path, and two constructs enforcing one rule with different answers is what produced this
  finding three rounds running.
- Domain predicate rewritten as a set intersection over `participants(actor, onBehalf)`, which
  computes the same set from the same two columns as the key, so parity is structural.
- `MapPgError`'s `23505` switch (`infrastructure/errors.go:82-89`) gains both new constraint
  names; without the second, a lost race surfaces as `ErrUnknownDB` → 500.
- Six mandatory tests, each asserting domain rejection **and** DB rejection with the service
  bypassed: same actor twice · `A→B` then `B` direct · `B` direct then `A→B` · `A→B` then `C→A`
  · `A→B` and `A→C` same stage · three distinct people accepted.

### (c)-2 — documents/UPDATE discriminator → **persisted-state enforcement, GUC deleted**

Operator delegated the choice: *"O que recomenda avaliando com o objetivo de nunca errar"*.
Recommendation given and applied: option (a), the persisted-state discriminator. Outcome
**(a) restructure now**.

The deciding fact, found while answering: the return path already writes the authoritative
instance status **before** it touches `documents` (`decision_service.go:643` precedes `:681`,
same tx) and already sets `cancel_in_progress` at `:661`. The proposed GUC would have been a
third marker for a fact the database already holds.

Applied to the plan:
- The `documents` UPDATE arm derives from `approval_instances.status = 'changes_requested'`
  bound to the row, plus `ai.xmin::text::bigint = pg_catalog.pg_current_xact_id()::text::bigint`
  — proof the instance row was written by *this* transaction, which is unforgeable and closes
  the hole where a historical `changes_requested` instance would authorize forever. PG13+;
  the repo runs PG16.
- `set_config('metaldocs.approval_return_in_progress', …)` and the `authz.Require(CapDocumentEdit)`
  that accompanied it are both **deleted** from `decision_service.go:661-670`.
- `enforce_document_transition`'s own use of `cancel_in_progress` (`baseline:533`) is explicitly
  out of scope and unchanged.
- Section (2b)'s "labelled transitional local maximum with a `documents` write port as
  successor" is withdrawn in full: a Go port does not change what PostgreSQL observes, so the
  successor could never have discharged the GUC. The ten-site boundary violation is now recorded
  as standalone architecture debt with no dependency on the tripwire, and scheduling it is a
  separate operator decision argued on module-boundary grounds.

---

## Round-5 findings — applied dispositions

| # | Finding | Disposition |
|---|---|---|
| 1 | signer identity (BLOCKER) | **applied** — see (c)-1 above |
| 2 | pre-authz normalization unimplementable (BLOCKER) | **applied, and the fix reversed the design.** Verification found `ErrInstanceNotVisible` is produced by the *read* path (`read_service.go:961`), asserted by `read_service_tenant_grade_view_integration_test.go:121,134,244`, and branched on by the FE cockpit — so deleting its code was Class 19 repurposing of a shared entry. The global mapper, the code and every producer are now left untouched; a typed `RecordDecisionPreflight` + `ErrDecisionSubjectUnresolvable` boundary emits the literal canonical 404 in the handler and logs the cause there. No vocabulary fanout at all. Reviewer's three sub-claims all confirmed: `WriteError` logs only ≥500 (`errors.go:555-565`), `responseTitle` derives the title from the error (`errors.go:549`) so one code ≠ one response, and `errors_test.go:43,64` test `ErrNoActiveInstance`/`ErrInstanceCompleted` — the author had cited them unread |
| 3 | Task 1 cannot compile; wrong envelope shape (BLOCKER) | **applied** — all `NewFixture` snippets replaced with test-local helpers over the real constructors; envelopes are `[{"cap":"…"}]` per `baseline:498`; `seedStageInstance` specified; `WithTemplateSubject` and an exported `SeedWithCapsIdentity` added to the Files list. Plain `SeedWithCaps` is insufficient — it seeds no tenant GUC, so RLS filters the row instead of the CHECK rejecting it |
| 4 | shared stale sentinel → 412 (BLOCKER) | **applied** — the distinct `ErrIfMatchPreconditionFailed` was already in place; this round adds the missing half: `updateApprovalRoute`/`deactivateApprovalRoute` require `If-Match` and declare no 412 (`openapi.yaml:4184,4198-4210,4226`), recorded as pre-existing and out of scope with its anchors |
| 5 | neither route-coverage option is executable (BLOCKER) | **applied** — both options withdrawn with the reasons (`main.go:490` is a stdlib mux; `/api/v1/metrics` at `main.go:852-859` is a deliberate off-contract mount). The gap is stated, not papered over; spawned as its own work item |
| 6 | the labelled successor cannot delete the GUC (MAJOR) | **applied** — see (c)-2 above |
| 7 | delete `StageActorSlot` (MINOR) | **accepted, deferred with an owner** — the frontend lane, in the milestone that next touches route-admin; re-raise rather than re-defer if that milestone lands without it |

**Found by the author while applying, not by the reviewer:** `scripts/api-lint/async-tenant-tables.txt`
is a hand-typed mirror of the FORCE-RLS table set and has drifted ten names (seven missing —
including `approval_signoffs` and `approval_review_verdicts` — and three phantom). Task 1 now
adds only this change's two rows and records the drift as out of scope; the durable fix
(generate it from the schema) is spawned as its own work item.

---

## Loop state

**RESUMABLE.** Both §2 escalations are answered and recorded above, and all seven round-5
findings have a written disposition. Round 6 attacks the new material: the signer-identity
relation and its trigger, the `xmin`-bound documents arm, the preflight boundary, and the
Task 1 helper contract.

Evidence artifacts: `agent__r5.log`, `agent__r5.last.md` in this session's scratchpad;
prompt at `prompt-approval-review-r5.md`.
