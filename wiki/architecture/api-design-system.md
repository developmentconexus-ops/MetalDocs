# API Design System

> **Status:** accepted 2026-05-10
> **Last verified:** 2026-05-10
> **Scope:** API design conventions for v1 — error envelope, pagination, idempotency, two-tier authz, list filtering, naming.
> **Out of scope:** Adoption — Plan 2 migrates handlers, paths, and module names. Frontend wiring is in Plan 1 only for the shared parser; per-page adoption is Plan 2.
> **Key files:**
> - `api/openapi/v1/openapi.yaml:5382` — Problem schema definition
> - `internal/platform/problem/problem.go` — Go helper (RFC 9457 envelope)
> - `internal/platform/problem/codes.go` — error code taxonomy
> - `internal/platform/pagination/cursor.go` — cursor primitive (sort + filter_hash validation)
> - `internal/platform/idempotency/middleware.go` — idempotency middleware (Stripe-model raw-byte SHA-256)
> - `internal/platform/idempotency/postgres_store.go` — idempotency store (4-col PK defense-in-depth)
> - `internal/modules/iam/authz/authz.go:44` — tier-2 `Require` (sets `metaldocs.asserted_caps` GUC)
> - `migrations/0142b_role_capabilities_v2_enforce.sql:138-172` — Postgres tripwire trigger
> - `frontend/apps/web/src/lib/api/problem.ts` — frontend Problem parser + ApiError
> - `scripts/api-lint/spec_rules.go` — envelope/pagination/authz-drift rules
> - `scripts/api-lint/code_rules.go` — authz-call-present + tripwire-pairing rules

---

## 1. Why

MetalDocs exposes a growing HTTP surface across 7 modules. Without a shared contract, each module invents its own error shape, pagination style, and auth call sequence — making the frontend fragile and the codebase look hand-rolled rather than production-grade. This document defines the conventions that govern every endpoint so that (a) the frontend never reinvents a pattern, (b) unmigrated modules (`approval`, `taxonomy`, `iam`, `platform/auth`) can adopt them via a single plan, and (c) drift between spec, generated code, and runtime is detectable in CI rather than at runtime. The spec (`api/openapi/v1/openapi.yaml`) is the single source of truth; handlers implement; CI verifies.

---

## 2. The 7 conventions

| # | Convention | Decision | Code / CI surface |
|---|---|---|---|
| 1 | Error envelope | RFC 9457 Problem Details + `code` + `errors[]` | `Problem` schema in spec; `problem.go` helper; `ENVELOPE-DRIFT` CI rule |
| 2 | Pagination | Cursor-only, opaque token, uniform shape | `cursor.go`; `PAGINATION-DRIFT` CI rule |
| 3 | Idempotency | Stripe-model `Idempotency-Key`, shared store, 24h TTL | `middleware.go` + `postgres_store.go`; migration `0147_idempotency_keys.sql` |
| 4 | Body validation | oapi-codegen per-op generated validation (not reflection middleware) | `cfg.yaml` `validate-against-spec: true`; strict-decode `DisallowUnknownFields` |
| 5 | Two-tier authz | `security` + `x-authz-area` in spec; enforced by Postgres tripwire | `authz.Require`; tripwire trigger; `AUTHZ-DRIFT` + `authz-call-present` + `tripwire-pairing` CI rules |
| 6 | List filtering | Stripe-flat params + per-resource allowlist + typed parser | oapi-codegen `parameters` per op; `UNKNOWN_FILTER` code |
| 7 | Mini-conventions | UUIDv4, ISO 8601 UTC, snake_case, plural kebab paths, positive booleans, null-vs-absent, UUID-only path IDs, Bearer JWT | Codified in spec; enforced by codegen type alignment |

---

## 3. Error envelope (RFC 9457)

All 4xx/5xx responses MUST use `Content-Type: application/problem+json` with this shape:

```json
{
  "type": "https://errors.metaldocs.io/<category>",
  "title": "<short human-readable title>",
  "status": 400,
  "detail": "<longer explanation>",
  "instance": "/api/v1/templates",
  "code": "VALIDATION_ERROR",
  "errors": [
    { "field": "key", "code": "REQUIRED", "message": "key is required" },
    { "field": "name", "code": "OUT_OF_RANGE", "message": "name must be ≤ 200 chars" }
  ]
}
```

- `type`, `title`, `status`, `detail`, `instance` — RFC 9457 standard fields.
- `code` — required extension; machine-readable; drawn from the catalog in `internal/platform/problem/codes.go`.
- `errors[]` — optional extension; field-level errors for `VALIDATION_ERROR`. Each entry carries `field` (JSON pointer or dot path), `code`, and `message`.

**Go helper:** `internal/platform/problem/problem.go` — `problem.New(status, code, title)`, chainable `.WithDetail()`, `.WithFieldError()`, `.Write(w)`.

