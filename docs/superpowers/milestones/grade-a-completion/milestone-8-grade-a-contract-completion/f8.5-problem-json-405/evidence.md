# Feature F8.5 — Evidence (problem+json 404/405 interceptor; D-03)

> **Milestone:** 8  ·  **Feature:** `f8.5-problem-json-405`  ·  **Closed:** 2026-06-20
> **Contract:** `spec.md` (approved 2026-06-20). Plan: `plan.md`. No ADR (implements existing D-03).
> **Commit:** recorded at commit time below.

## What was implemented

- **Interceptor middleware** ([`internal/platform/middleware/method_not_allowed.go`](../../../../../internal/platform/middleware/method_not_allowed.go)) —
  `MethodNotAllowedJSON(next)` wraps the writer in `problemInterceptor`. On `WriteHeader`, a `404`/`405`
  whose `Content-Type` starts `text/plain` (the Go 1.22 `ServeMux`/`http.Error` signature) is **swallowed**
  (status captured, body discarded); after the handler returns the middleware re-emits it via
  `problem.Write` — `405 → CodeMethodNotAllowed`, `404 → CodeNotFound`. The mux-set `Allow` header is left
  on `w.Header()` and survives (`problem.Write` overwrites only `Content-Type`). `Flush`/`Hijack` delegate
  to the underlying writer so SSE / websocket-upgrade routes are unaffected.
- **Chain wiring** ([`apps/api/cmd/metaldocs-api/chain.go:25`](../../../../../apps/api/cmd/metaldocs-api/chain.go),
  [`main.go:594`](../../../../../apps/api/cmd/metaldocs-api/main.go)) — added named link `method_not_allowed`
  as the **innermost** layer (nearest the mux), so it catches the stdlib envelope before any outer layer and
  outer layers see canonical problem+json. Order is asserted by `chain_test.go` (REQ-MW-7) — updated in lockstep.

## Consumer contract satisfied

Method mismatch on a registered path → `405 application/problem+json` (RFC 9457) with the correct `Allow`
header; unknown path → `404 application/problem+json`. Hand-coded problem+json 404/405 (already canonical)
never match the `text/plain` guard and pass through byte-for-byte. No `text/plain` error envelopes remain on
method-routed endpoints.

## Verification

| Check (spec Validation Gate) | Named test / command | Result | Real vs fixture |
|------------------------------|----------------------|--------|-----------------|
| `DELETE` on GET-only route → problem+json + `Allow` | `TestMethodNotAllowedJSON_RewritesStdlib405` (real Go 1.22 method mux) | **PASS** | real |
| Unknown path → problem+json 404 | `TestMethodNotAllowedJSON_RewritesStdlib404` | **PASS** | real |
| Hand-coded problem 405 passes through unchanged (no double-write) | `TestMethodNotAllowedJSON_PassesThroughHandcodedProblem` | **PASS** | real |
| 200 success body/content-type intact | `TestMethodNotAllowedJSON_PassesThroughSuccess` | **PASS** | real |
| Streaming interface preserved | `TestMethodNotAllowedJSON_PreservesFlusher` | **PASS** | real |
| Chain order normative incl. new link | `TestAPIChainOrder_REQMW7` | **PASS** | real |
| Build + tests green | `go build ./...` ; `go test -count=1 ./internal/platform/middleware/... ./apps/api/cmd/metaldocs-api/...` | exit 0, both `ok` | real |
| Static | `go vet ./internal/platform/middleware/... ./apps/api/cmd/metaldocs-api/...` | no findings | — |

TDD red captured before implementation: `undefined: MethodNotAllowedJSON` (compile-fail) across all five
middleware tests; green after `method_not_allowed.go` added.

## Acceptance vs spec Validation Gate

| Acceptance criterion | Met? | Evidence |
|----------------------|------|----------|
| `DELETE` on GET-only route → problem+json + Allow | yes | `_RewritesStdlib405` |
| Unknown path → problem+json 404 | yes | `_RewritesStdlib404` |
| Real handler problem 405 passes through | yes | `_PassesThroughHandcodedProblem` |
| Build + tests green | yes | build/test/vet clean |

## Review disposition

- Spec-compliance review: PASS — global interceptor (not per-route in-body guards, per interview Q1); both
  404 and 405 rewritten (Q2); rewrite gated on `text/plain` + status so real handler bodies are never
  corrupted (Q3, mirrors `recovery.go` best-effort). Non-goals honored (no route migration, no new routes,
  no change to existing hand-coded problem 405s).
- Code-quality review: PASS — interceptor preserves `Flusher`/`Hijacker`; chain order kept normative and
  test-asserted; reuses `problem` package + existing `Code` constants; innermost placement justified inline.

## Bounded defers

None.
