# API Design System

> **Status:** accepted 2026-05-10
> **Last verified:** 2026-06-08 (Phase E2 spec hygiene: root-level `security: [sessionCookie]` makes every op secure-by-default in-doc with explicit `security: []` on public ops; top-level `tags` + every op tagged/`operationId`'d/`summary`'d; one pagination convention — cursor canonical, bounded/offset lists `x-pagination-exempt` with reason, `limit` max clamped 100 in spec+runtime; 64 zero-error ops now document their actual error modes via the shared `#/components/responses/*` set; two non-standard methods fixed (`DELETE /iam/area-memberships` → path-param `…/{user_id}/{area_code}`, `DELETE /approval/routes/{id}` → `POST …/{id}/deactivate`); `bearerAuth` myth corrected to the real `sessionCookie` scheme; `AUTHZ-DRIFT` + `PAGINATION-DRIFT` both driven to 0 and flipped BLOCKING. Prior: Phase E1 payload-casing big-bang — every declared JSON property is `snake_case` end-to-end, blocking `CASING-DRIFT` (one exemption `MDDM_NATIVE_EXPORT_ROLLOUT_PCT`); Phase D error-envelope unification — RFC 9457 `Problem` only, `ENVELOPE-DRIFT` BLOCKING; Phase C dead-path prune; 2026-05-21)
> **Scope:** API design conventions for v1     error envelope, pagination, idempotency, two-tier authz, list filtering, naming.
> **Out of scope:** Adoption     Plan 2 migrates handlers, paths, and module names. Frontend wiring is in Plan 1 only for the shared parser; per-page adoption is Plan 2.
> **Key files:**
> - `api/openapi/v1/openapi.yaml:4332` — Problem schema definition
> - `internal/platform/problem/problem.go`     Go helper (RFC 9457 envelope)
> - `internal/platform/problem/codes.go`     error code taxonomy
> - `internal/platform/pagination/cursor.go`     cursor primitive (sort + filter_hash validation)
> - `internal/platform/idempotency/middleware.go`     idempotency middleware (Stripe-model raw-byte SHA-256)
> - `internal/platform/idempotency/postgres_store.go`     idempotency store (4-col PK defense-in-depth)
> - `internal/modules/iam/authz/authz.go:44`     tier-2 `Require` (sets `metaldocs.asserted_caps` GUC)
> - `migrations/0142b_role_capabilities_v2_enforce.sql:138-172`     Postgres tripwire trigger
> - `frontend/apps/web/src/lib/api/problem.ts`     frontend Problem parser + ApiError
> - `scripts/api-lint/spec_rules.go`     envelope/pagination/authz-drift rules
> - `scripts/api-lint/code_rules.go`     tripwire-pairing + registry-binding rules (the `authz-call-present` rule was deleted in Phase F, FD-1)

---

## 1. Why

MetalDocs exposes a growing HTTP surface across 7 modules. Without a shared contract, each module invents its own error shape, pagination style, and auth call sequence     making the frontend fragile and the codebase look hand-rolled rather than production-grade. This document defines the conventions that govern every endpoint so that (a) the frontend never reinvents a pattern, (b) unmigrated modules (`approval`, `taxonomy`, `iam`, `platform/auth`) can adopt them via a single plan, and (c) drift between spec, generated code, and runtime is detectable in CI rather than at runtime. The spec (`api/openapi/v1/openapi.yaml`) is the single source of truth; handlers implement; CI verifies.

---

## 2. The 7 conventions

| # | Convention | Decision | Code / CI surface |
|---|---|---|---|
| 1 | Error envelope | RFC 9457 Problem Details + `code` + `errors[]` | `Problem` schema in spec; `problem.go` helper; `ENVELOPE-DRIFT` CI rule |
| 2 | Pagination | Cursor-first target; module-level temporary exceptions allowed while migration is incomplete | `cursor.go`; `PAGINATION-DRIFT` CI rule |
| 3 | Idempotency | Stripe-model `Idempotency-Key`, shared store, 24h TTL | `middleware.go` + `postgres_store.go`; migration `0147_idempotency_keys.sql` |
| 4 | Body validation | oapi-codegen per-op generated validation (not reflection middleware) | `cfg.yaml` `validate-against-spec: true`; strict-decode `DisallowUnknownFields` |
| 5 | Two-tier authz | `security` + honest `x-authz-area`/`x-authz-area-none` markers in spec; enforced by Postgres tripwire | `authz.Require`; tripwire trigger; `AUTHZ-DRIFT` + `tripwire-pairing` CI rules |
| 6 | List filtering | Stripe-flat params + per-resource allowlist + typed parser | oapi-codegen `parameters` per op; `UNKNOWN_FILTER` code |
| 7 | Mini-conventions | UUIDv4, ISO 8601 UTC, snake_case, plural kebab paths, positive booleans, null-vs-absent, UUID-only path IDs, session-cookie auth | Codified in spec; payload casing enforced by the blocking `CASING-DRIFT` rule (Phase E1); codegen type alignment |

