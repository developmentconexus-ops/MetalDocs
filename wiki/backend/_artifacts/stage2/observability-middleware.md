# Stage 2 Evaluation — Observability & Middleware Chain

> **Theme:** observability-middleware
> **Findings covered:** F-01, F-16, F-17
> **RF mapping:** RF-1 (observability depth), RF-2 (middleware chain), RF-9 (timeouts / graceful shutdown)
> **REQ mapping:** REQ-MW-1, REQ-MW-2, REQ-MW-4, REQ-MW-5, REQ-MW-7, REQ-OBS-1..3, REQ-REL-1, REQ-REL-2
> **Evaluator role:** senior staff engineer — verdict another reviewer can sign off without re-doing the research
> **Written:** 2026-06-11

---

## How to read this document

Each section:

1. **Code confirmation** — file:line anchor re-verified against current code; any discrepancy with the register is noted.
2. **Standard** — the external reference the finding is judged against (RFC/spec/guide with identifier).
3. **Verdict** — one of KEEP / SIMPLIFY / REFACTOR / DELETE / DEFER, with rationale.
4. **Smallest correct fix** — minimum change that reaches the professional bar; anything larger than this is over-engineering for this codebase.
5. **Effort / blast-radius** — S/M/L + contained / module / cross-module / system.
6. **ADR needed?** — yes/no and why.
7. **Over-engineering check** — explicit note where a common "fix" would be disproportionate.

---

## F-01 — Middleware Chain Canon Inversion

### Code confirmation

Register anchor `main.go:598-602` confirmed current:

```
// main.go:598
var presenceWrapped http.Handler = httpObs.Wrap(rateLimiter.Wrap(mux))
if presenceBump != nil {
    presenceWrapped = presenceBump.Wrap(presenceWrapped)
}
handler := cors.Wrap(originProtection.Wrap(authMiddleware.Wrap(iamMiddleware.Wrap(presenceWrapped))))
```

Actual outermost → innermost order:

```
cors → originProtection → authMiddleware → iamMiddleware → [presenceBump] → httpObs → rateLimiter → mux
```

Three structural problems confirmed against source:

1. **`httpObs` (metrics + access log) sits inside authn and IAM.** `internal/platform/observability/http.go:59` — the Wrap function is called at line 598, *after* `authMiddleware.Wrap` and `iamMiddleware.Wrap` enclose it. Result: any request rejected by CORS (403), origin protection (403), or authn (401/403) produces zero metric increment and zero access log line. Platform/ops-config wiki §4 flow 2 confirms this explicitly.

2. **No outermost panic-recovery or request-ID middleware.** Neither `main.go:598-602` nor any platform package provides an outermost `http.Handler` that: (a) wraps `defer recover()`, (b) generates a request ID before the inner chain runs. The trace ID is created at `http.go:61-65` — inside `httpObs`, which is layer 6. A panic in CORS or authn produces an unstructured TCP reset with no log line and no request ID.

3. **No pre-auth IP-keyed rate limit for the login path.** The per-route token-bucket limiter (`platform/ratelimit/middleware.go`) is fully implemented but never mounted at the login endpoint; `RegisterRoutes` is called without rate-limit injection at `main.go:501`. The existing account-lockout logic (`authn/config.go:88-104`) applies per-account, not per-IP. An attacker probing many accounts from one IP with a single wrong guess per account is not throttled.

The register does not over-state the severity. All three are confirmed by direct code read.

### Standard

**Canonical middleware order** — `wiki/standards/backend-canon.md §2.3` and `wiki/architecture/backend-target-architecture.md §2.1` both define the required order as: panic-recovery → request-ID/trace-context → access-log/metrics → CORS → pre-auth IP rate limit → authn → authz → identity-keyed limits → presence/side-effects → handler. REQ-MW-1 through REQ-MW-5 are the specific normative requirements.

**Metrics completeness (RED method)** — Tom Wilkie's RED method (cited in Google SRE Workbook, "Monitoring Distributed Systems") requires *all* requests to be counted in the rate/error/duration signals, including rejections at the outermost layers. A metrics layer that sits inside authn produces materially misleading error-rate numbers: a 401 storm appears as zero errors.

**OWASP ASVS 4.0 §V2.2.1** (maps to NIST SP 800-63 §5.2.2): "Verify that anti-automation controls are effective at mitigating breached credential testing, brute force, and account lockout attacks. Such controls include … rate limiting … IP address restrictions … Verify that no more than 100 failed attempts per hour is possible on a single account." Per-account lockout alone does not satisfy the IP-address-restriction arm of this control. CWE-307 (Improper Restriction of Excessive Authentication Attempts) is the CWE mapping.

