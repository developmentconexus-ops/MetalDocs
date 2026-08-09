# `required` observed green and red on a real PR (plan Task 10)

**PR:** #97 (`ci/restructure-phases-1-5` → `main`)
**Date observed:** 2026-08-09
**Verdict:** `required` was observed **green for the right reason** (every gating job succeeded) and **red for the right reason** (a gating job failed while `test-integration` was skipped without running). Phase 4 (ruleset swap) is unblocked.

## Green run

- Head SHA: `40d8eacc`
- Run: https://github.com/developmentconexus-ops/MetalDocs/actions/runs/31300791707 — conclusion **success**
- Check-run sweep on the head SHA: 42 checks total; every context in ruleset 20560142's 21 required contexts reported **success**; `required` reported **success**. The only failures were three non-required, known-red contexts:
  - `Perf suite (reduced — PR gate)` — reduced perf gate, red by design pending baseline work (not a required context)
  - `hardening` (gosec/govulncheck) — advisory until the Phase 4 promotion this report gates
  - `E2E smoke (approval flows)` — broken `go run ./cmd/api` invocation in the legacy workflow; nightly.yml carries the fixed copy; workflow is deleted in Phase 5

## Red run

- Head SHA: `1304ba79`
- Run: https://github.com/developmentconexus-ops/MetalDocs/actions/runs/31300439165 — conclusion **failure**
- `verify` **failed** (real defects from the gosec-triage diff: line-pinned allowlist drift + governance comment-only false positives + gofmt misalignment — all subsequently fixed in `40d8eacc`)
- `test-integration` reported **skipped** with `completedAt < startedAt` (zero duration — it never started and consumed no runner minutes)
- `required` reported **failure**: the jq gate requires literal `"success"` from all four jobs — `test-integration` reporting `skipped` fails the gate exactly like `verify`'s `failure` does; this run failed on both counts at once

**Deviation from the plan's Step 4, recorded:** the plan prescribed a deliberate gofmt break to manufacture a red run. A *real* red run occurred first (the gosec-triage diff tripped three genuine gates, gofmt among them) and exercised exactly the semantics Step 4 exists to prove — a cheap gate failing, the staging edge holding, and `required` red for the right reason. That evidence is strictly stronger than a manufactured break, so no artificial red commit was pushed. Step 5's "green returns after revert" is the green run above, produced by fixing the real defects rather than reverting a fake one.

## The four timings

| Measurement | Value |
|---|---|
| Workflow start → `test-integration` start (green run) | **6m 38s** (07:17:01Z → 07:23:39Z) — the staging edge: the 12-minute suite waits for lint-go + verify + security |
| `test-integration` duration (green run) | 12m 25s (07:23:39Z → 07:36:04Z) |
| Total wall-clock, green run | 19m 13s (07:17:01Z → 07:36:14Z) |
| Pre-restructure baseline (legacy `test-full.yml`, same push, starts at minute 0 unconditionally) | 12m 00s (07:17:01Z → 07:29:01Z) |

**Answer to the original complaint:** the restructure trades ~7 minutes of green-path latency (19m13s vs 12m00s, because the suite now waits for the cheap gates) for zero suite-minutes burned on broken PRs. On the red run the whole workflow finished in **6m 26s** with `test-integration` unstarted; the legacy layout would have run the full 12-minute suite to completion against a PR that already failed lint. Broken PRs are the common case during iteration; that is where the minutes were being lost.

## Statement

`required` has now reported both **success** (all gates green, 40d8eacc) and **failure** (gate red with `test-integration` skipped-not-run, 1304ba79) on a real PR, each for the correct reason. The aggregator requires literal `"success"` from all four jobs — no allowance for `skipped` from any of them, including `test-integration` — and that exact-set-equality logic was exercised live in both directions. The precondition spec §8 sets for the ruleset swap is met.
