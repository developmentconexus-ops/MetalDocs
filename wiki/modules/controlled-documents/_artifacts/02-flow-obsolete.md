# Phase 2 — Flow trace: Obsolete / Supersede

Lifecycle state-transitions on `controlled_documents`. Both ops share `changeStatus` plumbing.

### 1. Entry point

| Op | Method | Route | Handler | File:line |
|---|---|---|---|---|
| Obsolete | PUT | `/api/v1/controlled-documents/{id}/obsolete` | `(*Handler).ObsoleteControlledDocument` | `internal/modules/registry/delivery/http/handler.go:87`, `routes.go:328` |
| Supersede | PUT | `/api/v1/controlled-documents/{id}/supersede` | `(*Handler).SupersedeControlledDocument` | `internal/modules/registry/delivery/http/handler.go:88`, `routes.go:337` |

### 2. Call chain

**Obsolete**
1. OpenAPI wrapper binds path `id` UUID; calls handler — `api.gen.go:1018-1033`.
2. Handler delegates: `h.svc.Obsolete(...)` (`routes.go:328-331`).
3. Service maps to `changeStatus(..., CDStatusObsolete)` (`application/service.go:293-295`).
4. `changeStatus` loads doc, enforces active guard via `doc.IsActive()`, returns `ErrCDNotActive` if false (`service.go:309-316`; predicate `domain/controlled_document.go:44-46`, sentinel `:39`).
5. Repository mutation: `docs.UpdateStatus(...)` (`service.go:317`) → SQL UPDATE (`infrastructure/repository.go:184-197`).

**Supersede**
1. OpenAPI wrapper binds path `id` UUID; calls handler — `api.gen.go:1098-1113`.
2. Handler delegates: `h.svc.Supersede(...)` (`routes.go:337-340`).
3. Service maps to `changeStatus(..., CDStatusSuperseded)` (`service.go:297-299`).
4. Same `changeStatus` guard (same file:line).
5. Same repository mutation (`UpdateStatus`).

### 3. State changes

| Op | Entity | From | To | Trigger | Guard | Capability required |
|---|---|---|---|---|---|---|
| Obsolete | ControlledDocument | `active` only | `obsolete` | `Obsolete` op via `changeStatus` | `if !doc.IsActive() { ErrCDNotActive }` (`service.go:314-316`) | (unclear: not enforced in registry path; resolver mapping outside module) |
| Supersede | ControlledDocument | `active` only | `superseded` | `Supersede` op via `changeStatus` | same guard | (unclear: same) |

Status string constants: `domain/controlled_document.go:13-15`.

### 4. SQL touched

```sql
UPDATE controlled_documents
   SET status = $1, updated_at = $2
 WHERE tenant_id = $3 AND id = $4
```
(`infrastructure/repository.go:186`)

**Tripwire pairing: VIOLATION.** No `authz.Require` call anywhere in `routes.go:328-343`, `service.go:293-317`, `repository.go:184-197`, `handler.go:87-88`. No `metaldocs.tenant_id` GUC set before UPDATE — tenant scoped via `$3` arg only.

Capability names `registry.obsolete` / `registry.supersede` not found in `internal/modules/registry/` path. (unclear: whether single `registry.create` cap covers all mutations, or a distinct cap exists in resolver elsewhere).

### 5. Response shape

**Obsolete**
- 2xx: `204 No Content` (`routes.go:334`)
- Errors:
  - `400 VALIDATION_ERROR` — UUID binding failure in wrapper (`api.gen.go:1026-1029`; handler error writer `handler.go:74-76`)
  - `404 CONTROLLED_DOCUMENT_NOT_FOUND` — `ErrCDNotFound` mapping (`repository.go:194`; `routes.go:412-414`)
  - `409 CONTROLLED_DOCUMENT_NOT_ACTIVE` — `ErrCDNotActive` mapping (`service.go:314-316`; `routes.go:414-416`)

**Supersede**
- 2xx: `204 No Content` (`routes.go:343`)
- Errors: identical mapping (same handler tail)

**Envelope:** `{"code":"...","message":"..."}` (`httpresponse/response.go:14-16`). Same legacy envelope as the rest of the module — not RFC 9457.

### 6. Cross-references

- **Idempotency:** no — middleware attached only to POST create/revision (`handler.go:80-82`), not the PUT lifecycle routes (`:87-88`). Replayed PUTs may re-emit 204 silently or return 409 if already transitioned (consequences not captured in this trace).
- **Audit emission:** **NO** for obsolete/supersede. `changeStatus` performs get + guard + update only; no `s.govLogger.Log(...)` call (`service.go:309-317`). Gap relative to create path which does emit (`service.go:267-271`).
- Only callers of `changeStatus`: `Obsolete` and `Supersede` (`service.go:293-299`).
