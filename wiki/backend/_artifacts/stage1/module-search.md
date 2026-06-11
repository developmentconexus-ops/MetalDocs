# Stage-1 Audit Artifact — module-search

> **Generated:** 2026-06-10
> **Branch at audit:** qa/iam-area-membership
> **Audit scope:** `internal/modules/search/` — all files read exhaustively.
> **Read-only:** no source files were modified.

---

## 1. Identity & purpose

The `search` module provides a single cross-cutting read surface that lets an authenticated, tenanted caller search across controlled documents (and their linked document versions) using free-text and structured filters. There is no dedicated search index: results are produced by a live SQL JOIN against `public.documents` and `public.controlled_documents` at query time, which guarantees zero staleness at the cost of query weight.

Per-document visibility is enforced entirely at the data layer. The SQL query applies the same five-case visibility predicate used by the controlled-documents list reader (AD-3 / ADR 0022): company-scoped CDs are visible to every tenant member; restricted CDs are visible to the owner, to users with an area grant, or to users with a direct user grant; standalone documents (no linked CD) are visible only to their creator. Unauthenticated callers are short-circuited at the service layer before the reader is ever called.

Template search is intentionally deferred (tech-debt T-002). The module is classified L2/active with one major and one minor outstanding debt item.

---

## 2. File inventory

| File | Role |
|---|---|
| `internal/modules/search/domain/model.go` | Domain types: `Document`, `Query`, `Classification`/`Status` consts, `NewDocument`/`NewQuery` constructors, three sentinel errors |
| `internal/modules/search/domain/port.go` | `Reader` interface — one method: `ListDocuments(ctx, query, limit, offset)` |
| `internal/modules/search/application/service.go` | Application service: authn guard, actor injection, limit capping, single `ListDocuments` call; exports `Service`, `NewService`, `ErrTenantRequired` |
| `internal/modules/search/application/service_test.go` | Unit tests for service: authn guard, actor/tenant forwarding, limit cap, unauthenticated short-circuit |
| `internal/modules/search/delivery/http/handler.go` | HTTP delivery: route registration (`GET /api/v1/search/documents`), query-param parsing, JSON response marshalling; exports `Handler`, `NewHandler`, `Searcher` interface, `SearchDocumentResponse` struct |
| `internal/modules/search/delivery/http/handler_test.go` | Unit tests for handler: auth guards, tenant context, filter mapping, response shape |
| `internal/modules/search/infrastructure/v2documents/reader.go` | SQL reader: 13-param parameterised query against `public.documents LEFT JOIN public.controlled_documents`; enforces unified visibility predicate; exports `Reader`, `NewReader` |
| `internal/modules/search/infrastructure/v2documents/reader_test.go` | Sqlmock unit tests: tenant filter, actor-binding assertion |
| `internal/modules/search/infrastructure/v2documents/reader_visibility_integration_test.go` | Build-tagged (`integration`) integration test: seeds DB fixtures, proves all five visibility cases end-to-end |
| `internal/modules/search/application/.gitkeep` | Empty marker — no artefact significance |
| `internal/modules/search/delivery/http/.gitkeep` | Empty marker — no artefact significance |
| `internal/modules/search/domain/.gitkeep` | Empty marker — no artefact significance |
| `internal/modules/search/infrastructure/.gitkeep` | Empty marker — no artefact significance |

Total: 9 substantive Go files across 4 packages. No god file; largest is `reader.go` at ~190 lines.

---

## 3. Public surface

### Exported Go types consumed by the wiring layer (`apps/api/cmd/metaldocs-api/main.go`)

| Symbol | Package | Consumed by |
|---|---|---|
| `NewReader(db *sql.DB) *Reader` | `infrastructure/v2documents` | `main.go:219` — constructs the SQL reader |
| `NewService(reader domain.Reader) *Service` | `application` | `main.go:219` — constructs the service |
| `NewHandler(service Searcher) *Handler` | `delivery/http` | `main.go:220` — constructs the HTTP handler |
| `(*Handler).RegisterRoutes(mux *http.ServeMux)` | `delivery/http` | `main.go:284` — registers route onto shared mux |

