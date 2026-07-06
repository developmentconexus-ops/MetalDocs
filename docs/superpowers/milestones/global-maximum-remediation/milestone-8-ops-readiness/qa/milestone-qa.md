# Milestone 8 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (up-front spec) + `../validation-contract.md` (D4 binding) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-06  ·  **Verdict:** see C7 — **PASS**.
> Program: `global-maximum-remediation` · Milestone: M8 (ops readiness) · Commit range judged: `0afc8944..d51b6195` (excl. untracked `docs/release/`).

## Inputs loaded (fail-fast gate)

All present and readable: `milestone.md`; `validation-contract.md` (committed `218e2d12` pre-impl); F8.1/F8.2/F8.3 `spec.md`+`plan.md`+`evidence.md`; program `README.md`; governing `mission.md` (§7 M8). Aggregate diff read. No self-judged verdict existed (`qa/` did not exist on disk and no `qa/milestone-qa.md` in the commit range — verified). Nothing blind.

## C1 — Spec & plan conformance (per feature)

Every feature: `spec.md` approved pre-code (2026-07-05 operator-delegated via M8 `/goal` brief + binding `validation-contract.md` §1–§3 committed `218e2d12`); interview record populated (each has an explicit "none needed — why" or verified Q&A rows); `plan.md` execution-shaped (task A/B/C, files, TDD ordering, test strategy — not a re-spec); `evidence.md` acceptance table maps row-for-row to the spec Validation Gate; non-goals + rabbit-holes respected (verified in C6).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F8.1 dockerfiles | ✅ compose refs Dockerfiles at unchanged paths (115/184/219); `check-system-runnable.ps1` passes unchanged vs containerized api:8081 | ✅ 3 images build; boot+runnable PASS; non-root uid=10001 live; `.dockerignore` covers `.env`/`.env.*`/`docs/`/`third_party/`/`**/*.exe`; DEPLOY.md Compose-v1 + K8s re-homed | ✅ no K8s/Helm, no CI gate, no web/docx image, dev PowerShell path untouched | evidence.md + C2/C3 |
| F8.2 rate-limiter | ✅ chain links call unchanged `ratelimit.Middleware` API; only construction wires the store; `chain.go` byte-untouched (REQ-MW-7 order preserved) | ✅ cross-replica redis 10/30, memory contrast 20/30, guard 7/7, suite green, live 2-replica admitted=10 limited=20, ADR 0071 Accepted | ✅ no per-route env quota (stale `config.go:12` comment corrected, not implemented); no chain-order/429-shape change; worker/jobs untouched; no Redis outside ratelimit | evidence.md + C2/C3 |
| F8.3 metrics-backup | ✅ scraper gets exposition families; JSON `/api/v1/metrics` unchanged (pin tests green, live 401 not 404); on-call runbook top-to-bottom executable | ✅ /metrics 200 live w/ 4 binding groups; openapi diff empty; correlation triple one id; runbook executed end-to-end (deviations labeled) | ✅ no Grafana/collector/worker-jobs endpoint; openapi.yaml unedited; no new backup impl | evidence.md + C2/C3 |

C1 clean. Env-var-name note: F8.2 evidence prose used shorthand `REDIS_ADDR`/`MULTI_REPLICA`; the actual compose file and `store_config.go` constants both use the full `METALDOCS_RATELIMIT_STORE` / `METALDOCS_RATELIMIT_REDIS_ADDR` / `METALDOCS_MULTI_REPLICA` — verified identical, **no split-brain** (loose prose only, not a real name mismatch).

## C2 — Gates re-run, isolated (validator ran these, not trusted from transcript)

