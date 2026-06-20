# Feature F8.5 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.5-problem-json-405`
> **Status:** Drafting (seed from post-M7 re-audit Middleware Major #2, D-03; **interview + approval pending**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*

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
