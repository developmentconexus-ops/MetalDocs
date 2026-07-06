# M8 validation contract (D4 — binding, authored before implementation)

> **Program:** global-maximum-remediation · **Milestone:** M8 (ops readiness: production images,
> distributed rate limiting, metrics/correlation/backup posture)
> **Authored:** 2026-07-05, **before any F8.x implementation** (mission D4). Committed before the
> first code change.
> **Binding:** the `milestone-validator` checks the implementation against this document **section by
> section**. Any divergence between shipped code and this contract is **HS-7**: fix the code to the
> contract, or re-open this contract **with operator approval** — never silently edit the contract to
> match the code (mission §9 HS-7).
> **D7 gate:** not required for M8 (mission D7 names M5/M6/M7 only). ADR requirement: F8.2's shared-
> store decision is durable → **ADR 0071** ships with F8.2.
>
> **Load-bearing clauses:** §0 baseline corrections (Dockerfiles/backup-scripts EXIST — scope is
> hardening + proof, not creation), §2.3 cross-replica arithmetic, §3.2 contract-surface decision for
> `/metrics`, §4 live-drive script.

---

## 0. Runtime-truth basis (facts this contract is built on)

All claims traced to source 2026-07-05 (investigator sweep + inline reads). Runtime truth beats docs.
If any anchor has moved at implementation time, the code wins and this section is re-stamped (HS-7 if
it changes an acceptance bar).

### 0.1 Images & compose

- `deploy/compose/docker-compose.yml` (273 lines) services: postgres, redis (`redis:7`, :6379,
  lines 21–35), docx-renderer (`apps/docx-renderer/Dockerfile`), api (`deploy/docker/api.Dockerfile`,
  line 115), worker (`deploy/docker/worker.Dockerfile`, 184), jobs (`deploy/docker/jobs.Dockerfile`,
  219), web (`frontend/apps/web/Dockerfile`), gateway/nginx. Compose-level `healthcheck:` blocks exist
  (lines 14, 30, 73, 106, 174, 249).
- The 3 Go Dockerfiles **exist** (commit `baf6e1b7`, F-19-deployment) — multi-stage
  (`golang:1.25-alpine` → `alpine:3.21`), static build, **run as root** (no `USER`), no image-level
  hardening. `api.Dockerfile` also copies `db/migrations`. **Baseline correction to the review/brief
  claim "paths don't exist".** F8.1 scope is therefore *hardening + build/run proof*, not creation.
- No repo-root `.dockerignore` verified at authoring time — F8.1 must ensure build context excludes
  `frontend/**/node_modules`, `docs/`, `backups/`, `.git`.
- `ops/DEPLOY.md` is a K8s-targeted, Approval-v2-scoped deploy note (`kubectl set image/rollout/env`,
  lines 43–123) — contradicts the Compose stack. No general deploy doc exists.
- Local dev path stays PowerShell scripts (`scripts/start-api.ps1` etc.); containers are the deploy
  path. `scripts/check-system-runnable.ps1` probes `http://127.0.0.1:8081` (login → `/auth/me` →
  target route → blank-template seed check) — reusable against a containerized api published on 8081.

### 0.2 Rate limiter

- `internal/platform/ratelimit/middleware.go:39` — in-memory `sync.Map`, keys
  `"<route>:user:<id>"` / `"<route>:ip:<addr>"`; fixed-window quotas hardcoded in
  `config.go:78–91` (`DefaultConfig`); login quota `10` hardcoded at `main.go:421`.
- Two mounted layers: `preAuthLimiter` (`main.go:428`, IP-keyed, login route only, via `chain.go:32`
  `pre_auth_login_rate_limit`) and `globalLimiter` (`main.go:438`, via `chain.go:36` `rate_limit`,
  `GlobalEnvelopeWrap` `middleware.go:271`).
- Config surface today: only `METALDOCS_TRUSTED_PROXY_CIDRS` wired. `config.go:12` doc comment claims
  `METALDOCS_RLIMIT_<ROUTE_KEY>` env overrides — **stale/aspirational, not implemented** (comment must
  be corrected in F8.2, behavior NOT implemented — rabbit-hole refusal).
- **Defect:** N api replicas ⇒ N independent budgets ⇒ effective limit silently ×N.
- Redis: compose service exists; **no Go Redis client in `go.mod`** (`go 1.25.0`).

