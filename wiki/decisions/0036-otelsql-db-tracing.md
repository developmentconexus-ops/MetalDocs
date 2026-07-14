# ADR 0036 — Driver-Level DB Tracing via `otelsql`

> **Status:** Accepted 2026-06-16
> **Last verified:** 2026-06-16
> **Scope:** Standard approach for adding OTel child spans to all `database/sql` queries in `metaldocs-api`. Covers `internal/platform/db/postgres/connect.go` only — not repository layer, not service layer.
> **Key files:**
> - `internal/platform/db/postgres/connect.go` — `Open` func; `otelsql.Open` replaces `sql.Open`

## Context

Feature F2.3 (M2 observability) requires DB call child spans to appear under the HTTP envelope
span so an SRE can answer "where is this request slow?" at the DB layer. Prior to F2.3,
`MetricsHandler` and all application flows produced only one span per request (the `otelhttp` HTTP
envelope). DB queries were completely invisible in traces.

MetalDocs uses `database/sql` throughout (`pgx/v5/stdlib` driver registered as `"pgx"`). The OTel
SDK is already bootstrapped at `main.go:112` (`observability.SetupOTel`). The question is: where and
how to inject DB-level spans?

Three alternatives were evaluated:

| Alternative | Ruling |
|-------------|--------|
| **Manual `tracer.Start()` per repository function** | Ruled out: requires every repository author to add spans; creates an ongoing discipline problem; misses existing hot paths; no coverage inheritance for future repos. High maintenance, low reliability. |
| **Repository-layer base struct with embedded tracer** | Ruled out: requires touching all repository constructors; still author-opt-in; mixes tracing concern into the persistence layer instead of the composition root. HS-2 risk (interface changes across all repos). |
| **`otelsql` driver wrapper at composition root** (`internal/platform/db/postgres/connect.go`) | **Selected.** One change at the `sql.Open` seam; all queries traced automatically — past and future. No repository or service changes required. Composition-root injection matches M2 architectural constraint. Matches industry pattern (Stripe, Datadog, GitHub Go stacks). |

## Decision

Replace `sql.Open("pgx", dsn)` with `otelsql.Open("pgx", dsn, ...)` in `internal/platform/db/postgres/connect.go`. The `github.com/XSAM/otelsql` package wraps the already-registered `"pgx"` driver and emits OTel child spans for every `database/sql` call. Span names follow the library's own conventions (`sql.connector.connect`, `sql.conn.ping`, `sql.conn.query`, `sql.conn.exec`, `sql.conn.prepare`, `sql.conn.begin_tx`, `sql.conn.reset_session`, `sql.rows`, etc. — `vendor/github.com/XSAM/otelsql/methods.go:25-47`). Return type stays `*sql.DB`; all callers are unchanged.

`OTEL_TRACES_EXPORTER` env var continues to gate the entire OTel pipeline — when unset (dev/production without tracing configured), the `otelsql` wrapper emits spans to the NOOP tracer; zero overhead, zero output.

The `_ "github.com/jackc/pgx/v5/stdlib"` blank import remains in `connect.go` to ensure the `"pgx"` driver is registered before `otelsql` wraps it.

## Consequences

- **All DB queries traced by default.** New repositories and services inherit coverage automatically — no per-author action required.
- **Connection DSN excluded from span attributes** (`otelsql.WithDBConnectStringAttribute(false)` or equivalent) to prevent credential leakage in traces.
- **Span cardinality is bounded.** `otelsql` defaults use operation type as span name (e.g. `sql.conn.query`, `sql.conn.exec`), not query text. Full query text is an opt-in attribute (`otelsql.WithSQLCommenter` / `otelsql.WithAttributes`) — **not enabled by default** in this ADR to avoid cardinality explosion and PII risk.
- **One new dependency** (`github.com/XSAM/otelsql`) added to `go.mod`. Maintained by the community; tracks OTel SDK releases.
- **Manual semantic spans** (`cd.create`, `signoff.record`) are layered on top of driver-level spans — they provide business-domain labels that SQL text cannot. This ADR governs only the driver layer; semantic spans are an independent concern.