### HTTP routes

| Method | Path | Auth | Authz | Handler |
|---|---|---|---|---|
| `GET` | `/api/v1/search/documents` | Required (IAM session via `authMiddleware`) | None beyond authenticated principal; visibility enforced by SQL predicate per AD-3 | `handler.go:49-51` |

### Query parameters accepted by the handler

| Parameter | Type | Notes |
|---|---|---|
| `q` | string | Free-text; LIKE `%lower(q)%` against `d.name` |
| `document_type` | string | Maps to `profile_code_snapshot`; takes precedence over `document_profile` when both supplied |
| `document_profile` | string | Maps to `profile_code_snapshot`; fallback if `document_type` is empty |
| `document_family` | string | Resolved via subquery on `metaldocs.document_profiles` |
| `process_area` | string | Maps to `process_area_code_snapshot` |
| `department` | string | Maps to `cd.department_code` |
| `owner_id` | string | Maps to `d.created_by` |
| `status` | string | UPPER-normalised; matched against `d.status` |
| `expiry_before` | RFC 3339 datetime | `d.effective_to <= $9` |
| `expiry_after` | RFC 3339 datetime | `d.effective_to >= $10` |
| `limit` | integer | Capped to 100; defaults to 20 |
| `subject` | string | **Accepted and forwarded to `Query.Subject` but the SQL reader ignores it** (not a `public.documents` column) |
| `businessUnit` | string | **camelCase** — not renamed in the snake_case standardisation pass; forwarded to `Query.BusinessUnit` but SQL reader ignores it |
| `classification` | string | Forwarded to `Query.Classification` but SQL reader ignores it |
| `tag` | string | Forwarded to `Query.Tag` but SQL reader ignores it |

Parameters `subject`, `businessUnit`, `classification`, and `tag` are **not listed in `api/openapi/v1/openapi.yaml`** (lines 802-882) and are silently no-ops at the data layer.

### Response shape

`200 OK` — `application/json`:
```json
{ "items": [ <SearchDocumentResponse> ] }
```

`SearchDocumentResponse` fields: `document_id`, `title`, `document_type`, `document_profile`, `document_family`, `document_sequence`, `document_code`, `process_area`, `subject`, `owner_id`, `business_unit`, `department`, `classification`, `status`, `tags`, `effective_at`, `expiry_at`, `created_at`.

Fields `subject`, `business_unit`, `classification`, and `tags` are always zero-valued in practice (empty string / null array) because the SQL reader never populates them.

---

## 4. Logic flows

### Flow 1 — Authenticated document search (happy path)

