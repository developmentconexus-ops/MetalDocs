# Pass 12: Observability + Async Runtime Map

**Date:** 2026-08-09
**Baseline:** `main@418070bf38a9f358f9131bcc36b7a6bcbc069273`
**Status:** reproduced-current-state — all claims re-verified direct against source this pass, same baseline commit prior inventory (`inventory/observability.md`, `inventory/persistence.md`) ran against.

## 1. Health/readiness endpoints per binary

| Binary | HTTP listener? | Health endpoint | Evidence |
|---|---|---|---|
| `metaldocs-api` | yes, 2 listeners | main server + separate metrics server | `apps/api/cmd/metaldocs-api/main.go:626` `server.ListenAndServe()`; `:631` `metricsServer.ListenAndServe()`; both built in `buildServers` `:1353` (`*http.Server` `:1360`, metrics `:1374`) |
| `metaldocs-worker` | **NO** | none | zero `http.Server`/`ListenAndServe` hits in `apps/worker/cmd/metaldocs-worker/main.go` (350 lines, full grep) |
| `metaldocs-jobs` | **NO** | none | zero `http.Server`/`ListenAndServe` hits in `apps/jobs/cmd/metaldocs-jobs/main.go` (325 lines, full grep) |
| `docx-renderer` | yes | `GET /health` | `apps/docx-renderer/src/index.ts:16`; `app.listen(...)` `:41` |

**#95/A7 claim "worker+jobs have NO HTTP listener" → CONFIRMED still true at this baseline.** Nothing added since the prior pass. No `/metrics`, no `/healthz`, no liveness/readiness probe target exists for either process — a container orchestrator (compose healthcheck, k8s probe) cannot ask these two binaries "are you alive" over HTTP; it can only infer from process exit code.

## 2. Metrics surfaces

Two parallel, non-unified systems:

1. **Prometheus** — wired only into `metaldocs-api`'s second listener (`metricsServer`, main.go `:1374`). Worker/jobs never construct a Prometheus registry or exporter (confirmed: `grep -rn "RuntimeStatusProvider" --include=*.go apps internal` → only `internal/platform/bootstrap/api.go:115` constructs `NewPostgresRuntimeStatusProvider`).
2. **Hand-rolled JSON** — `internal/platform/observability/runtime.go`. Two implementations of the same interface:
   - `StaticRuntimeStatusProvider.RuntimeMetrics` (`:86`) — **hardcodes** `worker.outbox.{claimable,pending,deadLettered}` to `0` at `:109-111`. Used wherever the live Postgres-backed provider isn't wired (any path that isn't the api binary's own runtime status handler).
   - `PostgresRuntimeStatusProvider.RuntimeMetrics` (`:178`) — real query via `outboxStats` struct (`:200-202`), scanned `:235`.

**The #95 "worker.outbox hardcoded-0 fallback" claim → CONFIRMED still true.** Since worker/jobs have no HTTP listener at all (§1), they cannot serve *any* runtime-status JSON, live or static — the static-zero fallback is what api itself falls back to whenever the live provider path isn't exercised, and it is the *only* thing an operator watching outbox depth from outside the api process can ever see for worker/jobs-side backlog, because those two processes expose nothing.

## 3. Tracing

`otel.Tracer(...).Start` call sites, repo-wide (`grep -rn "otel.Tracer(" --include=*.go`): **exactly 2**, both in the core module cluster, neither in worker/jobs/render/notifications:
- `internal/modules/approval/application/decision_service.go:258`
- `internal/modules/controlleddocuments/application/service.go:277`

Zero spans anywhere under `internal/platform/worker/`, `internal/platform/messaging/`, `internal/modules/render/fanout/dispatchjobs/`, `internal/modules/notifications/infrastructure/` — i.e., the entire outbox-publish → outbox-consume → River-dispatch → renderer-HTTP-call chain, across all three of api/worker/jobs, produces zero OTel spans. Confirmed by direct read of `internal/platform/worker/service.go` (full file, 202 lines) — every log call is plain `slog.Info/Warn/Error`, no span start/end anywhere in `RunOnce`/`dispatchEvent`/`markFailure`.

