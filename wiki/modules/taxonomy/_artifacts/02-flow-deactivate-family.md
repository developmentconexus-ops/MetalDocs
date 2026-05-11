# Phase 2 — Data-flow trace: DELETE /api/v2/taxonomy/families/{code} (deactivateFamily)

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| Route registration | `mux.HandleFunc("DELETE /api/v2/taxonomy/families/{code}", h.deactivateFamily)` | `internal/modules/taxonomy/delivery/http/handler.go:68` |
| Handler | `(*Handler).deactivateFamily` | `internal/modules/taxonomy/delivery/http/routes_families.go:90-96` |
| OpenAPI op / generated stub | n/a — no OpenAPI spec | — |

## 2. Call chain

```
1. routes_families.go:90 (*Handler).deactivateFamily
     reads path param "code"; on error maps via writeFamilyError; on success 204
     → calls: family_service.go:48 (*FamilyService).Deactivate
2. family_service.go:48 (*FamilyService).Deactivate
     orchestrates: load → guard → mutate → persist (NO tx)
     → calls: family_repository.go:19 (*FamilyRepository).GetByCode
3. family_service.go:53 (*FamilyService).Deactivate
     → calls: family_repository.go:91 (*FamilyRepository).HasActiveProfiles
4. family_service.go:60 (*FamilyService).Deactivate
     → calls: domain/family.go:22 (*DocumentFamily).Deactivate (in-memory state transition)
5. family_service.go:63 (*FamilyService).Deactivate
     → calls: family_repository.go:75 (*FamilyRepository).Update (UPDATE document_families)
```

Tx boundary: **none**. `FamilyRepository` holds `*sql.DB` (`family_repository.go:11-13`); three discrete connections used (SELECT family + EXISTS profiles + UPDATE family).

## 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `DocumentFamily.IsActive` (`document_families.is_active`) | `TRUE` | `FALSE` | `(*DocumentFamily).Deactivate` (`domain/family.go:22-27`) + `Update` SQL | `taxonomy.manage` (tier-1 only, via `apps/api/cmd/metaldocs-api/permissions.go:174-180`); constant at `internal/modules/iam/domain/capabilities.go:16` |

## 4. SQL touched

| File:line | Verb | Table | Tenant predicate | authz.Require before SQL? |
|---|---|---|---|---|
| `family_repository.go:20-27` | SELECT | `metaldocs.document_families` | `WHERE code = $1` (table has no tenant_id — global) | **VIOLATION/N/A — no in-tx authz call** |
| `family_repository.go:92-99` | SELECT EXISTS | `metaldocs.document_profiles` | `WHERE family_code = $1 AND archived_at IS NULL`; **no tenant_id predicate** — counts profiles across ALL tenants | **VIOLATION/N/A — no in-tx authz call** |
| `family_repository.go:76-80` | UPDATE | `metaldocs.document_families` | `WHERE code = $4`; no tenant_id (global) | **VIOLATION/N/A — no in-tx authz call** |

**Race window (TOCTOU).** `HasActiveProfiles` check + `Update` are separate `sql.DB` calls (not wrapped in tx, no row lock). Concurrent INSERT into `document_profiles` referencing this family can succeed between the EXISTS check and the UPDATE, leaving an inactive family with active profiles → orphaned-FK state (FK still valid since family row still exists; semantic invariant violated).

**Cross-tenant blast radius.** `HasActiveProfiles` checks across every tenant. A `qms_admin` (per migration 0169) or `system_admin` (per migration 0165) in tenant A can deactivate a globally-shared family in use by tenant B's profiles only if tenant B has no profiles — but family deactivation itself affects every tenant's UI/registry.

## 5. Response shape

- `204 No Content` on success (`routes_families.go:95`).
- Error envelope (non-RFC9457): `{"code": "...", "message": "..."}` (`internal/platform/httpresponse/response.go:14-16`).

| HTTP | Code | Source |
|---|---|---|
| 404 | FAMILY_NOT_FOUND | `routes_families.go:101-103` |
| 409 | FAMILY_ALREADY_INACTIVE | `routes_families.go:103-105` |
| 409 | FAMILY_HAS_PROFILES | `routes_families.go:105-107` |
| 500 | INTERNAL_ERROR | `routes_families.go:111-114` |

## 6. Cross-references

- Idempotency: **no**. Already-inactive family returns `409 FAMILY_ALREADY_INACTIVE`. Re-issuing DELETE is not safe.
- Pagination: n/a.
- Audit emission: **no**. `FamilyService` has no `govLogger` dependency (`family_service.go:11-13`). Module wiring (`module.go:30`) creates `FamilyService` without the logger that `ProfileService`/`AreaService` receive. Mutations on globally-shared catalog are unobserved by any audit sink.