1. **HTTP layer — auth guard** (`handler.go:59-67`): `iamdomain.UserIDFromContext` is called directly on the request context. If empty, respond `401` immediately. `tenant.FromContext` is then called; if missing, respond `401`.
2. **HTTP layer — parameter parsing** (`handler.go:69-107`): `limit` is parsed as integer; `expiry_before` / `expiry_after` as RFC 3339 via `parseOptionalDateTimeQuery`; remaining filters read as plain query strings. A `searchdomain.Query` is assembled and passed to `h.service.SearchDocuments`.
3. **Service layer — actor injection** (`service.go:33-53`): `authn.UserIDFromContext` is called on the context (the platform wrapper around `iamdomain.UserIDFromContext`). If the actor is absent or empty after trim, the function returns an empty slice immediately without touching the reader. `q.ActorUserID` is set to the trimmed actor string. `domain.NewQuery(q)` validates that `TenantID` is non-empty (returns `ErrQueryTenantEmpty` → mapped to `ErrTenantRequired` by the service). `effectiveLimit` clamps `limit` to `[1, 100]` with a default of 20. `s.reader.ListDocuments(ctx, normalized, limit, 0)` is called with a fixed offset of 0 (no pagination at service layer — offset is always 0).
4. **Infrastructure layer — SQL execution** (`reader.go:21-182`): The 13-argument parameterised query is executed against `deps.SQLDB`. `$1`=tenantID, `$2`=lower(Text), `$3`=UPPER(Status), `$4`=lower(profileFilter), `$5`=lower(DocumentFamily), `$6`=lower(ProcessArea), `$7`=Department, `$8`=OwnerID, `$9`=ExpiryBefore (timestamptz or NULL), `$10`=ExpiryAfter (timestamptz or NULL), `$11`=limit, `$12`=offset (always 0), `$13`=ActorUserID. The visibility predicate (lines 77-106) is part of the WHERE clause — no post-fetch filtering. A `document_profiles` subquery appears twice in the SELECT list and once in the WHERE clause (lines 34-41, 59-66, 34-41 duplicate).
5. **Infrastructure layer — row scan** (`reader.go:145-177`): Each row is scanned into a `searchdomain.Document`. `DocumentType` is assigned the same value as `DocumentProfile` (`doc.DocumentType = doc.DocumentProfile`, line 172). `EffectiveAt` and `ExpiryAt` are converted from `sql.NullTime` to `*time.Time` via `cloneNullTime`. `Subject`, `BusinessUnit`, `Classification`, `Tags` remain zero-valued.
6. **HTTP layer — response marshalling** (`handler.go:113-137`): Results are mapped to `SearchDocumentResponse` slice and written as `{"items": [...]}` with `httpresponse.WriteJSON` (`200 OK`, `Content-Type: application/json`).

### Flow 2 — Unauthenticated caller short-circuit

1. **HTTP layer** (`handler.go:59`): `iamdomain.UserIDFromContext` returns empty string. Handler writes `401` via `httpresponse.WriteError` and returns.
2. Reader is never called. No DB round-trip.

### Flow 3 — Missing tenant context (authenticated but no tenant)

1. **HTTP layer** (`handler.go:63-67`): `tenant.FromContext` returns an error (tenant not set by middleware). Handler writes `401` with code `AUTH_UNAUTHORIZED`.
2. Reader is never called.

### Flow 4 — Authenticated caller with no document grants (visibility predicate excludes all rows)

1. Flow proceeds identically to Flow 1 through to the SQL execution step.
2. The visibility predicate in the WHERE clause (`reader.go:77-106`) excludes all rows for the caller: no company-scoped CDs exist, the caller is not an owner, has no area grants, has no user grants, and is not the creator of any standalone document.
3. `rows.Next()` is never entered; `out` remains `nil`.
4. **Service layer** (`service.go:57-59`): `nil` slice is replaced with `[]domain.Document{}`.
5. Handler marshals `{"items": []}` and responds `200 OK`. No error path is taken — zero results is not an error.

### Flow 5 — Visibility predicate detail: restricted CD with area grant

The full multi-step join at `reader.go:82-96` is:
1. Check `cd.visibility_scope = 'restricted'`.
2. Correlated EXISTS on `public.controlled_document_area_grants cdag` where `cdag.controlled_document_id = cd.id` and `cdag.tenant_id = cd.tenant_id`.
3. Inner correlated EXISTS on `public.user_process_areas upa` where `upa.user_id = $13`, `upa.area_code = cdag.area_code`, `upa.effective_to IS NULL` (active membership required).
4. If that fails, a second correlated EXISTS on `public.controlled_document_user_grants cdug` where `cdug.controlled_document_id = cd.id` and `cdug.user_id = $13`.

This predicate is a verbatim port of the one in `internal/modules/controlleddocuments/infrastructure/repository.go:128`.

---

## 5. Dependencies

### Outbound — packages this module imports

