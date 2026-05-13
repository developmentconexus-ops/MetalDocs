# Phase 5 — Industry comparison: registry

Patterns drawn from `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. One section per applicable row; rejected rows recorded explicitly.

## IP-001 — RFC 9457 Problem Details (errors)

- Source: https://www.rfc-editor.org/rfc/rfc9457.html (RFC 9457, 2023-07)
- Quote: "A problem details object can be extended with additional members."
- **Registry state:** NOT applied. Errors are emitted by `internal/platform/httpresponse/response.go:14-15` as the legacy envelope `{"code":"...","message":"..."}`. Both create (`routes.go:417..441`) and lifecycle (`routes.go:412-416`) paths use this envelope. `Content-Type: application/problem+json` is never set.
- **Gap:** T-row Major — uniform spec compliance gap aligned with peer modules (templates T-005, documents T-001, audit T-002).

## IP-002 — Stripe-style idempotency (idempotency)

- Source: https://docs.stripe.com/api/idempotent_requests (accessed 2026-05-10)
- Quote: "Keys are eligible to be removed from the system after they're at least 24 hours old."
- **Registry state:** APPLIED on two routes (`POST /controlled-documents`, `POST .../revisions`). Middleware at `internal/platform/idempotency/middleware.go:22`; store at `postgres_store.go:19`; route wiring `delivery/http/handler.go:80-82`. Body-hash check → conflict 422 (`middleware.go:50`); replay returns recorded 2xx (`postgres_store.go:36`). Retention policy: (unclear: no TTL/sweeper visible in `internal/platform/idempotency/`); record as Minor "no 24h sweeper" if absent.
- **Note:** PUT lifecycle routes (`obsolete`, `supersede`) are NOT guarded. By Stripe model PUTs are naturally idempotent, but the active-only guard (`ErrCDNotActive`) means a replay after success returns 409 instead of replaying the original 204 — surprising for callers retrying past a timeout. Capture as Minor.

## IP-004 — Defense-in-depth authz

- Source: NIST SP 800-95 §4.3 (2007)
- Quote: "Multiple layers of access control reduce single-point bypass risk."
- **Registry state:** PARTIAL. Tier-1 capability check via IAM middleware + permission resolver (`apps/api/cmd/metaldocs-api/main.go:173-174`, `:386`, `permissions.go:186-187`; capability `CapRegistryCreate`). Tier-2 (`authz.Require` in-tx) and tier-3 (Postgres `enforce_capability_asserted` tripwire) are **absent** for registry-owned tables. Persistence audit (04-persistence §5) records 5 tripwire violations on `Create`, `CreateTx`, `UpdateStatus`, `EnsureCounter`, `NextAndIncrement`.
- **Gap:** Critical for `controlled_documents` mutations (UpdateStatus/CreateTx) — tier-1 bypass via direct DB or alternate module import would not be caught. Cross-ref `wiki/decisions/0007-two-tier-authz.md`, `wiki/concepts/authz-tiers.md`.

## IP-005 — OpenAPI source-of-truth + codegen

- Source: https://learn.openapis.org/best-practices.html (OAI 3.0.3, 2020)
- Quote: "The OpenAPI Specification … is the standard for HTTP APIs."
- **Registry state:** APPLIED. Module generates server stubs via `oapi-codegen` (`internal/modules/registry/api/api.gen.go`; cfg at `api/cfg.yaml`; spec partial at `api/openapi/v1/partials/registry.yaml`). All 8 routes registered through the generated `ServerInterfaceWrapper`. Drift caveat: spec lives under `v1/` even though URL is `/api/v1/` — Minor naming drift. Spec/handler drift on 422 `template_invalid` (declared in spec at `registry.yaml:73`; no handler branch) — captured in T-row.

## IP-006 — Forward-only migrations

- Source: https://martinfowler.com/articles/evodb.html (Fowler, 2016)
- Quote: "Each change to the database is described by a migration script."
- **Registry state:** APPLIED. 7 numbered migrations (0124, 0126, 0127, 0128, 0167, 0182, 0183 — see 04-persistence §6). Migration 0182 drops the legacy `profile_sequence_counters` and creates `cd_sequence_counters`; this is forward-only but DROP-on-merged-table — log as maint:migration-cleanup row in refactor backlog if any code/tests still reference legacy table name (Phase 6 grep).

## IP-008 — Row-level tenant_id

- Source: https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy (accessed 2026-05-10)
- Quote: "Add tenant_id to every multi-tenant table and index it first."
- **Registry state:** APPLIED at column level. `controlled_documents.tenant_id NOT NULL`, `cd_sequence_counters` PK includes `tenant_id`. Indexes `ix_controlled_documents_tenant_area`, `ix_controlled_documents_tenant_profile` lead with `tenant_id` (04-persistence §4).
- **Gap:** tenant_id is enforced only as query arg in every WHERE clause (`repository.go:26`, `:36`, `:46`, `:184`); not via `SET LOCAL metaldocs.tenant_id` GUC + RLS. A repo method that omits the tenant predicate has no DB-level backstop. Pair with IP-004 gap → Critical tenant-leak risk on mutation paths.

## IP-003 — Cursor pagination

- **not-applicable: registry — List endpoint uses simple LIMIT/OFFSET via `CDFilter` (`domain/port.go:19`); cursor-pagination not required by current consumers.** Logging in tech-debt would be over-scoping.

## IP-007 — Structured logs with correlation id

- **not-applicable: observability not yet wired in MetalDocs (per index notes).** Defer until platform-wide observability lands.

## Summary

| Pattern | Status in registry | Severity if gap |
|---|---|---|
| IP-001 RFC 9457 | NOT applied | Major (T-row) |
| IP-002 Idempotency | Applied (POST routes only) | Minor — PUT lifecycle gap |
| IP-004 Defense-in-depth authz | Partial (tier-1 only) | Critical (tier-3 tripwire missing on owned tables) |
| IP-005 OpenAPI codegen | Applied | Minor — v1/v2 path naming + 422 spec/handler drift |
| IP-006 Forward-only migrations | Applied | n/a |
| IP-008 Row-level tenant_id | Applied at column level; no GUC/RLS | Critical-adjacent (paired with IP-004) |
| IP-003 Cursor pagination | n/a | — |
| IP-007 Observability | n/a (deferred) | — |
