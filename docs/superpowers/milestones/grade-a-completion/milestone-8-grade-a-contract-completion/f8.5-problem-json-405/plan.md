# Feature F8.5 — Plan

> Engine: subagent-driven-development equivalent, inline TDD (small DB-free middleware).
> Input: approved `spec.md`.

## Plan

### Files touched
- **CREATE** `internal/platform/middleware/method_not_allowed.go` — `MethodNotAllowedJSON(next http.Handler) http.Handler` + unexported `problemInterceptor` response-writer wrapper.
- **CREATE** `internal/platform/middleware/method_not_allowed_test.go` — DB-free `httptest` tests against a real Go 1.22 method-prefixed `ServeMux`.
- **EDIT** `apps/api/cmd/metaldocs-api/main.go` — wire `platformmw.MethodNotAllowedJSON` into the chain. Placement: innermost (closest to mux) so it sees the stdlib mux's raw 404/405 before any outer layer, and runs for every route. Add as the last entry of `apiChain(...)` (after the global limiter envelope wrap) — i.e. nearest the mux.

### Design (interceptor)
`problemInterceptor` embeds `http.ResponseWriter`:
- `WriteHeader(code)` — if `code ∈ {404,405}` **and** `Content-Type` begins `text/plain` → set `intercepted=true`, capture `status`, **swallow** (do not flush). Else delegate.
- `Write(b)` — if `intercepted` → discard (return `len(b), nil`). Else delegate.
- `Flush()` / `Hijack()` — delegate when the underlying writer supports the interface (preserve SSE / websocket upgrade).
After `next.ServeHTTP`, if `intercepted`: re-emit via `problem.Write` — 405 → `CodeMethodNotAllowed` (the mux-set `Allow` header is still on `w.Header()`, preserved); 404 → `CodeNotFound`. Hand-coded `application/problem+json` 404/405 never match the `text/plain` guard → pass through untouched.

### Test strategy (TDD, red first)
1. `TestMethodNotAllowedJSON_RewritesStdlib405` — mux `GET /thing`; `DELETE /thing` → 405, `Content-Type: application/problem+json`, `Allow: GET` preserved, body `code=METHOD_NOT_ALLOWED`.
2. `TestMethodNotAllowedJSON_RewritesStdlib404` — unknown path → 404 problem+json, `code=NOT_FOUND`.
3. `TestMethodNotAllowedJSON_PassesThroughHandcodedProblem` — handler emits `problem.Write(405)`; assert single problem+json body unchanged (no double-write, Allow intact if set).
4. `TestMethodNotAllowedJSON_PassesThroughSuccess` — 200 handler body + content-type intact.
5. (interface) `TestMethodNotAllowedJSON_PreservesFlusher` — wrapper exposes `http.Flusher` when underlying does.

### Ordering
test (red, compile-fail) → implement middleware (green) → wire main.go → `go build ./...` + `go test ./internal/platform/middleware/...` + `go vet` → review → evidence.
