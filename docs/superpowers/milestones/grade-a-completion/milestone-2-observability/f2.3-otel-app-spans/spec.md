# Feature F2.3 — Spec

> **Milestone:** 2 — Composition / observability  ·  **Folder:** `f2.3-otel-app-spans`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-16 — leandrotca.work — approved 2026-06-16.

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

Driven inline (brainstorming engine flow; one question at a time, persisted below). Seed: M2
`milestone.md` F2.3 row + wiki module research on critical business flows.

| # | Question | Answer |
|---|----------|--------|
| 1 | DB instrumentation approach — `otelsql` driver wrapper (auto-instruments all `database/sql` calls at composition root, one bootstrap change) vs manual `tracer.Start()` per repository function? | **`otelsql` driver wrapper (A).** Industry consensus (Stripe, Datadog, GitHub, Shopify). One `sql.Register` at `internal/platform/db/postgres/connect.go` wraps the `pgx` stdlib driver; every query traced automatically including future ones. Composition-root injection, not in-package wiring — matches M2 architectural constraint. Manual spans are for *semantic* critical-flow labels only (not for DB queries). |
| 2 | Which 2 critical request flows get manual semantic spans? | **`RecordSignoff` + `AtomicCreateControlledDocument` (Create).** Rationale: `RecordSignoff` (`decision_service.go:150`) is the highest DB-intensity flow (6+ tables, idempotency replay path invisible without a span, ISO SoD + quorum logic, PDF dispatch outbox in-tx, P0 if stuck). `Create` (`controlleddocuments/service.go:147`) is the entry point to the entire document lifecycle — sequence allocation (`cd_sequence_counters`) is a lock-contention hotspot, cross-module port call (CD → `documents.CreateDocumentTx`) in-tx latency invisible in the HTTP envelope. Together they cover the two highest-value patterns: a multi-step state-machine transition (signoff/quorum) and an atomic multi-table creation (sequence allocation + FK cascade). Based on full wiki module audit of all domain flows. |
| 3 | Test strategy — `tracetest.SpanRecorder` (OTel SDK's own in-process recorder), custom mock exporter, or OTLP-to-collector in tests? | **`tracetest.SpanRecorder`** (`go.opentelemetry.io/otel/sdk/trace/tracetest`). OTel SDK's canonical test package. Zero network, deterministic, no mock to maintain. Test creates `TracerProvider` with the recorder, runs the handler/service, asserts `sr.Ended()` for span names + attributes. Used by Grafana, Honeycomb, and the OTel SDK itself. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - **Primary — SRE / operator viewing a trace** in Grafana, Honeycomb, or Datadog. They need to
    see child spans under the HTTP envelope — at minimum DB query spans and semantic business-flow
    spans — to answer: "Where is this request slow?" and "Is this a signoff or a CD creation?"
  - **Secondary — handler integration test** asserting the span tree shape for the two named flows.

- **Contract:**
  1. Every `database/sql` query executed in the `metaldocs-api` process emits an OTel child span
     under the active trace context. Span names follow `github.com/XSAM/otelsql` defaults:
     `sql.connector.connect`, `sql.conn.ping`, `sql.conn.query`, `sql.conn.exec`,
     `sql.conn.prepare`, `sql.conn.begin_tx`, `sql.conn.reset_session`, `sql.rows`, etc.
     (`vendor/github.com/XSAM/otelsql/methods.go:25-47`). No filtering — all queries traced.
  2. `ControlledDocumentService.Create` wraps its body in a manual span named **`cd.create`**
     (`internal/modules/controlleddocuments/application/service.go:147`). The span carries at
     minimum the attribute `document.profile_code` (string). The span is a child of the incoming
     request context; if no trace is active, the span is a root span (not an error).
  3. `DecisionService.RecordSignoff` wraps its body in a manual span named **`signoff.record`**
     (`internal/modules/documents/approval/application/decision_service.go:150`). The span carries
     at minimum the attribute `signoff.verdict` (string: `"approved"` or `"rejected"`). Same
     parent-or-root rule as above.
  4. Both manual spans set `status = Error` and record the error if the wrapped function returns
     a non-nil error.
  5. The `otelsql` driver wrapper is registered in `internal/platform/db/postgres/connect.go` at
     the `sql.Register` seam — not inside any repository or service. `pgdb.Open` continues to
     return `*sql.DB`; callers are unchanged.
  6. Existing `otelhttp` HTTP envelope spans, W3C propagation, and env-gated `SetupOTel` are
     **unmodified**. OTel remains inert (NOOP) when `OTEL_TRACES_EXPORTER` is unset.
  7. The existing `"items"`, `"runtime"`, and `"scheduler"` keys in `/api/v1/metrics` are
     **unmodified**. This feature touches no metrics payload.

- **Source of truth for the contract:**
  - OTel bootstrap → [`internal/platform/observability/otel.go:39`](../../../../../internal/platform/observability/otel.go) (`SetupOTel`).
  - DB driver registration → [`internal/platform/db/postgres/connect.go:12`](../../../../../internal/platform/db/postgres/connect.go) (`Open`).
  - CD Create service → [`internal/modules/controlleddocuments/application/service.go:147`](../../../../../internal/modules/controlleddocuments/application/service.go).
  - Signoff service → [`internal/modules/documents/approval/application/decision_service.go:150`](../../../../../internal/modules/documents/approval/application/decision_service.go).

## What this feature implements

Add app-level OTel child spans to a pragmatic A− bar. Concrete changes, scoped to four files:

1. **`internal/platform/db/postgres/connect.go`**
   - Add `github.com/XSAM/otelsql` dependency (`v0.42.0`).
   - Replace `sql.Open("pgx", dsn)` with `otelsql.Open("pgx", dsn, otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL))`.
   - `openWithOptions` (the shared helper) uses `otelsql.Open`; `OpenWithTracerProvider` additionally passes `otelsql.WithTracerProvider(tp)`.
   - No change to pool settings, ping, or return type.

2. **`internal/modules/controlleddocuments/application/service.go`**
   - Add `tracer.Start(ctx, "cd.create")` span wrapping `Create` body.
   - Attribute: `attribute.String("document.profile_code", cmd.ProfileCode)`.
   - `defer span.End()`, set error status on non-nil return.

3. **`internal/modules/documents/approval/application/decision_service.go`**
   - Add `tracer.Start(ctx, "signoff.record")` span wrapping `RecordSignoff` body.
   - Attribute: `attribute.String("signoff.verdict", string(req.Decision))`.
   - `defer span.End()`, set error status on non-nil return.

4. **New test files:**
   - `internal/platform/db/postgres/connect_otel_test.go` — asserts that `Open` produces a `*sql.DB` whose driver emits an `otelsql`-generated span when a query executes (via `tracetest.SpanRecorder`).
   - `internal/modules/controlleddocuments/application/service_otel_test.go` — asserts `cd.create` span name + `document.profile_code` attribute on a successful `Create` call (stub dependencies; `tracetest.SpanRecorder`).
   - `internal/modules/documents/approval/application/decision_otel_test.go` — asserts `signoff.record` span name + `signoff.verdict` attribute on a successful `RecordSignoff` call (stub dependencies; `tracetest.SpanRecorder`).

## Non-goals (mandatory)

- **No full distributed-tracing rollout.** Only `cd.create` and `signoff.record` get manual spans.
  No sweep of all handlers, services, or repositories. HS-2 trigger if scope drifts.
- **No baggage propagation across external HTTP calls.** W3C `traceparent` is already wired via
  `SetupOTel`; this feature does not add outbound propagation to third-party clients.
- **No exporter reconfiguration.** Console / OTLP / NOOP selection stays 100% env-var driven
  (`OTEL_TRACES_EXPORTER`) — no code change to `SetupOTel`. The feature is inert in production
  when OTel is disabled.
- **No span naming taxonomy redesign.** `otelsql` defaults + two named manual spans. No catalog,
  no registry, no per-team conventions. HS-2 trigger.
- **No metrics payload change.** `/api/v1/metrics` shape is not touched.
- **No change to `otelhttp` HTTP envelope span.** Already working; no modification.
- **No new endpoint.** All observability flows through the existing OTel pipeline.
- **No touch to M3 / M4 findings.** Code quality and module-port work stays in their milestones.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| 1. `OpenWithTracerProvider` produces a `*sql.DB` whose driver emits at least one `sql.*` span (driver-level auto-instrumentation confirmed). No live DB required — `github.com/XSAM/otelsql` emits `sql.connector.connect` even on connection failure. | `go test ./internal/platform/db/postgres/ -run TestOpen_EmitsOTelSpan -count=1` — unreachable DSN, calls `PingContext`, asserts `SpanRecorder.Ended()` has ≥ 1 span with name prefix `sql.`. Always runs (no skip). | fixture (in-process recorder) |
| 2. `Create` emits a `cd.create` span with `document.profile_code` attribute. | `go test ./internal/modules/controlleddocuments/application/ -run TestCreate_EmitsCdCreateSpan -count=1` — stub deps, call `Create`, assert span name `"cd.create"` and attribute `document.profile_code` present in `SpanRecorder.Ended()`. | fixture |
| 3. `RecordSignoff` emits a `signoff.record` span with `signoff.verdict` attribute. | `go test ./internal/modules/documents/approval/application/ -run TestRecordSignoff_EmitsSignoffRecordSpan -count=1` — stub deps, call `RecordSignoff`, assert span name `"signoff.record"` and attribute `signoff.verdict` present. | fixture |
| 4. Error path — `Create` returning error sets span status to `Error`. | `go test ./internal/modules/controlleddocuments/application/ -run TestCreate_SpanStatusError_OnFailure -count=1` — stub `Create` to return error; assert `span.Status().Code == codes.Error`. | fixture |
| 5. Whole-repo regression — no existing test broken by driver wrapper. | `go test ./...` exits 0; no FAIL lines. | fixture |
| 6. Runtime proof — trace tree captured with OTel console exporter showing child spans under HTTP envelope. | Start API with `OTEL_TRACES_EXPORTER=console`; trigger `POST /api/v1/controlled-documents` (CD create) via curl with auth; capture stdout trace JSON; paste verbatim snippet showing `cd.create` child span + at least one `sql.*` grandchild span (e.g. `sql.connector.connect`, `sql.conn.exec`) into `evidence.md`, labeled **real-provider**. | **real-provider** |

> TDD: rows 1–4 are failing tests, written first. Row 5 is regression guard. Row 6 is runtime
> evidence from a real `start-api.ps1` run with console exporter.

## ADR needed?

- [x] Durable decision — recorded as **ADR 0036** [`wiki/decisions/0036-otelsql-db-tracing.md`](../../../../../wiki/decisions/0036-otelsql-db-tracing.md) — Accepted 2026-06-16.
