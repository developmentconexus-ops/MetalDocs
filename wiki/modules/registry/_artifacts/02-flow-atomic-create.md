# Phase 2 — Flow trace: AtomicCreateControlledDocument

`POST /api/v1/controlled-documents`

### 1. Entry point

| OpenAPI op | Generated server stub | Handler |
|---|---|---|
| `POST /api/v1/controlled-documents` `operationId: atomicCreateControlledDocument` — `api/openapi/v1/partials/registry.yaml:2`, `api/openapi/v1/partials/registry.yaml:51` | Route registration to wrapper: `internal/modules/registry/api/api.gen.go:1244`; wrapper method: `internal/modules/registry/api/api.gen.go:875`; interface method: `internal/modules/registry/api/api.gen.go:720` | `internal/modules/registry/delivery/http/routes.go:43` (`AtomicCreateControlledDocument`) |

Note: live OpenAPI spec is under `api/openapi/v1/partials/registry.yaml` even though the HTTP path is `/api/v1/...` — there is no `v2` openapi tree.

### 2. Call chain (16 layers)

1. `internal/modules/registry/api/api.gen.go:1244` route map — generated POST route binds to wrapper.
   → calls: `internal/modules/registry/api/api.gen.go:875` `ServerInterfaceWrapper.AtomicCreateControlledDocument`
2. `internal/modules/registry/delivery/http/handler.go:79` mux registration — wraps generated handler with tenant + idempotency middleware.
   → calls: `internal/modules/registry/delivery/http/handler.go:80` `idempotency.Require(...)(generated.AtomicCreateControlledDocument)`
3. `internal/platform/idempotency/middleware.go:22` `Require` — validates `Idempotency-Key`, hashes body, checks replay/conflict, records 2xx replay.
   → calls: `internal/platform/idempotency/postgres_store.go:36` `Store.CheckReplay`; then `internal/modules/registry/api/api.gen.go:875`; then `internal/platform/idempotency/postgres_store.go:67` `Store.RecordReplay`
4. `internal/modules/registry/delivery/http/handler.go:49` `injectTenant` — resolves tenant from `X-Tenant-ID` (fallback dev tenant) into context.
   → calls: `internal/modules/registry/delivery/http/handler.go:55` `context.WithValue(...)`
5. `internal/modules/registry/api/api.gen.go:875` generated wrapper — parses required header `Idempotency-Key` and forwards to handler.
   → calls: `internal/modules/registry/api/api.gen.go:909` `siw.Handler.AtomicCreateControlledDocument(...)`
6. `internal/modules/registry/delivery/http/routes.go:43` handler — decodes/validates request and builds `CreateControlledDocumentCmd`.
   → calls: `internal/modules/registry/delivery/http/routes.go:61` `h.svc.Create(...)`
7. `internal/modules/registry/application/service.go:104` `RegistryService.Create` — validates profile/area, selects code path, coordinates atomic create.
   → calls: `internal/modules/registry/infrastructure/repository.go:303` `TaxonomyProfileReader.GetByCode`; `internal/modules/registry/infrastructure/repository.go:343` `TaxonomyAreaReader.GetByCode`
8. `internal/modules/registry/application/service.go:153` transaction start for auto-code path; `setAuthzGUC` primes `metaldocs.tenant_id`/`metaldocs.actor_id`, then `authz.Require(registry.create, tenant)` appends the asserted cap before guarded writes.
   → calls: `internal/modules/registry/application/service.go:153` `s.db.BeginTx(...)`
9. `internal/modules/registry/application/service.go:164` sequence allocation.
   → calls: `internal/modules/registry/infrastructure/repository.go:239` `PostgresSequenceAllocator.NextAndIncrement(...)`
10. `internal/modules/registry/infrastructure/repository.go:239` `NextAndIncrement` — ensures counter row then `UPDATE ... RETURNING`.
    → calls: `internal/modules/registry/infrastructure/repository.go:245` `ensureCounterViaExec`; `internal/modules/registry/infrastructure/repository.go:250` `exec.QueryRowContext(UPDATE cd_sequence_counters...)`