| Imported package | Import path | Why |
|---|---|---|
| `domain` (self) | `metaldocs/internal/modules/search/domain` | Domain types and port interface |
| `iam/domain` | `metaldocs/internal/modules/iam/domain` | `UserIDFromContext` (handler), `WithAuthContext` (tests), `Role` (tests) |
| `platform/authn` | `metaldocs/internal/platform/authn` | `UserIDFromContext` — presence-aware actor resolver (service) |
| `platform/httpresponse` | `metaldocs/internal/platform/httpresponse` | `WriteError`, `WriteJSON` (handler) |
| `platform/tenant` | `metaldocs/internal/platform/tenant` | `FromContext` — tenant ID extraction (handler) |
| `database/sql` | stdlib | SQL driver interface (reader) |
| `context`, `errors`, `fmt`, `net/http`, `strconv`, `strings`, `time` | stdlib | Various |
| `github.com/DATA-DOG/go-sqlmock` | external | SQL mock in unit tests only |
| `metaldocs/tests/integration/testdb` | internal test helper | Integration test fixture helpers |

### Outbound — tables read by `reader.go`

| Table | Schema | Purpose |
|---|---|---|
| `documents` | `public` | Primary document rows |
| `controlled_documents` | `public` | CD metadata, visibility scope, owner, department |
| `controlled_document_area_grants` | `public` | Area-based restricted CD grants |
| `controlled_document_user_grants` | `public` | Direct user restricted CD grants |
| `user_process_areas` | `public` | Active area memberships for area-grant resolution |
| `document_profiles` | `metaldocs` | Family code lookup (cross-schema subquery) |

### Inbound — who imports this module (verified with grep)

| Importer | File | What it uses |
|---|---|---|
| `apps/api/cmd/metaldocs-api/main.go` | `main.go:58-60, 219-220, 284` | `NewReader`, `NewService`, `NewHandler`, `RegisterRoutes` |

No other production code imports the search module's packages. The module is wired exclusively via the binary entry point.

---

## 6. Persistence

### Tables

| Table | Schema | Access | Notes |
|---|---|---|---|
| `public.documents` | `public` | SELECT only | Primary entity; `archived_at IS NULL` filter always applied |
| `public.controlled_documents` | `public` | SELECT only (LEFT JOIN) | Visibility scope, department, code, sequence |
| `public.controlled_document_area_grants` | `public` | SELECT only (correlated EXISTS) | Area-based restricted visibility |
| `public.controlled_document_user_grants` | `public` | SELECT only (correlated EXISTS) | Direct user restricted visibility |
| `public.user_process_areas` | `public` | SELECT only (correlated EXISTS) | Active area memberships; `effective_to IS NULL` active filter |
| `metaldocs.document_profiles` | `metaldocs` (legacy schema) | SELECT only (scalar subquery) | Family code resolution; cross-schema reference |

### Query patterns

- Single parameterised `SELECT` per search call (`reader.go:28-109`).
- `document_profiles` family-code subquery appears **twice** in the query: once in the SELECT list (lines 34-41) and once in the WHERE clause filter (lines 59-66). The queries are structurally identical, meaning two correlated subquery executions per row for the family filter.
- `LIKE '%' || $2 || '%'` pattern for free-text (`reader.go:56`) — leading wildcard, not index-friendly on large datasets.
- `ORDER BY d.created_at DESC, d.id DESC` — deterministic ordering.
- No write path. No outbox. No triggers. Fully read-only persistence behaviour.

### Migration files

- `db/migrations/0232_drop_document_access_policies.sql` — dropped `document_access_policies` table (T-003 closure); removes the last persistence surface that the old ABAC path depended on. No search-specific migration beyond this.

---

## 7. Config & environment

The search module consumes no configuration keys or environment variables directly. It receives a `*sql.DB` handle injected at construction time (`main.go:219`) which is sourced from `deps.SQLDB` — the platform-level database connection pool configured by `internal/platform/bootstrap`. No feature flags, no timeouts, no configurable constants beyond the hard-coded `defaultLimit = 20` and `maxLimit = 100` at `service.go:12-15`.

