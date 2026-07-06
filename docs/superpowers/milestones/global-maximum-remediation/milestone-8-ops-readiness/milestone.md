# Milestone 8 — Ops Readiness

> **Program:** global-maximum-remediation  ·  **Governing spec:** `../mission.md` §7 M8
> **Status:** PASSED — milestone-validator VERDICT: PASS (2026-07-06, `qa/milestone-qa.md`); HS-1 operator gate pending
> **Authored:** 2026-07-05 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** M8 is, **which features** it contains,
> **what each feature implements**, and **what gets validated**. No execution steps — the "how" lives
> in each feature's `plan.md`. The close QA (`qa/milestone-qa.md`) validates M8 against *this* file and
> the binding `validation-contract.md` (D4).

## Objective

Clear the three named pre-customer ops blockers (review 778f494a §9, findings 16–18) so that **an
operator can build, run, scale, observe, and recover the system without discipline-dependent
knowledge**: after M8, `docker compose build && docker compose up` produces a runnable containerized
stack from production-standard images; rate limits stay **correct when the API scales to N replicas**
(no silent N× multiplication); a Prometheus scraper gets real metrics from `/metrics`; a request's
trace id **provably** links its log lines to its trace; and a written backup/restore runbook has been
**executed end-to-end** against the local stack.

**Baseline correction (runtime truth beats the review/brief, verified 2026-07-05):** the three
`deploy/docker/*.Dockerfile` files **exist** (basic multi-stage, added by F-19-deployment `baf6e1b7`)
but are below production standard (run as root, no image healthcheck, no non-root user); backup/restore
**scripts** exist (`scripts/backup-postgres.ps1`, `restore-postgres.ps1`, `validate-backup.ps1`,
`run-backup-restore-gate.ps1`) but no wiki runbook documents them and no end-to-end execution proof
exists; Redis 7 is **already a compose service** but no Go client consumes it. Full fact basis:
`validation-contract.md` §0.

**Bars this milestone moves:**
- **Deploy target unambiguous** — one documented deployment story (Docker Compose for v1); the K8s-only
  `ops/DEPLOY.md` contradiction resolved in writing.
- **Horizontal-scale correctness** — the rate limiter enforces one shared budget across replicas
  (criterion: 2-instance test shows combined admits ≤ quota), or is structurally prevented from running
  multi-replica misconfigured.
- **Observability floor** — Prometheus `/metrics` scrapes; log↔trace correlation demonstrated with a
  captured pair (log line + trace share one trace id).
- **Recoverability proven, not asserted** — backup → destroy → restore → verify cycle executed against
  the local stack with recorded output.

## Appetite & rabbit holes

**Appetite:** 3 features, ops-layer only; no product behavior changes; no schema migrations; one new Go
dependency class (Redis client + limiter) allowed for F8.2.

**Rabbit holes (do not chase):**
- **No Kubernetes manifests / Helm** — v1 target is Compose; K8s is a post-v1 decision (ADR 0071 records the trigger).
- **No Grafana/alerting/OTel-collector deployment** — M8 exposes metrics + proves correlation; dashboards and alert rules are post-v1 ops build-out.
- **No per-route rate-quota env config revival** — the stale `METALDOCS_RLIMIT_<ROUTE_KEY>` doc comment (`config.go:12`) gets corrected, not implemented; quotas stay code-defined.
- **No frontend/web or docx-renderer image work** — both already have their own Dockerfiles; out of the 3-binary scope.
- **No CI docker-build gate** — YAGNI pre-v1; compose build proof at milestone close suffices; trigger = first external deploy pipeline.
- **No S3/offsite backup automation** — runbook covers local pg_dump/restore class; offsite is a production-hosting decision.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F8.1 | `f8.1-dockerfiles` | Harden the 3 existing `deploy/docker/*.Dockerfile` to production standard (multi-stage kept; add non-root `USER`, pinned base images, `.dockerignore` coverage, healthcheck story consistent with compose); resolve the DEPLOY.md-vs-compose target contradiction by declaring **Docker Compose the v1 deployment target** in a rewritten ops deploy doc (K8s-scoped Approval-v2 content re-homed or archived, not silently deleted). **Consumer:** operator deploy flow (`docker compose build/up`) + `check-system-runnable.ps1`. | `docker compose build` succeeds for api/worker/jobs from clean tree; containerized stack boots and `check-system-runnable.ps1` passes against it; images run as non-root (`docker inspect` / `whoami` proof); deploy doc names Compose as v1 target and contains zero contradicting K8s instructions. |
| F8.2 | `f8.2-rate-limiter` | Distributed rate limiting: store abstraction in `internal/platform/ratelimit` with two backends — existing in-memory (single-replica default) and Redis GCRA-class (`redis_rate`) sharing one budget across replicas; backend selected by explicit config; **startup guard** fails-fast a configuration that declares multi-replica intent without a shared store; ADR 0071 records the decision + K8s/scale-out trigger. Existing middleware semantics (route keying, headers, problem+json 429) unchanged. **Consumer:** middleware chain (`chain.go` pre-auth + envelope layers) on N api replicas. | Cross-replica correctness test: 2 limiter instances against one Redis admit ≤ quota combined (and in-memory admits 2× — the defect pinned as contrast); backend selection + startup guard covered by tests (multi-replica intent + memory store → refuse to boot); compose api service wired to Redis backend; existing ratelimit unit/integration tests stay green; ADR 0071 Accepted. |
| F8.3 | `f8.3-metrics-backup` | (a) Prometheus `/metrics` endpoint (promhttp) exposing the existing observability counters + Go runtime + DB pool gauges, mounted on the api mux alongside the kept JSON `/api/v1/metrics` (no consumer break); (b) log↔trace correlation verified live — one request driven with tracing on, captured slog line's trace id == captured span/`X-Trace-Id`; (c) backup/restore runbook `wiki/runbooks/backup-restore.md` wrapping the existing scripts, **executed end-to-end** (backup → restore to scratch DB → validate) against the local stack. **Consumers:** Prometheus scraper; on-call operator. | `curl /metrics` returns Prometheus text format containing named metric families (contract §3 lists them); correlation proof captured (log line + trace id pair); runbook steps executed with real output recorded in evidence (restore smoke passes); JSON endpoint still serves its shape (existing tests green). |

