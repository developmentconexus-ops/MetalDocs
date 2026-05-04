# Module: taxonomy

> **Last verified:** 2026-05-02
> **Scope:** Document Families, Profiles (Tipos Documentais / Perfis Documentais), Areas, and their inter-relationships. Covers backend domain, infrastructure, service, HTTP delivery, and frontend admin UI.
> **Out of scope:** Code generation rules (see `concepts/controlled-documents.md`); approval routing (see `modules/approval.md`).
> **Key files:**
> - `internal/modules/taxonomy/domain/family.go:8` — DocumentFamily struct + sentinel errors + Deactivate()
> - `internal/modules/taxonomy/domain/port.go:34` — FamilyRepository interface
> - `internal/modules/taxonomy/infrastructure/family_repository.go:11` — FamilyRepository SQL impl
> - `internal/modules/taxonomy/application/family_service.go:11` — FamilyService (Create, Get, List, Update, Deactivate)
> - `internal/modules/taxonomy/delivery/http/routes_families.go:20` — HTTP handlers (list, create, get, update, deactivate)
> - `internal/modules/taxonomy/delivery/http/handler.go:28` — familyService interface + Handler struct + route registration
> - `internal/modules/taxonomy/module.go:26` — wire familyRepo → familyService → handler
> - `migrations/0161_grant_families_write_privileges.sql` — GRANT on document_families to metaldocs_app
> - `frontend/apps/web/src/features/taxonomy/TaxonomyAdminPage.tsx:10` — admin UI, Famílias tab is default
> - `frontend/apps/web/src/features/taxonomy/FamilyList.tsx:13` — family table + deactivate action
> - `frontend/apps/web/src/features/taxonomy/FamilyEditDialog.tsx:12` — create/edit modal
> - `frontend/apps/web/src/features/taxonomy/ProfileEditDialog.tsx:29` — family dropdown (now dynamic via fetchFamilies)
> - `frontend/apps/web/src/features/taxonomy/types.ts:63` — DocumentFamily, CreateFamilyRequest, UpdateFamilyRequest
> - `frontend/apps/web/src/features/taxonomy/api.ts:98` — fetchFamilies, createFamily, updateFamily, deactivateFamily

---

## Entities

### DocumentFamily (globally scoped)

Groups profiles into broad document categories (e.g. "Qualidade", "Recursos Humanos").

**Key invariants:**

- `document_families` has **no `tenant_id`** — families are global, not per-tenant.
- Deactivation uses `is_active BOOLEAN` (not `archived_at TIMESTAMPTZ` as profiles/areas use).
- Family `code` is **immutable** after creation — enforced by `ErrFamilyCodeImmutable` and service logic.
- Deactivation is **guarded**: blocked if any active profiles reference the family (`HasActiveProfiles` check returns `ErrFamilyHasProfiles`).
- Sentinel errors live in `domain/family.go:16`: `ErrFamilyNotFound`, `ErrFamilyAlreadyInactive`, `ErrFamilyCodeImmutable`, `ErrFamilyHasProfiles`.

**DB table:** `metaldocs.document_families` — columns: `code`, `name`, `description`, `is_active`, `created_at`.

**Domain method:**

```go
// internal/modules/taxonomy/domain/family.go:23
func (f *DocumentFamily) Deactivate() error {
    if !f.IsActive {
        return ErrFamilyAlreadyInactive
    }
    f.IsActive = false
    return nil
}
```

### Profile (Tipo Documental / Perfil Documental)

Document category within a family. Tenant-scoped (`tenant_id`). Has a code prefix (e.g. `DC`, `POP`) that drives `{doc_code}`. Archived via `archived_at TIMESTAMPTZ` (not `is_active`).

**Default template binding:** Each profile may have a `default_template_version_id`. The document creation wizard uses this to clone content. Without the binding, the wizard has no template to offer.

**FK:** `document_profiles.family_code REFERENCES document_families(code)` — a profile cannot be created without a valid, active family.

### Area (ProcessArea)

