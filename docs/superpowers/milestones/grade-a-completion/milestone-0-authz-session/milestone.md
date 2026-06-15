# Milestone 0 — Auth / authz / session correctness

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec (drafting)
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the 4 skeptic-confirmed auth/authz/session defects (mission §5: B1–B4) at their shared
root, lifting the **authz** and **sessions** dimensions and removing the latent
premature-access security bug. This is sequenced first (D3) because security/correctness must
land before any FE-visible or systemic work.

**Bar:** each defect carries a regression test that **fails before and passes after** the fix;
no authz regression across the suite; the authz correctness fix (F0.1) lands at the shared authz
predicate, **not** per-caller (ADR 0022 / authz-root-cause memory). Criterion that proves the bar
moved: a future-dated membership is denied by `authz.Require`, and the existing authz test corpus
plus whole-repo `go test ./...` stay green.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F0.1 | `f0.1-authz-effective-from` | `authz.Require` honors `effective_from <= now()` (matching `ResolveEligibleActors`), at the shared authz layer — `iam/authz/authz.go:123`. Fix the predicate once, not per-caller. | Integration test: a **future-dated** membership is **denied**; a **current** membership is granted; existing authz tests stay green. |
| F0.2 | `f0.2-manual-code-create-identity` | The manual-code controlled-document create branch seeds the tx identity so non-admin creates pass the PEP/PDP — `controlleddocuments/application/service.go:173`. | Integration test: a **non-system-admin** manual-code create **succeeds**; the **system-admin** path still works. |
| F0.3 | `f0.3-tenant-grade-view` | `CapDocumentView` no longer narrows a tenant-grade view to area-grade when an area code is present — `documents/approval/application/read_service.go:68`. | Test: a **tenant-role-only** viewer can read a document that carries a real area code. |
| F0.4 | `f0.4-changepassword-cookie` | Self-service `ChangePassword` emits an expired session cookie (mirrors `AdminResetPassword`) — `auth/delivery/http/handler.go:153`. | Handler test: the response **sets an expired cookie**; sessions are revoked. |

For each feature, "what to validate" is objectively checkable — a named test that passes and an
observed runtime behavior. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What that gate
enforces for this milestone:

1. **Per-feature acceptance** — every feature above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored.
2. **Workflow-class QA checklist** — [`wiki/quality/backend-api-qa-checklist.md`](../../../../wiki/quality/backend-api-qa-checklist.md)
   with an **authz-correctness lens** (deny-by-default, no privilege widening, no symptom-patch).
3. **Regression** — whole-repo `go test ./...` green; this is M0 so there is no prior milestone to
   regress, but the **existing authz/session test corpus** must stay green.
4. **Quality-bar / root-cause check** — F0.1 is confirmed fixed at the **shared authz predicate**
   (`authz.Require`), not patched per-caller; the other three are fixed at their named site without
   widening any grant.
5. **No unplanned scope** — anything implemented beyond these four features is recorded with rationale.

## Dependencies & constraints

- Depends on: nothing (M0 is first). HEAD is `5ce0cffb`; the audited findings were at `02ed1c24` with
  no Go diff since (mission §4), so every `file:line` above is current — re-verify at feature start.
- Architectural constraints respected:
  - **Authz root-cause (ADR 0022 / authz-root-cause-over-symptom memory):** never symptom-patch authz;
    F0.1 fixes the shared predicate.
  - **H-PRE-1 advisory-lock hazard:** an authz-recording read must not be called inside a lock-holding
    atomic tx — keep any new read off-tx (relevant if F0.2's identity seed touches authz recording).
  - **No schema/migration redesign** (mission Non-Goals) — these are code-level correctness fixes.
  - **Skill routing:** backend HTTP/handler/contract → `metaldocs-backend-api`; prereq repair →
    `runtime-contract-prereq`.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | This milestone's boundary — operator review gate after the validator PASS; no next milestone / no merge without approval. |
| HS-2 | If fixing F0.1/F0.2 implies redesigning the shared authz/PEP-PDP model beyond honoring `effective_from` / seeding identity — stop, report the boundary + minimum prerequisite plan, do not symptom-patch. |
| HS-3 | If a prerequisite boundary fails (build / runnable / auth-session / route / contract truth) — repair via `runtime-contract-prereq`, rerun the failed checkpoint, resume the feature. |
| HS-4 | If `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | If a fix uncovers a finding F5.1 missed, or scope drifts off these four features — stop, surface the deviation, replan before continuing. |