Order intentional: F8.1 first (containerized stack is the substrate the F8.2/F8.3 live proofs run on),
F8.2 second (needs compose Redis wiring from a working build), F8.3 last (its live QA drive doubles as
the milestone-close drive).

## Milestone validation definition

Close gate run by the **`milestone-validator` subagent** (separation of powers — judges and writes
`qa/milestone-qa.md`; main session flips status only on PASS), per the binding C1–C7 checklist
(`.claude/skills/milestone/references/milestone-end-validation.md`). For M8:

1. **Per-feature acceptance** — every feature meets its "what to validate"; each feature's consumer
   contract (`spec.md`) honored. Checked **section-by-section against `validation-contract.md`** (D4)
   — any divergence is **HS-7**.
2. **Workflow-class QA** — `wiki/quality/backend-api-checklist.md` (metrics route addition),
   docs checklist (`wiki/quality/*` per docs governance) for ADR 0071 + deploy doc + runbook;
   ops proofs per contract §1–§4.
3. **Regression** — M0–M7 gates still pass: `go build ./...`, targeted ratelimit/observability suites,
   M7's RLS-truth posture untouched (M8 adds no tenant-scoped queries), contract-sync + openapi lint
   green if the contract is touched (metrics endpoint is **outside** `/api/v1` product contract — see
   contract §3.2).
4. **Root-cause check** — limiter fix is the shared-store class, not a per-route patch; Dockerfile fix
   is image-standard hardening, not a dev-shim; backup proof is an executed cycle, not a doc citation.
   Confirmed fixed, not symptom-patched (HS-2).
5. **No unplanned scope** — anything beyond F8.1–F8.3 recorded with rationale; rabbit-hole list above
   is the drift baseline.
6. **Live QA drive** (runtime-visible milestone, D4) — containerized stack up; hit a rate-limited route
   across 2 api replicas; scrape `/metrics`; run the backup/restore runbook; proof captured.

## Dependencies & constraints

- **Depends on:** M5 (River jobs binary is one of the 3 images), M7 passed (HS-1 ratified 2026-07-05).
  No dependency on M9.
- **Quality goals (ranked):** correctness-under-scale > operational simplicity > performance.
- **Architectural constraints:** modular-monolith + 4-binary boundaries hold (no new binary); platform
  code stays in `internal/platform/*`; middleware chain order inherited (rate-limit links keep their
  positions); RFC 9457 for 429s unchanged; `/metrics` carries **no tenant data** (aggregate counters
  only); contract-first rule — no `/api/v1` route added/changed without openapi regen (Prometheus
  endpoint deliberately mounted outside the versioned API surface); no schema migrations; `.env` never
  read/printed/committed; PowerShell startup scripts remain the local dev path (containers are the
  deploy path, not a dev replacement); full integration suite NOT run locally (targeted `-run`
  filters); commits local, **never push**.
- **Risks:** (1) Docker build context bloat (repo has huge frontend/node_modules) → mitigate with
  `.dockerignore` in F8.1; (2) Redis becomes an api boot dependency → mitigate: memory backend stays
  the single-replica default, Redis required only when configured for multi-replica (guard makes
  misconfig loud); (3) Windows-local docker compose environment may be unavailable/slow → surface
  early in F8.1; if the box cannot run compose at all, that is HS-3 territory (prerequisite repair),
  not a silent defer.

## Applicable hard-stops

- **HS-1** (milestone boundary) — after validator PASS, STOP; operator gate; no M9 without approval.
- **HS-2** (fix implies redesign outside boundary) — e.g. if correct distributed limiting demands
  changing middleware chain order or authn coupling; stop, surface, no symptom-patch.
- **HS-3** (prerequisite boundary fails) — build/runnable/docker-daemon broken → repair first.
- **HS-4** (validator FAIL) — open named fix feature, re-run lifecycle, re-dispatch validator.
- **HS-6** (scope drift) — e.g. discovering the compose stack needs unplanned services to boot;
  stop, surface, replan.
- **HS-7** (impl deviates from `validation-contract.md`) — fix code to contract, or re-open contract
  WITH operator approval; never silently edit contract to match code.
