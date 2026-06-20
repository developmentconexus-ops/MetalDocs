# Feature F8.5 — Spec (APPROVED)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.5-problem-json-405`
> **Status:** APPROVED 2026-06-20
> **Approved before code:** YES — 2026-06-20. Runtime truth confirmed: `problem.Write` sets
> `application/problem+json` (`problem.go:77-87`); `problem.CodeMethodNotAllowed` / `CodeNotFound`
> exist (`codes.go:21-22`); `WriteMethodNotAllowed` is the canonical 405 envelope incl. `Allow`
> (`httpresponse/response.go:20-25`); `recovery.go` establishes the best-effort response-rewrite
> precedent; chain wired at `apps/api/cmd/metaldocs-api/main.go:583` (`buildChain(mux, apiChain(...))`).
> Go 1.22 method-routed `ServeMux` emits `405`/`404` via `http.Error` → `Content-Type: text/plain;
> charset=utf-8` + `X-Content-Type-Options: nosniff`, with `Allow` set before the body on a 405 —
> this is the exact signature the interceptor keys on. Interceptor must preserve `http.Flusher` /
> `http.Hijacker` so streaming/upgrade routes are unaffected.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Fix per-route (in-body method checks) or one global interceptor? | **One global interceptor** — Go 1.22 method-prefixed `ServeMux` rejects the method before dispatch, so in-body guards can't run. A response-rewriting middleware fixes every method-routed endpoint at once. |
| 2 | Rewrite 404 too? | Yes — stdlib `404 text/plain` from `ServeMux` is the same envelope defect; convert both, preserve `Allow` on 405. |
| 3 | Risk of corrupting already-written bodies? | Interceptor only rewrites when status is 405/404 AND content-type is stdlib `text/plain` AND body not yet flushed by a real handler; otherwise passes through (mirror recovery.go best-effort guard). |

## Consumer contract (FIRST)

- **Consumer(s):** any client hitting a registered path with an unsupported method; FE error handling expects the canonical envelope.
- **Contract:** method mismatch → `405 application/problem+json` (RFC 9457) with the correct `Allow` header; unknown path → `404 application/problem+json`. Consistent across all modules (no `text/plain` 405s).
- **Source of truth:** `httpresponse.WriteMethodNotAllowed` + the `problem` package (`problem.go:77-87`); D-03 ("no bare-status error responses").

## What this feature implements

A new middleware in `internal/platform/middleware/` that wraps the mux, detects stdlib `405`/`404` `text/plain` responses, and rewrites them to `application/problem+json` (preserving `Allow`). Wired into the chain in `apps/api/cmd/metaldocs-api/main.go:580-591`.

## Non-goals (mandatory)

- No migration of method-prefixed routes to in-body checks (interceptor supersedes that need).
- No change to existing problem+json 405s already emitted by hand-coded modules.
- No new route registration.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| `DELETE` on GET-only route → problem+json + Allow | middleware test against a method-prefixed mux | real |
| Unknown path → problem+json 404 | middleware test | real |
| Real handler 405s (problem+json) pass through unchanged | middleware test with pre-written problem body | real |
| Build + tests green | `GOFLAGS=-mod=mod go build ./... && go test -count=1 ./internal/platform/middleware/...` | real |

## ADR needed?

- [x] No durable decision — skip (implements existing D-03 standard).
