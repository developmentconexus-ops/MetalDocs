# API Design System Spec — MetalDocs

> **Date:** 2026-05-10
> **Status:** Design (pre-plan)
> **Sub-project:** #1 of 6 in the API Foundation roadmap
> **Scope:** Pareto cut — 8 core conventions + mini-conventions + 3 cross-cutting gaps. Foundation for the `metaldocs-backend` skill (#2), API inventory wiki (#3), module rename (#4), spec coverage rollout (#5), and screens × API patterns analysis (#6).
> **Validated:** Codex senior-engineer pass (2026-05-10) — all 10 decisions accepted, 8 footguns + 3 gaps integrated.

---

## 1. Goal

Define the conventions that govern every HTTP endpoint MetalDocs exposes, so that:

1. The frontend never reinvents a pattern (errors, pagination, list params, idempotency).
2. The 4 unmigrated backend modules (`approval`, `taxonomy`, `iam`, `platform/auth`) can land on spec-driven codegen without re-litigating shape decisions.
3. The codebase reads as production-grade SaaS to anyone who clones the repo to self-host their own QMS.
4. Drift between spec, generated code, and runtime behavior is detectable in CI rather than runtime.

**Non-goal:** Rewrite working code. ADR-0012 (contract-first) already established codegen for 3 modules. This spec layers conventions on top of that foundation; it does not replace it.

---

## 2. Scope

**In scope (Pareto cut):**

| # | Convention | Decision |
|---|---|---|
| 1 | Error envelope | RFC 9457 Problem Details + `code` + `errors[]` |
| 2 | Pagination | Cursor-only, opaque token, uniform |
| 3 | Idempotency | Stripe-model `Idempotency-Key`, shared store, 24h TTL |
| 4 | Body validation | oapi-codegen per-operation generated validation (NOT reflection middleware) |
| 5 | RBAC declared in spec | `security` + custom `x-authz-area` extension |
| 6 | URL versioning | Static `/api/v1/...` for all paths |
| 7 | List filtering | Stripe-style flat params + per-resource allowlist |
| 8 | Sorting | JSON:API prefix style + `id` tie-breaker |
| 9 | Mini-conventions | UUIDv4, ISO 8601 UTC, snake_case, plural kebab paths, positive booleans, null-vs-absent rule, UUID-only path IDs, Bearer JWT |
| 10 | Module rename + spec file | `templates_v2/` → `templates/`; single `openapi.yaml` retained |
| A | OpenAPI compatibility gate | CI lint rules for envelope + pagination + authz declaration |
| B | Error code taxonomy | Fixed catalog of machine-readable `code` values |
| C | Cursor contract | Server-side encoding spec + tie-breaker rule |

**Explicitly deferred (out of scope):**

- ETag / `If-Match` / conditional requests
- Rate limiting infrastructure
- Audit-log requirements (separate concern, owned by audit module)
- Telemetry / observability hooks (no telemetry stack wired yet)
- Snapshot-date versioning header (`MetalDocs-Version: <date>`) — reserved as future escape hatch
- Public API quality bar (SDK gen, semver guarantees, rate-limit infra). Open-source-quality codebase ≠ public-API product.

---

## 3. Architecture

### 3.1 Spec is single source of truth

`api/openapi/v1/openapi.yaml` (OpenAPI 3.0.3) governs every HTTP contract. New endpoints are spec-first; handlers implement. ADR-0012 already established this for `registry`, `templates_v2`, `documents`. This spec extends the same rule to all conventions defined here.

### 3.2 Codegen layers

| Layer | Tool | Output | Role |
|---|---|---|---|
| Backend types + handlers | `oapi-codegen` v2 | `internal/modules/<x>/api/api.gen.go` | Request/response Go types, `ServerInterface`, per-op validation |
| Backend authz middleware | Custom oapi-codegen template | Generated route decoration | Reads `security` + `x-authz-area`, emits `CapabilityService` + `authz.Require` calls |
| Frontend types | `openapi-typescript` v7 | `frontend/apps/web/src/lib/api-types/index.d.ts` | Typed `paths` for `lib/api/client.ts` |
| Frontend error parser | Hand-written | `frontend/apps/web/src/lib/api/problem.ts` | Parses RFC 9457 envelope; replaces current `ApiError` shape |

### 3.3 Drift detection (CI)

`.github/workflows/api-contract.yml` extends current 3 jobs with new lint job (Gap A):

| Existing | New |
|---|---|
| `backend-codegen-drift` | `envelope-drift` (every error response uses RFC 9457 schema) |
| `frontend-codegen-drift` | `pagination-drift` (every list op declares `cursor`/`limit`/page response) |
| `openapi-lint` | `authz-drift` (every operation declares `security`; tier-2 ops declare `x-authz-area`) |

---

## 4. Convention specs

### 4.1 Error envelope (Decision 1)

All error responses (4xx, 5xx) MUST emit `Content-Type: application/problem+json` with this shape:

```json
{
  "type": "https://errors.metaldocs.io/<category>",
  "title": "<short human-readable title>",
  "status": 400,
  "detail": "<longer human-readable explanation>",
  "instance": "/api/v1/templates",
  "code": "VALIDATION_ERROR",
  "errors": [
    { "field": "key", "code": "REQUIRED", "message": "key is required" },
    { "field": "name", "code": "OUT_OF_RANGE", "message": "name must be ≤ 200 chars" }
  ]
}
```

- `type`, `title`, `status`, `detail`, `instance` — RFC 9457 standard fields.
- `code` — extension. Required. Machine-readable, drawn from the taxonomy in §6.
- `errors[]` — extension. Optional. Field-level errors for `VALIDATION_ERROR`. Each entry has `field` (JSON pointer or dot path), `code` (REQUIRED / INVALID_FORMAT / OUT_OF_RANGE / etc.), `message`.

**Footgun fix (Codex #1):** Contract test in CI asserts every 4xx/5xx response in `openapi.yaml` schemas to a `Problem` ref. Lint job `envelope-drift` enforces.

**Frontend impact:** `lib/api/client.ts` `ApiError` rewritten to parse new shape. `resolveErrorMessage` switches on `code`. Wizard form steps consume `errors[]` for per-field error rendering. ~150 LOC + tests.

### 4.2 Pagination (Decision 2)

All list endpoints MUST use cursor pagination.

**Request:**
```
GET /api/v1/documents?cursor=<opaque>&limit=20
```

- `cursor` — opaque base64url-encoded JSON. See §6 (Cursor contract).
- `limit` — integer, default 20, max 100. Server enforces hard cap.
- `cursor` and `limit` are query-string params, NOT body.

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

- `next_cursor: null` ⇒ no more pages.
- `has_more: false` ⇒ same.

**Total count opt-in (Codex #2 fix):**

- Default: no total computed (avoid `COUNT(*)` cost on big tables).
- Opt-in: `?include=total` ⇒ response includes `page.total_count: <int>`.
- Each list endpoint documents whether it supports `include=total`. Endpoints over high-volume tables (audit, documents) MAY refuse with 400 `INCLUDE_NOT_SUPPORTED`.

**Admin/export iteration (Codex #2 fix):** A separate convention will be defined when first export feature lands. Not in this spec.

**Frontend impact:** TanStack `useInfiniteQuery` is the primary pattern. Existing `listTemplates` + `listDocuments` migrate.

### 4.3 Idempotency (Decision 3)

All unsafe writes (POST, PUT, PATCH, DELETE) MUST support optional `Idempotency-Key` header.

**Behavior:**

| Client sends | Server behavior |
|---|---|
| No `Idempotency-Key` | Non-idempotent. Caller's choice. |
| `Idempotency-Key: <uuid>` first time | Execute, store `(key, body_hash, status, response_body, expires_at)`, return response. |
| Same key + same canonical body hash within TTL | Replay stored response. Status `200 OK` even if original was `201`. |
| Same key + different body hash | `422 IDEMPOTENCY_KEY_REUSED`. Don't execute. |
| Same key after TTL expired | Treat as new request. Execute. |

**Implementation notes:**

- **Body hash:** SHA-256 over raw request bytes (Stripe model). Clients MUST send byte-identical retries. Existing `internal/platform/idempotency/middleware.go` already implements this. MetalDocs is internal-only with a single TypeScript client — deterministic serialization is guaranteed by the client; canonical-JSON normalization is unnecessary complexity. Reconsider only if a public API ships.
- **TTL:** 24h flat. Document this. Future op-specific TTLs require ADR.
- **Long-running ops:** Sync handlers must finish before TTL window closes. Async ops (PDF render) use job-id pattern, not idempotency key.

**Storage (already shipped — migration `0147_idempotency_keys.sql`):**

```sql
-- metaldocs.idempotency_keys (existing, do not replace)
PRIMARY KEY (tenant_id, actor_user_id, route_template, key)
-- 4-col PK is intentional defense-in-depth:
--   actor_user_id prevents User A replaying User B's response
--   route_template prevents same key colliding across endpoints
```

**Plan 2 reconciliation scope (NOT Plan 1):**

1. Switch middleware error envelope from legacy `{code, message}` → RFC 9457 `Problem`.
2. Rename conflict code `IDEMPOTENCY_KEY_CONFLICT` → `IDEMPOTENCY_KEY_REUSED` (taxonomy alignment).
3. Add `Idempotent-Replay: true` header on replay path.
4. Delete `internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go` — duplicate of platform store.

**Frontend impact:** Mutation hooks auto-generate `crypto.randomUUID()` per attempt. Helper in `lib/api/mutations.ts` ensures TanStack Query retries reuse same key.

### 4.4 Body validation (Decision 4 — REVERSED per Codex)

**Original plan (rejected):** Reflection-based shared middleware that walks decoded struct, checks required fields nil.

**Codex pushback (#4):** Generic middleware misses nested required fields, drifts from OpenAPI, can't see operation-specific schema rules.

**Final decision:** Use oapi-codegen's per-operation generated validation.

- Spec declares `required: [...]` on every schema.
- oapi-codegen config `generate.embedded-spec: true` + `validate-against-spec: true` (or post-decode validator hook).
- Per-op validator emits 400 + RFC 9457 `errors[]` with `field` from JSON pointer + `code: REQUIRED`.
- Spec-aware: nested required fields, format constraints, min/max, enum membership all enforced.

**Strict-decode still mandatory:** `DisallowUnknownFields` enforced at HTTP boundary (current `readStrictJSON` pattern). Unknown field → 400 `UNKNOWN_FIELD` with `errors[].field`.

**Why this works better than reflection:**

- Operation-specific validation (e.g., `template.key` regex `^[a-z0-9-]{1,64}$`) is in the spec, generated into validators.
- No drift: spec change → regenerate → validator updates. CI catches mismatch.
- Reuses existing oapi-codegen machinery; no new custom middleware to maintain.

**Codegen flag changes:**
- `generate.std-http-server: true` (already on).
- `generate.strict-server: true` (already on, gives validator hooks).
- Required fields stay value-typed (oapi-codegen default). The per-op validator handles "missing vs empty" distinction.

### 4.5 RBAC declared in spec (Decision 5)

Two-tier authz model documented in `wiki/decisions/0007-two-tier-authz.md`. Both tiers declared in spec.

**Tier 1 — capability (HTTP-level):**
```yaml
paths:
  /api/v1/templates:
    post:
      security:
        - bearerAuth: [template.create]
```
Codegen middleware reads `security`, calls `CapabilityService.Require("template.create", userID)`.

**Tier 2 — area (in-tx):**
```yaml
paths:
  /api/v1/templates/{id}/publish:
    post:
      security:
        - bearerAuth: [template.publish]
      x-authz-area:
        source: "$request.body#/areas[*]"
        capability: "template.publish"
```
Codegen template emits `authz.Require(ctx, capability, areaCodes)` after body decoded.

**Footgun fix (Codex #5):** New CI lint job `authz-drift`:

1. Every operation in `openapi.yaml` declares `security` (or explicit `security: []` to opt out).
2. Operations matching pattern `/api/v1/{module}/{id}/{verb}` where `verb` is a state transition (`publish`, `submit`, `approve`, `reject`) MUST declare `x-authz-area`.
3. Generated handler files contain the authz middleware call (regex check on `api.gen.go`).

Failure → CI red. Drift impossible.

**Self-host operator benefit:** Redoc renders `security` requirements + `x-authz-area` extension natively. Operator reads spec, sees capability matrix per op, no source code grep.

### 4.6 URL versioning (Decision 6)

All paths under `/api/v1/...`. No `/api/v2/...` exists today; not introduced.

| Old | New |
|---|---|
| `/api/v2/templates` | `/api/v1/templates` |
| `/api/v2/documents` | `/api/v1/documents` |
| `/api/v2/controlled-documents` | `/api/v1/controlled-documents` |
| `/api/v2/approvals` | `/api/v1/approvals` |
| `/api/v1/auth` | `/api/v1/auth` (unchanged) |
| `/api/v1/feature-flags` | `/api/v1/feature-flags` (unchanged) |

**Migration timing:** Path rewrite is mechanical (ServeMux pattern + frontend constants). Lands as part of sub-project #4 (module rename) after this spec lands. NOT part of this sub-project's plans.

**Future major version:** New `/api/v1/` is THE stable major version. If a future breaking change is genuinely unavoidable, choose:
- (a) Date-snapshot header `MetalDocs-Version: 2027-XX-XX` (Stripe pattern), OR
- (b) `/api/v2/...` parallel path (GitHub legacy).
Decision deferred to that day. ADR will record.

### 4.7 List filtering (Decision 7)

Stripe-style flat query params. Per-resource allowlist + typed parser (Codex #7 fix).

**Pattern:**

```
?status=draft                       # equality (single value)
?status=draft&status=published      # multi-value (repeat key)
?created_after=2026-01-01T00:00:00Z # named range, ISO 8601
?created_before=2026-12-31T23:59:59Z
?q=annual%20audit                   # free-text search (FTS-backed when avail)
?area=quality&area=engineering      # multi-value enum
```

**Naming convention:**
- Equality: `<field>=value`
- Multi-value: repeat `<field>=v1&<field>=v2`
- Range: `<field>_after`, `<field>_before` (ISO 8601 for timestamps; numeric otherwise)
- Free-text search: `q=<term>`

**Per-resource allowlist (Codex #7 fix):** Each list operation in spec declares allowed filter keys via `parameters`. Unknown keys → 400 `UNKNOWN_FILTER`. Typed parsing per param schema (string/integer/boolean/date-time/enum). Enforced by oapi-codegen-generated validation.

**Frontend helper:** `lib/api/list-params.ts` builds query strings from a typed object. Returns frozen URLSearchParams.

**Reserved keys:** `cursor`, `limit`, `sort`, `include` — reserved by pagination/sorting/total. Endpoints MUST NOT use these as filter keys.

### 4.8 Sorting (Decision 8)

JSON:API prefix style.

```
?sort=-updated_at             # single field, desc
?sort=name                    # single field, asc
?sort=-status,name            # multi-key: status desc, then name asc
```

- Leading `-` → desc. No prefix → asc.
- Comma separates multi-key.
- Per-endpoint allowlist of sortable fields. Unknown field → 400 `INVALID_SORT_FIELD`.
- Default sort documented per endpoint when `?sort` absent.

**Footgun fix (Codex #8):** Cursor encoding (§6) MUST include `id` tie-breaker when sort field is non-unique (e.g., `updated_at` can collide). Cursor decode validates that current request's sort matches cursor's stored sort; mismatch → 400 `INVALID_CURSOR`.

### 4.9 Mini-conventions (Decision 9)

Codify current patterns. No churn.

| Topic | Decision |
|---|---|
| ID format | UUIDv4 (`gen_random_uuid()`). UUIDv7 considered, rejected for Pareto cut (no measurable perf wall). |
| Timestamps | ISO 8601 UTC, mandatory `Z` suffix (`2026-05-10T14:30:00Z`). No tz offsets in wire format. Frontend converts to local. |
| JSON case | `snake_case`. Matches DB columns. openapi-typescript already generates matching field names. |
| Resource paths | Plural lowercase, kebab-case for compound (`/templates`, `/controlled-documents`). |
| Booleans | Positive form (`is_archived`, `is_published`). Never `is_not_active`. Prefer `archived_at: timestamp\|null` for state-with-when. |
| Null vs absent | Required+nullable → field always present, value `null` or actual value. Optional → field omitted entirely when absent. |
| ID in URL path | UUID only. No slugs in paths. Slugs/keys are display-layer (e.g. template `key` field). |
| Auth scheme | `Authorization: Bearer <jwt>`. Declared in `securitySchemes.bearerAuth`. |

### 4.10 Module rename + spec file (Decision 10)

**Rename plan (executed in sub-project #4, NOT this sub-project):**

| Backend dir | Frontend feature | API path | Notes |
|---|---|---|---|
| `internal/modules/templates_v2/` → `templates/` | `features/templates/` (already correct) | `/api/v2/templates` → `/api/v1/templates` | Atomic single-PR rename. |
| `internal/modules/documents/` (already correct) | `features/documents/` (already correct) | `/api/v2/documents` → `/api/v1/documents` | Path migration only. |
| All `documents_v2`, `_v2` references in code/tests/docs/migrations | — | — | Purge in same PR. Migration 0168 already dropped the `documents_v2` table. |

**Footgun fix (Codex #10):** Rename PR is atomic. No shims, no re-exports, no compatibility URLs. CI must pass with all paths updated in same change.

**Spec file:** Single `api/openapi/v1/openapi.yaml` retained. Splitting into per-module fragments stitched at codegen time deferred until 100+ paths (currently ~30–40). Reconsider trigger: per-PR diff exceeds 1000 lines or codegen runtime exceeds 30s.

---

## 5. Cross-cutting (Codex gaps)

### 5.1 OpenAPI compatibility gate (Gap A)

CI workflow `.github/workflows/api-contract.yml` adds three lint jobs to the existing three:

| New job | Rule |
|---|---|
| `envelope-drift` | Every response with status ≥ 400 references `#/components/schemas/Problem`. |
| `pagination-drift` | Every operation tagged `list` declares query params `cursor`, `limit`; response schema includes `page.next_cursor`, `page.has_more`. |
| `authz-drift` | Every operation declares `security` (or explicit `security: []`); state-transition ops declare `x-authz-area`. |

Implementation: bespoke linter script in `scripts/api-lint/` (Go or Node — language TBD in plan). Runs after `redocly lint`.

### 5.2 Error code taxonomy (Gap B)

Fixed catalog. Documented in `wiki/concepts/error-codes.md` (new doc, created in plan).

| Code | HTTP | Meaning |
|---|---|---|
| `VALIDATION_ERROR` | 400 | Body/params failed schema or constraint validation. `errors[]` populated. |
| `REQUIRED` | (errors[].code) | Field-level: required field missing. |
| `INVALID_FORMAT` | (errors[].code) | Field-level: format violation (regex, ISO 8601, UUID, etc.). |
| `OUT_OF_RANGE` | (errors[].code) | Field-level: numeric or length range violation. |
| `INVALID_ENUM` | (errors[].code) | Field-level: value not in declared enum. |
| `UNKNOWN_FIELD` | 400 | Strict-decode rejected unknown JSON field. |
| `UNKNOWN_FILTER` | 400 | Filter key not in operation's allowlist. |
| `INVALID_SORT_FIELD` | 400 | Sort field not allowed for endpoint. |
| `INVALID_CURSOR` | 400 | Cursor decode failed or filter/sort mismatch. |
| `INCLUDE_NOT_SUPPORTED` | 400 | `?include=<x>` not supported by endpoint. |
| `UNAUTHENTICATED` | 401 | No JWT or invalid JWT. |
| `FORBIDDEN_CAPABILITY` | 403 | User lacks capability (tier-1). |
| `FORBIDDEN_AREA` | 403 | User lacks area-scoped permission (tier-2). |
| `NOT_FOUND` | 404 | Resource does not exist or caller cannot see it. |
| `ALREADY_EXISTS` | 409 | Unique-constraint violation (e.g., key collision). |
| `STATE_TRANSITION_INVALID` | 409 | Workflow state transition not allowed (e.g., publish a draft that has missing tokens). |
| `CONCURRENT_MODIFICATION` | 409 | Optimistic-lock failure (`expected_lock_version` mismatch). |
| `IDEMPOTENCY_KEY_REUSED` | 422 | Same `Idempotency-Key` with different body. |
| `IDEMPOTENCY_REPLAY` | 200/2xx | Stored response replayed. (Returned with original status; informational header `Idempotent-Replay: true`.) |
| `RATE_LIMITED` | 429 | Reserved. Not implemented in this spec. |
| `INTERNAL_ERROR` | 500 | Unhandled server error. Generic. |

**Adding new codes:** PR with ADR justifying the code. Renaming or removing existing codes = breaking change → requires `MetalDocs-Version` snapshot bump (future).

### 5.3 Cursor contract (Gap C)

**Server-side encoding (opaque to clients):**

```json
{
  "v": 1,
  "sort": [{"field": "updated_at", "dir": "desc"}, {"field": "id", "dir": "asc"}],
  "anchor": {"updated_at": "2026-05-10T14:30:00Z", "id": "550e8400-e29b-41d4-a716-446655440000"},
  "filter_hash": "<sha256 of sorted filter params>"
}
```

- `v: 1` — encoding version. Bump on breaking encoding change.
- `sort` — full sort spec, including `id` tie-breaker (always last).
- `anchor` — last row's values for each sort field.
- `filter_hash` — fingerprint of all filter params. Mismatch on next request → 400 `INVALID_CURSOR`.

**Encoding:** `base64url(json(<payload>))`. No encryption — opacity only, not security.

**Defaults / limits:**
- `limit` default `20`, max `100`.
- Server rejects `limit > 100` with 400 `OUT_OF_RANGE`.

**Tie-breaker rule:** When sort field is non-unique (anything except `id`), cursor MUST include `id` as final sort key with deterministic direction (`asc` by convention).

**Filter/sort change between requests:** Reject cursor with 400 `INVALID_CURSOR`. Client must restart pagination from beginning.

**Helper:** `internal/platform/pagination/cursor.go` (encode/decode/validate). Used by every list handler.

---

## 6. Implementation impact

### 6.1 Blast radius

| Surface | Impact |
|---|---|
| Backend handlers (3 migrated) | All regenerated. Path rewrites. Authz decoration generated. ~50 mechanical edit sites + 200 LOC custom codegen template. |
| Backend handlers (4 unmigrated) | Spec authoring + bootstrap + handler migration. Sub-project #5 territory. This spec defines the conventions they adopt. |
| Frontend `lib/api/client.ts`, `ApiError`, `resolveErrorMessage` | Rewrite for RFC 9457 envelope. ~150 LOC + tests. |
| Frontend list pages | Migrate to `useInfiniteQuery` + cursor. `templatesV2.ts` + `documents.ts` + any other list call. ~100 LOC across features. |
| Frontend wizard forms | Consume `errors[]` for per-field errors. ~50 LOC across StepIdentity, StepConfirmation, NovoDocumento steps. |
| Wiki | New: `concepts/error-codes.md`, `architecture/api-design-system.md` (this spec promoted). Updated: `architecture/api-contract.md`, `concepts/error-ux.md`. |
| Migrations | New: `0184_idempotency_keys.sql`. |
| CI | New jobs: `envelope-drift`, `pagination-drift`, `authz-drift`. |

### 6.2 Migration order (rollout sequence)

This spec produces ~2 implementation plans. Suggested ordering for the writing-plans phase:

**Plan 1 — Foundation primitives (~1 week):**
1. RFC 9457 envelope + Problem schema in spec
2. Error code taxonomy doc + constants
3. Idempotency table + middleware + canonical hash helper
4. Cursor helper + Problem helper
5. Authz custom codegen template
6. CI lint jobs
7. Frontend `lib/api/problem.ts` + `ApiError` rewrite

**Plan 2 — Adoption + path migration (~1 week):**
1. Migrate `registry`, `templates_v2`, `documents` to new envelope + cursor + idempotency
2. Path rewrite `/api/v2/...` → `/api/v1/...`
3. Module rename `templates_v2/` → `templates/` (sub-project #4 mechanical)
4. Frontend list pages → `useInfiniteQuery`
5. Frontend wizard forms → `errors[]` consumption
6. Delete `PostgresSignoffIdempStore`

Sub-project #5 (spec coverage for `approval`/`taxonomy`/`iam`/`platform`) consumes these conventions; runs as separate sub-project plans, one per module.

### 6.3 Backward compatibility

**No public clients today.** Internal-only frontend ships in lockstep. Breaking changes acceptable. No deprecation window required.

**Single-PR atomicity:**
- Spec changes + codegen + handler migration + frontend changes land together per Plan.
- Path rewrite (Plan 2 step 2) is atomic across server route registration + frontend `lib/api/client.ts` constants.

---

## 7. Open risks

| Risk | Mitigation |
|---|---|
| Custom oapi-codegen authz template breaks on generator upgrade | Pin oapi-codegen version. CI runs codegen on every PR; upgrades require explicit ADR + template revalidation. |
| Idempotency table grows unbounded if cleanup job misses | Background cleanup runs hourly: `DELETE WHERE expires_at < now()`. Monitor row count. Alert at 1M rows. |
| Cursor encoding v1 needs breaking change someday | `v: 1` field reserved. Bump to `v: 2`, decoder rejects mismatched version with 400 `INVALID_CURSOR`. Clients restart pagination — same behavior as filter/sort change. |
| `errors[]` field-level codes proliferate ad-hoc | Field-level `code` values (REQUIRED / INVALID_FORMAT / etc.) governed by same taxonomy in §5.2. New codes via PR + ADR. |
| 4 unmigrated modules ship inconsistent patterns until #5 lands | Sub-project #5 plans explicitly adopt this spec's conventions per module. Until then, their endpoints carry "pre-migration" status in spec. CI lint jobs scoped to migrated modules until all are in. |

---

## 8. Validation

**Codex senior-engineer pass (2026-05-10):** All 10 decisions accepted directionally. 8 footguns integrated as fixes. 3 missing decisions (Gaps A/B/C) added to scope. Reversed Decision 4 from reflection middleware to per-op generated validation.

**User decisions on disputed items:**
- UUIDv4 over UUIDv7 — accept Pareto cut argument.
- 24h flat TTL over op-specific — accept simplicity.

---

## 9. See also

- `wiki/architecture/api-contract.md` — operational reference for current codegen pattern (will be updated post-implementation)
- `wiki/decisions/0012-contract-first-api.md` — ADR establishing spec-as-source-of-truth
- `wiki/decisions/0007-two-tier-authz.md` — two-tier authz model (CapabilityService + authz.Require)
- `wiki/concepts/error-ux.md` — current frontend error parsing (to be rewritten)
- `wiki/backlog/contract-first-followups.md` — 4 unmigrated modules (sub-project #5)

---

## 10. Roadmap

This spec is **sub-project #1** of 6 in the API Foundation roadmap:

| # | Sub-project | Status |
|---|---|---|
| 1 | API Design System spec | **THIS DOC** — design phase complete pending user review |
| 2 | `metaldocs-backend` skill + `backend-code-reviewer` agent | Blocked on #1 |
| 3 | API inventory wiki + Redoc page | Blocked on #1 |
| 4 | Module naming refactor (`templates_v2` → `templates`, all paths to `/api/v1/`) | Blocked on #1 (mechanical work after spec lands) |
| 5 | Spec coverage rollout (`approval`, `taxonomy`, `iam`, `platform`) | Blocked on #1 + #2 (4 separate plans) |
| 6 | Screens × API patterns matrix | Optional, runs alongside #5 if drift found |

Each sub-project is its own spec → plan → implementation cycle.
