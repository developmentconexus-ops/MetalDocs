# Milestone 2b — Approval Kernel Backend

> **Program:** approval-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md`
> **Status:** Spec approved
> **Authored:** 2026-07-07 — before any feature in this milestone began.

## Runtime-truth corrections to the plan (locked before execution)

The implementation plan (`docs/superpowers/plans/2026-07-07-m2b-approval-kernel-backend.md`) was
written before baseline verification. A dedicated verification pass (2026-07-07, this session)
found the following drifts from the plan's assumed schema — these corrections are binding for
every feature below, superseding the plan's text wherever they conflict:

1. **Table name:** the instance-scoped stage table is `approval_stage_instances`, not
   `approval_instance_stages`. F1's migration and all instance-stage code touches
   `approval_stage_instances`.
2. **`content_hash_at_submit` lives on TWO tables today**, neither of which is the future pin site:
   - `approval_instances.content_hash_at_submit` (text NOT NULL, set once at submit) — exists but
     is **not** read by either hash-chain check path today.
   - `documents.content_hash_at_submit` (nullable) — this is what both
     `active_instance_reader.go:52-75` and `postgres_approval_repository.go` (`LoadActiveDocumentContentHash`,
     lines 1128-1154) actually `COALESCE` against `document_revisions.content_hash`.
   F6 must repoint **both** call sites off `documents.content_hash_at_submit` entirely and onto the
   new `approval_instances.frozen_content_hash` column (added in F1, populated in F5) for any
   document with an active/completed instance; `documents.content_hash_at_submit`'s COALESCE-to-head-revision
   remains only for the true no-instance/draft case, expressed as a separate status-explicit
   branch — never one polymorphic COALESCE spanning both.
3. **No pre-existing "unified status transition function"** exists for the approval domain (the
   plan's F4 step referenced one "from GMR M4 versioning-kernel" — that GMR milestone's unified
   transition function belongs to `controlleddocuments` document versioning, a different bounded
   context, not `documents/approval`). Approval status transitions today are plain domain-method
   validation (`domain/instance.go` `AdvanceStage`/`RejectHere`/`SkipStage`/`Cancel`) plus a DB CHECK
   constraint (`approval_instances_status_check`, `approval_stage_instances_status_check`) — no
   plpgsql transition-validation function. F4 extends this existing app-checks-first + DB-CHECK-last
   pattern (add `changes_requested` to both the Go enum and the CHECK) rather than presuming an
   external function to extend.
4. Current `approval_instances.status` CHECK values: `in_progress`, `approved`, `rejected`,
   `cancelled`. Current `approval_stage_instances.status` CHECK values: `pending`, `active`,
   `completed`, `skipped`, `rejected_here`, `cancelled`. F4 adds `changes_requested` to the instance
   enum only (stage-level re-entry reuses `active`).
5. `approval_signoffs.decision` CHECK is `approve`/`reject` today — no `signature_meaning` column
   yet (F1 adds it, F7 wires it end-to-end). Confirmed no drift here vs. the plan.
6. SoD is already enforced at two layers today — DB trigger `enforce_signoff_sod()` (author cannot
   sign own document) + app-level `domain.CheckSoD()` call in `decision_service.go:300`. F7's "single
   predicate" unification collapses these to one rule text reused by both layers (it does not
   invent SoD from scratch).
7. `HasUnresolvedComments` (decision_service.go:381-388 gate at final approve) is **document-scoped**
   today, confirmed. F5 replaces this call site with a new **instance-scoped** predicate at the
   freeze boundary, per spec §2.2's comment-resolution-scope note.

## Objective

Remediate the approval backend to the ratified design (spec §2-§7, §11): separate review
(collaborative, no signature) from approval (signature) stage kinds; a real content-freeze
boundary producing one canonical, no-fallback hash chain (pin at freeze → echo at signoff →
verify at publish); versioned-immutable route definitions (routes stop being permanently frozen
after first use, per superseded ADR 0018 §1/§3); explicit `approval.review`/`approval.oversee`
capabilities replacing the generic `/approval/` tier-1 prefix fallback; signature meaning (21 CFR
11.50(a)(3)); SLA due-date surfacing; visibility gating (404, not 403, across the eligibility
boundary); and delegation-of-authority. The quality bar this milestone must move: the approval
system's non-negotiable-invariant coherence (AuthZ = capabilities never roles; DB enforces
invariants; no-fallback on integrity-critical reads) — measured by the grep-zero checks and lint
parity in the validation definition below, not by adjectives.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|---|---|---|---|
| F1 | `f1-stage-kind-schema-expand` | `stage_kind` domain enum (Go + DB CHECK, aligned) on `approval_route_stages`/`approval_stage_instances`; `due_in_days`/`due_at`; `approval_instances.frozen_content_hash`/`cancel_reason`; `approval_signoffs.signature_meaning`. Expand-only migration `0286`. | `TestStageKindValues`/`TestStageKindValidate` pass; testdb integration test proves the DB CHECK rejects an unknown `stage_kind` on insert; `go build ./...` clean. |
| F2 | `f2-route-versioning-pool-validation` | Versioned-immutable `approval_routes` (version, `is_active`/`active`, partial-unique active-per-profile); `enforce_route_immutable` becomes a tripwire on referenced/inactive rows only; empty stage pool at submit → `ErrEmptyStagePool` → 422. Migration `0287`. ADR: route versioning (supersedes ADR 0018 §1/§3). | Integration test: route v1 → instance A pins v1 → route update creates v2 active, v1 untouched and `is_active=false` → instance A still resolves v1 → direct SQL UPDATE of a referenced/inactive route row raises P0001 → submit against an empty-pool route returns 422 problem+json with `ErrEmptyStagePool`. `oapi-codegen` regen committed. Existing route-admin tests stay green. |
| F3 | `f3-capabilities-review-oversee` | `CapApprovalReview`/`CapApprovalOversee` walked through all 10 capability-wiring touchpoints; delete the generic `/approval/` tier-1 fallback (`permissions.go:250-253`); explicit tier-1 row per runtime approval verb; tier-2 `authz.Require` gating; seed grants + regenerated tripwire arms. Migration `0288`. | `TestCapabilityRegistrySize` count is exactly +2 over its pre-F3 value; classification/area-grade guard tests pass; both authz drift/parity lints green; grep-zero: no `/approval/` generic prefix block remains in `permissions.go`. |
| F4 | `f4-review-verdicts-changes-requested` | `POST .../stages/{stageId}/review-verdict` (`ready`/`request_changes`, mandatory comment on `request_changes`); `changes_requested` instance status (extends the existing Go-enum + DB-CHECK pattern per correction #3 above — no external transition fn to extend); delete `SkipStage` + `ErrCannotSkipLastStage`; `cancel` requires `reason`. Migration `0289`. | Domain + integration tests: ready-verdict advances quorum like today's approve; request_changes (no comment) → domain error; request_changes → `changes_requested`, route/version pin retained, re-submit re-enters same instance/same stage; verdict on an approval-kind stage → error; concurrent verdicts on the same stage serialize (one 409 or both recorded per quorum, never a lost update); cancel without reason → 422, with reason → persisted. Grep-zero: zero references to `SkipStage` anywhere in the tree. |
| F5 | `f5-freeze-boundary` | Freeze executor at last-review-stage→first-approval-stage transition (and at submit for approval-only routes): markup gate (OOXML scan rejects `w:ins`/`w:del`/`w:pPrChange`/comment marks), **instance-scoped** unresolved-comment gate (replaces F5's removal of the document-scoped gate at `decision_service.go:381-388`), canonical hash computed + pinned to `approval_instances.frozen_content_hash`. | `markup_gate_test.go` table-driven over real minimal docx fixtures (clean → pass; `w:ins` → `ErrUnresolvedTrackedChanges`; comment mark → `ErrUnstrippedComments`); freeze integration test: clean review completion freezes + locks the document (edit-save endpoint 409s post-freeze); unresolved instance comment blocks freeze with a typed 409; approval-only route freezes at submit; freeze is idempotent (re-entry no-op, hash unchanged); concurrent freeze attempts — one wins CAS, other 409. |
| F6 | `f6-no-fallback-hash-chain` | Repoint `LoadActiveDocumentContentHash`/`active_instance_reader.go` off `documents.content_hash_at_submit` COALESCE onto `approval_instances.frozen_content_hash` for any document with an active/completed instance, with a separate explicit branch for the true draft/no-instance case — per correction #2. Absent/NULL pin at signoff/publish → typed `ErrNoActiveContentHash`, fail closed. | Integration test: signoff against an instance with NULL frozen pin never falls through to a head-hash comparison, gets `ErrNoActiveContentHash` → 409; publish with NULL pin fails closed; active-document endpoint for a frozen instance returns the pin, matching what signoff compares (FE echo contract preserved). Grep-zero: zero production `COALESCE` hits on either the `documents.content_hash_at_submit`-vs-head-revision pattern in the two named call sites (tests may still use the pattern to assert its absence). |
| F7 | `f7-signature-meaning-sod-unification` | `signature_meaning` (`approval`/`rejection`) persisted + rendered in the signoff DTO/manifest; single exported `ViolatesSoD(authorID, actorID, stageKind)` predicate reused by both the app-level check and one DB trigger (collapses the existing dual-site enforcement per correction #6 — same rule text both places, not new SoD logic). Migration `0290`. | Table-driven SoD test covers author-signoff, author-review-verdict, delegate pass-through-for-F9 cases; direct SQL insert violating SoD still raises P0001 via the single trigger; signoff record + manifest DTO expose `signature_meaning`. |
| F8 | `f8-sla-visibility-worklist` | `due_at` set on stage activation from `due_in_days`; `document-review-surfacer`-pattern periodic job surfaces overdue approval stages (alert-only, ADR 0068); visibility predicate — instance/unpublished-revision reads allowed only for author, stage-pool member (any stage), `CapApprovalOversee`, or `document.edit` holder, else **404**; worklist filters (`stage_kind`, `due_before`, doc type) + `?scope=oversee`. | Visibility matrix integration test: consumer→404, author→200, pool member→200, oversee→200, cross-tenant→404. Surfacer test follows the GMR M6 per-tenant-seed + explicit-`tenant_id`-predicate pattern (no cross-tenant leakage). Worklist filter params round-trip through contract-first regen. |
| F9 | `f9-delegation` | `approval_delegations` table (tenant, delegator, delegate, window, reason, audit; `ends_at > starts_at`, no self-delegation). Eligibility union — active delegation makes delegate eligible wherever delegator is eligible; signoff/verdict records both identities (`on_behalf_of`). Delegate inherits delegator's SoD constraints. Migration `0291`. ADR: delegation. | Integration test: eligibility union correct (delegate can act in delegator's pool during window only); dual identity persisted + surfaced in the manifest; SoD inheritance blocks a delegate acting for the document's author on that document; expired window → ineligible; overlapping windows for the same delegator → both stay active (union semantics); self-delegation rejected at the DB constraint. |
| F10 | `f10-adrs-wiki-gates-live-qa` | 4 ADRs under `wiki/decisions/` (route versioning [written in F2], freeze boundary + review layer + choke-point concurrency, `approval.oversee` + visibility, delegation [written in F9]) — all `accepted`, ADR-status CI gate green; `wiki-curator` dispatch for `wiki/modules/documents.md` + `wiki/architecture/backend-api-structure.md` + ADR 0018 superseded annotations; full gate suite; mandatory live QA walkthrough. | See "Milestone validation definition" below — this feature's acceptance IS the milestone-end gate, run by the `milestone-validator` subagent, not self-certified here. |

For each feature, "what to validate" above is objectively checkable — a named test, a grep-zero
count, a response code, a build/lint result. No feature closes on "works"/"looks right"; each
feature's own `spec.md`/`evidence.md` restates its row's criteria with the actual test names and
command output.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — every feature row above meets its declared "what to validate", and
   each feature's `spec.md` consumer contract was honored (producer matches consumer, not guessed).
2. **Full-suite gates, re-run from clean state:**
   - `go build ./...` clean.
   - `go test ./...` (full suite) green.
   - Both authz drift/parity lints green (capability tripwire generation pipeline, GMR M2 machinery).
   - `api-lint` clean (contract-first — no hand-added routes).
   - `TestCapabilityRegistrySize` — exactly **+2** over its pre-milestone value (`approval.review`,
     `approval.oversee`), never more, never less.
   - Grep-zero checks: zero `SkipStage` references anywhere in the tree; zero production `COALESCE`
     on the `content_hash_at_submit`-vs-head-revision pattern at the two F6 call sites; zero
     generic `/approval/` tier-1 prefix fallback block in `permissions.go`.
3. **Live QA walkthrough (mandatory — compile ≠ work), against the running local stack**
   (`.\scripts\start-api.ps1`): submit an instance with a review stage → record a review verdict
   `request_changes` (mandatory comment) → author resolves the requested changes and re-submits →
   freeze fires at the review→approval transition (markup gate rejects a dirty buffer if tested,
   passes a clean one; instance-scoped comment gate) → signoff with `signature_meaning` recorded
   and the pin echoed back matching the active-document endpoint → publish verifies the pin →
   visibility 404 matrix exercised as consumer/author/oversee-capability users → worklist filters
   (`stage_kind`, `due_before`, `?scope=oversee`) return the expected sets → cancel with a mandatory
   reason persists it. Separately: an approval-only route (no review stage) freezes at submit.
4. **Workflow-class QA** — backend-api checklist (contract↔handler parity, authz tier-1/tier-2
   coherence, multi-tenant isolation on every new table incl. `approval_delegations`, async/idempotency
   on the surfacer's notification consumer, DB-invariant tripwires on every new trigger/CHECK).
5. **Regression** — this is the first milestone in the `approval-remediation` program; regression
   target is the prior program's terminal state (GMR program, `backend-module-boundary-hardening`)
   — confirm neither program's gates broke.
6. **Root-cause check** — confirm W1 (route immutability) and W2 (floating hash pin) are fixed by
   the structural changes named in the system-impact analysis (versioned routes; freeze + no-fallback
   hash chain), not symptom-patched (e.g. not "widen the trigger condition" or "add another COALESCE
   branch").
7. **No unplanned scope** — anything implemented beyond the 10 feature rows above is recorded with
   rationale in the aggregate diff review.

## Dependencies & constraints

- Depends on: nothing external — this is the program's first milestone. Downstream: M2c (FE screen)
  is blocked on this milestone's operator HS-1 approval; do not start M2c work, and do not touch
  `frontend/apps/web` beyond regenerated `api-types`, until M2b passes and is approved.
- Feature order is F1→F10 strictly (each feature depends on the prior — F1's schema underlies
  everything; F9 delegation is intentionally the last separable slice per spec §4 so it cannot
  destabilize the freeze/versioning choke-point work).
- Architectural constraints respected: contract-first (every route change lands in
  `api/openapi/v1/openapi.yaml` + `oapi-codegen` regen before handlers); expand/contract migrations
  only (no destructive drops mid-milestone); H-PRE-1 (no authz-recording read inside a lock-holding
  tx — freeze and stage-transition CAS sections respect this); `authz.Require` needs a writable tx
  (G1); watchdog stays alert-only (ADR 0068) — the new SLA surfacer is new machinery beside it, not
  a change to it; `testdb` factory is the canonical framework for every new DB integration test;
  no `.env` reads/prints/commits; PowerShell scripts for local startup only.
- Runtime-truth corrections above are binding — they supersede the plan document's text on schema
  names/locations wherever the two disagree.

## Applicable hard-stops

- **HS-1** — this milestone's close (validator PASS) triggers an operator review gate; no start of
  M2c and no push without explicit approval.
- **HS-2** — if any feature's fix implies redesign outside `documents/approval`/`iam` (e.g. a shared
  API consumer outside this module needs a breaking change, or the freeze boundary needs to reach
  into `controlleddocuments` lifecycle semantics beyond triggering existing transitions) — stop,
  report the boundary, do not symptom-patch.
- **HS-3** — if a prerequisite (build/runnable/auth-session/route/contract truth) fails mid-feature,
  repair the prerequisite first via the relevant runtime-contract-prereq skill, then resume.
- **HS-4** — validator FAIL opens a named fix feature (`f10.x-<slug>` or a numbered fix slug),
  re-runs that feature's lifecycle, re-dispatches the validator. Milestone stays active.
- **HS-6** — any scope drift (e.g. the eigenpal suggestion-mode capability risk noted in the
  system-impact analysis bleeding into this backend milestone, or FE changes beyond regenerated
  `api-types`) — stop, surface the deviation, replan before continuing.
