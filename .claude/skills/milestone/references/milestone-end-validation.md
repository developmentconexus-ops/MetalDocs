# Milestone-End Validation — Binding Gate (C1–C7)

> **This is the binding checklist the `milestone-validator` subagent runs at milestone close.**
> It is **rigid** — every check is mandatory; "N/A" is allowed only with a written justification.
> The validator follows this file exactly, then writes its verdict to the milestone's
> `qa/milestone-qa.md`. **Separation of powers:** the validator *judges and writes the verdict only*.
> It never edits source, never fixes findings, and never flips milestone/program status. The **main
> session** flips status — and only on a PASS verdict. A judge that closes its own milestone is a
> forbidden-list violation (C6).

## Inputs the validator must load first

1. The milestone spec — `milestone-<n>-<slug>/milestone.md` (the up-front contract).
2. Every feature's `spec.md`, `plan.md`, `evidence.md` in that milestone.
3. The program `README.md` (status table, prior milestones, terminal acceptance).
4. The governing program spec linked from the README.
5. The aggregate milestone diff (all commits/changes since the prior milestone's close).

If any input is missing or unreadable → **FAIL fast** (cannot judge blind). Name what is missing.

## The two dimensions

Every milestone is judged on **both**, independently:

- **Code-wise** — is the work correct, senior-level, contract-clean, with no split-brain (two
  sources of truth for one fact), no dead code, no guessed contracts?
- **Function-wise / QA** — does it actually do, end-to-end, what the milestone spec and each
  feature plan promised? Fixture-only proof is not end-to-end proof.

A milestone passes only if **both** dimensions pass. A clean diff that doesn't do the job FAILs;
a working result built on a split-brain or a guessed contract FAILs.

## The checklist

### C1 — Spec & plan conformance (per feature)

**C1 binds on artifacts, not on which skill produced them.** The validator does not check whether
`superpowers:brainstorming` / `superpowers:writing-plans` / `superpowers:subagent-driven-development`
were used — it checks that the artifacts those skills are supposed to produce exist with the
required structure. Skill absent + equivalent inline output present = PASS. Skill present + thin
artifact = FAIL.

For **each** feature, verify all of:

1. **`spec.md` exists and is approved before code.** The `Approved before code:` line is filled
   with a date + operator. An empty approval line → **C1 fail** (work started without a contract).
2. **`spec.md` Interview record is populated** — either a Q&A table with at least one row, OR an
   explicit "none needed — why" justification. An empty interview record on a feature that touches
   a contract → **C1 fail** (consumer contract was guessed, violating fail-closed).
3. **Consumer contract declared in `spec.md` was honored** — the producer matches the contract the
   consumer depends on, not the reverse. Cross-check by reading the consumer site referenced in
   `spec.md`'s "Source of truth for the contract" line.
4. **`plan.md` exists and is execution-shaped** — task list, files touched, test strategy, ordering.
   A `plan.md` that is just a restatement of `spec.md` → **C1 fail** (no plan was written, only
   re-spec'd).
5. **`evidence.md` acceptance table matches `spec.md` Validation Gate** row-for-row.
6. **Non-goals respected** — nothing built that `spec.md` declared out of scope (and nothing in the
   milestone's rabbit-hole list).
7. **Any deviation carries a written rationale** in `evidence.md`.

Missing `spec.md` / `plan.md` / `evidence.md` for any feature in `milestone.md`'s Features table →
**C1 fail**, name the missing artifact(s).

### C2 — Gates re-run, isolated
Re-run **each feature's named tests and proof commands yourself**, fresh, from a clean state — do
not trust the evidence file's transcript. Record the actual command + actual output. A test that
passed during implementation but fails on isolated re-run → **C2 fail** (flaky or
environment-coupled is not green).

### C3 — Senior review of the AGGREGATE milestone diff
Review the whole milestone's diff as one unit (not per-feature, which the implementer already saw).
Look for what only shows up in aggregate: duplicated logic across features, a contract defined two
ways, a fact stored in two places (split-brain), dead code left by a superseded approach, a feature
that quietly broke another. Senior-engineer bar: would a staff engineer approve this diff?

### C4 — Platform / workflow-class QA + regression
Run the canonical QA checklist for the milestone's workflow class (backend-api / screen /
workflow-async / release close-out — whichever the milestone declared). **Then run regression
against every previously-completed milestone** — confirm their gates still pass. A new milestone
that regresses an old one → **C4 fail**.

### C5 — Quality-bar re-measure + "could it be built better" retrospective
If the milestone declared a quality bar (a grade, a closed defect class, a cleaned contract),
**re-measure it here** and confirm the **root cause is fixed, not symptom-patched**. Then a short
retrospective: given what is now known, is there a materially better construction? If yes, record it
(it becomes input to the next milestone or a defer) — it does not by itself FAIL the milestone unless
the current construction is actually unsound.

### C6 — Forbidden-list (any hit → FAIL)
- Suite-level "all green" reported **as a pass** without per-feature acceptance mapped to evidence.
- Fixture / mock output **passed off as real-provider** proof.
- A consumer contract that was **guessed** rather than read from the consumer (fail-closed was
  violated).
- **Split-brain**: one fact with two sources of truth.
- **Self-judged close**: the actor that built the milestone also flipped its status, or the
  validator edited/fixed code instead of only judging.
- **Scope drift**: work delivered beyond the milestone spec without a recorded rationale.
- **Symptom-patch**: a declared quality bar "moved" by masking the symptom, root cause intact.

### C7 — Verdict
Write **PASS** or **FAIL**, with per-check (C1–C6) evidence — the actual commands and outputs, not
adjectives. On **FAIL**: name the specific failed check(s) and open a **named fix feature**
(`f<n>.x-<slug>`) describing the minimum work to clear it. The milestone stays **active**; the main
session does **not** advance. On **PASS**: state it plainly; the main session may then flip status
and present the HS-1 operator gate.

## Output contract

The validator writes exactly one file — `qa/milestone-qa.md` — using
`templates/milestone-qa.md`. It writes nothing else. It returns to the main session a one-paragraph
summary ending in the verdict token `VERDICT: PASS` or `VERDICT: FAIL`.
