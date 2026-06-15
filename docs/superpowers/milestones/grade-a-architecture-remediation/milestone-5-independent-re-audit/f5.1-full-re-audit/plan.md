# F5.1 — Plan (how the re-audit runs)

## Engine

A single **Workflow** fan-out (the one milestone authorized for it, governing spec §10). Pipeline:

1. **Phase `Audit`** — 10 dimension auditors run concurrently (one per scorecard dimension). Each reads
   the **actual code at HEAD `02ed1c24`** (not wiki claims), grades A–F vs industry standards, and
   returns a structured finding set. The 3 formerly-C dimensions additionally return an explicit
   `meetsAMinus` boolean with cited evidence.
2. **Phase `Verify`** — every **Critical/Major** finding is sent to an independent adversarial skeptic
   (refute-by-default: "default to refuted unless the code proves it real"). Skeptic returns
   confirmed / downgraded / refuted + reasoning. Minor findings are recorded, not skeptic-gated.
3. **Phase `Class-counts`** — two dedicated grep-backed agents re-measure **H-D** (handler emits ⊄
   declared OpenAPI, or OpenAPI ⊄ FE codegen) and **H-G** (cross-module raw SQL vs another module's
   owned table; hardcoded domain-state). They return the exact commands + counts so the validator can
   reproduce. These run in parallel with the audit.
4. **Phase `Synthesize`** — one agent assembles `wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md`:
   10-row scorecard vs 2026-06-13 baseline, per-finding skeptic verdicts, re-measured H-D/H-G, and the
   §6 PASS / micro-wave-needed verdict.

## Models (token discipline + workflow-model-balancing)

- Dimension auditors, skeptics, synthesis → **sonnet** (review-grade reasoning).
- H-D / H-G class-count agents → **sonnet** (must read code, not just count).
- **Never fable** workers. Concurrency ≤ the cap (10 auditors + 2 counters < 15).

## Adversarial discipline

Skeptic prompt defaults to **refuted** — a finding survives only if the code, at the cited line, proves
it. This is what makes "0 new Critical/Major" trustworthy rather than a raw-finding tally.

## Output → acceptance

The synthesized report is the F5.1 deliverable. `evidence.md` records the Workflow run, the report path,
the scorecard delta, and the verdict (PASS → F5.3; any miss → F5.2 micro-wave per HS-5). The
milestone-validator independently re-greps H-D/H-G to confirm 0/0.

## Bound

If a dimension or the class-counts come back ambiguous (auditor + skeptic disagree, or a count is
non-zero), that is **not** patched here — it surfaces to the operator as an F5.2 trigger (HS-5).
