# Approval shape — one decision surface

**Date:** 2026-08-05
**Module:** `approval` (owner) · touches `iam`, `documents`, `notifications`, `taxonomy` (read-only)
**Gate:** `docs/superpowers/analysis/2026-08-04-approval-shape-system-impact.md` — verdict 🟡 Yellow,
ten locked constraints in its §10. This spec is written inside those rails and does not re-derive them.
**Status:** design approved by the operator, section by section, 2026-08-05.

---

## 1. Why

The operator's objection: a route variant that only signs makes no sense, because a signer who finds a
defect has to review and return the document anyway.

Investigation confirmed the objection and relocated it. The behaviour already exists — an
`approval`-kind stage can record `request_changes` today (`review_verdict_service.go:177`). What is
broken is the expression:

- **Two near-identical records.** `approval_review_verdicts` is `approval_signoffs` minus five
  signature columns, with the same uniqueness rule and the same `enforce_approval_sod` trigger. The
  SoD invariant is therefore enforced in two schemas and can drift in one.
- **Two services duplicating predicates.** `ReviewVerdictService` and `DecisionService` each call the
  same `ResolveEligibleIdentity`, the same `CheckSoD`, and each carry a byte-equivalent
  `emitEligibilityRejection`.
- **Two "no" actions on the same stage, no rule.** `request_changes` returns the document to the
  author; a signoff with `decision='reject'` kills the instance. Identical preconditions, divergent
  terminality, and the distinction carried entirely by which screen the actor opened.
- **A dead field.** Each stage declares `required_capability`, and both services ignore it in favour
  of a hard-coded constant.
- **An invariant held by configuration.** `RecordVerdict` requires `approval.review` even on an
  approval stage, and `approval.review` is a per-profile grant — so a tenant can configure a signer
  whose only "no" is terminal rejection.

These are catalogued as defect classes 9 (second copy of a critical path), 23 (invariant delegated to
configuration) and 24 (two operations, one precondition, divergent terminality) in
`docs/engineering/defect-class-catalog.md`.

The concepts are sound and regulatorily required — ISO 13485 §4.2.4 wants documents reviewed *and*
approved; 21 CFR Part 11 §11.50 wants the meaning of each signature recorded. What is legacy is the
belief that two sound concepts need two mechanisms.

## 2. Decisions taken

Six forks, each decided by the operator during design:

| # | Question | Decision |
|---|---|---|
| 1 | Does an approver keep a terminal rejection? | **No.** One "no": return to the author. Negative termination leaves the stage decision and becomes cancellation (`CancelInstance`, which already has its own capability and `cancel_reason`). |
| 2 | Where does the signature requirement live? | **On the stage**, as `requires_signature`. `stage_kind` dies. Rule R1 of the 2026-07-10 model (per-profile signature policy) is superseded — it is Class 23 by construction. |
| 3 | What does `controlado` require of a route? | **≥1 stage with `requires_signature`.** No mandatory review stage: every stage can now return, so an approval-only route keeps its promise. No existing route breaks. |
| 4 | How are the two tables unified? | **Clean schema.** Originally "widen in place", revised once the operator confirmed there is no production data — that option's only advantage was evidence continuity. |
| 5 | Who decides the required capability? | **Derived from `requires_signature`.** `required_capability` and its snapshot are deleted. No configuration surface on the authz path. |
| 6 | Does R5 "Aprovar já" (fast-forward) survive? | **No.** One deliberate act per stage. `fast_forward_service.go`, its endpoint and its screen are deleted. |

Plus one addition accepted after a clarification: **signature meaning becomes a per-stage closed
enum** (§3.3).

## 3. The model

### 3.1 Route stage

- `approval_route_stages.requires_signature boolean NOT NULL` replaces `stage_kind`.
- `approval_stage_instances.requires_signature_snapshot boolean NOT NULL` is its snapshot.
- `required_capability` and `required_capability_snapshot` are **deleted** — from the route table, the
  stage-instance table, the domain structs, the OpenAPI contract and the route builder. The code
  already ignored them.

### 3.2 Decision record

One table, `approval_decisions`, replacing both `approval_signoffs` and `approval_review_verdicts`:

| Column | Note |
|---|---|
| `id`, `approval_instance_id`, `stage_instance_id`, `actor_user_id`, `actor_tenant_id` | as today |
| `outcome` | **closed:** `approve` \| `return_for_changes` |
| `comment` | nullable; mandatory for `return_for_changes` (§3.5) |
| `decided_at` | replaces `signed_at` / `verdict_at` |
| `actor_display_name_snapshot` | non-empty CHECK, as today |
| `on_behalf_of_user_id` | ADR 0077 delegation, as today |
| `requires_signature_snapshot` | denormalized from the stage — see §3.4 |
| `signature_method`, `signature_payload`, `content_hash`, `signature_meaning` | **the signature block; all four nullable together** |