**Pre-auth rate limiting** — Bearer CLI rule `go_gosec_http_http_slowloris` and CWE-400 (Uncontrolled Resource Consumption) cover the case where no ReadTimeout is set, but the same resource-exhaustion argument applies at the application layer: without a pre-auth IP-keyed limit on `/api/v1/auth/login`, a multi-account credential-stuffing attack has no middleware-level barrier.

### Verdict: REFACTOR (P1)

The chain inversion is a correctness defect against REQ-MW-1, REQ-MW-4, and a security gap against REQ-MW-5. The three sub-problems are independent changes that share the same file:

**Sub-problem A (metrics outside auth) — REFACTOR, effort S.**
Move `httpObs.Wrap` to outermost-viable position: immediately outside CORS, inside the not-yet-existing panic-recovery wrapper. Single line change in `main.go:598-602`. Blast radius: contained — no module logic touched, only composition order.

```go
// target order
handler := panicRecovery.Wrap(
    requestID.Wrap(
        httpObs.Wrap(
            cors.Wrap(
                originProtection.Wrap(
                    authMiddleware.Wrap(
                        iamMiddleware.Wrap(presenceWrapped)))))))
```

**Sub-problem B (no panic-recovery / outermost request-ID) — REFACTOR, effort S.**
Add a minimal `platform/middleware/recovery.go` (≤40 lines): `defer recover()` → write 500 problem+json to a `statusWriter`, emit one slog line with the trace ID already in context. The request-ID middleware is already embedded in `httpObs`; moving `httpObs` to outermost position (sub-problem A) covers the request-ID propagation requirement without a separate middleware. If the team wants the trace ID available to CORS rejections too, a dedicated 5-line `requesttrace` wrapper can precede `httpObs`. That is optional; sub-problem A is sufficient for REQ-MW-1/2 in practice (CORS rejections are low value for tracing).

**Sub-problem C (no pre-auth IP rate limit on login) — REFACTOR, effort S.**
The `platform/ratelimit` middleware already implements IP-keyed token-bucket limiting. It needs to be instantiated with a tighter budget (e.g., 10 req/min/IP on the `/api/v1/auth/login` path) and inserted at the chain position between CORS/origin-protection and authn. This is a wiring change in `main.go` + one route-specific config. The existing per-route limiter pattern (`platform/ratelimit/middleware.go`) supports this exactly; no new code is needed.

**Chain-order test (REQ-MW-7):** write a unit test that builds the composed handler and asserts the layer order via reflection or a sentinel middleware that records call order. ~30 lines. Effort S.

**Blast radius:** The composition change in `main.go:598-602` is a contained, single-file change. No module business logic is touched. The new `platform/middleware/recovery.go` is a new file with no downstream imports.

### Smallest correct fix

1. Create `internal/platform/middleware/recovery.go` — panic-catch + 500 problem+json + slog. (~40 lines)
2. Edit `main.go:598-602` — reorder to: `panicRecovery → httpObs → cors → originProtection → preAuthRateLimit(login) → authn → iam → presenceBump → rateLimiter → mux`.
3. Add chain-order test to `main_test.go`. (~30 lines)

Total estimated diff: ~100 lines new, ~5 lines changed.

### Effort / blast-radius

Effort: **S** (all changes are in one composition file plus one new platform file).
Blast radius: **contained** — only `main.go` wiring and a new platform middleware package.

### ADR needed?

No. The target order is already normative in `backend-target-architecture.md §2.1`. This is implementation catch-up, not a design decision.

### Over-engineering check

Do not add a full middleware framework (chi, gorilla, etc.) to solve this. The existing `http.Handler` wrapping pattern works correctly. A purpose-built recovery middleware of 40 lines is the right scope. The pre-auth login rate limit does not require Redis or a distributed counter at this scale; the existing `platform/ratelimit` in-process token bucket is adequate until horizontal scale demands cross-replica coordination.

---

## F-16 — Server Timeout and Graceful-Shutdown Gaps

### Code confirmation

Register anchor `main.go:613-617` confirmed current:

```go
server := &http.Server{
    Addr:              addr,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
}
```

`ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are absent. Only `ReadHeaderTimeout: 5s` is set.

Graceful shutdown at `main.go:660-662`:

```go
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
defer shutdownCancel()
if err := server.Shutdown(shutdownCtx); err != nil {
```

15-second drain budget is present; `SIGTERM` is handled via `signal.NotifyContext` at `main.go:149`. Shutdown joins `schedulerWG` and `workerWG`.

The WebSocket presence stream at `/api/v1/iam/presence/stream` is noted `[runtime-unverified]` in the wiki: hijacked connections are not drained by `server.Shutdown` per the Go stdlib documentation: "Shutdown does not attempt to close nor wait for hijacked connections."

Secondary gap at `observability/runtime.go:281-325`: `applyDependencyChecks` creates individual `context.WithTimeout(ctx, 2s)` per check in a sequential loop. With N=1 (Gotenberg only) today, worst-case is 5s total (3s DB ping + 2s Gotenberg). Not a risk now; misleading contract for future additions.

Tertiary gap (`config/attachments.go:63-68`): `METALDOCS_ATTACHMENTS_SIGNING_SECRET` is required even for `StorageProviderMemory`, which never uses HMAC-signed URLs. Out of scope for this theme; flagged for completeness.

### Standard

**Cloudflare "The complete guide to Go net/http timeouts"** (https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/): explicitly states that `http.ListenAndServe` (and equivalent bare `http.Server` without timeouts) is unsuitable for production internet servers. Without `ReadTimeout`, a slow-body client holds the goroutine and the connection indefinitely. Without `WriteTimeout`, a slow reader holds the response buffer open.

**CWE-400 (Uncontrolled Resource Consumption)** / **OWASP API4:2023 Unrestricted Resource Consumption**: a server without `ReadTimeout` is vulnerable to Slowloris-class DoS. The Slowloris attack (CWE-400) works by opening many partial connections and never completing them. `ReadHeaderTimeout: 5s` (already set) mitigates the header-phase of Slowloris. The missing `ReadTimeout` leaves the body-read phase open; practically most API endpoints do not stream large bodies, so the exposure here is lower than a file-upload service, but it remains a correctness gap.

**REQ-REL-1** (backend-target-architecture.md §8): "Every outbound call (DB, Redis, MinIO, renderer, gotenberg) has an explicit timeout. Unbounded waits are review-blocking." The same principle applies inbound: the HTTP server is the entry point and must have bounded connection lifetimes.

**REQ-REL-2**: "metaldocs-api drains gracefully on SIGTERM: stop accepting, finish in-flight within a deadline, close pools." The 15s `server.Shutdown` budget satisfies this requirement for standard HTTP connections. The WebSocket presence stream (hijacked) is the exception.

**Go stdlib net/http Server documentation** (`pkg.go.dev/net/http#Server`): "If ReadTimeout is zero, there is no timeout." Confirmed for `WriteTimeout` and `IdleTimeout` as well.

### Verdict: REFACTOR (P1)

Three independent changes at different urgency levels:

**A — Set ReadTimeout, WriteTimeout, IdleTimeout — REFACTOR, effort S (P1).**
This is a 3-line change. Recommended values for a JSON API:
- `ReadTimeout: 30s` — allows large request bodies (document uploads); if uploads go through presigned MinIO URLs (which they should per REQ-BLOB-1), 10–15s is sufficient.
- `WriteTimeout: 60s` — covers large JSON responses; the export endpoints may produce sizeable payloads.
- `IdleTimeout: 90s` — standard keep-alive drain; Go doc: "If IdleTimeout is zero, ReadTimeout is used."

Note: the existing `ReadHeaderTimeout: 5s` is correct and should remain. Adding the others is strictly additive.

**B — WebSocket presence drain on SIGTERM — DEFER.**
The presence hub at `main.go:306` needs `server.RegisterOnShutdown` to signal open WebSocket connections to close, or the hub's `Run` goroutine (already stopped via root context cancellation at `main.go:670`) must close active upgrade connections before the shutdown deadline. The current code stops the hub via root context cancellation (line 670), which *does* stop the hub's send loop, but the OS-level TCP connections remain open until clients time out. In Kubernetes, SIGKILL follows SIGTERM after the termination grace period, so this is a DEFER not a P0. The fix is medium effort (needs connection tracking in the hub); the blast radius is module-level (presence module).

**C — Sequential dependency checks — SIMPLIFY, effort S (P3).**
Change `applyDependencyChecks` to run checks concurrently under a single shared budget (e.g., 3s total across all checks) using `errgroup` or goroutines + a `time.After`. Currently not a production risk (N=1), but the fix is small and makes the contract honest before a second dependency is added.

**D — Signing secret required for memory provider — DEFER (P3).**
The unconditional requirement is security-conservative (ensures secret is always set, preventing accidental misconfiguration when switching providers). DEFER: document the invariant with a comment; do not relax it until there is operational evidence of friction.

### Smallest correct fix (addressing the P1)

Edit `main.go:613-617`:

```go
server := &http.Server{
    Addr:              addr,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      60 * time.Second,
    IdleTimeout:       90 * time.Second,
}
```

Total diff: 3 lines added. No module logic touched.

For readiness probe sequential hazard (sub-problem C): ~10-line change in `observability/runtime.go:281-325` to run checks in goroutines under one shared context.

### Effort / blast-radius

Effort: **S** (main.go timeout fields) + **S** (readiness probe concurrency).
Blast radius: **contained** (main.go only for timeout; observability/runtime.go for probe).

### ADR needed?

No. REQ-REL-1 already requires explicit timeouts. This is a missing implementation, not a design decision.

### Over-engineering check

Do not add a per-handler timeout middleware for the P1 fix. The `http.Server`-level timeouts cover all routes uniformly. Per-route `http.TimeoutHandler` is appropriate for routes with known tight budgets (e.g., health checks at 1s), but that is a separate, optional P3 concern. The minimum fix (3 lines) is the right scope.

---

## F-17 — No OpenTelemetry Exporter / No W3C Trace Propagation

### Code confirmation

`go.mod` grep for `go.opentelemetry.io` returns zero matches — confirmed: no OTel dependency anywhere in the module. `internal/platform/observability/http.go` imports `metaldocs/internal/platform/requesttrace` and `metaldocs/internal/modules/auth/domain` only — no OTel imports confirmed.

The existing trace mechanism:
- `requesttrace.Normalize` reads `X-Trace-Id` header (line 61); generates a UUID if absent.
- Trace ID is stored in context and emitted to every `slog` access log line.
- The `outbox_events` table carries a `trace_id` column (referenced at `main.go:768, 801, 824, 831-833`), allowing async hops to link back to the originating request.

This is a custom, non-standard request-correlation mechanism. It provides:
- Per-request correlation *within* a single binary (log lines linkable by `trace_id`).
- Outbox → worker correlation if the worker reads the `trace_id` column.

It does not provide:
- W3C `traceparent` header propagation to/from upstream proxies or downstream services.
- OTel spans (start/end timestamps, structured attributes, parent/child span relationships).
- OTLP export to any backend.
- Prometheus metrics endpoint (current metrics surface is a custom JSON endpoint behind auth, at `/api/v1/metrics`).

The `normalizeRoute` function at `http.go:178-208` covers only 6 `if` branches (documents, document-profiles, workflow/documents, iam/users/roles, iam/users/reset-password, iam/users/unlock). All other parameterized routes (templates, taxonomy, approval, controlled-documents, etc.) log raw IDs. This inflates metric cardinality and leaks document/user IDs into log lines.

### Standard

**W3C Trace Context Level 2** (https://www.w3.org/TR/trace-context-2/): the industry standard for cross-service trace propagation. The `traceparent` header format carries version, trace-id (128-bit), parent-span-id (64-bit), and trace-flags. It is implemented by every major observability vendor and all OTel SDKs. Using a custom `X-Trace-Id` header is not interoperable: proxies, APMs, and service meshes cannot correlate spans.

**OpenTelemetry specification** (https://opentelemetry.io/docs/languages/go/): the vendor-neutral SDK standard for structured telemetry. REQ-OBS-3 directly cites W3C traceparent and OTel as the target. The Go SDK minimum setup for OTLP trace export is ~50 lines using `autoexport` (environment-configured): `OTEL_TRACES_EXPORTER=otlp`, `OTEL_EXPORTER_OTLP_ENDPOINT=<backend>`. No backend is required in the codebase; the endpoint is injected at runtime.

**Google SRE "Four Golden Signals"** (https://sre.google/sre-book/monitoring-distributed-systems/): Latency, Traffic, Errors, Saturation. The existing in-process RED counters at `/api/v1/metrics` partially cover Traffic (request rate), Errors, and Latency (p50/95/99). They do not cover Saturation (goroutines, DB pool depth, queue depth exposed externally), and they are not scrapeable by standard monitoring infrastructure (Prometheus, Grafana) because the endpoint is JSON-only and auth-gated. REQ-OBS-2 explicitly requires a Prometheus-compatible exposition.

**REQ-OBS-3**: "One trace context propagates edge → api → outbox → worker → docx-renderer via W3C `traceparent`, carried through the outbox row so async hops join the originating trace. OpenTelemetry is the export standard." Current state: partially addressed for the `outbox → worker` hop (trace_id in outbox row), not addressed for W3C propagation or span creation.

**OWASP ASVS 7.4.1** (Logging): every log line in a request path must carry correlation data. The existing `slog` + `trace_id` approach satisfies this for authenticated paths; the gap (metrics invisible for unauth paths, noted in F-01) is a REQ-OBS-1 violation, not a REQ-OBS-3 one.

### Verdict: REFACTOR (P2)

**Split into three independent sub-problems with different priority levels:**

**A — W3C traceparent propagation (inbound + outbound) — REFACTOR, effort M (P2).**

Replace the custom `X-Trace-Id` header read/write in `observability/http.go:61-65` with W3C `traceparent` propagation using the OTel Go SDK propagator. The existing `requesttrace` package provides the context key; the SDK's `propagation.TraceContext{}` propagator handles the header parsing. The `X-Trace-Id` custom header can be kept as a secondary emit for backward compatibility during the transition period. The outbound hop to `docx-renderer` needs `traceparent` injected into the fanout HTTP client headers.

This is the highest-value change because it makes the existing log correlation compatible with any OTel-aware infrastructure immediately, with no backend required.

**B — OTel SDK setup + OTLP exporter — REFACTOR, effort M (P2).**

Add to `go.mod`: `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk/trace`, `go.opentelemetry.io/contrib/exporters/autoexport`. Create `internal/platform/observability/otel.go` (~60 lines): `initTracerProvider(ctx)` → `autoexport.NewSpanExporter` → `sdktrace.NewTracerProvider` → `otel.SetTracerProvider`. Wire into `main.go` at startup before the first span. Export is entirely configuration-driven via `OTEL_TRACES_EXPORTER` and `OTEL_EXPORTER_OTLP_ENDPOINT` env vars — no backend is compiled in.

Rationale for P2 not P1: the binary runs without OTLP infrastructure today; the feature degrades gracefully to noop exporter when `OTEL_TRACES_EXPORTER=none` (the `autoexport` default when the env var is absent). The existing slog-based correlation covers the minimum ops need. The P2 classification is appropriate for a company that does not yet have an OTel backend running; this gives the implementation lead time for infrastructure prep.

**C — Prometheus metrics endpoint — REFACTOR, effort M (P2, same program as B).**

The existing `/api/v1/metrics` endpoint emits a custom JSON shape behind auth. Prometheus scraping requires an unauthenticated (or fixed-credential) `/metrics` endpoint exposing the Prometheus text format. Replace or augment the current endpoint with a `promhttp.Handler()` backed by OTel SDK metrics (which can emit Prometheus format via `go.opentelemetry.io/otel/exporters/prometheus`). The existing in-process ring-buffer aggregation becomes unnecessary after OTel metrics are wired.

**D — `normalizeRoute` coverage — REFACTOR, effort S (P2).**

The 6-branch custom route normalizer at `http.go:178-208` misses all parameterized routes in templates, taxonomy, approval, controlled-documents. Fix: expand the `normalizeRoute` function to cover every parameterized route pattern, or — better — after OTel is wired, use the `otelhttp.NewHandler` wrapper which extracts route labels from `r.Pattern` (available in Go 1.22+ `net/http.Request.Pattern`). The latter eliminates the handwritten normalizer entirely. Given the repo is on a recent Go version, this is the preferred path.

Note: route ID leakage into logs (e.g., `document_id` field in the access log) is intentional and useful for debugging. PII analysis: document IDs and profile codes are internal UUIDs/codes, not personal data. No change needed for the `extractRouteContext` emit.

### Smallest correct fix

For P2 (W3C + OTel), the bounded program is:

1. `go get go.opentelemetry.io/otel go.opentelemetry.io/otel/sdk/trace go.opentelemetry.io/contrib/exporters/autoexport go.opentelemetry.io/otel/propagators/b3 go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
2. `internal/platform/observability/otel.go` — `InitTracerProvider(ctx) (shutdown func(), err error)` (~60 lines)
3. `main.go` — call `InitTracerProvider` at startup, defer shutdown
4. `observability/http.go:59-65` — replace custom X-Trace-Id read with `otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))`; keep `X-Trace-Id` emit for backward compat
5. Fanout HTTP client — inject `otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))` on outbound calls to docx-renderer
6. Expand `normalizeRoute` to cover all parameterized routes (or use Go 1.22 `r.Pattern`)

Total estimated diff: ~150 lines new, ~20 lines changed.

### Effort / blast-radius

Effort: **M** (new SDK dependency, wiring across 3 files, fanout client change).
Blast radius: **cross-module** — `platform/observability`, `main.go`, fanout HTTP client in the approval/render area need coordinated changes.

### ADR needed?

No new ADR required. REQ-OBS-3 already mandates W3C traceparent and OTel. This is implementation work against an existing requirement. If the team defers OTel in favor of a non-OTel backend (e.g., Datadog agent with custom header), that deviation from REQ-OBS-3 needs an ADR.

### Over-engineering check

Do not add manual span creation to every application service function in the first pass. The minimum OTel value is: (1) HTTP middleware auto-instrumentation via `otelhttp.NewHandler` (gives every route a root span and propagates context), (2) OTLP export wiring so spans reach a backend. Database-level spans (pgx OTel plugin) and outbox-worker spans are a second pass. The 12-factor OTel setup via `autoexport` achieves this in ~60 lines with zero per-route code changes — do not expand scope to full manual instrumentation in the same PR.

The existing custom in-process RED metrics ring buffer (`observability/http.go:156-301`) does not need to be deleted immediately. It provides useful data at `/api/v1/metrics` and has no consumers that need migration. It can coexist until a Prometheus/OTel metrics endpoint is confirmed working, then be removed in a follow-up.

---

## Summary table

| Finding | Verdict | Priority | Effort | Blast radius | ADR needed |
|---|---|---|---|---|---|
| F-01: Middleware chain inversion (all 3 sub-problems) | REFACTOR | P1 | S | contained | No |
| F-16A: Missing ReadTimeout / WriteTimeout / IdleTimeout | REFACTOR | P1 | S | contained | No |
| F-16B: WebSocket presence drain on shutdown | DEFER | P3 | M | module | No |
| F-16C: Sequential readiness probe checks | SIMPLIFY | P3 | S | contained | No |
| F-17A: W3C traceparent propagation | REFACTOR | P2 | M | cross-module | No (REQ-OBS-3 exists) |
| F-17B: OTel SDK + OTLP exporter | REFACTOR | P2 | M | cross-module | Only if non-OTel backend chosen |
| F-17C: Prometheus metrics endpoint | REFACTOR | P2 | M | contained | No |
| F-17D: normalizeRoute coverage | REFACTOR | P2 | S | contained | No |

**Recommended sequencing within this theme:**

1. F-16A (3-line timeout fix) — lowest risk, highest severity; do this in any PR touching `main.go`.
2. F-01 sub-problems A+B (reorder chain + panic recovery) — prep work that F-17 instrumentation depends on (spans need to be created after request-ID is in context).
3. F-01 sub-problem C (pre-auth login rate limit) — security gap; standalone PR.
4. F-17 (OTel program) — requires infrastructure decision (which backend); block on that, not on code readiness.

---

## Sources

- W3C Trace Context Level 2: https://www.w3.org/TR/trace-context-2/
- OpenTelemetry Go exporters: https://opentelemetry.io/docs/languages/go/exporters/
- OpenTelemetry Go getting started: https://opentelemetry.io/docs/languages/go/getting-started/
- Cloudflare "The complete guide to Go net/http timeouts": https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
- Go net/http Server pkg docs: https://pkg.go.dev/net/http#Server
- Google SRE Book — Monitoring Distributed Systems: https://sre.google/sre-book/monitoring-distributed-systems/
- OWASP ASVS 4.0 V2.2 (anti-automation / brute force): https://github.com/OWASP/ASVS/blob/master/4.0/en/0x11-V2-Authentication.md
- OWASP API4:2023 Unrestricted Resource Consumption: https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/
- CWE-400 Uncontrolled Resource Consumption: https://cwe.mitre.org/data/definitions/400.html
- CWE-307 Improper Restriction of Excessive Authentication Attempts: https://cwe.mitre.org/data/definitions/307.html
- Bearer CLI Go Slowloris rule: https://docs.bearer.com/reference/rules/go_gosec_http_http_slowloris/
- Go graceful shutdown + hijacked connections: https://pkg.go.dev/net/http#Server.Shutdown ("Shutdown does not attempt to close nor wait for hijacked connections")