### 0.3 Metrics / tracing / logs

- JSON metrics: `GET /api/v1/metrics` (`main.go:769`) served by
  `observability.MetricsHandler` (`http.go:190–209`), envelope `MetricsResponse{items, runtime,
  scheduler, db_pool}` — **mirrors an OpenAPI schema** (comment `http.go:174–182`), i.e. it IS part of
  the versioned contract and has pin tests (`http_typed_test.go`, `http_scheduler_test.go`,
  `http_dbpool_test.go`). Per-route items: requests, errors, duration total/avg/p50/p95/p99.
- No Prometheus-format endpoint. `prometheus/client_golang` present in `go.mod` only as an
  **indirect** dep (via OTel bridge/exporter modules) — no direct usage.
- OTel: `SetupOTel` (`otel.go:39–81`), gated by `OTEL_TRACES_EXPORTER` (`"none"`/unset ⇒ no-op) and
  `OTEL_EXPORTER_OTLP_ENDPOINT`; exporter via `autoexport` (`otlp|console|none`).
- Logs: every request logs `slog.Info("http_request", "trace_id", traceID, …)` (`http.go:157–167`);
  `traceID` resolution order: active OTel span → `X-Trace-Id` header → `requesttrace.Resolve`
  (`http.go:94–112`). **Correlation is plausible-by-code but unproven live** — that proof is F8.3(b).

### 0.4 Backup/restore

- Scripts exist and are non-trivial: `scripts/backup-postgres.ps1` (pg_dump custom format, SHA-256
  checksum, refuses `metaldocs_app` as backup user, JSON result object), `restore-postgres.ps1`,
  `validate-backup.ps1`, `run-backup-restore-gate.ps1` (194 lines, evidence JSON + restore smoke).
  Commits `0443aef8`, `6d847952`.
- **Gap:** no `wiki/runbooks/` backup/restore doc; the only prose mention is one pg_dump checklist
  line in `ops/DEPLOY.md:13`. No recorded end-to-end execution proof against the current stack.

---

## 1. F8.1 — production Dockerfiles + deploy-target truth

### 1.1 Image standard (each of api/worker/jobs)

Observable end-state per Dockerfile:
- Multi-stage kept; build stage pinned (`golang:1.25-alpine` acceptable; runtime base pinned by tag —
  digest-pinning optional, record either way).
- Runtime stage creates and switches to a **non-root user** (`USER` directive present; container
  `whoami`/uid proof ≠ 0).
- `.dockerignore` exists at repo root excluding at minimum: `.git`, `frontend/`, `docs/`, `backups/`,
  `node_modules` anywhere, `third_party/` (unless build-required — verify), local env files
  (`.env*` — MUST be excluded; constraint §10 mission).
- api image keeps `db/migrations` copy **only if** runtime actually reads it (verify; if nothing reads
  it in-container, drop the copy and record why — no cargo-cult layers).
- No secrets, no `.env`, no build-arg secrets baked into any layer (`docker history` spot-check).

### 1.2 Build + run proof

- `docker compose build api worker jobs` (or full `docker compose build`) exits 0 from clean tree —
  full command + tail of output recorded in evidence.
- `docker compose up` (db + redis + docx-renderer + api at minimum) reaches healthy;
  `scripts/check-system-runnable.ps1` run against the **containerized** api on 8081 passes all its
  checkpoints — transcript captured.
- Non-root proof: `docker compose exec api whoami` (or `id -u`) ≠ root, for all 3 images (worker/jobs
  via `docker run --rm --entrypoint id <image>` if not long-running-exec-able).

### 1.3 Deploy-target truth

- A deploy doc (either rewritten `ops/DEPLOY.md` or a new `ops/` doc that supersedes it — one of the
  two, stated explicitly) declares **Docker Compose as the v1 deployment target**, documents the
  build/up/down/upgrade flow actually proven in §1.2, and contains **zero live kubectl instructions**
  presented as the current path. The Approval-v2 K8s content is re-homed/archived with a pointer, not
  silently deleted (docs governance).
- Acceptance: reading the doc top-to-bottom yields no Compose-vs-K8s contradiction; wiki index/stamps
  updated per governance.

## 2. F8.2 — distributed rate limiter

### 2.1 Structure (global maximum, not patch)