11. `internal/modules/registry/application/service.go:243` controlled-document insert in tx.
    → calls: `internal/modules/registry/infrastructure/repository.go:137` `PostgresControlledDocumentRepository.CreateTx(...)`
12. `internal/modules/registry/application/service.go:247` document initializer port call (cross-module).
    → calls: `internal/modules/registry/domain/document_initializer.go:31` `DocumentInitializer.CloneTemplate(...)` (interface), implemented by `internal/modules/documents/application/cd_initializer.go:25`
13. `internal/modules/documents/application/cd_initializer.go:25` `CDDocumentInitializer.CloneTemplate` — adapts to documents service tx flow.
    → calls: `internal/modules/documents/application/cd_initializer.go:35` `svc.cloneIntoTx(...)`
14. `internal/modules/documents/application/service.go:394` `cloneIntoTx` — resolves template and creates document rows in same tx.
    → calls: `internal/modules/documents/application/service.go:476` `repo.CreateDocumentTx(...)`
15. `internal/modules/documents/repository/repository.go:73` `CreateDocumentTx` — inserts `documents`, `editor_sessions`, `document_revisions`, updates pointers/snapshots/placeholders.
    → returns to registry tx
16. `internal/modules/registry/application/service.go:257` commit; then post-commit governance emit loop `internal/modules/registry/application/service.go:268` via logger wired at `internal/modules/registry/module.go:31` (`taxonomyapp.NewDBGovernanceLogger`) and executed at `internal/modules/taxonomy/application/governance_logger.go:18`.

### 3. State changes

| Entity | From | To | Trigger | Capability required |
|---|---|---|---|---|
| `metaldocs.idempotency_keys` | no completed row for `(tenant, actor, route, key)` | completed replay row inserted/updated | `RecordReplay` after 2xx (`internal/platform/idempotency/middleware.go:67`, `postgres_store.go:67`) | `CapRegistryCreate` resolved at `apps/api/cmd/metaldocs-api/permissions.go:186-187`, enforced via IAM middleware (`main.go:173-174`, `:386`) |
| `cd_sequence_counters.next_seq` | `N` (or missing row) | `N+1` (and row ensured) | `NextAndIncrement` (`repository.go:214`, `:251`) | `registry.create` asserted in tx |
| `controlled_documents` | no row | new active row | `CreateTx` (`repository.go:146`) | `registry.create` asserted in tx |
| `documents` | no row | new draft row + pointer/snapshot updates | `CreateDocumentTx` (`documents/repository/repository.go:101`, `:135`, `:145`) | `document.create` for INSERT; `document.edit` before guarded document UPDATEs |
| `editor_sessions` | none | active session inserted | `CreateDocumentTx` (`:112`, `:128`) | same |
| `document_revisions` | none | initial revision inserted (template passthrough key) | `CreateDocumentTx` (`:120`) | same |
| `document_placeholder_values` | missing required rows | seeded placeholder rows (idempotent insert) | `CreateDocumentTx` (`:165`) | same |
| `governance_events` | no event | `numbering.override`/`template.override` rows (when applicable) | post-commit `govLogger.Log` (`service.go:268`, `taxonomy/governance_logger.go:25`) | same |

### 4. SQL touched