| Feature | Command re-run (fresh, `-count=1`) | Real output | Pass? |
|---------|------------------------------------|-------------|-------|
| build floor | `go build ./...` | exit 0 | ✅ |
| F8.2 | `go test -count=1 ./internal/platform/ratelimit/ -run 'TestCrossReplicaSharedBudget\|TestStoreConfig_MultiReplicaGuard' -v` | `redis shared budget: 10/30 admitted (quota 10, tol +1)`; `memory contrast: 20/30 — double the intended quota (defect pinned)`; guard 7/7 PASS; `ok ... 2.796s` — **real Redis 127.0.0.1:6379** | ✅ |
| F8.3 | `go test -count=1 ./internal/platform/observability/ -run 'Prometheus\|TraceID\|Wrap' -v` | `TestHTTPObservability_Wrap_SetsTraceIDResponseHeader` PASS, `...EchoesInboundTraceIDOnResponse` PASS, `TestPrometheusHandler_ExposesBindingMetricFamilies` PASS, `TestMetricsHandler_JSONUnaffectedByPrometheusInstrumentation` PASS; `ok ... 6.6s` | ✅ |
| F8.3 | `go test -count=1 ./apps/api/cmd/metaldocs-api/ -run Metrics -v` | `TestMetricsEndpoint_BypassesAuthChain_REQF83` PASS; `ok ... 6.681s` | ✅ |
| F8.3 JSON pin | `go test -count=1 ./internal/platform/observability/ -run 'TestMetricsHandler\|Scheduler\|DBPool'` | all typed/scheduler/dbpool/method-not-allowed pins PASS | ✅ |
| F8.3 live | `curl http://127.0.0.1:8081/metrics` | `HTTP 200`, `content-type: text/plain; version=0.0.4`; families present: `metaldocs_http_requests_total`, `metaldocs_http_request_errors_total`, `metaldocs_http_request_duration_seconds` (histogram), `go_goroutines`, `process_cpu_seconds_total`; 85 `metaldocs_` samples — **unauth, containerized api** | ✅ |
| F8.1/F8.3 live | `curl /api/v1/health/live` → 200; `curl /api/v1/metrics` → 401 (present, auth-gated, not 404) | JSON endpoint intact | ✅ |
| F8.1 live | `docker exec metaldocs-api id` | `uid=10001(metaldocs) gid=10001(metaldocs)` | ✅ |
| F8.1 regression | `check-system-runnable.ps1` vs containers | 5/5 PASS: blank-template-object, login-endpoint 200, login-session 1 cookie, auth-me 200, target-route /health/ready 200 | ✅ |

Every named feature test + live proof re-run from clean state matches the evidence transcripts. One-shot live drives not re-run per the environment brief (2-replica drive, backup/restore, OTEL correlation — scaffolding torn down) are judged from labeled evidence + code-path consistency (C3).

## C3 — Senior review of the aggregate milestone diff

Reviewed non-vendor source diff as one unit (~3.2k LOC incl. milestone docs; vendor churn = `redis_rate`/`go-redis`/`gopher-lua`/`atomic`, the ADR-0071-recorded deps).

