# F-R1 Plan

## Files touched
1. `internal/platform/config/server.go` — add `MetricsAddr` field + `METRICS_ADDR` parse/validate.
2. `apps/api/cmd/metaldocs-api/main.go` — remove the rootMux `/metrics` dispatch from the public
   server; start a dedicated metrics `http.Server` on `MetricsAddr`; drain both on shutdown.
3. `apps/api/cmd/metaldocs-api/main.go` `shutdownServer(...)` — accept + Shutdown the metrics server;
   treat its bind error as fatal.
4. `apps/api/cmd/metaldocs-api/metrics_endpoint_test.go` — rewrite to the split contract.
5. `deploy/compose/docker-compose.yml` — `METRICS_ADDR` env on `api`; no host publish for it.
6. `ops/DEPLOY.md` — document the two-listener topology + scrape target.

## Order (TDD)
1. **Config first** — add `MetricsAddr`, default `:9090`, validate; unit-covered by build + existing config tests.
2. **Rewrite the test to the new contract** (fails against current main = red).
3. **main.go** — split the listeners; shutdown wiring → test green.
4. `go build ./...`, `go test -run TestMetricsEndpoint`.
5. Compose + DEPLOY docs.
6. Live drive via `start-api.ps1`; curl both ports.

## Test strategy
- Replace `buildComposedServerHandler` (which mounted /metrics) with:
  - `buildPublicHandler(httpObs)` — apiChain/buildChain harness, **no** /metrics route (mirrors new main).
  - `buildMetricsHandler(httpObs)` — `platformmw.Recovery(mux{GET /metrics → PrometheusHandler})` (mirrors new metrics server).
- Warm up one request through the public chain (httpObs.Wrap records a sample) so the family-present
  assertions on the metrics scrape are non-vacuous.
- Assert public `GET /metrics` = 401 (fail-closed chain), metrics `GET /metrics` = 200 + families +
  no `route="/metrics"` self-count.

## Risk / rollback
- Low. Additive config + a second listener; public surface strictly shrinks (loses /metrics).
- Rollback = revert the 6 files; no schema/contract/authz touched.
