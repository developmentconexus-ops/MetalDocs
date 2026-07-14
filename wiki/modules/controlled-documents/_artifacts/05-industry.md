# Phase 5 — Industry comparison: controlled-documents

Last verified: 2026-06-11

Patterns drawn from `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. One section per applicable row; rejected rows recorded explicitly.

## IP-001 — RFC 9457 Problem Details (errors)

- Source: https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07)
- Quote: "A problem details object can be extended with additional members."
- **Module state:** APPLIED — closed Plan 7 (T-003). `internal/platform/httpresponse/response.go` `WriteError` now delegates to `problem.Write(w, problem.New(status, code, message))`. All controlled-documents routes using `httpresponse.WriteError` emit `Content-Type: application/problem+json`. `ErrTemplateProfileMismatch` additionally calls `problem.Write` directly at `delivery/http/routes.go:561` → 422 `template_invalid` (T-007 also closed Plan 7).
- **Status:** No gap. T-003 and T-007 closed.

## IP-002 — Stripe-style idempotency (idempotency)

- Source: https://docs.stripe.com/api/idempotent_requests (accessed 2026-05-10)
- Quote: "Keys are eligible to be removed from the system after they're at least 24 hours old."
- **Module state:** APPLIED on two routes (`POST /controlled-documents`, `POST .../revisions`). Middleware at `internal/platform/idempotency/middleware.go:22`; store at `postgres_store.go:19`; route wiring `delivery/http/handler.go:80-82`. Body-hash check → `RequestHash` called at `middleware.go:104`; conflict 422 emitted at `middleware.go:116-118`; replay (cache-hit, `"completed"` status) returns recorded 2xx at `postgres_store.go:139-151`. Retention policy: (unclear: no TTL/sweeper visible in `internal/platform/idempotency/`); record as Minor "no 24h sweeper" if absent.
- **Note:** PUT lifecycle routes (`obsolete`, `supersede`) are NOT guarded. By Stripe model PUTs are naturally idempotent, but the active-only guard (`ErrCDNotActive`) means a replay after success returns 409 instead of replaying the original 204 — surprising for callers retrying past a timeout. Capture as Minor.

## IP-004 — Defense-in-depth authz

- Source: NIST SP 800-95 §4.3 (2007)
- Quote: "Multiple layers of access control reduce single-point bypass risk."
- **Module state:** APPLIED for create and lifecycle mutation paths — T-001 and T-004 closed Plan 5. Tier-1: IAM middleware + permission resolver. Tier-2: `authz.Require` in-tx in `application/service.go` for all create, obsolete, and supersede paths; also present at repo layer for `Create` (`repository.go:341`) and `CreateTx` (`repository.go:362`). Tier-3: Postgres `trg_require_cap_asserted` on `controlled_documents` and `cd_sequence_counters` (migration 0231).
- **Remaining gap (T-006 open):** `GetActiveDocument` (`routes.go:266`) has no `authz.Require` / capability check on the read path — only a visibility EXISTS check. See T-006. T-005 (no GUC + RLS tenant backstop) also open.
- Cross-ref `wiki/decisions/0007-two-tier-authz.md`, `wiki/concepts/authz-tiers.md`.

## IP-005 — OpenAPI source-of-truth + codegen

- Source: https://learn.openapis.org/best-practices.html (OAI 3.0.3, 2020)
- Quote: "The OpenAPI Specification … is the standard for HTTP APIs."
- **Module state:** APPLIED. Module generates server stubs via `oapi-codegen` (`internal/modules/controlleddocuments/api/api.gen.go`; cfg at `api/cfg.yaml`; spec partial at `api/openapi/v1/partials/controlled-documents.yaml`). All 8 routes registered through the generated `ServerInterfaceWrapper`. Drift caveat: spec lives under `v1/` even though URL is `/api/v1/` — Minor naming drift. 422 `template_invalid` spec/handler drift — closed T-007 Plan 7.

## IP-006 — Forward-only migrations

- Source: https://martinfowler.com/articles/evodb.html (Fowler, 2016)
- Quote: "Each change to the database is described by a migration script."
- **Module state:** APPLIED. 13 numbered migrations documented in 04-persistence §6 (0124 through 0231). Migration 0182 drops the legacy `profile_sequence_counters` and creates `cd_sequence_counters`; forward-only throughout.

## IP-008 — Row-level tenant_id

- Source: https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy (accessed 2026-05-10)
- Quote: "Add tenant_id to every multi-tenant table and index it first."
- **Module state:** APPLIED at column level. `controlled_documents.tenant_id NOT NULL`, `cd_sequence_counters` PK includes `tenant_id`. Indexes `ix_controlled_documents_tenant_area`, `ix_controlled_documents_tenant_profile` lead with `tenant_id` (04-persistence §4).
- **Partial gap (T-005 open):** tenant_id is enforced as query arg in WHERE clauses. `SET LOCAL metaldocs.tenant_id` GUC is now set via `setAuthzGUC` before in-tx mutations; but no RLS policy exists. A repo method that omits the tenant predicate has no RLS backstop. See T-005.

## IP-003 — Cursor pagination

- **not-applicable: controlled-documents — List endpoint uses simple LIMIT/OFFSET via `CDFilter` (`domain/port.go:19`); cursor-pagination not required by current consumers.** Logging in tech-debt would be over-scoping.

## IP-007 — Structured logs with correlation id

- **not-applicable: observability not yet wired in MetalDocs (per index notes).** Defer until platform-wide observability lands.

## Summary

| Pattern | Status in controlled-documents | Severity if gap |
|---|---|---|
| IP-001 RFC 9457 | APPLIED — T-003 + T-007 closed Plan 7 | n/a |
| IP-002 Idempotency | Applied (POST routes only) | Minor — PUT lifecycle gap (T-008 note) |
| IP-004 Defense-in-depth authz | Applied for mutations (T-001 + T-004 closed Plan 5); T-006 read-path gap open | Major (T-006) |
| IP-005 OpenAPI codegen | Applied | Minor — v1/v2 path naming; 422 drift closed T-007 |
| IP-006 Forward-only migrations | Applied | n/a |
| IP-008 Row-level tenant_id | Applied at column level; no GUC/RLS (T-005 open) | Major (T-005) |
| IP-003 Cursor pagination | n/a | — |
| IP-007 Observability | n/a (deferred) | — |