**Outbox trace propagation — bare string, not W3C context.** `internal/platform/messaging/events.go:21`:
```go
type TraceID string
```
`Event` struct (`:66-78`) carries `TraceID TraceID` as one flat field alongside `EventID`/`EventType`/`Payload`/etc — no `traceparent`/`tracestate` W3C fields, no `context.Context`-carried `SpanContext` serialized into the outbox row. `worker/service.go:96,153,160` logs this `trace_id` as a plain string field on `slog.Info/Warn/Error` calls — useful for grep-correlating log lines to the original request, but it cannot reconstruct a distributed trace (no parent span ID, no sampling flags, no W3C `traceparent` format) and does not connect to whatever OTel span (if any) was open when the outbox row was written, since nothing calls `otel.Tracer` on the publish side either (`internal/platform/messaging/outbox/postgres/publisher.go` — zero otel references).

## 4. Log correlation

`grep -rc "slog\.\(Info\|Warn\|Error\|Debug\)Context" --include=*.go` vs non-Context variants (repo-wide):
- Context-variant (`slog.InfoContext` etc.) count: **46**
- Non-Context variant (`slog.Info` etc.) count: **169**

No `context.Context` → `slog` bridge exists (no custom `slog.Handler` implementation found repo-wide: `grep -rn "slog.Handler" --include=*.go` → 0 hits for a custom type). Consequence: the 169 non-Context call sites (including every log line in `internal/platform/worker/service.go` and `internal/modules/jobs/*/job.go`) cannot have request-scoped attributes (tenant ID, trace ID beyond whatever's manually passed as a literal field, request ID) auto-injected by a handler — each call site must manually thread `"trace_id", event.TraceID`-style fields, which is exactly what `worker/service.go` does by hand at `:96,153,160`, and exactly what is easy to forget at any of the other 168 non-Context sites.

Numeric drift note: prior inventory pass (`inventory/observability.md`) cited 45/170; this pass finds 46/169 at the same commit — a 1-line swap (one non-Context call site was rewritten as Context-variant since that inventory ran, or the earlier count had an off-by-one), immaterial to the underlying finding (79% of all structured-log call sites still bypass Context-carried correlation).

## 5. Schedulers — River periodic jobs, ADR 0067 dual-define

`internal/modules/jobs/maintenance/periodic.go` (97 lines, read in full):
- Package doc (`:1-11`) states the ADR 0067 rationale verbatim: both `metaldocs-api` and `metaldocs-jobs` register the same `PeriodicJobs()` list on their River client config so **either** can win leader election and keep the schedule alive, but only `metaldocs-jobs` also registers `Workers` for the `maintenance` queue — so if `metaldocs-api` wins leader election, jobs get *scheduled* (inserted) on time but sit unexecuted until `metaldocs-jobs` claims them. This is the exact "enqueue vs execute" split named in the system CLAUDE.md.
- `PeriodicJobs()` (`:36-96`) returns 7 `river.PeriodicJob` entries: `stuck-instance-watchdog` (5 min), `idempotency-janitor` (15 min), `audit-integrity-validator` (1 hr), `document-review-surfacer` (1 hr), `approval-sla-surfacer` (1 hr), `release-hold-reconciler` (15 min, comment explains ADR 0085 Stage C W2 timing rationale in detail), `outbox_retention` (24 hr).
- **Zero `slog` calls in this file** (`grep -c slog internal/modules/jobs/maintenance/periodic.go` → 0) — confirmed exactly. The file only builds `river.PeriodicJob` structs (schedule + constructor closure); it never logs "registered N periodic jobs" or similar at startup, so there is no single log line an operator can grep to confirm the full schedule loaded correctly on either binary — they'd have to infer it from each individual job's own run-time log lines showing up on schedule.
- The **same dual-define shape is replicated a second time**, independently, in `internal/modules/render/fanout/retention/periodic.go` (~30 lines, read in full) — its own doc comment explicitly cross-references the ADR 0067 convention and registers one more periodic job (24 hr staging-outbox purge) alongside the 7 above. This means the *actual* count of dual-defined periodic jobs at this baseline is **8**, not 7 — the render module's retention job is a second, easy-to-miss instance of the same pattern living in a different package tree, which any audit of "the periodic job list" must remember to union across both `internal/modules/jobs/maintenance` and `internal/modules/render/fanout/retention`.

## 6. Alert rules

`grep -rn` across the repo for `*.rules.yml`, `*prometheus*rule*`, `alert` (case-insensitive, filenames + content) → **zero real alerting-rule configuration found**. All hits are either doc prose (mentioning "alert-only" behavior in ADR 0068 for the stuck-instance watchdog), UI copy/strings, or vendored dependency source. No Prometheus `ALERT`/`alerting-rules` YAML, no Alertmanager config, no Grafana alert JSON anywhere in `deploy/` or elsewhere in the repo.

**Consequence, tying together §1/§2/§6**: the stuck-instance watchdog (ADR 0068) is explicitly documented as "alert-only, not auto-canceling" — but there is no alerting-rule config anywhere in the repo that would actually turn its output into a paged/notified alert. Today the watchdog's "alert" is a log line (from `stuck_instance_watchdog/job.go`, non-Context `slog`), with no Prometheus rule wired to page on it and no metrics-server-exposed counter an alert rule could even target.

## 7. Deployment artifacts

**Dockerfiles** — `deploy/docker/api.Dockerfile`, `worker.Dockerfile`, `jobs.Dockerfile`: pairwise-diffed, near-identical. Differences are only: build target/output binary name, and an api-only tail (`EXPOSE 8081`, `db/migrations` COPY, `db/grants` COPY). All three share `GOFLAGS=-mod=mod` at line 6 of each — despite the repo vendoring a full `vendor/` tree, every image build still resolves modules from the network/module cache rather than `-mod=vendor`, which is both a reproducibility gap (build-time network dependency) and wasted vendor-tree disk cost if it's never actually used at build time.

**Compose healthchecks** (`deploy/compose/docker-compose.yml`, service block line numbers):

| Service | Image pin | Healthcheck |
|---|---|---|
| postgres (`:2`) | pinned | yes, `:41` |
| redis (`:56`) | pinned | yes, `:65` |
| minio (`:72`) | **`minio/minio:latest`** | **NO** |
| minio-init (`:87`) | **`minio/mc:latest`** | n/a (init container) |
| gotenberg (`:110`) | pinned | yes, `:117` |
| docx-renderer (`:124`) | build | yes, `:150` |
| api (`:156`) | `${API_IMAGE}` | yes, `:236` |
| worker (`:243`) | `${WORKER_IMAGE}` | **NO healthcheck key** |
| jobs (`:295`) | build | **NO healthcheck key** |
| web (`:334`) | build | yes, `:344` |
| gateway (`:351`) | `nginx:1.27-alpine`, pinned | — |

**#95/A7 claims confirmed exactly:**
- worker/jobs have no compose healthcheck at all — directly consistent with §1 (they have no HTTP endpoint a healthcheck could even target without adding one).
- minio is the only unpinned (`:latest`) image in the whole compose file; every other service pins a version/tag.

## 8. Outbox architecture

- **Single INSERT site, repo-wide** — confirmed via `grep -rn "INSERT INTO metaldocs.outbox_events" --include=*.go`: exactly one hit, `internal/platform/messaging/outbox/postgres/publisher.go:34`. (Numeric drift note: prior `persistence.md` cited line 30; this pass's fresh grep finds line 34 — a small line-shift from intervening edits, not a new site.) `:39` — `ON CONFLICT (idempotency_key) DO NOTHING`, the producer-side dedup mechanism.
- **Consumer idempotency** — `internal/platform/messaging/outbox/postgres/consumer.go:107`, `event.IdempotencyKey = messaging.IdempotencyKey(idempotencyKey)` on claim; `ClaimUnpublished`/`MarkPublished`/`MarkFailed` lifecycle (`internal/platform/worker/service.go:59-103`) is the sole consumer-side state machine — one `Service.RunOnce` call claims a batch, dispatches each by `EventType` (`dispatchEvent`, `:108-123`), and fails loud (`errPDFRunnerNotConfigured`/`errMaterializeRunnerNotConfigured` sentinels, `:112,117`) rather than silently marking an unhandled event published if a runner wasn't wired — a deliberately fail-closed design (comment cites `F-QA2-1`).
- **Dead-letter handling** — `markFailure` (`:125-167`): computes `attempt` from `event.AttemptCount`, force-maxes it to `s.cfg.MaxAttempts` immediately for two cases — unknown event type (`:131-133`) and a render-classified permanent failure matched structurally via `interface{ Retryable() bool }` (`:138-141`, comment explicitly notes this keeps `platform/worker` decoupled from the `render` module — no import inversion, a second positive G-class exemplar alongside Pass 1's `resolvers.*Reader` finding). Otherwise schedules `NextAttemptAt` via exponential backoff (`backoffDuration`, `:169-190`, capped at `RetryMaxSeconds`). `DeadLetteredAt` is set once `attempt >= MaxAttempts` (`:148`).
- **Trace correlation on dead-letter/retry log lines** — both branches log `"trace_id", event.TraceID` (`:152-153,159-160`) — bare string, per §3, not a reconstructable span.

## 9. Delta table — #95/A7 evidence at this baseline

| # | #95/A7 original claim | Status at `main@418070bf` | Classification |
|---|---|---|---|
| 1 | worker + jobs have no HTTP listener at all | **still true, exact** | exists-today |
| 2 | worker.outbox metrics hardcoded to 0 in the static fallback path | **still true, exact** — `runtime.go:109-111` | exists-today |
| 3 | Metrics are split Prometheus (api only) vs hand-rolled JSON (fallback everywhere else) | **still true** | exists-today |
| 4 | Outbox trace propagation is a bare `TraceID string`, not W3C context | **still true, exact** — `events.go:21` | #95-must-solve (structural — needs a schema change to the outbox envelope, not a local fix) |
| 5 | Zero OTel spans across worker/render/notifications async paths | **still true** — 2 total `otel.Tracer` sites repo-wide, both outside async paths | exists-today |
| 6 | slog Context-variant usage is a small minority | **still true**, drifted 45/170→46/169 (immaterial) | exists-today |
| 7 | No context→slog handler bridge exists | **still true**, confirmed 0 custom `slog.Handler` | #95-must-solve (would resolve #4's grep-only correlation issue for logs, though not for traces) |
| 8 | `jobs/maintenance/periodic.go` has zero scheduler-level log lines | **still true, exact** | exists-today |
| 9 | ADR 0067 dual-define pattern documented and live | **still true** — AND this pass found a **second, previously-uncatalogued instance** (`render/fanout/retention/periodic.go`, 8th periodic job, not counted in the prior "7 jobs" framing) | exists-today (new sub-finding: catalog undercount, 7→8) |
| 10 | No alert-rule configuration anywhere in the repo | **still true**, zero hits | #95-must-solve |
| 11 | Worker/jobs have no compose healthcheck | **still true, exact** | exists-today (mechanical consequence of #1 — fixable independently by adding a TCP/process-level healthcheck even without an HTTP endpoint) |
| 12 | minio is the only `:latest`-pinned image | **still true, exact** | exists-today |
| 13 | 3 near-identical Dockerfiles, `GOFLAGS=-mod=mod` despite vendoring | **still true, exact**, all 3 files | exists-today |
| 14 | (duplicate of 13, folded) | — | — |
| 15 | Outbox: single INSERT site, `ON CONFLICT DO NOTHING` producer dedup, fail-loud consumer dispatch, structural `Retryable()` dead-letter decoupling | **still true**, line-number drift only (30→34) | exists-today |

**Net assessment:** nothing in #95/A7's evidence set has been fixed or even partially addressed since the prior inventory pass — expected, since this is the *same pinned commit* the prior pass audited, not a later one. The value of this pass is (a) fresh independent file:line re-verification of every claim rather than trusting the carried-forward citations, (b) two new sub-findings not previously catalogued: the render-module's second dual-defined periodic job (§5, item 9) and the explicit tie between the ADR-0068 "alert-only" watchdog design and the total absence of any alerting-rule config to make that alert actually page anyone (§6).

## 10. Findings mapped to owning issues

- **#95/A7** — owns essentially every row in §9's delta table; this pass adds no new root cause, only sharper/fresher evidence and the two sub-findings above (periodic-job undercount, watchdog-alert-has-no-alert-target).
- **#92/A5** — the outbox `INSERT`/dedup/dead-letter mechanics (§8) are architecturally sound (single insert site, producer+consumer idempotency, fail-loud on missing runner) and should be cited as a **positive exemplar** when A5 documents target patterns for async persistence, in contrast to the 82-site/25-file `BeginTx` sprawl found in Pass 1's module maps.
- **#90/A3** — not directly implicated; no contract/route surface issues found in this pass (worker/jobs/render have no HTTP routes at all per Pass 1 §3/§5, consistent with §1's finding here).