Organizational unit. Tenant-scoped. Code (e.g. `RH`, `QUA`, `PROD`) becomes part of the controlled-document number. Archived via `archived_at TIMESTAMPTZ`.

---

## Scoping comparison

| Entity | Scoped by | Inactive pattern |
|--------|-----------|-----------------|
| DocumentFamily | global (no tenant_id) | `is_active = FALSE` |
| DocumentProfile | tenant_id | `archived_at IS NOT NULL` |
| ProcessArea | tenant_id | `archived_at IS NOT NULL` |

---

## API routes

All routes under `/api/v2/taxonomy/`. Registration: `internal/modules/taxonomy/delivery/http/handler.go:50`.

### Families

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/families` | `listFamilies` | `?includeInactive=true` to include deactivated |
| POST | `/families` | `createFamily` | code + name required; code immutable |
| GET | `/families/{code}` | `getFamily` | 404 → `FAMILY_NOT_FOUND` |
| PATCH | `/families/{code}` | `updateFamily` | name + description only; code ignored from body |
| DELETE | `/families/{code}` | `deactivateFamily` | 204 on success; 409 if has active profiles |

**Error codes** (`routes_families.go:97`): `FAMILY_NOT_FOUND` (404), `FAMILY_ALREADY_INACTIVE` (409), `FAMILY_HAS_PROFILES` (409), `FAMILY_ALREADY_EXISTS` (409 — PG unique violation), `VALIDATION_ERROR` (400).

Note: DELETE is used for deactivation (not POST /archive) because families have no audit-trail requirement.

### Profiles

| Method | Path | Notes |
|--------|------|-------|
| GET | `/profiles` | `?includeArchived=true` |
| POST | `/profiles` | requires `familyCode` |
| GET | `/profiles/{code}` | |
| PATCH | `/profiles/{code}` | |
| DELETE | `/profiles/{code}` | archives |
| PUT | `/profiles/{code}/default-template` | binds template version |

### Areas

| Method | Path | Notes |
|--------|------|-------|
| GET | `/areas` | `?includeArchived=true` |
| POST | `/areas` | |
| GET | `/areas/{code}` | |
| PUT | `/areas/{code}` | full update |
| DELETE | `/areas/{code}` | archives |

---

## Wiring (module.go)

```go
// internal/modules/taxonomy/module.go:26
familyRepo := infrastructure.NewFamilyRepository(deps.DB)
familyService := application.NewFamilyService(familyRepo)
handler := thttp.NewHandler(profileService, areaService, familyService)
```

---

## Frontend admin UI

`TaxonomyAdminPage.tsx` — three tabs: **Famílias** (default), **Perfis**, **Áreas**. All three are loaded in a single `Promise.all` on mount and on filter change.

`FamilyList.tsx` — table with columns: code, name, description, status (Ativa/Inativa), actions (Editar / Desativar). The Desativar button is hidden for already-inactive rows.

`FamilyEditDialog.tsx` — modal for create (code editable) and edit (code read-only). Create enforces code to lowercase.

`ProfileEditDialog.tsx` — family selector (`<select>`) is populated via `fetchFamilies()` (active families only, no `includeInactive` param). Previously hardcoded; changed to dynamic in this feature.

---

## DB migration

`migrations/0161_grant_families_write_privileges.sql` — grants `SELECT, INSERT, UPDATE, DELETE` on `metaldocs.document_families` to the `metaldocs_app` runtime user. Without this migration the API returns Postgres permission errors on all family endpoints.

---

## Invariants summary

1. Family code is immutable after creation.
2. A profile requires a valid family (`family_code` FK).
3. A family with active profiles cannot be deactivated (`HasActiveProfiles` guard in `FamilyService.Deactivate`).
4. Families are globally shared — all tenants see the same set.
5. Profile/area archiving uses `archived_at`; family deactivation uses `is_active`.

---

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Steps 1 + 4
- [concepts/controlled-documents.md](../concepts/controlled-documents.md) — code generation
- [modules/templates-v2.md](templates-v2.md)
- [architecture/data-model.md](../architecture/data-model.md) — document_families table entry