**Frontend helper:** `frontend/apps/web/src/lib/api/problem.ts` — `parseProblem(res)` + `ApiError` class (holds `.code`, `.errors[]`, `.status`). `resolveErrorMessage` dispatches on `code`.

**Error codes** (full catalog in `internal/platform/problem/codes.go`):

| Code | HTTP | Meaning |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Body/params failed validation; `errors[]` populated |
| `UNKNOWN_FIELD` | 400 | Strict-decode rejected unknown JSON field |
| `UNKNOWN_FILTER` | 400 | Filter key not in operation allowlist |
| `INVALID_SORT_FIELD` | 400 | Sort field not allowed for endpoint |
| `INVALID_CURSOR` | 400 | Cursor decode failed or filter/sort mismatch |
| `INCLUDE_NOT_SUPPORTED` | 400 | `?include=<x>` not supported by endpoint |
| `UNAUTHENTICATED` | 401 | No or invalid JWT |
| `FORBIDDEN_CAPABILITY` | 403 | User lacks capability (tier-1) |
| `FORBIDDEN_AREA` | 403 | User lacks area-scoped permission (tier-2) |
| `NOT_FOUND` | 404 | Resource does not exist or caller cannot see it |
| `ALREADY_EXISTS` | 409 | Unique-constraint violation |
| `STATE_TRANSITION_INVALID` | 409 | Workflow state transition not allowed |
| `CONCURRENT_MODIFICATION` | 409 | Optimistic-lock failure |
| `IDEMPOTENCY_KEY_REUSED` | 422 | Same key, different body |
| `INTERNAL_ERROR` | 500 | Unhandled server error |

Field-level codes (`errors[].code`): `REQUIRED`, `INVALID_FORMAT`, `OUT_OF_RANGE`, `INVALID_ENUM`.

---

## 4. Cursor pagination

All list endpoints MUST use cursor pagination. No offset-based list endpoints.

**Request:**
```
GET /api/v1/documents?cursor=<opaque>&limit=20
```

- `cursor` — opaque `base64url(json(<payload>))`. Absent on first page.
- `limit` — integer, default 20, max 100. Server clamps via `pagination.ClampLimit`.

**Response:**
```json
{
  "data": [...],
  "page": {
    "next_cursor": "<opaque|null>",
    "has_more": true,
    "limit": 20
  }
}
```

`next_cursor: null` / `has_more: false` → no more pages. Opt-in total: `?include=total` → `page.total_count` if endpoint supports it; otherwise 400 `INCLUDE_NOT_SUPPORTED`.

**Cursor payload** (opaque to clients, server-internal):
```json
{
  "v": 1,
  "sort": [{"field": "updated_at", "dir": "desc"}, {"field": "id", "dir": "asc"}],
  "anchor": {"updated_at": "2026-05-10T14:30:00Z", "id": "550e8400-..."},
  "filter_hash": "<sha256 of sorted filter params>"
}
```

`id` is always appended as a tie-breaker (`asc`) when sort field is non-unique. Filter/sort change between pages → 400 `INVALID_CURSOR`; client restarts pagination.

**Go helper:** `internal/platform/pagination/cursor.go` — `Encode`, `Decode`, `ValidateMatch`, `HashFilters`, `AppendIDTieBreaker`, `ClampLimit`.

---

## 5. Idempotency

All unsafe writes (POST, PUT, PATCH, DELETE) MUST support optional `Idempotency-Key` header.

| Client sends | Server behavior |
|---|---|
| No header | Non-idempotent; caller's choice |
| First time with key | Execute, store `(key, body_hash, status, response_body, expires_at)`, return response |
| Same key + same body hash within TTL | Replay stored response; add `Idempotent-Replay: true` header |
| Same key + different body hash | 422 `IDEMPOTENCY_KEY_REUSED` |
| Same key after TTL | Treat as new request |

**Body hash:** SHA-256 over raw request bytes (Stripe model — `internal/platform/idempotency/middleware.go`). MetalDocs is internal-only with a single TS client; deterministic serialization is guaranteed; canonical-JSON normalization is not needed.

**TTL:** 24h flat. Future op-specific TTLs require an ADR.

**Storage PK** (`internal/platform/idempotency/postgres_store.go`, table `metaldocs.idempotency_keys`):
```sql
PRIMARY KEY (tenant_id, actor_user_id, route_template, key)
```
4-col PK is intentional defense-in-depth: `actor_user_id` prevents User A replaying User B's response; `route_template` prevents the same key colliding across endpoints.

**Plan 2 reconciliation** (not Plan 1): switch middleware error envelope from legacy `{code, message}` to RFC 9457 Problem; rename conflict code to `IDEMPOTENCY_KEY_REUSED`; add `Idempotent-Replay: true` header; delete duplicate `PostgresSignoffIdempStore`.

---

## 6. Two-tier authz

