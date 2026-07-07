# Program: Approval Remediation

> **Governing spec:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` (ratified, commits e4a0717a, 046f0633, 68a0b3b8)
> **System-impact analysis:** `docs/superpowers/analysis/2026-07-07-approval-remediation-m2b-system-impact.md` (🟡 Yellow, no open hard-stop)
> **Status:** Milestone 2b validator PASS, HS-1 operator gate pending
> **Owner / operator:** MetalDocs operator (Leandro)

Remediate the approval workflow system to professional eQMS grade: separate review
(collaborative) from approval (signature) stages, a real content-freeze boundary with a
no-fallback canonical hash chain, versioned-immutable route definitions, explicit
capabilities (`approval.review`, `approval.oversee`) replacing the generic tier-1 prefix
fallback, signature meaning (21 CFR 11.50(a)(3)), SLA surfacing, visibility gating, and
delegation-of-authority — then, in a second milestone, replace the standalone approval
cockpit with an editor-shell + sidebar surface. Terminal acceptance for the program is
M2c's live-QA-verified FE screen sitting on a milestone-validator-passed M2b backend, both
gated by the operator (HS-1) at each boundary.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 2b | `milestone-2b-approval-kernel-backend` | Backend workflow/permission/contract remediation (W1-W13, P1-P8) inside `documents/approval` + `iam` | passed (validator PASS; HS-1 pending) | [PASS](milestone-2b-approval-kernel-backend/qa/milestone-qa.md) |
| 2c | `milestone-2c-approval-screen-fe` | Cockpit → editor-shell + sidebar reuse, worklist single destination, suggestion UX | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) /
`blocked` (hard-stop open). The **Gate result** column links the milestone-validator's
verdict (`qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| | | | |

## Program close-out / reconciliation

Fill in only when the last milestone (2c) has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner (incl. W12 parallel-stage DAG routing, spec §10).
- [ ] Terminal acceptance passed — link the evidence.
- [ ] Operator sign-off: <date / name>