---

## 8. Concurrency & async

None. The module is purely synchronous and request-scoped. There are no goroutines, channels, timers, background workers, outbox writes, or job enqueues within the search module. All concurrency concerns (connection pooling, request isolation) are delegated to `database/sql`'s pool and the HTTP server's goroutine-per-request model.

---

## 9. Error handling & observability

### Error handling

| Location | Pattern | Notes |
|---|---|---|
| `handler.go:55-57` | `405 Method Not Allowed` via bare `w.WriteHeader` | No response body; does not use `httpresponse.WriteError` |
| `handler.go:59-67` | `401` via `httpresponse.WriteError` (→ RFC 9457 `application/problem+json`) | Auth guard |
| `handler.go:72-77` | `400` via `httpresponse.WriteError` (→ RFC 9457) | Invalid `limit` |
| `handler.go:79-88` | `400` via `httpresponse.WriteError` (→ RFC 9457) | Invalid `expiry_before` / `expiry_after` |
| `handler.go:108-111` | `500` via `httpresponse.WriteError` (→ RFC 9457) | Any service/reader error; no detail exposed |
| `service.go:36-39` | Returns `ErrTenantRequired`; no logging | Unauthenticated caller short-circuit |
| `reader.go:139-141` | `fmt.Errorf("v2 list documents: %w", err)` — wraps DB error | Query execution error |
| `reader.go:153-167` | `fmt.Errorf("v2 scan document: %w", err)` — wraps scan error | Row scan error |
| `reader.go:178-180` | `fmt.Errorf("v2 list documents rows: %w", err)` — wraps rows.Err | Post-iteration error |

Error wrapping is consistent in the reader. The 405 case at `handler.go:55-57` is a minor inconsistency (no RFC 9457 body), but it is a low-traffic path.

### RFC 9457 compliance

`httpresponse.WriteError` delegates to `platform/problem.Write` which sets `Content-Type: application/problem+json` and writes a `Problem` struct with `title`, `status`, and `code` fields (`problem.go:77-87`). This is the standard error path for all 4xx/5xx responses from the handler **except** the 405 (bare `WriteHeader` only).

### Observability

No structured logging calls (`slog`, `log`, or equivalent) are present in the search module's own code. No metrics instrumentation. No tracing spans. Error propagation relies on the caller (handler) mapping all errors to a generic `500`; no error detail is surfaced downstream. The platform-level request tracing middleware (wired at `main.go:602`) provides request-level trace IDs but the search module itself adds no spans.

---

## 10. Legacy / duplication / smell flags

- **SMELL-1 — camelCase query parameter `businessUnit` survives snake_case standardisation pass**
  - WHERE: `handler.go:99`
  - WHY: All other query parameters use snake_case (matching the refactor commit `16b72f81f` "snake_case query + path parameters"). `businessUnit` was not renamed. The OpenAPI spec does not declare this parameter at all (lines 802-882), so it is an undocumented, off-contract, camelCase survivor. Confirmed by git blame: the rename commit touched the handler file but `businessUnit` remained.

- **SMELL-2 — Four query parameters accepted by handler but silently no-op at data layer: `subject`, `businessUnit`, `classification`, `tag`**
  - WHERE: `handler.go:96-103`, `domain/model.go:72-86`, `reader.go:23-27`
  - WHY: `Query.Subject`, `Query.BusinessUnit`, `Query.Classification`, `Query.Tag` are populated from HTTP params and flow through the service to the reader, but the SQL query has no `$N` bindings for them. The reader comment at line 23-27 explicitly states these columns "live on the decommissioned `metaldocs.documents` schema". The fields remain in `Document` and `Query` structs but are dead cargo for the v2 reader. Tests in `handler_test.go` (`TestHandleSearchDocumentsMapsAdvertisedFilters`) assert these params are forwarded to the query struct, and `TestHandleSearchDocumentsResponseIncludesAdvertisedFields` asserts the response includes `subject`, `business_unit`, `classification`, `tags` — cementing the dead-filter behaviour in test expectations.