---

## 3. Error envelope (RFC 9457)

RFC 9457 `Problem` is the **only** error shape (AD-2, api-contract-hardening Phase D). The legacy `ApiErrorEnvelope` (`{error:{code,message,details,trace_id}}`) is retired — deleted from the spec and the frontend. Error responses reference one shared set of `#/components/responses/*` definitions (`BadRequest`/`Unauthorized`/`Forbidden`/`NotFound`/`Conflict`/`UnprocessableEntity`/`InternalServerError`), each `application/problem+json` → `Problem`; operations that document failures also document a `500`. A bespoke (non-Problem) error body may opt out only via a reviewed `x-error-envelope-exempt: "<reason>"` marker (mirrors `x-pagination-exempt-reason`).

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
    { "field": "name", "code": "OUT_OF_RANGE", "message": "name must be     200 chars" }
  ]
}
```

- `type`, `title`, `status`, `detail`, `instance`     RFC 9457 standard fields.
- `code`     required extension; machine-readable; drawn from the catalog in `internal/platform/problem/codes.go`.
- `errors[]`     optional extension; field-level errors for `VALIDATION_ERROR`. Each entry carries `field` (JSON pointer or dot path), `code`, and `message`.

**Go helper:** `internal/platform/problem/problem.go`     `problem.New(status, code, title)`, chainable `.WithDetail()`, `.WithFieldError()`, `.Write(w)`.

**Frontend helper:** `frontend/apps/web/src/lib/api/problem.ts`     `parseProblem(res)` + `ApiError` class (holds `.code`, `.errors[]`, `.status`). `resolveErrorMessage` dispatches on `code`.

**Error-code coverage (FE):** every backend code must have a PT-BR mapping in `frontend/apps/web/src/lib/api/errorMessages.ts`. Enforced by `src/lib/api/__tests__/errorMessages.coverage.test.ts` against the snapshot `frontend/apps/web/src/lib/api/error-codes.generated.json`. After adding/renaming/removing a backend code, regenerate the snapshot: `go run ./scripts/dump-error-codes.go`. Unmapped codes fall back to an actionable PT-BR template (`Não foi possível concluir a ação. Código: <X>`) and emit a `console.warn` breadcrumb tagged `[api] unmapped error code` for observability.

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
| `IDEMPOTENCY_KEY_CONFLICT` | 422 | Same key, different body |
| `INTERNAL_ERROR` | 500 | Unhandled server error |

Field-level codes (`errors[].code`): `REQUIRED`, `INVALID_FORMAT`, `OUT_OF_RANGE`, `INVALID_ENUM`.
Exception note: `template_invalid` is intentionally lowercase for legacy contract compatibility on controlled-documents 422 responses.

---

## 4. Cursor pagination

Target state: list endpoints use cursor pagination.  
Current verified exception (2026-05-21): controlled-documents list still uses LIMIT/OFFSET (`wiki/modules/controlled-documents.md`   8.4). Offset-based behavior remains allowed for modules not yet migrated.

**Request:**
```
GET /api/v1/documents?cursor=<opaque>&limit=20
```

- `cursor`     opaque `base64url(json(<payload>))`. Absent on first page.
- `limit`     integer, default 20, max 100. Server clamps via `pagination.ClampLimit`.

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

`next_cursor: null` / `has_more: false`     no more pages. Opt-in total: `?include=total`     `page.total_count` if endpoint supports it; otherwise 400 `INCLUDE_NOT_SUPPORTED`.

**Cursor payload** (opaque to clients, server-internal):
```json
{
  "v": 1,
  "sort": [{"field": "updated_at", "dir": "desc"}, {"field": "id", "dir": "asc"}],
  "anchor": {"updated_at": "2026-05-10T14:30:00Z", "id": "550e8400-..."},
  "filter_hash": "<sha256 of sorted filter params>"
}
```

`id` is always appended as a tie-breaker (`asc`) when sort field is non-unique. Filter/sort change between pages     400 `INVALID_CURSOR`; client restarts pagination.

**Go helper:** `internal/platform/pagination/cursor.go`     `Encode`, `Decode`, `ValidateMatch`, `HashFilters`, `AppendIDTieBreaker`, `ClampLimit`.

---

## 5. Idempotency

Policy baseline: unsafe writes SHOULD support idempotency where replay safety is required by product/risk profile. Requirement can be route-specific (`required`, `optional`, or `not enabled yet`) during migration.

| Client sends | Server behavior |
|---|---|
| No header | Non-idempotent; caller's choice |
| First time with key | Execute, store `(key, body_hash, status, response_body, expires_at)`, return response |
| Same key + same body hash within TTL | Replay stored response; add `Idempotent-Replay: true` header |
| Same key + different body hash | 422 `IDEMPOTENCY_KEY_CONFLICT` |
| Same key after TTL | Treat as new request |

Current verified module truth (2026-05-21): controlled-documents requires `Idempotency-Key` on create/revision POST routes and does not currently enforce idempotency middleware on PUT lifecycle routes.

**Body hash:** SHA-256 over raw request bytes (Stripe model     `internal/platform/idempotency/middleware.go`). MetalDocs is internal-only with a single TS client; deterministic serialization is guaranteed; canonical-JSON normalization is not needed.

**TTL:** 24h flat. Future op-specific TTLs require an ADR.

**Storage PK** (`internal/platform/idempotency/postgres_store.go`, table `metaldocs.idempotency_keys`):
```sql
PRIMARY KEY (tenant_id, actor_user_id, route_template, key)
```
4-col PK is intentional defense-in-depth: `actor_user_id` prevents User A replaying User B's response; `route_template` prevents the same key colliding across endpoints.

**Plan 2 reconciliation** (not Plan 1): switch middleware error envelope from legacy `{code, message}` to RFC 9457 Problem; use conflict code `IDEMPOTENCY_KEY_CONFLICT`; add `Idempotent-Replay: true` header; delete duplicate `PostgresSignoffIdempStore`.

---

## 6. Two-tier authz

See also: `wiki/decisions/0007-two-tier-authz.md` and its 2026-05-10 codegen-rejection amendment.

**Tier 1     capability (HTTP-level):**
```yaml
security:
  - sessionCookie: []
