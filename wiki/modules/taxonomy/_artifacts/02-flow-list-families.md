# Phase 2 — Data-flow trace: GET /api/v1/taxonomy/families (listFamilies)

## 1. Entry point

| Layer | Symbol | File:line |
|---|---|---|
| API bootstrap | `taxonomyModule.RegisterRoutes(mux)` | `apps/api/cmd/metaldocs-api/main.go:201` |
| Module wiring | `(*Module).RegisterRoutes` | `internal/modules/taxonomy/module.go:36` |
| Route registration (ServeMux) | `mux.HandleFunc("GET /api/v1/taxonomy/families", h.listFamilies)` | `internal/modules/taxonomy/delivery/http/handler.go:64` |
| Handler | `(*Handler).listFamilies` | `internal/modules/taxonomy/delivery/http/routes_families.go:20` |
| OpenAPI op / generated stub | n/a — no OpenAPI spec / no generated stub | — |

## 2. Call chain

```
1. internal/modules/taxonomy/delivery/http/routes_families.go:20 (*Handler).listFamilies
     parses includeInactive query, writes {"items": ...} JSON
     → calls: internal/modules/taxonomy/application/family_service.go:19 (*FamilyService).List
2. internal/modules/taxonomy/application/family_service.go:19 (*FamilyService).List
     pass-through; NO tenant arg, NO authz
     → calls: internal/modules/taxonomy/infrastructure/family_repository.go:38 (*FamilyRepository).List
3. internal/modules/taxonomy/infrastructure/family_repository.go:38 (*FamilyRepository).List
     SELECT code, name, description, is_active, created_at FROM metaldocs.document_families [WHERE is_active=TRUE] ORDER BY code ASC
     → calls: database/sql.QueryContext
```

Tx boundary: none.

## 3. State changes

`none` — read-only.

## 4. SQL touched

| File:line | Verb | Table | Tenant predicate | authz.Require before SQL? |
|---|---|---|---|---|
| `internal/modules/taxonomy/infrastructure/family_repository.go:39-45` | SELECT | `metaldocs.document_families` | **none** — `document_families` has no `tenant_id` column (global catalog) | **VIOLATION/N/A — no in-tx authz call**; module has no `authz.Require` usage |

Upstream tier-1 capability gate: `apps/api/cmd/metaldocs-api/permissions.go:174-180` — `GET /api/v1/taxonomy/families*` → `iamdomain.CapDocView`. Any authenticated user with `doc.view` lists the global family catalog regardless of tenant.

## 5. Response shape

- `200 OK`: `{"items": [DocumentFamily, ...]}` via `writeJSON(w, http.StatusOK, ...)` (`routes_families.go:31`).
- Element: `domain.DocumentFamily` (`domain/family.go:8-14`) — `code, name, description, isActive, createdAt`.
- No `next_cursor` / pagination fields.
- Error envelope (non-RFC9457): `{"code": "...", "message": "..."}` (`internal/platform/httpresponse/response.go:14-16`).

| HTTP | Code | Source |
|---|---|---|
| 400 | VALIDATION_ERROR (bad `includeInactive`) | `routes_families.go:21-24` |
| 500 | INTERNAL_ERROR | `routes_families.go:26-29` |

## 6. Cross-references

- Idempotency: yes (GET, idempotent by HTTP).
- Pagination: **no** — full ordered list. Risk grows with family-count (no current cap; expected to remain small).
- Audit emission: **no** — `FamilyService.List` does not call govLogger nor `audit.Writer`. (Listing reads not typically audited, but governance log is also absent on Create/Update/Deactivate — see flow-deactivate-family artifact.)
