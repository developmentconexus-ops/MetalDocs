# Mission Terminal Acceptance — authoring §8 + the validator charter

The terminal acceptance is the mission's **definition-of-done, written before any milestone runs**. It is
the single most important section of `mission.md`, because it is the bar the whole program is judged
against — and a bar written *after* the work is no bar at all. Read this in Phase 3 when filling §8, and
again before dispatching the `mission-validator` at the end.

## Why up front

If "done" is defined after the milestones, the definition bends to fit whatever got built. Writing it first
is the same discipline `milestone` applies to a feature (consumer-contract before code), lifted to the
program: you commit to what success *is* while you're still honest about it, then build toward it. The
operator approves this bar in Phase 4 — it is part of what they're signing off on.

## The five fields (all required)

1. **Pass bar — "the mission shall be X".** A measurable end-state, not an adjective. Good: "the three
   formerly-C dimensions all ≥ A−, 0 new Critical/Major, defect classes H-D and H-G at 0." / "feature X is
   live end-to-end, its acceptance suite green, 0 regressions in the touched consumers." Bad: "the code is
   clean", "the feature works".
2. **What to validate.** The concrete checklist of conditions that *together* mean the bar is met. Each item
   objectively checkable. This is what the validator will tick off.
3. **How to validate.** The method and the exact shape of it: a multi-agent re-audit fan-out (with the
   dimension list and the skeptic-per-finding rule), a full E2E run (with the commands), an acceptance suite
   (named tests), CI grep-guards ("0 matches for the old pattern"). Be concrete enough that a fresh judge
   could execute it without guessing.
4. **Who validates.** Always the independent `mission-validator` subagent. Name it; link its charter.
5. **On miss (HS-5).** What happens when the bar isn't met: the missed criteria become a bounded remediation
   micro-milestone run through `milestone`, then the validator is re-dispatched; the operator decides
   continue vs replan at each loop. State it so the failure path is already agreed.

## Match the validation to the mission type

| Type | Typical terminal validation |
|------|-----------------------------|
| remediation | Independent re-audit fan-out re-grades the touched dimensions and re-measures the closed defect classes; pass bar = dimensions ≥ target grade AND classes at 0. |
| greenfield-build | Acceptance/E2E suite proves the spec's capabilities live; pass bar = suite green AND the §2 goals demonstrably met at runtime. |
| enhancement | Regression suite proves existing consumers still work AND a new-capability acceptance proves the enhancement; pass bar = 0 regressions + new acceptance green. |
| migration | CI grep-guard proves 0 old-pattern sites remain AND a behavior-parity check; pass bar = census fully drained + parity proven. |

## The `mission-validator` charter (what the agent is and isn't)

The terminal gate is run by a **fresh, independent** subagent (`.claude/agents/mission-validator.md`) for
the same reason `milestone` uses `milestone-validator`: the session that built the program is biased toward
passing it. The validator:

- **Judges and writes one file only** — `qa/mission-validation.md` (PASS/FAIL + per-criterion evidence).
- **Verifies, sized to its tools.** It has no `Agent` tool and cannot fan out. **Deterministic** §8 methods
  (suites, CI greps, single-pass review) it **runs itself** with `Bash`/`Grep`, not trusting the milestones'
  transcripts. **Fan-out** §8 methods (a multi-dimension re-audit) are run by the **main session**, which
  hands the validator the artifact to **judge** — and the validator independently **spot-checks** its
  load-bearing claims (re-grep a sample of "0 remaining" sites; re-run a named proof). This split is why §8's
  "how to validate" must say which methods are deterministic and which need a main-session fan-out.
- **Never** edits code, fixes findings, or flips program status. If it finds a gap, it records a finding and
  fails the criterion — it does not repair it. Status flips and the final Grade/Done declaration belong to
  the main session and the operator.
- **Fails closed.** A bar "met" by fixture-only proof, a criterion it couldn't actually run, or a
  suite-green without per-criterion mapping is a FAIL, not a pass.

It runs **once**, only after the last milestone has passed its own `milestone-validator` *and* cleared its
HS-1 operator gate. On FAIL → HS-5. On PASS → the operator signs off and the main session writes the §12
program close-out.