```
Authentication is the `sessionCookie` apiKey scheme (a `metaldocs_session` cookie issued by `POST /api/v1/auth/login`); there is **no** bearer/JWT scheme — MetalDocs is session-cookie only. A root-level `security: [- sessionCookie: []]` makes every operation authenticated by default (api-contract-hardening Phase E2, F-NO-GLOBAL-SEC); truly public ops (login, health/readiness probes, feature flags, signed-URL downloads) opt out with an explicit `security: []`. Capabilities are **not** carried as OpenAPI security scopes: the route→capability map lives in `apps/api/cmd/metaldocs-api/permissions.go` and the `CapabilityService` middleware enforces it before the handler runs. The `AUTHZ-DRIFT` lint rule (now BLOCKING, Phase E2) treats the inherited global `security` as satisfying each op.

**Tier 2     area (in-transaction).** The marker honestly declares **where the enforced area comes from** (api-contract-hardening Phase F, FD-1):

```yaml
# Area loaded from the DB row inside the tx (un-spoofable) — the common case:
x-authz-area:
  source: tx
  derived_from: documents.process_area_code_snapshot   # the DB column the area is read from
  note: "..."                                          # optional: where/how the tier-2 check runs

# Area is the request-supplied action target (e.g. "grant membership in area X"):
x-authz-area:
  source: body          # body | path
  field: area_code      # the request field / path param carrying the area
  # multi: false        # true → field is an array → authz.RequireAll