See also: `wiki/decisions/0007-two-tier-authz.md` and its 2026-05-10 codegen-rejection amendment.

**Tier 1 — capability (HTTP-level):**
```yaml
security:
  - bearerAuth: [template.create]
```
Existing `CapabilityService` middleware reads `security`, calls `CapabilityService.CanDo` before the handler runs.

**Tier 2 — area (in-transaction):**
```yaml
x-authz-area:
  source: body        # body | path
  field: area_code    # top-level field; dotted for nested
  multi: false        # true → field is array → authz.RequireAll
```
Handler calls `authz.Require(ctx, tx, capability, areaCode)` (`internal/modules/iam/authz/authz.go:44`) inside its transaction. This call does three things: permission check, `system_admin` bypass, and **sets the tx-scoped GUC `metaldocs.asserted_caps`**.

**The Postgres tripwire trigger is the real enforcer.** `migrations/0142b_role_capabilities_v2_enforce.sql:138-172` reads `metaldocs.asserted_caps` on every INSERT into `approval_instances` and `approval_signoffs`. If the required cap is absent, the database raises `ErrCapabilityNotAsserted` and rejects the write. The Go `authz.Require` call is how handlers prove the cap was checked; the transaction carries the proof.

**No code generation.** `x-authz-area` is a lint contract, not a codegen input. A `cmd/authzgen` spike was evaluated and rejected (see amendment in `wiki/decisions/0007-two-tier-authz.md`). The pre-tx wrapper cannot supply `tx`; the tripwire already provides the static guarantee.

**Escape hatches:**
- `x-authz-skip-area: "<reason>"` — explicit opt-out for routes with no area dimension (e.g. `/v1/auth/login`).
- `x-authz-custom: true` — silences lint for ops that compute area at runtime; handler must still contain at least one `authz.Require` call.

---

## 7. List filtering

Stripe-style flat query params with per-resource allowlist.

```
?status=draft
?status=draft&status=published
?created_after=2026-01-01T00:00:00Z
?created_before=2026-12-31T23:59:59Z
?q=annual%20audit
?area=quality&area=engineering
```

- Equality: `<field>=value`; multi-value: repeat key.
- Range: `<field>_after` / `<field>_before` (ISO 8601 for timestamps).
- Free-text: `q=<term>`.
- Unknown key → 400 `UNKNOWN_FILTER`.
- Reserved keys (`cursor`, `limit`, `sort`, `include`) MUST NOT be used as filter names.

Each list operation in the spec declares allowed filter keys via `parameters`. oapi-codegen-generated validation enforces typed parsing per param schema.

---

## 8. CI lint

Five rules in `scripts/api-lint/`, running in the `api-design-system-lint` job of `.github/workflows/api-contract.yml`. Currently `continue-on-error: true` — non-blocking until Plan 2 finishes adoption. Final Plan 2 step flips to `false`.

| Rule | File | What it checks |
|---|---|---|
| `ENVELOPE-DRIFT` | `spec_rules.go` | Every response ≥ 400 references `#/components/schemas/Problem` |
| `PAGINATION-DRIFT` | `spec_rules.go` | Every `list*` op declares `cursor`/`limit` params; response includes `page.next_cursor` + `page.has_more` |
| `AUTHZ-DRIFT` | `spec_rules.go` | Every op declares `security`; state-transition ops declare `x-authz-area` or escape hatch |
| `authz-call-present` | `code_rules.go` | Each op with `x-authz-area` has a matching `authz.Require` call in the handler body |
| `tripwire-pairing` | `code_rules.go` | Every mutating SQL in `*repository*.go` lives in a function that also calls `authz.Require` |

The three existing jobs (`backend-codegen-drift`, `frontend-codegen-drift`, `openapi-lint`) remain blocking and unchanged.

---

## 9. What's next

**Plan 2 — Adoption (~1 week):**
1. Migrate `registry`, `templates_v2`, `documents` to RFC 9457 envelope + cursor + idempotency.
2. Path rewrite `/api/v2/...` → `/api/v1/...` (atomic across server + frontend).
3. Module rename `templates_v2/` → `templates/`.
4. Frontend list pages → `useInfiniteQuery` + cursor.
5. Frontend wizard forms → `errors[]` field consumption.
6. Delete `PostgresSignoffIdempStore`.
7. Flip `continue-on-error: false` in CI.

Sub-projects #5 (spec coverage for `approval`/`taxonomy`/`iam`/`platform`) run as separate plans, each adopting these conventions per module.

---

## See also

- `wiki/architecture/api-contract.md` — operational guide: oapi-codegen wiring, module migration status, how to add a module
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz ADR + codegen-rejection amendment
- `wiki/decisions/0012-contract-first-api.md` — ADR establishing spec-as-source-of-truth
- `wiki/concepts/error-ux.md` — current frontend error parsing (legacy shape; Plan 2 rewrites)