Uniqueness stays `(stage_instance_id, actor_user_id)`. `enforce_approval_sod` is re-pointed at this
table. RLS: `tenant_isolation` policy + `FORCE ROW LEVEL SECURITY`, carried over unchanged.

The instance status `rejected` is removed from the status enum — but only after its **second**
producer is dealt with. Removing the human rejection is not enough: `ApplyEligibilityDrift` forces
`QuorumRejectedStage` with zero rejection votes when a `fail_stage` stage loses an eligible actor, and
`EvaluateQuorumResult` forces it again when the effective denominator collapses to zero
(`domain/drift.go`, `domain/quorum.go`). Both are *system* verdicts meaning "nobody can decide this
stage any more".

Their honest outcome is the same as a human "no": **return to the author**. So the forced outcome is
renamed `QuorumStageUnsatisfiable` and lands on the return path, carrying a system reason on the
governance event and writing no decision row — nobody decided. Only then are `rejected`, the stage
status `rejected_here`, and `Instance.RejectHere` genuinely unreachable, and dropped.

### 3.3 Signature meaning

21 CFR Part 11 §11.50 requires the *meaning* of a signature to be recorded with it — the norm's own
examples are review, approval, responsibility, authorship. A route may legitimately carry several
signature stages that mean different things: QA confirms the technical content, the area manager
approves for their sector, a director authorises issue. Recording all three as `approval` is correct
about the act and poor about the meaning.

So:

- **Closed vocabulary**, not free text: `review` · `approval` · `release`. Cross-boundary vocabulary
  (client, PDF manifestation, audit) without a registry is Class 19, the most expensive class to
  retrofit.
- Declared **per stage**: `approval_route_stages.signature_meaning`, snapshotted onto the stage
  instance and copied onto the decision row at the moment of signing — the same snapshot discipline
  already used for stage name and actor display name.
- Constraint: `signature_meaning` is non-null on a stage **iff** `requires_signature` is true.
- Display labels are PT-BR and live in the UI layer; the stored value is the stable identifier.

### 3.4 The central constraint

> A signature block is present **if and only if** `requires_signature_snapshot` **and**
> `outcome = 'approve'`.

Returning for changes never carries a signature: Part 11 asks for a declared meaning on an act of
approval, not on "send it back and fix it".

For that to be a real CHECK rather than an opaque trigger, the decision row carries
`requires_signature_snapshot` denormalized. A denormalized copy is a second copy of a fact
(Class 2), so it is paired with a trigger asserting it equals the owning stage's snapshot — **the copy
cannot lie**. This is the one place the design accepts denormalization, and it pays for it in the same
migration.

### 3.5 Comment

`comment` is mandatory when `outcome = 'return_for_changes'` — a DB CHECK, not an app guard. An
author who gets work back with no stated reason cannot act on it, and the accountability loop shipped
in the previous spec exists to stop exactly that kind of silent stall.

### 3.6 Route shape

`assert_route_shape` keeps its three call sites and its arms are restated in the new vocabulary:

- `livre` ⇒ **zero** stages (ADR 0087, unchanged);
- every other class ⇒ **≥1** stage;
- `controlado` **and any unknown class** (fail-closed) ⇒ ≥1 stage with `requires_signature`.

`simples` keeps only the ≥1-stage floor.

## 4. Flow

One sequence, one transaction:

```
load instance → require status in_progress → resolve area code
  → authz.Require(capability derived from the stage's requires_signature, area)
  → assert the addressed stage is the active one
  → eligibility: ResolveEligibleIdentity + ADR 0077 delegations
  → SoD: CheckSoD (author cannot decide own submission; nor a delegate of the author)
  → insert the decision row
  → apply the outcome
```

Eligibility and SoD now exist **once**. `emitEligibilityRejection` exists once.

**`approve`** goes through the existing `EvaluateQuorum` — review stages and signature stages use the
*same* quorum path, where today two code paths express the same intent. On quorum:

- stage completed; if a next stage exists it is activated, and `approval.pending` is enqueued in the
  **same transaction** with the newly activated stage's `EligibleActorIDs`;
- otherwise the instance completes through the already-extracted terminal helpers
  (`document_terminal_approval.go` / `template_terminal_approval.go`).

