# Program: Approval Remediation

> **Governing specs:** `docs/superpowers/specs/2026-07-07-approval-remediation-design.md` (M2b/M2c, ratified, commits e4a0717a, 046f0633, 68a0b3b8) · `docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` (M2d/M3/M4 extension, ratified 2026-07-08, commits ffdde76d, c6ea72f0)
> **System-impact analysis:** `docs/superpowers/analysis/2026-07-07-approval-remediation-m2b-system-impact.md` (🟡 Yellow, no open hard-stop) · M3 requires a fresh `developing-new-work` gate before its milestone.md (new top-level module boundary)
> **Status:** M2b passed (HS-1 approved 2026-07-07); M2c validator PASS with **recorded deviation** (DecisionFooter violated the F4 `stage_kind` contract — confirmed by operator-driven UI QA 2026-07-08; closure delegated to M2d's gate); program extended with M2d/M3/M4 per the 2026-07-08 coherence spec
> **Owner / operator:** MetalDocs operator (Leandro)

Remediate the approval workflow system to professional eQMS grade: separate review
(collaborative) from approval (signature) stages, a real content-freeze boundary with a
no-fallback canonical hash chain, versioned-immutable route definitions, explicit
capabilities (`approval.review`, `approval.oversee`) replacing the generic tier-1 prefix
fallback, signature meaning (21 CFR 11.50(a)(3)), SLA surfacing, visibility gating, and
delegation-of-authority — then, in a second milestone, replace the standalone approval
cockpit with an editor-shell + sidebar surface. Extended 2026-07-08 (coherence spec) with
three milestones: M2d (server viewer-facts contract + single mode-adaptive document screen —
closes the M2c deviation), M3 (approval kernel extracted to a top-level module, templates
unified onto it, superseding ADR 0072), M4 (BPMN-aligned ActorSelector union for stage
assignment). Terminal acceptance for the extended program is M4's validator PASS on the
extracted kernel serving documents AND templates, UI-driven live QA at each FE-facing gate,
every milestone operator-approved (HS-1).

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 2b | `milestone-2b-approval-kernel-backend` | Backend workflow/permission/contract remediation (W1-W13, P1-P8) inside `documents/approval` + `iam` | passed (operator HS-1 approved 2026-07-07) | [PASS](milestone-2b-approval-kernel-backend/qa/milestone-qa.md) |
| 2c | `milestone-2c-approval-screen-fe` | Cockpit → editor-shell + sidebar reuse, worklist single destination, suggestion UX | validator PASS — deviation recorded (F4 contract violation), closes at M2d gate | [PASS](milestone-2c-approval-screen-fe/qa/milestone-qa.md) |
| 2d | `milestone-2d-workflow-coherence-fe` | Server `viewer` facts + one workspace-mode selector + single mode-adaptive screen (`/documents/:id`); closes the M2c deviation at the root | in-progress (operator-approved 2026-07-08; executing in dedicated session) | — |
| 3 | `milestone-3-approval-kernel-extraction` | Approval kernel → top-level `approval` module (subject-generalized routes); templates unified onto it; supersedes ADR 0072 | planned (requires `developing-new-work` gate first) | — |
| 4 | `milestone-4-actor-selectors` | Stage assignment generalizes to ActorSelector union (named_user / role_in_fixed_area / role_in_document_area / submit_choice) in the extracted kernel + route-builder UI + submit picker | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) /
`blocked` (hard-stop open). The **Gate result** column links the milestone-validator's
verdict (`qa/milestone-qa.md`); `passed` requires a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| M2c F0 (2026-07-07) | HS-2 | Plan F0-Step-3 wired the docx-XML markup gate (`ScanForUnresolvedMarkup`) into `executeFreeze`. Runtime truth: MetalDocs content = form-data JSON; docx only rendered externally (markup-free by construction); reviewer suggestions/comments persist as structured JSON (`document_comments`), never `w:ins`/`w:del`; in-tx blob fetch is forbidden (`reconstruct_service.go` precedent). The docx-XML scan checks a state that cannot occur at freeze. | Surfaced to operator; **path A** chosen (global maximum): freeze integrity already complete via the hash chain (`FrozenContentHash` pin+echo) + `HasUnresolvedInstanceComments` server gate + F6 client clean-buffer. `ScanForUnresolvedMarkup` (+test) deleted as misdirected dead code. Server-authoritative suggestion-resolution gate registered as a **bounded defer** (see close-out list). F0 re-scoped to wire enum `changes_requested` + route `stage_kind` + regen. |

## Program close-out / reconciliation

Fill in only when the last milestone (2c) has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner (incl. W12 parallel-stage DAG routing, spec §10; **server-authoritative suggestion-resolution freeze gate** — trigger: a requirement that the *server* must independently prove no unaccepted suggestion exists, at which point model suggestions as structured resolvable state like `document_comments` and gate freeze on that JSON state; today suggestion resolution is client-authoritative via eigenpal + caught by the hash chain — HS-2, F0).
- [ ] Terminal acceptance passed — link the evidence.
- [ ] Operator sign-off: <date / name>