- A **store abstraction** inside `internal/platform/ratelimit` (platform layer; no module imports):
  the counting/decision primitive gets an interface with two implementations —
  (a) existing in-memory (default; single-replica semantics preserved bit-for-bit for current tests),
  (b) Redis-backed GCRA-class (`github.com/go-redis/redis_rate/v10` + `redis/go-redis/v9`, or
  equivalent — recorded in ADR 0071) sharing one budget across processes.
- Middleware semantics unchanged: same route keying (`<route>:user:<id>` / `:ip:<addr>`), same two
  chain positions (§0.2), same 429 problem+json + headers. Only the counting store varies.
- Config: explicit env (names recorded in ADR/spec, e.g. `METALDOCS_RATELIMIT_STORE=memory|redis` +
  Redis addr); **fail-fast startup guard** — a config that declares multi-replica intent (env flag,
  e.g. `METALDOCS_MULTI_REPLICA=true`, exact name fixed in the feature spec) with `store=memory`
  refuses to boot with a clear error. Compose api service sets the Redis store.
- Stale `config.go:12` comment corrected (env-override claim removed).

### 2.2 ADR 0071

- Records: the per-replica defect, the shared-store decision, algorithm class (GCRA/sliding window vs
  today's fixed window — behavior delta stated), memory-as-default rationale, the startup guard, and
  the **named scale-out/K8s trigger** for revisiting (e.g. first multi-node deployment). Status
  Accepted, ≤3-line status field (M9 hygiene rule respected from birth).

### 2.3 Cross-replica correctness proof (the load-bearing test)

Integration test (real Redis via testcontainer/compose Redis, or miniredis if faithful to the chosen
lib — fixture-vs-real labeled in evidence):
- Two **independent** limiter/middleware instances (simulating 2 replicas) share one Redis and one
  quota Q for the same key.
- Driving R > Q requests split across both instances within one window admits **≤ Q total** (allowing
  the algorithm's documented burst tolerance, stated in the test), remainder 429.
- **Contrast pin:** same drive against two in-memory instances admits up to 2×Q — the defect made
  visible, so a regression back to per-replica counting fails loudly.
- Guard test: multi-replica intent + memory store ⇒ constructor/boot returns error (unit test).
- Existing ratelimit tests (`export_test.go` etc.) stay green with memory default.

### 2.4 Live proof (feeds §4)

- 2 api replicas via compose (scale or duplicated service) against one Redis: hammer one rate-limited
  route past quota; observed combined 200-count ≤ quota + documented burst; 429s carry the standard
  problem+json. Captured transcript.

## 3. F8.3 — Prometheus /metrics + correlation + backup runbook

### 3.1 /metrics endpoint

- Prometheus text-format endpoint served by `promhttp` against a registry populated with:
  - `metaldocs_http_requests_total{route,method}` and `metaldocs_http_request_errors_total{route,method}`
    (from the existing per-route counters — bridged or natively instrumented; mechanism per feature
    spec, no double-counting),
  - request duration histogram `metaldocs_http_request_duration_*{route,method}` (histogram preferred
    over bridging stored p50/p95/p99 gauges; if gauges are bridged instead, ADR-note why),
  - Go runtime collectors (`go_*`, `process_*` where the collector supports the platform),
  - DB pool gauges (from the existing `DBPoolStats` provider).
  Exact family names fixed in the F8.3 spec; the four groups above are the binding minimum.
- Mounted on the **api** binary. Path `/metrics`, **outside** `/api/v1` (see §3.2). Worker/jobs
  metrics exposure is OUT of M8 scope (rabbit hole: no fleet-wide scrape topology pre-v1; ADR 0071 or
  the runbook names the trigger).
- Security posture stated in the feature spec: `/metrics` carries no tenant data (aggregate only);
  exposure decision (bind/gateway rules) recorded — at minimum, not proxied to the public gateway by
  default (verify nginx/gateway config doesn't expose it; record disposition).

### 3.2 Contract-surface decision (binding)

- `/metrics` is an infrastructure endpoint **outside the versioned product contract** — NOT added to
  `api/openapi/v1/openapi.yaml`; mounted directly on the mux like `/health` class endpoints. Rationale:
  contract-first governs the product API surface; Prometheus exposition format is governed by the
  Prometheus spec, not our OpenAPI. (If the implementer finds `/health` etc. ARE in the OpenAPI spec,
  STOP — HS-7 — and reconcile before mounting.)
- `GET /api/v1/metrics` (JSON) **stays**, wire-shape unchanged (it is contract-bound, §0.3); its pin
  tests stay green untouched.

### 3.3 Log↔trace correlation proof

- Live drive with tracing enabled (`OTEL_TRACES_EXPORTER=console` acceptable — real exporter, stdout
  span dump): one authenticated request to a real route.
- Captured artifacts: (a) the emitted span (trace id visible), (b) the `http_request` slog line for
  that request. **Acceptance: identical trace id string in both**, plus the `X-Trace-Id` response
  header matching. Recorded verbatim in evidence.
- If the slog trace id does NOT match the active span id (resolution-order bug), that is a defect to
  fix inside F8.3 (root cause in `http.go:94–112` resolution), not a re-definition of the proof.

### 3.4 Backup/restore runbook

- `wiki/runbooks/backup-restore.md`: prerequisites (dedicated backup role — scripts refuse
  `metaldocs_app`, §0.4), backup procedure (script invocation + expected JSON result + checksum),
  restore procedure (target DB, `restore-postgres.ps1` flags), validation (`validate-backup.ps1` /
  `run-backup-restore-gate.ps1`), and recovery decision guidance (when to restore vs repair). Written
  against the **existing scripts** — no new backup implementation (global-max judgment: the scripts
  are sound; the missing piece is the operator-facing doc + executed proof).
- **Executed end-to-end against the local stack**: backup taken → restored into a scratch database
  (never over the live dev DB) → validation gate passes (restore smoke). Full command transcript +
  the evidence JSON captured. Fixture-vs-real: this is REAL (live local Postgres), no mocks.
- `ops/DEPLOY.md:13`'s pg_dump line updated to point at the runbook (or superseded by §1.3 rewrite).

## 4. Milestone live QA drive (D4, close evidence)

Sequence (single recorded session, transcripts in `qa/` or feature evidence):
1. `docker compose build` — clean-tree build of the 3 Go images (F8.1).
2. `docker compose up` core stack; `check-system-runnable.ps1` against containerized api — PASS.
3. Non-root spot-checks on the 3 containers.
4. Scale api to 2 replicas (one Redis); drive a rate-limited route past quota → combined admits ≤
   quota(+burst), 429 problem+json shape verified (F8.2 §2.4).
5. `curl http://<api>/metrics` → Prometheus text format, §3.1 families present (grep proof).
6. Traced request → capture span + log line + response header, one trace id (§3.3).
7. Execute backup/restore runbook end-to-end (§3.4) — scratch-DB restore + validation gate PASS.
8. Regression floor: `go build ./...` green; targeted ratelimit + observability test suites green;
   JSON `/api/v1/metrics` pin tests green.

Any step impossible on this box (e.g. Docker unavailable) is HS-3 (repair prerequisite) or an
operator-surfaced deviation — never a silent defer.

## 5. Forbidden / out of scope (validator fails on sight)

- K8s manifests/Helm; Grafana/alerting/OTel-collector deployment; worker/jobs scrape endpoints.
- Implementing `METALDOCS_RLIMIT_<ROUTE_KEY>` per-route env quotas (comment correction only).
- Changing middleware chain order, 429 shape, or route keying semantics.
- Adding `/metrics` to `api/openapi/v1/openapi.yaml`; breaking/altering the JSON
  `/api/v1/metrics` wire shape.
- New backup implementation replacing the existing scripts.
- Frontend/web or docx-renderer image changes; CI docker-build gate.
- Committing `docs/release/`, `.env*`, or any plans-dir force-add; any push to origin.
- Schema migrations; tenant-scoped queries (M8 is ops-layer only).

## 6. Feature → evidence map (filled at close; shape fixed now)

| Feature | Proof artifacts required in evidence.md |
|---------|------------------------------------------|
| F8.1 | build transcript; runnable-check transcript vs containers; non-root proofs; deploy doc diff; .dockerignore |
| F8.2 | cross-replica test output (real/fixture labeled); contrast pin output; guard test; live 2-replica drive; ADR 0071 link; existing-suite green |
| F8.3 | /metrics scrape capture; correlation triple (span/log/header); runbook file; end-to-end execution transcript + evidence JSON; JSON-endpoint pin tests green |