The stage-activation notification closes an accepted follow-up from the accountability loop
(multi-stage routes never told stage 2+ approvers anything). It is in scope here because this design
rewrites the activation path; leaving it out would mean reopening the same code twice.

**`return_for_changes`** has no quorum — one decision collapses the instance. Instance →
`changes_requested`; `SET LOCAL metaldocs.cancel_in_progress`; `authz.Require(document.edit, area)`;
`CanTransitionDocumentStatus(under_review → draft)`; document → `draft` with
`revision_version + 1`. Zero rows affected ⇒ stale-revision error. This is today's behaviour,
unchanged, and now the only "no".

Approval never writes documents' rules: the status flip continues to go through the published
transition check and the existing GUC gate.

## 5. Authorization

Derived, not configured:

| Stage | Subject | Capability required for **both** outcomes |
|---|---|---|
| `requires_signature = true` | document | `document.signoff` |
| `requires_signature = true` | template | `template.approve` |
| `requires_signature = false` | either | `approval.review` |

The subject arm is not a new idea — `decision_service.go:358` already switches the required capability
on `instance.Subject.Kind`, and the DB tripwire (ADR 0083, migration 0300) discriminates the same way
by the parent instance's `subject_kind`. Deriving the capability from `requires_signature` alone would
silently drop that arm; the derivation is a 2×2, and the tripwire is what would have caught it.

Because both outcomes on a signature stage require the same capability, **a signer can always
return** — R3 stops being a grant an administrator can withhold and becomes a property of the code.
This is the direct closure of the Class 23 defect.

No new capability is expected, so `TestCapabilityRegistrySize` should not move. Verify against the
merge base rather than the working tree.

Tier-1: the new route gets an explicit row in `apps/api/cmd/metaldocs-api/permissions.go`; the three
retired routes' rows are deleted in the same commit. The generic `/api/v1/approval/` prefix fallback
no longer exists, so a missing row falls through to `VisibilitySessionRequired` — silent privilege
escalation.

## 6. Contract

`POST /api/v1/approval/instances/{instance_id}/stages/{stage_id}/decisions` replaces, at the same
stage-scoped shape the retired routes already use:

- `POST …/stages/{stage_id}/review-verdict`
- `POST …/stages/{stage_id}/signoffs`
- `POST …/stages/{stage_id}/fast-forward`

The addressed stage stays a **path parameter**, as today — it is what the request identifies, not a
field inside it. `POST /api/v1/templates/{id}/versions/{n}/signoff` (the template kernel entry point,
`templates/delivery/http/routes_approval_kernel.go`) keeps its own path and is re-pointed at the
unified service; renaming it is template-module surface and out of scope here.

All three are deleted from `api/openapi/v1/openapi.yaml`, from `permissions.go`, from the generated
Go, and from the frontend, in one change set. No optional-for-compat fields, no aliases.

Regeneration is **full**: `go generate` for every module's `api` package plus `npm run gen:api`.
Partial regen is forbidden drift — a description-only edit still churns generated doc comments.

Request body carries `outcome`, `comment`, an optional signature block, and the existing
optimistic-concurrency field.

The read surface changes with it: the worklist's `stage_kind` query filter and the `stage_kind` field
on both the worklist item and the instance-detail stage become `requires_signature` (boolean), and the
instance detail's two parallel histories — `signoffs` and the ADR 0079 verdict history — become one
`decisions` list. The route-stage request/response schemas lose `required_capability` (today a
**required** property, so this is a breaking edit the route builder must follow) and gain
`requires_signature` + `signature_meaning`. Idempotency stays per-handler (`signoff_idemp.go` generalized),
not a middleware-chain link.

## 7. Events and errors

Governance event types collapse to `decision.recorded` and `decision.returned`, following the dotted
taxonomy of ADR 0089. The verdict-specific and signoff-specific types are deleted, not aliased.

Error codes for the retired routes are deleted; new codes are registered through `problem.Register`,
whose duplicate-registration panic at init is the mechanical guard against a half-migrated error
surface.

Two domain errors **cease to be expressible**: `ErrVerdictReadyOnApprovalStage` and
`ErrVerdictWrongStageKind` have no state that can produce them once `stage_kind` is gone. Deleting an
error together with its possibility is rung 1.

## 8. Frontend

The two decision screens converge into one surface: the same view offers both outcomes, and the
signature ceremony appears only when the stage requires it, labelled with that stage's meaning. The
fast-forward affordance is removed; "you are also eligible on the next stage" may remain as a hint,
not as a second write path.

Generated API types are regenerated in the same change set (`npm run gen:api`).

