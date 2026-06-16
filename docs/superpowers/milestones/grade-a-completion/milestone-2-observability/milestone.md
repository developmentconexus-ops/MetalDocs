# Milestone 2 — Composition / observability

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec (drafting)
> **Authored:** 2026-06-16 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the 4 skeptic-confirmed composition/observability defects in mission §5 (D1–D4) and drive the
**composition** dimension to **≥ A−** on a fresh F5.1 re-audit. After this milestone, the composition
root injects real observability everywhere: the scheduler emits **structured JSON logs** via the
injected `slog.Default()` (no hardcoded text handler), **per-job run/error/skip counters** are
scrapeable on the `/api/v1/metrics` route, **app-level OTel spans** exist on DB calls and critical
request flows (not just the HTTP envelope), and the `/api/v1/metrics` payload carries **DB-pool
stats** with an accurate endpoint name/comment. The consumer of this work is the
operator/SRE reading logs, scraping metrics, and tracing a request through the app.

**Bar (re-measured at close):** the F5.1 §6 composition pass-bar — **composition ≥ A−** on the
terminal re-audit (mission §8). Per-feature: every defect has runtime proof (a JSON log line, a
metrics line, a span tree, a pool-stats key) that fails before and passes after. **0** skeptic-
confirmed new Critical/Major on composition lines a fresh auditor reads. Fix lands at the
**composition root / injection seam**, not via in-package side-effects or symptom patches around the
old text handler / shadow metrics path.

## Appetite + rabbit holes

**Appetite:** 4 features (D1–D4 → F2.1–F2.4), one focused session each. Pragmatic A− bar, not
exhaustive instrumentation (mission Non-Goals: "no gold-plating observability").

**Rabbit holes (do NOT chase in this milestone):**
- **Full distributed-tracing rollout** (every package, every handler, baggage propagation across
  external HTTP). Out of scope — F2.3 is meaningful spans on DB + critical flows only, to a pragmatic
  A− bar. *Reason:* mission §2 Non-Goals.
- **Prometheus exporter / OpenMetrics format redesign.** Out of scope — F2.2/F2.4 wire counters and
  pool stats to the **existing** `/api/v1/metrics` payload shape; no new endpoint, no exporter swap.
  *Reason:* HS-2 boundary; would be an observability architecture redesign.
- **Log schema redesign** (event taxonomy, structured field catalog). Out of scope — F2.1 uses the
  already-injected `slog.Default()`; log keys are whatever that handler already produces.
  *Reason:* HS-2 boundary.
- **Replacing the scheduler library / job model.** Out of scope — F2.1/F2.2 patch the existing
  scheduler. *Reason:* HS-2 boundary; defeats the milestone slice.
- **Touching M3 quality-tail or M4 module-ports sites.** Out of scope — those have their own
  milestones (mission §7, locked order D3). *Reason:* sequencing isolation.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F2.1 | `f2.1-scheduler-slog` | Scheduler uses the injected `slog.Default()` JSON handler instead of a hardcoded `slog.NewTextHandler` at `internal/modules/jobs/scheduler/scheduler.go:131` (mission §5 D1). Consumer: SRE tailing logs expects a JSON line per job run, key-shaped like the rest of the app. | Runtime proof: scheduler log lines are JSON at runtime (sample captured in `evidence.md`); `grep -RIn 'NewTextHandler' internal/modules/jobs/` returns 0; unit/handler tests for the scheduler still pass. |
| F2.2 | `f2.2-scheduler-metrics` | Wire `MetricsSnapshot()` (per-job run/error/skip counters) at `internal/modules/jobs/scheduler/scheduler.go:273` into the existing `/api/v1/metrics` scrape payload (mission §5 D2). Consumer: the `/api/v1/metrics` route caller scrapes per-job counters. | Runtime proof: `curl /api/v1/metrics` payload includes per-job counters with non-zero keys after a forced run (sample in `evidence.md`); handler/integration test asserts the keys exist; no new endpoint added (existing payload shape). |
| F2.3 | `f2.3-otel-app-spans` | Add meaningful app-level OTel spans on **DB calls** + **critical request flows** to a pragmatic A− bar at `internal/platform/observability/otel.go:95` (mission §5 D3, HS-2 candidate). Consumer: operator viewing a trace expects child spans under the HTTP-envelope span (not just the envelope). | Runtime proof: a captured trace tree (exporter to stdout / OTel collector mock — labeled fixture-vs-real per mission §8) shows child spans on at least one DB path and at least one critical request flow (e.g. CD checkpoint create, document publish); spans carry app-meaningful names; integration test asserts the span tree shape for one named flow. |
| F2.4 | `f2.4-metrics-completeness` | Add DB-pool stats (`InUse`, `Idle`, `WaitCount`, etc.) to `/api/v1/metrics` at `internal/platform/bootstrap/api.go:99`; correct the misleading "Prometheus" comment / endpoint name at `cmd/metaldocs-api/permissions.go:95` (mission §5 D4). Consumer: SRE scraping pool stats; reader of the file expects the comment to match reality. | Runtime proof: `curl /api/v1/metrics` payload includes pool-stat keys (sample in `evidence.md`); the comment/endpoint name at the cited line accurately describes the format (no "Prometheus" claim if not Prometheus); handler test asserts pool-stat keys present. |