- **SMELL-3 — `normalized.Limit = q.Limit` is a dead assignment**
  - WHERE: `service.go:50`
  - WHY: Line 50 assigns `q.Limit` (the raw, unclamped value) back onto `normalized.Limit`. Line 52 then computes `effectiveLimit(q.Limit)` (the clamped value) and passes it as the `limit` argument to `ListDocuments`. The `normalized` query struct's `Limit` field is never read by the reader (the reader uses the explicit `limit` parameter, not `query.Limit`). The assignment is harmless but dead, introduced in commit `adf72c5d8` and not cleaned up.

- **SMELL-4 — Dual authn accessor: `iamdomain.UserIDFromContext` (handler) vs `authn.UserIDFromContext` (service)**
  - WHERE: `handler.go:59`, `service.go:33`
  - WHY: The handler calls `iamdomain.UserIDFromContext` directly to guard the route, while the service calls the platform wrapper `authn.UserIDFromContext` (which adds a trim+empty-check and returns a `(string, bool)` pair). The platform wrapper's doc comment (`platform/authn/context.go:14-19`) explicitly calls out that it exists to prevent 27 hand-rolled copies of this check. The handler pre-dates or bypasses the canonical accessor. The dual call also means the authn check runs twice per request — once in the handler (returns 401 if missing) and again in the service (returns empty slice if missing). No correctness bug, but it is a pattern inconsistency.

- **SMELL-5 — `document_profiles` family subquery duplicated inside a single SQL statement**
  - WHERE: `reader.go:34-41` (SELECT list) and `reader.go:59-66` (WHERE clause)
  - WHY: The correlated scalar subquery that resolves `family_code` from `metaldocs.document_profiles` is textually identical and executed twice per row: once to project the value, once to filter it. A lateral join or CTE would execute it once. At low row counts this is inconsequential; at scale it doubles the cross-schema subquery cost.

- **SMELL-6 — `DocumentType` assigned from `DocumentProfile` at scan time (aliasing)**
  - WHERE: `reader.go:172`
  - WHY: `doc.DocumentType = doc.DocumentProfile` makes `DocumentType` a duplicate of `DocumentProfile` in every result. The `Query` struct also accepts both `DocumentType` and `DocumentProfile` as separate inputs, and the reader's parameter-binding logic at lines 111-113 merges them (`DocumentType` takes precedence). Having two struct fields that always carry the same value after a scan, and two query fields that collapse to one SQL binding, signals that the distinction was never fully resolved after the taxonomy was introduced.

- **SMELL-7 — No pagination: offset is hardcoded to 0**
  - WHERE: `service.go:53`, `domain/port.go:9`, `reader.go:21`
  - WHY: The `Reader` interface accepts `offset int`, the SQL uses `$12`, and the service always passes `0`. The HTTP handler exposes only `limit`, not `offset` or a cursor. There is no pagination contract. For a search endpoint on a growing document corpus this is a functional gap, not merely cosmetic.

- **SMELL-8 — OpenAPI spec / handler response shape divergence: legacy fields in response struct absent from schema**
  - WHERE: `handler.go:24-43` (`SearchDocumentResponse`), `api/openapi/v1/openapi.yaml:4785-4820` (`SearchDocumentItem`)
  - WHY: The handler's response struct contains `Subject`, `BusinessUnit`, `Classification`, `Tags` (with JSON keys `subject`, `business_unit`, `classification`, `tags`). None of these fields appear in the `SearchDocumentItem` OpenAPI schema. Clients relying on the spec type-generation will not have these fields in their generated types. The fields are always zero-valued at runtime, so the divergence is doubly misleading: the spec omits them, and the implementation emits them as empty.

