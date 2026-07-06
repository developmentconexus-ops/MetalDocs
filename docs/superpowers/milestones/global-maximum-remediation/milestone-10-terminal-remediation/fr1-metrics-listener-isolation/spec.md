# F-R1 — Metrics listener isolation (Dim 9 → CONFIRMED)

> Consumer-contract-first. Approved **before** any code.
> **Approval:** APPROVED 2026-07-06 (operator "Go" on the M10 remediation; contract self-reviewed
> against `validation-contract.md` C3). — *filled before first F-R1 commit.*

## Problem (the DEBT this closes)

`/metrics` is registered on a top-level dispatch mux that fronts the **public** API `http.Server`
(`main.go:851-854`). Auth is bypassed *by route* (rootMux sits ahead of the chain), which is correct
for a credential-less Prometheus scrape — but the endpoint is served on the **same TCP listener** as
the public API. Compose host-publishes that listener (`docker-compose.yml:119-120`,
`"${APP_PORT}:8081"`), and the API port is reachable on the host directly (not only via the nginx
gateway, which proxies `/api/` and `/` but is a *separate* container). So any environment that exposes
`APP_PORT` beyond localhost exposes `/metrics` unauthenticated. Isolation currently depends on ops
discipline (ingress/firewall config), not on process topology — exactly the discipline-dependent-
invariant class this mission exists to eliminate.

## Consumers & the contract each requires

| Consumer | Required shape (the contract) |
|----------|------------------------------|
| **Prometheus scraper** (infra network) | Can `GET /metrics` **unauthenticated** on a **dedicated** listener address (`METRICS_ADDR`, default `:9090`) and receive `200` + Prometheus `text/plain` exposition with the `metaldocs_http_*` families and Go/process collectors. |
| **Public API client** (host / gateway) | The public listener **cannot** serve `/metrics` at all — `GET /metrics` on the public port is not an unauthenticated scrape surface (it enters the auth chain and fails closed / 404s). This is a **structural** guarantee: the public server's handler has no `/metrics` route registered. |
| **Operator reading DEPLOY** | Learns the two-listener topology: public port (published) vs metrics port (infra-only, **not** host-published); the scrape target is `api:9090` on the compose network. |
| **Composition root (main.go)** | Reads `METRICS_ADDR` from `ServerConfig`; starts a second `http.Server`; both listeners drain on graceful shutdown. |

## Contract details

- **Config:** `ServerConfig` gains `MetricsAddr string`. `LoadServerConfig()` reads `METRICS_ADDR`;
  unset → `:9090`; if set it must be a valid listen address (`host:port` or `:port`, port 1..65535),
  else a config error (fail-fast at boot, same discipline as `APP_PORT`).
- **Topology:** the public `http.Server.Handler` is the chained API handler **only** — no top-level
  `/metrics` dispatch mux. `PrometheusHandler()` is registered on a **separate** mux served by a
  **second** `http.Server` bound to `MetricsAddr`, wrapped in `platformmw.Recovery`. Grep-invariant:
  exactly one `PrometheusHandler()` call site in `main.go`, and it is on the metrics server, not the
  public mux.
- **Lifecycle:** both servers start; a fatal bind error on **either** is fail-fast; graceful shutdown
  drains both within the existing 15s shutdown budget.
- **Deploy:** compose adds `METRICS_ADDR` to the `api` service env and does **not** add a host
  `ports:` mapping for the metrics port (infra-network reachable only). `ops/DEPLOY.md` documents the
  split + scrape target.

## Non-goals (mandatory)

- No change to metric families, the Prometheus registry, `PrometheusHandler`/`MetricsHandler`
  internals, or the JSON `/api/v1/metrics` route.
- No auth added to the metrics listener — it stays unauthenticated (isolation is by network, not
  credentials); that is the deliberate design for a scrape target on an infra-only port.
- No Kubernetes, no ServiceMonitor, no TLS on the metrics port (post-v1; compose stack only).
- No change to the nginx gateway (it already does not proxy `/metrics`).

## Validation Gate

1. **TDD regression** (`metrics_endpoint_test.go`, rewritten):
   - public composed handler: `GET /metrics` → **not 200** (401 from the fail-closed auth chain).
   - dedicated metrics handler: `GET /metrics` → 200, `Content-Type` contains `text/plain`, body
     contains `metaldocs_http_requests_total`, `metaldocs_http_request_errors_total`,
     `metaldocs_http_request_duration_seconds`, and a `go_`/`process_` line; no `route="/metrics"`
     self-count series.
   - Named test: `TestMetricsEndpoint_DedicatedListener_REQF83`.
2. `go build ./...` → exit 0.
3. `go test ./apps/api/cmd/metaldocs-api/ -run TestMetricsEndpoint -count=1` → PASS.
4. **Live drive** (labeled real): start API via `.\scripts\start-api.ps1`; `curl` public port
   `/metrics` → not an unauth 200 scrape (401/404); `curl` `METRICS_ADDR` `/metrics` → 200 Prometheus.
5. Grep proof: single `PrometheusHandler()` call in `main.go`, on the metrics server.
6. Compose + DEPLOY reflect the split (metrics port not host-published; scrape target documented).

## ADR?

No new ADR. This is the structural completion of the M8 F8.3 observability decision (the scrape
surface already exists and is out-of-band per contract §3.2); F-R1 only moves it to its own listener.
The DEPLOY topology note + this spec are the durable record.