## 9. Migration

There is no production data — the operator confirmed the system is running on seeded data only, and
the 2026-07-29 baseline fold already wiped the dev database. So the migration is written for the
schema, not for a backfill:

1. Create `approval_decisions` in its final shape, with every constraint valid from the start.
2. Add `requires_signature` + `signature_meaning` to `approval_route_stages` and the matching snapshot
   columns to `approval_stage_instances`; drop `stage_kind`, `required_capability`,
   `required_capability_snapshot`.
3. Re-point `enforce_approval_sod`; `CREATE OR REPLACE assert_route_shape` with the §3.6 arms.
4. Narrow the instance status enum (drop `rejected`).
5. Drop `approval_signoffs` and `approval_review_verdicts`.
6. Reseed dev.

Migrations resolve every table they touch up front and qualify every name explicitly — no reliance on
`search_path` (Class 22).

## 10. Mechanical validation

The operator's standing requirement: nothing on this list is a convention.

| # | Guard | Rung |
|---|---|---|
| 1 | CHECK — signature block present ⟺ (`requires_signature_snapshot` ∧ `outcome='approve'`) | 1 |
| 2 | Trigger — denormalized `requires_signature_snapshot` must equal the owning stage's | 1 |
| 3 | CHECK — `signature_meaning` non-null on a stage iff `requires_signature`; closed value set | 1 |
| 4 | CHECK — `comment` non-empty when `outcome='return_for_changes'` | 1 |
| 5 | `enforce_approval_sod` re-pointed at `approval_decisions`, with a test proving it fires | 1 |
| 6 | `outcome` closed to two values; instance status enum no longer contains `rejected` | 1 |
| 7 | `assert_route_shape` test per arm: `livre`=0, others ≥1, `controlado` and unknown ≥1 signature stage | 3 |
| 8 | **Test: a profile holding `document.signoff` can always `return_for_changes`** — the R3 regression guard | 4 |
| 9 | Deletion guard — the three retired routes absent from the spec, `permissions.go`, generated Go, and the FE | 3 |
| 10 | `problem.Register` duplicate-registration panic covers the error surface | 2 |
| 11 | api-lint run as `-strict` (ladder amended 2026-08-05) + tripwire arms regenerated | 3 |

## 11. Tests

- Canonical framework only: `testdb` factory, `//go:build integration`, discipline R1–R4.
- `go vet -tags integration ./...` before any commit — a seam signature change that compiles untagged
  can still break tagged files.
- Tests belonging to the two deleted services are **deleted**, not adapted. Only contract and
  invariant guards survive, and any survivor must land on the canonical framework.
- Integration ladder policy: touched packages + guard suites by default; full `./...` because
  `db/migrations` is touched.

## 12. Evidence

`go build ./...` · `go test ./...` · `go vet -tags integration ./...` ·
`go test -tags=integration ./...` · `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` ·
`.\scripts\check-system-runnable.ps1` · live QA driving the unified endpoint end-to-end through the
container stack via gateway `:80`.

Report commands, outcomes, QA disposition and any bounded defers before claiming completion.

## 13. Docs and ADR

- **New ADR** stating: the unified decision record, the outcome enum, `requires_signature` as the
  ceremony trigger, per-stage signature meaning, the derived capability, and the regulatory mapping
  (Part 11 §11.50, ISO 13485 §4.2.4) showing unification does not weaken evidence.
- **Supersedes:** rules R1, R3 and R5 of the 2026-07-10 review/approval model; ADR 0082's stage-kind
  expression (the module promotion itself stands). ADR 0087's `livre` ruling is untouched. ADR 0083's
  tripwire arms are regenerated if any route/cap pair changes.
- **Wiki:** `wiki/modules/approval.md` and `approval-tech-debt.md` with refreshed `Last verified`;
  `wiki/architecture/api-contract.md` for the route retirement. Cite the REQ IDs from
  `wiki/architecture/backend-target-architecture.md` — confirm the exact IDs against the doc rather
  than quoting from memory.

## 14. Out of scope

Named so they are not silently absorbed:

- The other five accepted follow-ups from the accountability loop: `eligible_actor_ids = '[]'` legal
  but unsignable; template notifications with no deep link; ADR 0077 delegates absent from the
  recipient set; `stateTransitionPathRE` not forcing an area marker on approval ops; the repo-wide
  gofmt gate.
- Spec 2 — assignment strategy, escalation, absence/substitution, `approval.reassign`.
- Any change to `governance_class` itself (taxonomy-owned) or to documents' status machine.