- **SMELL-9 — 405 response missing RFC 9457 body**
  - WHERE: `handler.go:55-57`
  - WHY: All other error paths in the handler use `httpresponse.WriteError` (→ `application/problem+json`). The method-not-allowed case writes only a status code with no body, inconsistent with the module's own error pattern and the project-wide RFC 9457 contract (AD-2).

---

## 11. Wiki drift

### `wiki/modules/search.md` (Last verified: 2026-06-05)

1. **Doc line 141 (`T-001`):** Tech-debt register states `writeAPIError` helper at `handler.go:141` as evidence of legacy error envelope. **Code reality:** `writeAPIError` does not exist in the current `handler.go`. The handler uses `httpresponse.WriteError` throughout (which is RFC 9457 compliant). The tech-debt item's symptom description is stale — the handler was already migrated to `httpresponse.WriteError` in the Phase D commit (`c4c4d95d2` "error-envelope unification"). T-001 should be re-evaluated: either it is already resolved (the envelope is now RFC 9457), or the concern has shifted to the missing 405 body (SMELL-9).

2. **Doc claims `Fields present on the Document struct (Subject, BusinessUnit, Classification, Tags) are populated as zero values by the reader; the struct is shared with other surfaces that may populate them via different paths.`** Code reality: this is accurate for the reader, but the handler test `TestHandleSearchDocumentsResponseIncludesAdvertisedFields` (handler_test.go:145-206) tests that these fields round-trip through the response. There is no other surface that populates them in the current codebase — the "shared with other surfaces" qualifier is aspirational, not currently true. Minor — not a breaking drift.

3. **Doc (`search.md:141`) references `handler.go:141`:** Current `handler.go` has only 159 lines; line 141 is `parseOptionalDateTimeQuery`. The line anchor in the tech-debt register points to the wrong location. Minor staleness.

### `wiki/modules/search-tech-debt.md` (Last verified: 2026-06-05)

4. **T-001 evidence (`writeAPIError` at `handler.go:141`):** As noted above, `writeAPIError` is absent from the current handler. If T-001 is intended to track the 405 missing-body issue (SMELL-9), the evidence reference needs updating. If T-001 is intended to track the legacy envelope, it should be marked closed.

---

## 12. Open questions

1. **[runtime-unverified]** Does the `document_profiles` cross-schema subquery (`metaldocs.document_profiles`) execute on the same Postgres connection pool as `public.*` queries? The `metaldocs` schema is the decommissioned legacy schema. If the search_path or grants differ between schemas in production, the double subquery may behave differently under load or after a future schema-ownership migration.

2. **[runtime-unverified]** What is the p99 latency of `ListDocuments` for the largest tenant? The query has four correlated EXISTS subqueries per row (two for restricted CD visibility), a leading-wildcard LIKE, and a double cross-schema subquery. Without index coverage on `user_process_areas(user_id, area_code)` and `controlled_document_user_grants(controlled_document_id, user_id)`, the query may degrade non-linearly. The wiki notes this risk (`search.md:71`) but there is no evidence of index verification.

3. **[runtime-unverified]** The integration test (`reader_visibility_integration_test.go`) requires a live Postgres instance and is gated behind the `integration` build tag. It is unclear whether this test runs in CI or only locally. If CI does not run it, the visibility predicate has no automated regression coverage beyond the sqlmock unit tests (which mock the query shape, not the predicate semantics).

4. **Genuine unknown** — Are `subject`, `businessUnit`, `classification`, and `tag` query parameters intended to be removed (dead cleanup), re-implemented against v2 columns, or kept as no-ops for backward compatibility with existing API callers? The wiki does not document them at all, and they are absent from the OpenAPI spec. Decision required before any cleanup.

5. **Genuine unknown** — The `Query.Limit` field exists on the domain struct and is set by the service before passing `normalized` to the reader, but the reader uses the explicit `limit` argument parameter, not `query.Limit`. Is `Query.Limit` intended to eventually be the sole limit carrier (simplifying the interface), or is the explicit parameter the canonical path? The dual representation adds confusion without benefit.
