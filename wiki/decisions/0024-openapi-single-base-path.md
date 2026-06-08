# ADR 0024 — One OpenAPI base path (`servers.url: /api/v1`, relative path keys)

> **Status:** Accepted 2026-06-08 (formalises anchor decision AD-1 of the API Contract Hardening Program, shipped in Phase A 2026-06-05). Retroactive ADR.
> **Last verified:** 2026-06-08
> **Scope:** How the `/api/v1` prefix is expressed in the OpenAPI document and the FE client, and the CI gate that keeps it singular.
> **Out of scope:** Per-route capability/authz (ADR 0007/0022/0023); error envelope (ADR 0025).
> **Key files:**
> - `api/openapi/v1/openapi.yaml:7` — `servers: - url: /api/v1`; every path key relative (`/documents`, not `/api/v1/documents`)
> - `scripts/api-lint/spec_rules.go` — `PATH-BASE-PREFIX` rule (`checkBasePrefix`), blocking
> - `frontend/apps/web/src/lib/api/client.ts` — openapi-fetch `baseUrl: /api/v1` + the `apiUrl`/`rewriteRequest` guard
> - the 4 generated-router `BaseURL: "/api/v1"` mounts (taxonomy, controlleddocuments, documents ×2)

## Context

Before Phase A the spec set `servers.url: /api/v1` AND ~half the path keys re-included `/api/v1/`, so spec-respecting consumers built `/api/v1/api/v1/...` for half the API. The FE openapi-fetch client (baseUrl `/api/v1`) doubled the prefix on those paths → live 404s (documents list, blank-template). The spec was self-inconsistent.

## Decision (AD-1)

**There is exactly one declaration of the base path: `servers.url: /api/v1`. Every OpenAPI path key is relative.** A path key that re-includes `/api/` is a build-breaking lint violation (`PATH-BASE-PREFIX`, blocking). The FE openapi-fetch client carries the base once as `baseUrl`; its paths are relative; the `apiUrl` guard passes an already-`/api/`-prefixed string through unchanged so a hand-written `apiFetch('/api/v1/...')` is never doubled either. The four generated routers that mount via `HandlerFromMuxWithBaseURL` pass `BaseURL: "/api/v1"` so served routes stay `/api/v1/*` while the generated code matches the relative spec.

## Consequences

- The `/api/v1/api/v1/...` double-prefix bug class is structurally impossible — `PATH-BASE-PREFIX` fails CI on any re-prefixed key, and the FE guard is idempotent.
- oapi-codegen registers spec keys verbatim, so the `BaseURL` mount is the single place the prefix is re-applied at runtime; regenerating never silently drops it (a fresh `go generate` leaves `api.gen.go` unchanged → `backend-codegen-drift` green).
- New endpoints MUST be added with a relative key; reviewers and the lint enforce it.

## References
- `wiki/backlog/api-contract-hardening.md` Phase A — base-path normalization + the PATH-BASE-PREFIX gate (Evidence 2026-06-05)
- ADR [`0025-error-envelope-rfc9457.md`](0025-error-envelope-rfc9457.md), ADR [`0026-unified-authz-enforcement.md`](0026-unified-authz-enforcement.md) — the other two program anchors
