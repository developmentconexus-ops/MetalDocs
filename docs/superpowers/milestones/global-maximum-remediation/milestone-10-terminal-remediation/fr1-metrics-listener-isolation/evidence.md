# F-R1 Evidence — Metrics listener isolation (Dim 9 → CONFIRMED)

Closes the Dim-9 DEBT: `/metrics` no longer shares the public API listener. It is served on a
**dedicated** `http.Server` bound to `METRICS_ADDR` (default `:9090`); the public server has no
`/metrics` route. Isolation is now a property of process topology, not ops/ingress discipline.

## Changes

| File | Change |
|------|--------|
| `internal/platform/config/server.go` | `ServerConfig.MetricsAddr`; `LoadServerConfig` reads/validates `METRICS_ADDR` (default `:9090`, `host:port` form, port 1..65535). |
| `apps/api/cmd/metaldocs-api/main.go` | Removed the top-level rootMux `/metrics` dispatch from the public server. Added a dedicated metrics `http.Server` (`serverCfg.MetricsAddr`, `Recovery(mux{GET /metrics → PrometheusHandler})`). Both listeners start; both drain on shutdown; a fatal bind error on either is fail-fast. |
| `apps/api/cmd/metaldocs-api/main.go` (`shutdownServer`) | Signature now takes the metrics server + its error channel; selects on both listen errors; drains both. |
| `apps/api/cmd/metaldocs-api/main_test.go` | Three `shutdownServer` callers updated to the new signature. |
| `apps/api/cmd/metaldocs-api/metrics_endpoint_test.go` | Rewritten → `TestMetricsEndpoint_DedicatedListener_REQF83`: public handler must NOT 200 on `GET /metrics`; dedicated metrics handler serves 200 Prometheus, unauth, families present, no self-count. |
| `internal/platform/config/server_test.go` | Added `TestLoadServerConfig_MetricsAddr` (default/valid/invalid cases). |
| `deploy/compose/docker-compose.yml` | `api` gains `METRICS_ADDR: ":9090"`; the port is deliberately NOT host-published (only `${APP_PORT}:8081` is). |
| `ops/DEPLOY.md` | New "metrics listener isolation" section — two-listener table, scrape target `api:9090`, do-not-publish warning. |

## Structural proof — single Prometheus call site, on the metrics server

```
$ grep -n 'PrometheusHandler()' apps/api/cmd/metaldocs-api/main.go
887:	metricsMux.Handle("GET /metrics", httpObs.PrometheusHandler())
```

Exactly one call site; it is on `metricsMux` (the dedicated server), not on the public API mux. The
public `http.Server.Handler` is the chained API handler only.

## Gate output

```
$ go build ./...
BUILD_OK   (exit 0)

$ go test ./apps/api/cmd/metaldocs-api/ -run 'TestMetricsEndpoint|TestShutdownServer' -count=1
ok  	metaldocs/apps/api/cmd/metaldocs-api	5.637s

$ go test ./internal/platform/config/ -run TestLoadServerConfig -count=1
ok  	metaldocs/internal/platform/config	1.678s
```

TDD: the rewritten `TestMetricsEndpoint_DedicatedListener_REQF83` fails against the pre-F-R1 main
(public handler served `/metrics` = 200), passes after the split. First run also caught a real test
artifact — driving `GET /metrics` through the public chain records a legitimate `route="/metrics"`
401 sample, so the self-count assertion was reordered to scrape *before* the public-port probe (that
is the correct invariant: the scrape listener itself never self-counts).

## Live drive (labeled REAL — running compose stack, new image)

Rebuilt the `api` image with F-R1 and recreated the container (`docker compose --env-file .env -f
deploy/compose/docker-compose.yml build api && up -d api`; container reached `health=healthy`).
Probed both listeners **from inside the container** — exactly how an infra-network Prometheus reaches
`:9090` (the metrics port is not host-published):

```
=== PUBLIC :8081 /metrics (expect NOT 200) ===
  HTTP/1.1 401 Unauthorized
wget: server returned error: HTTP/1.1 401 Unauthorized

=== METRICS :9090 /metrics (expect 200 Prometheus) ===
go_goroutines 21
metaldocs_http_requests_total{method="GET",route="/api/v1/health/live"} 3
metaldocs_http_requests_total{method="GET",route="/metrics"} 1
--- response status ---
  HTTP/1.1 200 OK
  Content-Type: text/plain; version=0.0.4; charset=utf-8; escaping=underscores
```

- **Public port** `GET /metrics` → **401** (fail-closed auth chain; the scrape surface has left the
  public listener — pre-F-R1 this returned 200 unauthenticated).
- **Metrics port** `GET /metrics` → **200** `text/plain` Prometheus exposition with the
  `metaldocs_http_*` families and Go collectors.
- The `route="/metrics"` series present in the scrape is the count of the **public-port** 401 probe
  above (a real request to the public port), **not** a self-scrape of `:9090` — the metrics listener
  is not on the instrumented chain, so scraping it adds no `metaldocs_http_*` sample.

Note: pointing a *local* binary at the running compose Postgres failed SASL auth (the compose volume
was initialized with a password that differs from the current local-dev `.env` `PG*`); rather than
handle that secret, the live drive was run against the real container runtime, which is the stronger
proof anyway (actual deployment topology).

## Defers / follow-ups

None. The metrics listener stays intentionally unauthenticated (isolation is by network, per spec
non-goals); TLS/ServiceMonitor are post-v1 (compose stack only).
