# Milestone 10 — Terminal Remediation (HS-5)

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` §8 (Terminal acceptance) + §9 HS-5
> **Status:** Executing
> **Authored:** 2026-07-06 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M10 is, **which features** it contains,
> **what each feature implements**, and **what gets validated**. No execution steps — the "how" lives
> in each feature's `plan.md`. The close QA (`qa/milestone-qa.md`) validates M10 against *this* file and
> the binding `validation-contract.md` (D4).

## Why this milestone exists (HS-5 provenance)

The mission's terminal acceptance (`mission.md §8`) ran on 2026-07-06 (`qa/mission-validation.md`):
a fresh 10-dimension re-audit (`qa/terminal-reaudit-2026-07-06.md`) plus the independent
`mission-validator`. Verdict: **FAIL** on §8(a) — *"CONFIRMED on every in-scope dimension"* — because
three dimensions closed as **DEBT** rather than CONFIRMED:

- **Dim 9 (Observability):** `/metrics` is mounted on the **same TCP port** as the public API
  (`apps/api/cmd/metaldocs-api/main.go:851-864`). Auth is bypassed by route (rootMux ahead of the
  chain), but the endpoint is **not network-isolated** — Docker Compose publishes that port to the
  host (`deploy/compose/docker-compose.yml:120`), so any deployment that exposes `APP_PORT` beyond
  localhost exposes `/metrics` unauthenticated. New DEBT introduced by M8 F8.3.
- **Dim 10 (Governance):** M9 F9.1 established the ADR status-field rule (≤3 lines / ≤400 chars) and
  did a one-time sweep, but **no CI gate enforces it** — `documentation-governance.md:33-34`
  explicitly leaves CI wiring "optional." A convention without a structural gate is precisely the
  meta-defect class this whole mission was built to eliminate. New DEBT introduced by M9.
- **Dim 8 (Testing/QA):** `check-test-discipline.sh` is **RED at HEAD** (4 violations). Three predate
  the mission; **one is mission-introduced** — M9 F9.5 renamed `templates/repository/` →
  `templates/infrastructure/` but left the R2 allowlist path stale at
  `check-test-discipline.sh:59`, so a previously-sanctioned RLS probe now trips the gate.

HS-5 (`mission.md §9`) prescribes exactly this response: *a bounded remediation micro-milestone via
`milestone`, then re-dispatch [mission-validator]; operator decides continue vs replan at each loop.*
The operator approved proceeding on 2026-07-06 ("Go"). M10 is that bounded micro-milestone.

## Objective

Turn the three DEBT dimensions to **CONFIRMED** by closing each gap **structurally** (not by
documentation or one-off patch), so a re-run of `mission-validator` against `mission.md §8` passes.
Consistent with the mission thesis — *convert every discipline-dependent invariant into a
structurally-enforced one* — every fix installs or repairs a **gate/boundary**, never a convention:

- **Dim 9 → CONFIRMED:** the public API listener **cannot** serve `/metrics` — the endpoint lives on
  a **dedicated listener** on a separate address, so isolation is a property of the process topology,
  not of ops discipline or ingress config.
- **Dim 10 → CONFIRMED:** the ADR status-field rule is enforced by a **blocking CI gate** with
  recorded negative proof (fails on an over-budget status block) and positive proof (clean tree).
- **Dim 8 → CONFIRMED:** `check-test-discipline.sh` is **GREEN at HEAD** — all 4 violations resolved
  at root (repair-class per `wiki/quality/legacy-test-policy.md`), with the mission-introduced stale
  allowlist path corrected; the two remaining §8 uncovered-MUST defers (REQ-SEARCH-1, REQ-SEC-3) are
  ratified as bounded, absent-feature defers with named triggers (they are roadmap features, not
  cleanup — outside this micro-milestone's appetite).

## Appetite & rabbit holes (named refusals)

Appetite: **half a session.** All three fixes are mechanical/bounded; none requires an architecture
decision. Explicit rabbit holes — do **not** enter:

- **No** redesign of the observability stack, metric families, or the Prometheus handler itself
  (Dim 9 is a *listener topology* fix only).
- **No** implementation of REQ-SEARCH-1 (search reindex procedure) or REQ-SEC-3 (OWASP ASVS
  checklist) — these are absent **product features**, not hygiene; ratify as defers, do not build.
- **No** rewrite of the test-discipline rules (R1–R4) or the legacy-test-policy taxonomy — only bring
  the 4 current violators into compliance with the *existing* rules.
- **No** touching `api/openapi/`, DB migrations, authz capability code, or any other dimension's
  already-CONFIRMED surface.
- **No** mass test rewrite — repair exactly the 4 named violators; no drive-by.
- **No** Terminal Acceptance / mission-validator re-run *inside* feature work — that is the milestone
  close step, after all three features and the milestone-validator.

## Features (in order)

| ID | Outcome | Consumer | Owning module / surface |
|----|---------|----------|------------------------|
| **F-R1** | The public API listener cannot serve `/metrics`; a dedicated listener on `METRICS_ADDR` (default `:9090`) serves it unauthenticated; Compose no longer publishes the metrics surface on the public host mapping; DEPLOY documents the split. | Prometheus scraper (infra network) + any operator reading DEPLOY | `apps/api/cmd/metaldocs-api` + `internal/platform/config` + `deploy/` + `ops/DEPLOY.md` |
| **F-R2** | A blocking CI gate fails any PR that introduces an ADR status block > 3 lines or > 400 chars; passes on a clean tree; has recorded negative + positive proof. | CI / every future ADR author | `.github/workflows/governance-check.yml` + `scripts/` (sweep lifted from `documentation-governance.md:30`) |
| **F-R3** | `bash scripts/check-test-discipline.sh` exits 0 at HEAD (4 violations resolved at root); the M9-F9.5-introduced stale R2 allowlist path is corrected; REQ-SEARCH-1 / REQ-SEC-3 ratified as bounded absent-feature defers. | `check-test-discipline.sh` CI gate + future maintainers | `internal/modules/**` (3 test files) + `scripts/check-test-discipline.sh` + defer ledger |

**Order intentional:** F-R1 (production code, highest risk, live proof) first while attention is
freshest; F-R2 (pure CI addition, isolated) second; F-R3 (test-file repairs + ledger) last. No
feature depends on another's output, so any order is *correct* — this order is risk-front-loaded.

## What each feature validates (acceptance)

- **F-R1:** (a) new TDD regression test asserts the **public** composed handler returns **not-200**
  for `GET /metrics` (proving it left the public surface) **and** the dedicated metrics handler
  returns 200 + Prometheus `text/plain` with the three `metaldocs_http_*` families and a `go_`/
  `process_` line, unauthenticated; (b) `go build ./...` exit 0; (c) `go test` for the rewritten
  `TestMetricsEndpoint_*` green; (d) **live drive**: start the API, `curl` the public port `/metrics`
  → not reachable as an unauthenticated scrape (401/404), `curl` the metrics port `/metrics` → 200
  Prometheus; (e) Compose + DEPLOY reflect the split (metrics port not on the public host publish).
- **F-R2:** (a) the gate **fails** on a synthetic over-budget ADR status block (negative proof,
  captured command + output); (b) **passes** on the clean current tree (positive proof); (c) the gate
  is `pull_request`-triggered and blocking (no `continue-on-error`); (d) `documentation-governance.md`
  updated to state the rule is now CI-enforced (removing the "optional future extension" language).
- **F-R3:** (a) `bash scripts/check-test-discipline.sh` exits 0, output `test-discipline: clean`;
  (b) each repaired test still **compiles** (`go vet -tags integration` on the 3 packages) and its
  behavior is unchanged (repair = same assertions, sanctioned primitive); (c) the R2 allowlist path
  correction is shown as the F9.5-rename reconciliation it is; (d) REQ-SEARCH-1 / REQ-SEC-3 defers
  recorded in the evidence ledger with finding + trigger + owner.

## Quality goals (top-3 ranked) + hard constraints

1. **Structural over documentary.** Every fix is a gate/boundary the process enforces, not a rule a
   human must remember. (Dim 9 = process topology; Dim 10 = CI gate; Dim 8 = green gate + corrected
   allowlist.)
2. **Zero collateral.** No already-CONFIRMED dimension regresses; the aggregate M10 diff touches only
   the surfaces named above. Re-run of the 4 deterministic gates (`go build`, `req-trace`, `api-lint`,
   `check-module-boundaries`) stays green.
3. **Honest defers.** The two absent-feature MUSTs are ratified with triggers, never silently
   absorbed; fixture-vs-real proof labeled.

Hard constraints: never read/print/commit `.env`; PowerShell startup scripts only; commit after
verified work, **never push**; do not commit `docs/release/` or `docs/superpowers/plans/`; targeted
tests only (no full integration suite locally).

## Milestone validation definition (executable by a fresh `milestone-validator`)

1. **Per-feature conformance:** each of F-R1/F-R2/F-R3 has `spec.md` (consumer-contract-first,
   approval line filled before its code), `plan.md`, `evidence.md` with real command output.
2. **Re-run each gate from clean state:**
   - `go build ./...` → exit 0
   - `bash scripts/check-test-discipline.sh` → exit 0, `test-discipline: clean`
   - `go test ./apps/api/cmd/metaldocs-api/ -run TestMetricsEndpoint -count=1` → PASS
   - governance ADR-status gate: run its negative-proof repro (synthetic over-budget block → gate
     fails) and positive (clean tree → passes)
   - `go run ./scripts/req-trace` → still `stale=false`, exit 1, uncovered set unchanged
     `{REQ-AUTHN-1, REQ-AUTHN-3, REQ-SEARCH-1, REQ-SEC-3}`
   - `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → 0 violations
   - `powershell -File scripts/check-module-boundaries.ps1` → OK
3. **Regression against M0–M9:** confirm no commit in M10 touches `api/openapi/`, DB migrations,
   authz capability code, or any prior milestone's evidence; the aggregate diff is confined to the
   three feature surfaces + M10 docs.
4. **Live proof for F-R1** present and labeled real (not fixture): public-port `/metrics` not an
   unauthenticated scrape; metrics-port `/metrics` = 200 Prometheus.
5. Write `qa/milestone-qa.md` with the C1–C7 verdict. Separation of powers: the validator judges and
   writes the verdict only.

## Hard-stops (this milestone)

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Milestone boundary | Operator review gate; no mission-validator re-run / no merge without approval |
| HS-2 | A fix implies redesign outside its boundary (e.g. F-R1 needs an observability-stack change) | Stop; report boundary + minimum plan; no symptom-patch |
| HS-4 | `milestone-validator` returns FAIL | Open the named fix; re-run its lifecycle; re-dispatch validator |
| HS-6 | Scope drift (e.g. tempted to build REQ-SEARCH-1, rewrite discipline rules) | Stop; surface; replan before continuing |

## Terminal linkage

On M10 milestone-validator **PASS** + operator HS-1 approval, the main session re-dispatches
`mission-validator` for a **second** terminal-acceptance run against `mission.md §8`. Only that
second PASS (plus its own operator sign-off) closes the mission. M10 does not itself declare the
mission done.
