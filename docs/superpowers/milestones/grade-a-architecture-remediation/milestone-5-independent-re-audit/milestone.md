# Milestone 5 — Independent Re-Audit (authoritative Grade-A gate)

> **Program:** grade-a-architecture-remediation  ·  **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` (§6 M5, §7 HS-5)
> **Status:** Spec — awaiting operator Phase-2 agreement before F5.1
> **Authored:** 2026-06-15 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M5 is, **which features** it
> contains, **what each feature implements**, and **what gets validated**. It contains **no
> execution steps** — the "how" of F5.1's fan-out lives in `f5.1-*/plan.md`. The end-of-milestone
> QA (`qa/milestone-qa.md`) validates M5 against *this* document.

## Objective

Prove the backend reached **Grade A** by an **independent, fresh multi-agent re-audit** — the
program's authoritative sign-off. M0–M4 closed the four Grade-A blockers and the H-D / H-G **classes**
via owning-module ports, root-caused (not symptom-patched), each with evidence. M5 does **not** re-do
that work; it re-reads the post-M4 tree from scratch and asks whether the bar is independently
observable. Per the governing spec, **only M5 declares Grade A** — the M0–M4 focused slices are
indicative, never authoritative.

**Bar this milestone proves (governing spec §6 pass bar):**
1. The 3 formerly-C dimensions — **module-boundaries/DDD**, **contract/API**, **composition/observability** — all **≥ A−**, none below.
2. **0** new Critical/Major findings.
3. **H-D class = 0** tri-source contract drift; **H-G class = 0** reach-without-a-port **and** 0 hardcoded-domain-state.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F5.1 | `f5.1-full-re-audit` | Re-run the **full 10-dimension multi-agent audit method** on the post-M4 branch via the **Workflow** tool — independent fresh readers per dimension, **one adversarial skeptic per finding** (refute-by-default), dedup + severity-classify, written report under `wiki/backend/_artifacts/`. No source edits in this feature. | Audit report exists and is reproducible: every dimension has a graded verdict with cited `file:line` evidence; every surfaced finding carries a skeptic verdict (confirmed/refuted); the 3 formerly-C dimensions each carry an explicit ≥A−-or-not call; H-D and H-G class counts are re-measured by grep + build/test, not asserted. |
| F5.2 | `f5.2-micro-wave-loop` | **Conditional** — only if F5.1 finds any of the 3 dimensions `< A−`, any new Critical/Major, or a non-zero H-D/H-G class. Each such finding becomes a **bounded remediation micro-wave** (its own feature lifecycle: spec→gate→TDD→evidence), then F5.1 is re-run on the patched tree. Loop until the pass bar is met. **Operator decides continue vs replan at each loop iteration (HS-5).** | Each micro-wave has its own evidence row (root cause fixed, not symptom-patched); the subsequent re-audit shows the previously-failing dimension/finding cleared; the loop terminates with the §6 pass bar met. If a dimension loops **> 2×**, it is escalated as a design-boundary signal (HS-2/HS-5), not patched further. |
| F5.3 | `f5.3-operator-signoff` | Capture the **operator's final Grade-A declaration** and write the program close-out / reconciliation in the program `README.md`. The operator owns the sign-off; the agent never self-declares Grade A. | Program README reconciliation complete: every M0–M4(+M4b/M4c) feature has an evidence row; zero unplanned scope merged; every bounded defer has a written trigger + owner; M5 re-audit linked and passing §6; operator sign-off line filled (date + name). |

For F5.1/F5.2, "what to validate" is objectively checkable: a written audit report with graded
dimensions + cited evidence, grep/build/test class counts at 0, and (for micro-waves) a passing
re-audit on the patched tree. "Looks like A" is not acceptance — a cited, reproducible verdict is.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges and
writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the binding
C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. For M5 the gate
enforces:

1. **Per-feature acceptance** — F5.1 produced a reproducible, cited, adversarially-verified audit; any
   F5.2 micro-waves each closed at root cause with evidence and a clearing re-audit; F5.3 reconciliation
   + operator sign-off are present.
2. **Independence** — the re-audit was a fresh read (Workflow fan-out, skeptic-per-finding), **not** a
   re-assertion of M0–M4 evidence. The validator confirms the audit method (10 dimensions, adversarial
   verification) was actually run, not summarized from prior milestones.
3. **§6 pass bar met and re-measured** — 3 dimensions ≥ A−, 0 new Critical/Major, H-D = 0, H-G = 0 —
   the validator independently re-greps the class counts rather than trusting the report's number.
4. **Regression** — M0–M4 (and M4b/M4c) gates still pass; `go test ./...` green; no milestone regressed.
5. **No unplanned scope** — any source change beyond an authorized F5.2 micro-wave is recorded with rationale.
6. **Terminal reconciliation** — the program close-out checklist in `README.md` is fully satisfied.

## Dependencies & constraints

- **Depends on:** M0–M4 passed (M4 re-validated PASS 2026-06-15), M4b cluster-drop (`071931c9`) applied,
  M4c test-fixture framework merged. Post-M4 branch is the audit target.
- **Builds on the prior independent audit:** `wiki/backend/_artifacts/architecture-audit-2026-06-13.md`
  (the 06-13 read that re-flagged H-D/H-G as classes). F5.1 is the *re*-audit against that baseline.
- **Token discipline (governing spec §10):** M5 is the **one** milestone that uses a full Workflow
  fan-out — it is where fan-out pays. Worker-model balancing applies (sonnet for reasoning/review,
  haiku for mechanical, **never fable** workers, **≤15 concurrent** agents).
- **Architectural constraints:** F5.1 is read-only (no source edits). F5.2 micro-waves respect all
  CLAUDE.md hard-stops, H-PRE-1 advisory-lock rules, and contract-first regen order if any contract is touched.
- **No-merge:** the operator merges; the agent never does.

## Applicable hard-stops

- **HS-1** (every milestone boundary) — M5's own close gate: validator PASS → operator review → operator
  signs off Grade A. **No merge without approval.** This is also the program's terminal gate.
- **HS-5** (M5-specific) — re-audit finds any dimension `< A−`, any new Critical/Major, or non-zero
  H-D/H-G class → bounded remediation micro-wave (F5.2) + re-audit loop; **operator decides** continue
  vs replan at each iteration.
- **HS-2** — if an F5.2 micro-wave's fix implies a redesign outside its boundary (shared-API,
  cross-module auth model, storage/provider, workflow semantics), **stop**, report the boundary +
  minimum prerequisite plan; do not symptom-patch. A dimension that loops **> 2×** is treated as an
  HS-2/HS-5 design-boundary signal, not continued patching.
- **HS-3** — a prerequisite boundary (build/runnable/auth-session/route/contract truth) fails during
  the audit or a micro-wave → `runtime-contract-prereq`, repair, rerun the checkpoint, resume.
- **HS-4** — a micro-wave's own close-out finds symptom-patching (root cause not fixed) → replan that
  micro-wave feature, re-run its lifecycle.
- **HS-6** — scope drift / off-plan discovery mid-milestone → stop, surface, replan before continuing.