- **Rate-limiter refactor (global maximum, not patch):** the `sync.Map` counting logic moved **verbatim** from `middleware.go` into `memory_store.go` behind a new `Store` interface; `middleware.go` now delegates to `m.store.Allow(...)`. Keying (`<route>:user:<id>`/`:ip:<addr>`), quotas, 429 RFC-9457 shape, headers all unchanged — the middleware still owns the HTTP concern, only counting is abstracted. The `-235` churn is relocation, not deletion (existing full suite + live 2-replica drive confirm behavior preservation). This is the structure `validation-contract.md` §2.1 demanded.
- **Fail-open Redis policy** (`redis_store.go`, `store.go`) is documented, ADR-0071-§Decision-5-backed (availability > strict limiting), with a named compensating control (DB-backed account lockout bounds brute-force independent of Redis). Not a hidden defect.
- **`/metrics` mount** (`main.go` rootMux) served ahead of the chain so a credential-less scraper reaches it AND self-scrapes never feed `httpObs.Wrap`/rate-limiter; panic recovery still wraps it; not in openapi; gateway routes only `/api/` and `/` (not `/metrics`).
- **Prometheus vecs reuse** the already-computed route/method/duration in `Wrap` — no double-count, pinned by `TestMetricsHandler_JSONUnaffectedByPrometheusInstrumentation`. Separate storage (atomics vs vecs).
- **X-Trace-Id** set from the same resolved `traceID` that feeds slog + span → correlation triple is one variable, closed at root cause (header echo), not a patch.
- **Backup role** `BYPASSRLS` on a non-superuser SELECT-only role is the correct standard mechanism for a full-DB `pg_dump` over M7 FORCE-RLS tables; grants read-past-RLS only (no write/DDL), and governs `metaldocs_backup` — a different role from the app `metaldocs_ci`/`metaldocs_app`, so **M7 RLS-truth posture is untouched**.
- No duplicated logic, no dead code from a superseded approach, no fact stored twice, no feature breaking another. Staff-engineer bar: **met** ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api: metrics route addition + docs governance for ADR 0071 + deploy doc + runbook) | pass | `/metrics` outside `/api/v1` (contract §3.2), openapi untouched; ADR 0071 Accepted w/ ≤3-line status, wiki index entry; DEPLOY.md rewrite + verbatim K8s re-home w/ banner; runbook + runbooks index |
| Contract-first floor | pass | `git diff 0afc8944..d51b6195 -- api/openapi` = **empty** |
| Middleware chain order (REQ-MW-7) | pass | `git diff ... -- **/chain.go` = **empty** (not reordered) |
| Regression vs prior milestones (M0–M7) | all still pass | `go build ./...` green; ratelimit + observability targeted suites green; runnable check green; M7 RLS-truth posture unchanged (no tenant-scoped queries added; backup-role BYPASSRLS is a distinct non-app role); no schema migration in range |
| Forbidden scope sweep | clean | no K8s/Helm manifests; no worker/jobs metrics endpoint; no `docs/release/` commit in range |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Horizontal-scale correctness | N replicas → N independent budgets (silent ×N) | one shared budget across replicas | shared-store class fix (interface + Redis GCRA), NOT a per-route patch; live 2-replica admitted=10 (not 20); fail-fast guard converts misconfig into boot failure |
| Deploy target unambiguous | K8s-only DEPLOY.md contradicts Compose stack | Compose declared v1 target, zero live kubectl; K8s re-homed | doc review: DEPLOY.md §3 line + archive banner; contradiction resolved in writing |
| Observability floor | no Prometheus endpoint; correlation plausible-by-code only | `/metrics` scrapes live; log↔trace↔header one id proven | live curl 200 w/ 4 families; correlation triple `e8dc2036…` verbatim (code path single-sourced) |
| Recoverability | asserted, never executed | backup→scratch-restore→validate executed | real local Postgres; iam_users 127==127 exact; audit_events +8 = labeled live-append point-in-time drift, not loss |

Root cause fixed, not symptom-patched (HS-2 clear). **Retrospective:** the two F8.3 deviations (pg client-container because host lacks PG16 tools; audit_events +8 live-write drift) are honestly labeled, bounded, and each carries a written trigger/owner — they do not weaken the proofs (iam_users exact match anchors the restore; runbook prereq owns host tooling). Better-construction note (non-blocking, → next-ops-milestone input): fleet-wide worker/jobs scrape topology is a rabbit-hole-deferred follow-on with a named trigger (first multi-node deploy). Construction is sound.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — **clean** (each feature's acceptance mapped to a re-run command/live proof)
- [ ] Fixture/mock passed off as real-provider proof — **clean** (real Redis, real containers, real Postgres; the only labeled fixture-vs-real note is honest — memory contrast is deliberately in-memory to pin the defect)
- [ ] Consumer contract guessed rather than read from the consumer — **clean** (chain.go/compose-path/JSON-pin consumers read; producers match)
- [ ] Split-brain (one fact, two sources of truth) — **clean** (env-var names identical across code+compose; quotas single-sourced compiled-in; JSON vs Prometheus storage separate-by-design not duplicated fact)
- [ ] Self-judged close / validator edited or fixed code — **clean** (validator judged + wrote this file only; no source edits; status not flipped)
- [ ] Scope drift (work beyond the spec, no rationale) — **clean** (the one extra — compose gotenberg URL fix `b15a4480` — is boot-proof-required, within F8.1's stated non-goal boundary; backup-role BYPASSRLS is a live-surfaced F8.3 gap with rationale)
- [ ] Symptom-patch (bar moved by masking) — **clean** (shared-store class fix, image-standard hardening, executed backup cycle)

All unchecked = clean.

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-level Store abstraction with verbatim memory relocation, clean rootMux mount, no double-count, no split-brain, ADR-backed trade-offs) and **function-wise** (live /metrics scrape, live 2-replica shared-budget, executed backup/restore, correlation triple — end-to-end, not fixture-only). Contract-first and REQ-MW-7 floors held (openapi + chain.go byte-empty). M0–M7 regression floor green. No forbidden-list hit.
- Handed back to the main session to flip status and present the **HS-1** operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only the main session, only after HS-1