For each feature, "what to validate" is objectively checkable — a named test that passes plus an
observed runtime artifact (JSON log line, metrics key, span tree). No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — every F2.x above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored (real injected dependency, real metrics
   payload, real span tree — not a fixture-only proof). Fixture vs real-provider proof must be
   labeled per mission §8.
2. **Workflow-class QA checklist** — [`wiki/quality/backend-api-qa-checklist.md`](../../../../wiki/quality/backend-api-qa-checklist.md)
   with an **observability lens**: composition-root injection truth-table reconciled (logger,
   metrics, tracer); no in-package hardcoded handlers / counters; `/api/v1/metrics` payload shape
   reviewed; no shape regression for existing scrapers.
3. **Regression** — whole-repo `go test ./...` green; M0 (auth/authz/session) and M1
   (contract/H-D) gates still pass — re-run `go test ./...` and the report §6 grep commands on the
   H-D class (must remain 0).
4. **Quality-bar / root-cause check** — composition bar re-measured at milestone close: a focused
   read of D1–D4 lines shows the **injection seam** is the fix surface, not a per-call workaround.
   The validator confirms `NewTextHandler` is absent from scheduler, the metrics payload genuinely
   exposes per-job + pool counters at runtime, and the span tree is non-trivial. The terminal
   composition ≥ A− grade is judged at mission §8, not here — but a fail of the root-cause check
   here is a milestone FAIL.
5. **No unplanned scope** — anything implemented beyond F2.1–F2.4 (esp. anything touching M3
   quality-tail or M4 module-ports finding classes) is recorded with rationale or rolled back.
   Rabbit-hole list above is the scope-drift baseline.

## Dependencies & constraints

- **Depends on:** M0 passed (HS-1 approved 2026-06-15) and M1 passed (HS-1 approved 2026-06-15).
  HEAD includes M0 + M1 close commits; mission §5 `file:line` anchors must be **re-verified at
  feature start** (drift expected after M1 codegen regen).
- **Quality goals (top 3, ranked):**
  1. **Composition correctness** (real injected deps end-to-end) > 2. **Runtime observability
     value** (a trace/metric/log a human actually reads) > 3. **Implementation simplicity** (no
     new abstractions; reuse the composition root).
- **Architectural constraints:**
  - **Composition-root injection, not in-package wiring.** Every observability dependency
    (`slog.Default()`, metrics sink, `Tracer`) flows from `cmd/metaldocs-api` / `internal/platform/bootstrap`
    via the constructor seam. No `slog.New(...)` / `otel.Tracer(...)` inside leaf packages.
  - **No `/api/v1/metrics` payload shape regression** for existing scrapers — F2.2 and F2.4 add
    keys; never rename/remove existing ones.
  - **OTel span scope is pragmatic A−**, not exhaustive — DB calls + critical request flows. Gold-
    plating is an HS-2 trigger (see below).
  - **No schema/migration redesign** (mission Non-Goals).
  - **Skill routing:** backend wiring/composition → `metaldocs-backend-api`; prereq repair →
    `runtime-contract-prereq`; module-wiki sync after structural change →
    `metaldocs-module-doc-sync`. **No FE work** this milestone (mission Non-Goals).
- **Risks (named, with owner/mitigation):**
  - **R1 — F2.3 scope creep into a tracing-architecture overhaul.** *Mitigation:* HS-2 enforced;
    span list is bounded by `spec.md` Validation Gate before any code (no spans beyond the named
    flows). *Owner:* feature author at F2.3 spec time.
  - **R2 — `/api/v1/metrics` payload shape drift breaks an existing scraper.** *Mitigation:* F2.2
    and F2.4 contract-test the existing-key set before adding new keys; review diff on the payload
    snapshot test. *Owner:* feature author at F2.2/F2.4.
  - **R3 — Scheduler logger swap masks a latent panic in a job that relied on text-handler quirks
    (e.g. plain stderr write-through).** *Mitigation:* run all scheduled jobs once in a smoke run
    and capture log lines as evidence; if any job log goes silent, HS-3 the prereq. *Owner:*
    feature author at F2.1.
  - **R4 — OTel exporter side-effects (network noise, blocked shutdown) in tests.** *Mitigation:*
    F2.3 uses an in-process / stdout exporter for tests; production exporter unchanged. *Owner:*
    feature author at F2.3.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | This milestone's boundary — operator review gate after the validator PASS; no next milestone (M3) / no merge without approval. |
| HS-2 | If F2.3 grows into a tracing-architecture overhaul (cross-process baggage propagation, exporter swap, span-naming taxonomy redesign); or if F2.2/F2.4 imply a Prometheus/OpenMetrics exporter redesign; or if F2.1 implies a log-schema taxonomy redesign — stop, report the boundary + minimum prerequisite plan, do not symptom-patch. |
| HS-3 | If a prerequisite boundary fails (build / runnable / auth-session / route / contract truth — e.g. metrics route 401s after composition-root reshuffle, scheduler smoke run panics under the new logger) — repair via `runtime-contract-prereq`, rerun the failed checkpoint, resume the feature. |
| HS-4 | If `milestone-validator` returns FAIL — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | If a fix uncovers an observability defect F5.1 missed, or scope drifts off these four features (e.g. an M3 code-quality finding or an M4 boundary finding surfaces) — stop, surface the deviation, replan before continuing. |