```
Handler (or tx-layer service) calls `authz.Require(ctx, tx, capability, areaCode)` (`internal/modules/iam/authz/authz.go:44`) inside its transaction. This call does three things: permission check, `system_admin` bypass, and **sets the tx-scoped GUC `metaldocs.asserted_caps`**.

**The Postgres tripwire trigger is the real enforcer.** `migrations/0142b_role_capabilities_v2_enforce.sql:138-172` reads `metaldocs.asserted_caps` on every INSERT into `approval_instances` and `approval_signoffs`. If the required cap is absent, the database raises `ErrCapabilityNotAsserted` and rejects the write. The Go `authz.Require` call is how handlers prove the cap was checked; the transaction carries the proof.

**No code generation.** `x-authz-area` is a lint contract, not a codegen input. A `cmd/authzgen` spike was evaluated and rejected (see amendment in `wiki/decisions/0007-two-tier-authz.md`). The pre-tx wrapper cannot supply `tx`; the tripwire already provides the static guarantee. The old `authz-call-present` lint (which expected a handler-body `authz.Require(req.Body.AreaCode)`) was **deleted** in Phase F: MetalDocs derives the area in the tx-layer (`source: tx`), so that handler-body shape never existed and the rule was dormant by design. See `wiki/decisions/0023-authz-area-markers.md`.

**Markers for non-area ops:**
- `x-authz-area-none: "<reason>"`     genuinely area-less op (e.g. tenant-global templates; `POST /auth/login`). Replaces the retired negative `x-authz-skip-area`.
- `x-authz-custom: true`     ops that compute area at runtime by a bespoke path; handler must still contain at least one `authz.Require` call.

`AUTHZ-DRIFT` validates the marker shape: `source: tx` requires a non-empty `derived_from`; `source: body|path` requires a non-empty `field`; every state-transition POST must carry exactly one of `x-authz-area` / `x-authz-area-none` / `x-authz-custom`.

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
- Unknown key     400 `UNKNOWN_FILTER`.
- Reserved keys (`cursor`, `limit`, `sort`, `include`) MUST NOT be used as filter names.

Each list operation in the spec declares allowed filter keys via `parameters`. oapi-codegen-generated validation enforces typed parsing per param schema.

---

## 8. CI lint

Five rules in `scripts/api-lint/`, running in the `api-design-system-lint` job of `.github/workflows/api-contract.yml`. As of the **2026-05-12 verification snapshot**, this job was configured with `continue-on-error: true` (non-blocking during migration).

**Current freeze note (2026-05-21):** treat live workflow and contract checks as source of truth at verification time. Use `.github/workflows/api-contract.yml` and the blocking jobs listed in this section (`backend-codegen-drift`, `frontend-codegen-drift`, `openapi-lint`) to confirm current gate posture.

| Rule | File | What it checks |
|---|---|---|
| `ENVELOPE-DRIFT` | `spec_rules.go` | **BLOCKING (Phase D).** Every response     400 with a body resolves to `Problem` — inline or via a `#/components/responses/*` `$ref`. Exempt: description-only responses, `/health/*` probes, and reviewed `x-error-envelope-exempt` bodies |
| `CASING-DRIFT` | `spec_rules.go` | **BLOCKING (Phase E1).** Every declared schema `properties:` key (components/schemas + inline bodies) matches `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`. Exempt: RFC 9457 Problem fields (already lowercase) and the SCREAMING_SNAKE feature-flag key `MDDM_NATIVE_EXPORT_ROLLOUT_PCT` (env-var-mirrored constant). Does not walk free-form object content |
| `PAGINATION-DRIFT` | `spec_rules.go` | **BLOCKING (Phase E2).** Every `list*` op declares `cursor`/`limit` params with a `page.next_cursor` + `page.has_more` response, OR carries a reviewed `x-pagination-exempt: true` + non-empty `x-pagination-exempt-reason` (bounded/offset lists). `limit` max is clamped to 100 in spec and runtime |
| `AUTHZ-DRIFT` | `spec_rules.go` | **BLOCKING (Phase E2; markers reworked Phase F).** Every op is secure-in-doc via its own `security` (incl. `security: []` public opt-out) or the inherited root-level `security`. State-transition POSTs declare exactly one of `x-authz-area` / `x-authz-area-none` / `x-authz-custom`, and the `x-authz-area` shape is validated (`source: tx`→`derived_from`; `source: body\|path`→`field`) |
| `tripwire-pairing` | `code_rules.go` | Every mutating SQL in `*repository*.go` lives in a function that also calls `authz.Require` |

The three existing jobs (`backend-codegen-drift`, `frontend-codegen-drift`, `openapi-lint`) remain blocking and unchanged.

---

## 9. Historical Plan Snapshot (2026-05-12)

**RFC 9457 envelope rollout     DONE (Plan 7, 2026-05-11):**
All modules now emit `application/problem+json` on 4xx/5xx. Closes: controlled-documents T-003, approval T-001/T-003, documents T-001, auth T-003, iam T-006, audit T-002, templates T-005, taxonomy T-008.

**Plan 8     OpenAPI / contract-first completion (historical pending snapshot as of 2026-05-12):**
1. Add OpenAPI ops for documents (rename/duplicate/archive/comments/finalize), approval signoff/cancel, templates 12 hand-rolled routes, taxonomy 16 routes, audit operationId.
2. Regen via `make oapi`; rewire each module's `Register` to mount generated `ServerInterface`.
3. Flip `continue-on-error: false` in CI (historical planned step in this snapshot).

**Plan 9     Transactional + idempotency hardening (historical pending snapshot as of 2026-05-12):**
- Cursor pagination on list endpoints.
- `Idempotency-Key` middleware on `POST` create/mutate routes.
- Optimistic-lock enforcement on autosave.

Sub-projects (spec coverage for `approval`/`taxonomy`/`iam`) were tracked as separate plans in this snapshot.

---

## See also

- `wiki/architecture/api-contract.md`     operational guide: oapi-codegen wiring, module migration status, how to add a module
- `wiki/decisions/0007-two-tier-authz.md`     two-tier authz ADR + codegen-rejection amendment
- `wiki/decisions/0012-contract-first-api.md`     ADR establishing spec-as-source-of-truth
- `wiki/concepts/error-ux.md`     historical frontend error-parsing reference from the migration period (not the canonical freeze-truth source)

