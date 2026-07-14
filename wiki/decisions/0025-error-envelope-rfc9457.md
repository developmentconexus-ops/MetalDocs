# ADR 0025 — One error envelope: RFC 9457 Problem Details

> **Status:** Accepted 2026-06-08 (formalises anchor decision AD-2 of the API Contract Hardening Program; closes Plan 7's envelope rollout. Shipped across Phases D, E2, F). Retroactive ADR.
> **Last verified:** 2026-06-08
> **Scope:** The single error response shape for the entire HTTP API and the CI gates that keep it singular.
> **Out of scope:** Success-response shapes; base path (ADR 0024); authz (ADR 0023/0026).
> **Key files:**
> - `api/openapi/v1/openapi.yaml` — shared `#/components/responses/*` (BadRequest/Unauthorized/Forbidden/NotFound/Conflict/UnprocessableEntity/PreconditionFailed/Gone/TooManyRequests/BadGateway/InternalServerError), all → `#/components/schemas/Problem`
> - `internal/platform/problem/` — `problem.New` / `problem.Write` (the only error writer)
> - `scripts/api-lint/spec_rules.go` — `ENVELOPE-DRIFT` rule, blocking
> - `frontend/apps/web/src/lib/api/problem.ts` + `errors.ts` — the single FE Problem parser → `ApiError`

## Context

The API historically carried two error shapes: a legacy `ApiErrorEnvelope` (`{error:{code,message}}`) and RFC 9457 `Problem` (`application/problem+json`). They coexisted in the spec and across Go emitters; no operation documented a `500`; the FE had a legacy fallback branch. A single FE error handler could not work everywhere.

## Decision (AD-2)

**RFC 9457 `Problem` (`application/problem+json`) is the only error shape.** `ApiErrorEnvelope` is retired from the spec, generated code, and FE. Every error response (4xx/5xx) with a body resolves to the shared `#/components/responses/*` set, each referencing `Problem`. Every server error is written through `internal/platform/problem`. The FE parses exactly one shape (`parseProblem` → `ApiError`), with `code` carrying the RFC 9457 problem code. `ENVELOPE-DRIFT` (blocking) fails CI on any error response that does not resolve to `Problem`; there are **zero** `x-error-envelope-exempt` markers.

## Consequences

- One FE error parser handles the whole API; `resolveErrorMessage(code)` works uniformly.
- The bespoke template-publish 422 (`{valid, parse_errors, …}`) and the idempotency-middleware `{code,message}` body — the last two non-Problem emitters — were reconciled to Problem in Phase F (F3 + the AD-2 envelope-leak fix); the templates 422's per-token detail was found to be spec+FE fiction the backend never emitted and was removed, not migrated.
- 64 previously-error-less operations (Phase E2) plus the partial-coverage ops (Phase F · F2) now document their actual error modes via the shared refs; non-standard codes (410/429/502/412) got shared responses too.
- New endpoints inherit the shared error refs; a new bespoke error body is a blocking lint failure.

## References
- `wiki/backlog/api-contract-hardening.md` Phases D / E2 / F (F2, F3, envelope-leak)
- ADR [`0024-openapi-single-base-path.md`](0024-openapi-single-base-path.md), ADR [`0026-unified-authz-enforcement.md`](0026-unified-authz-enforcement.md)
- `wiki/architecture/api-design-system.md` §3 (error envelope)
