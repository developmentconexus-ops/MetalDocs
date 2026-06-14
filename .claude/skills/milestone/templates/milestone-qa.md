# Milestone <n> — QA & Verification

> **Validates against:** `../milestone.md` (the up-front milestone spec)
> **Run:** <date>  ·  **Verdict:** PASS | FAIL (→ hard-stop)
> Run only after every feature in the milestone is closed (each has a complete `evidence.md`).

## 1. Per-feature acceptance

Every feature meets the "what to validate" its milestone-spec row declared.

| Feature | Acceptance criteria | Pass? | Evidence (cmd / observed) |
|---------|---------------------|-------|---------------------------|
| F<n>.1 | <from milestone.md> | ✅/❌ | |
| F<n>.2 | | | |

## 2. Workflow-class QA checklist

Run the canonical checklist(s) named in `milestone.md` (backend-api / screen /
workflow-async / release close-out). Record outcome, not just "ran".

| Checklist | Outcome | Notes |
|-----------|---------|-------|
| <name> | pass / pass-with-defers | |

## 3. Regression (previously-completed milestones)

| Prior milestone | Re-checked how | Still passing? |
|-----------------|----------------|----------------|
| M<k> | <cmd / smoke> | ✅/❌ |

## 4. Quality-bar / root-cause check (if applicable)

The bar this milestone claimed to move, re-measured. Confirm **root cause fixed, not
symptom-patched** — name the class/instance and show the closing evidence.

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| <e.g. defect class F1> | open | closed | <focused audit slice link> |

## 5. Scope integrity

- [ ] No unplanned scope (anything beyond `milestone.md` is recorded with rationale).
- [ ] Every bounded defer has a written trigger and owner.

## Verdict & next step

- **Verdict:** PASS → present at HS-1 operator gate. / FAIL → raise HS-4 (replan feature)
  or HS-6 (replan milestone); do not proceed.
- **Operator gate (HS-1):** <pending / approved by … on …>