| File:line | Verb | Tables | Auth-area arg |
|---|---|---|---|
| `internal/platform/idempotency/postgres_store.go:42` | SELECT | `metaldocs.idempotency_keys` | route template scoped in store (`handler.go:41`) |
| `internal/modules/registry/infrastructure/repository.go:303` | SELECT | `metaldocs.document_profiles` | none in SQL; capability guard is upstream IAM middleware |
| `internal/modules/registry/infrastructure/repository.go:343` | SELECT | `metaldocs.document_process_areas` | same |
| `internal/modules/registry/infrastructure/repository.go:214` | INSERT ... ON CONFLICT DO NOTHING | `cd_sequence_counters` | same |
| `internal/modules/registry/infrastructure/repository.go:251` | UPDATE ... RETURNING | `cd_sequence_counters` | same |
| `internal/modules/registry/infrastructure/repository.go:47` | SELECT EXISTS | `controlled_documents` | same |
| `internal/modules/registry/infrastructure/repository.go:146` | INSERT ... RETURNING | `controlled_documents` | same |
| `internal/modules/documents/repository/repository.go:89` | SELECT | `metaldocs.iam_users` | same |
| `internal/modules/documents/repository/repository.go:95` | SELECT | `metaldocs.document_process_areas` | same |
| `internal/modules/documents/repository/repository.go:101` | INSERT ... RETURNING | `documents` | same |
| `internal/modules/documents/repository/repository.go:112` | INSERT ... RETURNING | `editor_sessions` | same |
| `internal/modules/documents/repository/repository.go:120` | INSERT ... RETURNING | `document_revisions` | same |
| `internal/modules/documents/repository/repository.go:128` | UPDATE | `editor_sessions` | same |
| `internal/modules/documents/repository/repository.go:135` | UPDATE | `documents` | same |
| `internal/modules/documents/repository/repository.go:145` | UPDATE | `public.documents` | same |
| `internal/modules/documents/repository/repository.go:165` | INSERT ... ON CONFLICT DO NOTHING | `public.document_placeholder_values` | same |
| `internal/modules/taxonomy/application/governance_logger.go:25` | INSERT | `governance_events` | same |
| `internal/platform/idempotency/postgres_store.go:69` | INSERT ... ON CONFLICT DO UPDATE | `metaldocs.idempotency_keys` | same |

**Tripwire pairing audit:** active for atomic create. HTTP tier-1 still resolves `registry.create` via IAM middleware + permission resolver (`apps/api/cmd/metaldocs-api/main.go:173-174`, `:386`, `permissions.go:186-187`). `RegistryService.Create` now sets authz GUCs and calls `authz.Require(registry.create, tenant)` before `cd_sequence_counters` and `controlled_documents` writes. Documents-side `CreateDocumentTx` asserts `document.create` for the INSERT and `document.edit` before guarded `documents` UPDATEs.

### 5. Response shape

- **2xx (201):** `AtomicCreateResponse { controlledDocument, document }` (spec `api/openapi/v1/partials/registry.yaml:247-254`; generated `api.gen.go:278-280`; handler writes 201 at `routes.go:81-84`).
- **400:**
  - `VALIDATION_ERROR` from route decode/required checks (`routes.go:46`, `:50`).
  - `IDEMPOTENCY_KEY_REQUIRED` / `IDEMPOTENCY_KEY_INVALID` from idempotency middleware (`middleware.go:27`, `:31`).
- **401:** declared in spec (`registry.yaml:71`); emitted by auth/iam middleware ((unclear: exact code string)).
- **409 domain conflicts** via `writeDomainError` (`routes.go:417..441`): `CONTROLLED_DOCUMENT_CODE_TAKEN`, `CONTROLLED_DOCUMENT_CODE_ARCHIVED`, `OVERRIDE_TEMPLATE_DELETED`, `OVERRIDE_TEMPLATE_NOT_PUBLISHED`, `TEMPLATE_PROFILE_MISMATCH`, `PROFILE_NO_DEFAULT_TEMPLATE`, `DEFAULT_TEMPLATE_OBSOLETE`, `PROFILE_ARCHIVED`, `AREA_ARCHIVED`.
- **422:** `IDEMPOTENCY_KEY_CONFLICT` (`middleware.go:50`). Spec declares `422 template_invalid` (`registry.yaml:73`) but handler has no template-invalid 422 branch (`routes.go:410-444`) — spec/handler drift.

**Envelope:** legacy `{"code":"...","message":"..."}` (`internal/platform/httpresponse/response.go:14-15`; `idempotency/middleware.go:101`). RFC 9457 Problem Details NOT used.

### 6. Cross-references

- **Idempotency:** yes — store at `internal/platform/idempotency/postgres_store.go:19`; middleware at `middleware.go:22`. Route-scoped wiring: `handler.go:41`, applied at `handler.go:80`.
- **Pagination:** n/a.
- **Audit emission:** yes — registry service accumulates governance events and emits post-commit (`service.go:124`, `:257`, `:268`). Logger: `taxonomyapp.NewDBGovernanceLogger(deps.DB)` (`module.go:31`). DB insert at `taxonomy/governance_logger.go:25`.
