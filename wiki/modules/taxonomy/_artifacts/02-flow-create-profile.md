# Phase 2 — Data-flow trace: POST /api/v2/taxonomy/profiles (createProfile)

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| OpenAPI op | n/a — no OpenAPI spec / no generated stub | — |
| Generated server stub | n/a — module uses raw `net/http` `ServeMux` | — |
| Handler | `(*Handler).createProfile` | `internal/modules/taxonomy/delivery/http/routes_profiles.go:53` |
| Route registration | `mux.HandleFunc("POST /api/v2/taxonomy/profiles", h.createProfile)` | `internal/modules/taxonomy/delivery/http/handler.go:52` |

## 2. Call chain

```
1. internal/modules/taxonomy/delivery/http/routes_profiles.go:53 (*Handler).createProfile
     decodes JSON, derives tenantID from header, builds domain.DocumentProfile
     → calls: internal/modules/taxonomy/delivery/http/routes_profiles.go:197 tenantIDFromRequest
2. internal/modules/taxonomy/delivery/http/routes_profiles.go:197 tenantIDFromRequest
     reads `X-Tenant-ID`; falls back to `tenant.DevTenantID` if empty
3. internal/modules/taxonomy/delivery/http/routes_profiles.go:85 h.profiles.Create(ctx, profile)
     → calls: internal/modules/taxonomy/application/profile_service.go:41 (*ProfileService).Create
4. internal/modules/taxonomy/application/profile_service.go:41 (*ProfileService).Create
     pass-through; NO validation, NO authz, NO governance log
     → calls: internal/modules/taxonomy/infrastructure/repository.go:102 (*ProfileRepository).Create
5. internal/modules/taxonomy/infrastructure/repository.go:102 (*ProfileRepository).Create
     INSERT INTO metaldocs.document_profiles via r.db.ExecContext (NO tx, NO authz GUC)
```

Tx boundary: **none**. `sql.DB.ExecContext` opens an implicit single-statement connection.

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `document_profiles` row | (no row) | row inserted, `archived_at = NULL` | this op | `taxonomy.manage` (tier-1 only) |

## 4. SQL touched

| File:line | Verb | Table | Tenant predicate | authz.Require before SQL? |
|---|---|---|---|---|
| `internal/modules/taxonomy/infrastructure/repository.go:104` | INSERT | `metaldocs.document_profiles` | none in WHERE; `tenant_id` is a column value from request-derived `p.TenantID` | **VIOLATION/N/A — no in-tx authz call**; `authz.Require(` not found under `internal/modules/taxonomy/` |

FK side-effect: `document_profiles.family_code REFERENCES document_families(code)` (`migrations/0023_init_document_family_and_profile_registry.sql:9-12`). Missing/inactive family → PG `23503` → handler maps to `409 FAMILY_NOT_FOUND`.

Upstream tier-1 capability gate: `apps/api/cmd/metaldocs-api/permissions.go:158-164` maps `POST /api/v2/taxonomy/profiles` → `iamdomain.CapTaxonomyManage` (`taxonomy.manage`). No tier-2 in-tx check exists.

Trust chain for `tenant_id`: client header `X-Tenant-ID` → `tenantIDFromRequest` → domain struct → SQL `INSERT`. No verification that the authenticated user belongs to the named tenant. Fallback to `tenant.DevTenantID` when header absent.

## 5. Response shape

- `201 Created` body: JSON encoding of `domain.DocumentProfile` (`domain/profile.go:8-21`). Fields: `code, tenantId, familyCode, name, description, alias, reviewIntervalDays, defaultTemplateVersionId, ownerUserId, editableByRole, archivedAt, createdAt`.
- Error envelope (non-RFC9457): `{"code": "...", "message": "..."}` via `internal/platform/httpresponse/response.go:14-16`.

| HTTP | Code | Source |
|---|---|---|
| 400 | VALIDATION_ERROR (invalid JSON) | `routes_profiles.go:55-57` |
| 400 | VALIDATION_ERROR (missing code) | `routes_profiles.go:80-82` |
| 400 | VALIDATION_ERROR (PG 23514) | `routes_profiles.go:187-188` |
| 400 | PROFILE_CODE_IMMUTABLE | `routes_profiles.go:185-186` |
| 404 | PROFILE_NOT_FOUND | `routes_profiles.go:177-179` |
| 409 | PROFILE_ARCHIVED | `routes_profiles.go:179-180` |
| 409 | TEMPLATE_NOT_PUBLISHED | `routes_profiles.go:181-182` |
| 409 | TEMPLATE_PROFILE_MISMATCH | `routes_profiles.go:183-184` |
| 409 | FAMILY_NOT_FOUND (PG 23503) | `routes_profiles.go:189-190` |
| 500 | INTERNAL_ERROR | `routes_profiles.go:191-193` |

## 6. Cross-references

- Idempotency: **no** — no `Idempotency-Key` handling. Duplicate POST with same code returns PG `23505` (no current mapping in `writeProfileError`; falls to `INTERNAL_ERROR 500`).
- Pagination: n/a (write).
- Audit emission: **no** for `createProfile`. `ProfileService.Create` does not call `govLogger.Log`. Sibling ops `SetDefaultTemplate` (`profile_service.go:77`) and `Archive` (`profile_service.go:98`) DO emit to `DBGovernanceLogger` → inserts into `governance_events` (legacy module-local sink; NOT `audit.Writer`).
