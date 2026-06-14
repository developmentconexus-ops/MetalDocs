# Milestone <n> — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** <date>  ·  **Verdict:** see C7.
> Run only after every feature is closed (each has a complete `evidence.md`). The validator judges and
> writes this file; the **main session flips status only on a PASS**. The validator never edits code,
> fixes findings, or flips status.

## C1 — Spec & plan conformance (per feature)

Each feature's evidence acceptance matches its `spec.md` Validation Gate; the **consumer contract was
honored** (producer matches consumer, not reverse); non-goals respected.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F<n>.1 | ✅/❌ | ✅/❌ | ✅/❌ | <link / note> |

## C2 — Gates re-run, isolated

Each feature's named tests + proof commands **re-run by the validator from clean state** (not trusted
from the evidence transcript).

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F<n>.1 | `<cmd>` | <key line> | ✅/❌ |

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff reviewed as one unit. Duplication / split-brain / dead code / one feature
breaking another / guessed contract.

- Findings: <none, or list with file:line>
- Staff-engineer bar met? ✅/❌

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (<name>) | pass / pass-with-defers / fail | |
| Regression vs prior milestones | <all still pass / which broke> | |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| <bar> | <state> | <state> | <focused proof — not symptom-patch> |

- Could it be built better? <retrospective note → defer / next-milestone input, or "no">

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence
- [ ] Fixture/mock passed off as real-provider proof
- [ ] Consumer contract guessed rather than read from the consumer
- [ ] Split-brain (one fact, two sources of truth)
- [ ] Self-judged close / validator edited or fixed code
- [ ] Scope drift (work beyond the spec, no rationale)
- [ ] Symptom-patch (bar "moved" by masking, root cause intact)

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS | FAIL**
- On **FAIL** — failed check(s): <Cx …>; minimum **fix feature** to open: `f<n>.x-<slug>` — <what it
  must do>. Milestone stays **active**; main session does not advance.
- On **PASS** — handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): <pending / approved by … on …>
> - Status flipped in `README.md`: <yes/no — only on PASS>
