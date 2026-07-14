# Stage-1 Audit Artifact — module-controlled-documents

> **Written:** 2026-06-10
> **Model:** claude-sonnet-4-6
> **Scope:** `internal/modules/controlleddocuments` — code read as-is on branch `qa/iam-area-membership`.
> **Read-only.** No source changes made.
> **Existing wiki docs checked for drift:** `wiki/modules/controlled-documents.md` (last verified 2026-06-08), `wiki/modules/controlled-documents-tech-debt.md` (last verified 2026-06-08).

---

## 1. Identity & purpose

The **controlled-documents** module owns the catalog of ISO 9001 / QMS-grade Controlled Documents (CDs). Each row in `public.controlled_documents` is a numbered slot binding a `(profile_code, process_area_code)` pair to one or more `documents` revisions. The module carries no document content — it owns identity, numbering, and lifecycle status (`active | obsolete | superseded`).

Its core responsibilities are:
1. Atomic, idempotency-safe creation of a CD slot and its first document revision inside a single `*sql.Tx` (ADR 0011).
2. Monotonic, per-`(tenant, profile, area)` sequence allocation that produces human-readable codes (`DC-RH-001`).
3. Lifecycle transitions (`obsolete`, `supersede`) with tier-2/3 authz and governance-event emission.
4. Visibility-scoped read access — company-wide or restricted to explicit area/user grant lists.
5. Serving an active-document lookup (`GET .../active-document`) used by the Documents flow to surface the current working or latest-published revision.

---

## 2. File inventory

### Root

| File | Role |
|---|---|
| `module.go` | DI entry point; constructs all infra, service, and handler instances; exposes `New`, `RegisterRoutes`, `Service`. 55 lines. |

### `api/`

| File | Role |
|---|---|
| `api/cfg.yaml` | oapi-codegen configuration: package `controlleddocumentsapi`, `std-http-server`, `strict-server`, `embedded-spec`, include-tags `controlled-documents`. |
| `api/gen.go` | `//go:generate` directive pointing at the OpenAPI spec. |
| `api/api.gen.go` | Generated server stubs, models, and embedded spec for the `controlled-documents` OpenAPI tag. |

### `application/`

| File | Role |
|---|---|
| `application/service.go` | Application-layer use-case orchestrator: `Create`, `CreateRevision`, `Obsolete`, `Supersede`, `List`, `Get`, `PreviewCode`, `PeekSeq`, `changeStatus`. 656 lines. |
| `application/integration_test.go` | Integration test stubs (all `t.Skip("requires live DB")` except `TestCrossProfileOverride_Rejected` which tests domain logic inline). Build tag `integration`. |
| `application/service_test.go` | Unit tests for service: 25+ cases covering manual code, auto code, override template, archive guard, actor-missing, revision conflict, template-artifact guard. Uses `go-sqlmock`. |
| `application/tenant_isolation_test.go` | Tenant isolation integration stubs; all skipped. Build tag `integration`. |

### `delivery/http/`

| File | Role |
|---|---|
| `delivery/http/handler.go` | HTTP handler struct, route registration via `controlleddocumentsapi.HandlerWithOptions`, idempotency middleware wiring, `injectTenant` middleware. 123 lines. |
| `delivery/http/routes.go` | Strict-server implementations for all 8 operations, request/response mapping, domain-error translation in `writeDomainError`. 658 lines. |
| `delivery/http/routes_contract_test.go` | In-process contract tests; exact coverage not read but confirms request/response shape. |

### `domain/`

| File | Role |
|---|---|
| `domain/controlled_document.go` | `ControlledDocument` entity, `CDStatus` enum, sentinel errors, `NewControlledDocument` constructor with validation, `AutoCode`, `IsActive`. |
| `domain/document_initializer.go` | `DocumentInitializer` port (owned here, implemented by `documents`); `CloneTemplateRequest` value object with constructor; `DocumentRef` value object. |
| `domain/port.go` | `ControlledDocumentRepository` interface (read + write), `CDFilter` struct, `DBTX` interface (`DBExecutor` alias). |
| `domain/resolution.go` | Pure `Resolve` function: selects override vs default template version, emitting typed errors. |
| `domain/sequence.go` | `SequenceAllocator` interface; `DBTX`/`DBExecutor` types. |
| `domain/visibility.go` | `Visibility` value object (`company` | `restricted` + area/user grant lists); `NewVisibility` constructor. |
| `domain/autocode_test.go` | Unit tests for `AutoCode`. |
| `domain/controlled_document_test.go` | Unit tests for `NewControlledDocument` validation. |
| `domain/resolution_test.go` | Unit tests for `Resolve`. |
| `domain/sequence_test.go` | Integration test for `PostgresSequenceAllocator.NextAndIncrement` concurrency. Build tag `integration`. |
| `domain/sequence_bench_test.go` | Benchmark for `NextAndIncrement`. |
| `domain/visibility_test.go` | Unit tests for `NewVisibility`. |

### `infrastructure/`

| File | Role |
|---|---|
| `infrastructure/repository.go` | All Postgres adapters: `PostgresControlledDocumentRepository` (CRUD + visibility grant hydration), `PostgresSequenceAllocator` (counter management), `PostgresTemplateVersionChecker` (template state read), `TaxonomyProfileReader`, `TaxonomyAreaReader`. 791 lines. |
| `infrastructure/repository_test.go` | Tests for infrastructure layer (content not read fully but file confirmed present). |

**Total: 24 files** (8 production Go, 3 config/generated, 13 test).

---

## 3. Public surface

### Exported types consumed outside the module

| Symbol | Location | Consumers |
|---|---|---|
| `Module`, `New`, `RegisterRoutes`, `Service` | `module.go:16-54` | `apps/api/cmd/metaldocs-api/main.go` (construction + route registration + setter injection) |
| `ControlledDocumentService`, `NewControlledDocumentService`, `WithDocumentInitializer` | `application/service.go:38-144` | `main.go` (DI cycle break via setter); `documents/application/cd_initializer.go` (implements port) |
| `CreateControlledDocumentCmd`, `CreateResult`, `CreateRevisionCmd` | `application/service.go:53-78,574` | `main.go` (`controlledDocumentDuplicatorAdapter`); `delivery/http/routes.go` |
| `ControlledDocument` (alias), `CDFilter` (alias) | `application/service.go:35-36` | `main.go`, `documents/delivery/http/handler.go`, test files |
| `ErrActorMissing`, `ErrTemplateArtifactMissing`, `ErrTemplateArtifactInvariantUnconfigured` | `application/service.go:80-87` | `delivery/http/routes.go` (`writeDomainError`) |
| `ControlledDocument` entity | `domain/controlled_document.go:18` | `documents/application/service.go`, `documents/application/cd_initializer.go`, `documents/delivery/http/handler.go` |
| `CDStatusActive/Obsolete/Superseded` | `domain/controlled_document.go:13-15` | Checked throughout all consumers |
| `AutoCode` | `domain/controlled_document.go:105` | `delivery/http/routes.go:192` |
| `DocumentInitializer` interface | `domain/document_initializer.go:61` | Implemented by `documents/application/cd_initializer.go`; injected from `main.go` |
| `CloneTemplateRequest`, `NewCloneTemplateRequest` | `domain/document_initializer.go:19-40` | `application/service.go`, `documents/application/cd_initializer.go` |
| `DocumentRef`, `NewDocumentRef` | `domain/document_initializer.go:51` | `application/service.go`, `delivery/http/routes.go`, `documents/application/cd_initializer.go` |
| `ControlledDocumentRepository`, `CDFilter` | `domain/port.go:8-35` | Implemented by `infrastructure/repository.go` |
| `DBTX`, `DBExecutor` | `domain/sequence.go:8-14` | `application/service.go`, `infrastructure/repository.go` |
| `SequenceAllocator` | `domain/sequence.go:16` | Implemented by `infrastructure/repository.go` |
| `TemplateResolutionInput/Result/Candidate`, `Resolve`, `ErrTemplate*` | `domain/resolution.go` | `application/service.go` |
| `Visibility`, `NewVisibility`, `VisibilityScope*` | `domain/visibility.go` | `application/service.go`, `infrastructure/repository.go` |
| `PostgresControlledDocumentRepository`, `NewPostgresControlledDocumentRepository` | `infrastructure/repository.go:22-27` | `main.go:357` (second standalone instance for search/resolver adapter) |
| `PostgresSequenceAllocator`, `PostgresTemplateVersionChecker`, `TaxonomyProfileReader`, `TaxonomyAreaReader` | `infrastructure/repository.go:513-669` | Constructed inside `module.go:New` only |

### HTTP routes

All 8 routes are registered through `controlleddocumentsapi.HandlerWithOptions` at `handler.go:95` with `BaseURL: "/api/v1"`. POST operations are wrapped with idempotency middleware inside `handler.go:77-93`.

| Method | Path | Handler file:line | Authz |
|---|---|---|---|
| GET | `/api/v1/controlled-documents` | `routes.go:30` | Read; actor filter applied in service via `authn.UserIDFromContext` + visibility SQL predicate |
| POST | `/api/v1/controlled-documents` | `routes.go:70` | `CapControlledDocumentCreate` via `authz.Require` in-tx (area-scoped, ADR 0022 Phase 7); tier-1 IAM middleware; tier-3 Postgres tripwire |
| GET | `/api/v1/controlled-documents/preview-code` | `routes.go:172` | `CapControlledDocumentCreate` checked by `PeekSeq` in-tx (ADR 0022 Phase 7) |
| GET | `/api/v1/controlled-documents/{id}` | `routes.go:245` | Visibility-based `CanRead` check in service |
| POST | `/api/v1/controlled-documents/{id}/revisions` | `routes.go:197` | `CapControlledDocumentCreate` via `authz.Require` in-tx (area-scoped); tier-1 IAM middleware |
| GET | `/api/v1/controlled-documents/{id}/active-document` | `routes.go:266` | Inline SQL visibility check; no `authz.Require` (T-006 open gap) |
| PUT | `/api/v1/controlled-documents/{id}/obsolete` | `routes.go:441` | `CapControlledDocumentObsolete` via `authz.Require` in-tx (area-scoped); tier-1 IAM middleware; tier-3 tripwire |
| PUT | `/api/v1/controlled-documents/{id}/supersede` | `routes.go:455` | `CapControlledDocumentSupersede` via `authz.Require` in-tx (area-scoped); tier-1 IAM middleware; tier-3 tripwire |

---

## 4. Logic flows

### Flow 1: Atomic CD create (`POST /api/v1/controlled-documents`)

The highest-value flow — the entire atomic guarantee for QMS identity issuance.

1. **Idempotency check** — `handler.go:79-91`: `idempotency.Require(h.idempCreate, actorOf)` middleware intercepts; calls `BeginReplay` with body hash. On replay with same key + body → returns recorded 201 immediately; same key + different body → 422 `IDEMPOTENCY_KEY_CONFLICT`.
2. **Handler decode** — `routes.go:70-128`: max 1 MiB body, strict JSON decode, required-field validation (`missingAtomicCreateField`). Actor extracted from `authn.UserIDFromContext`.
3. **Taxonomy validation** — `service.go:147-162`: `profiles.GetByCode` and `areas.GetByCode` (Postgres reads of `metaldocs.document_profiles`, `metaldocs.document_process_areas`). Archived profile/area → 409.
4. **Manual-code path** — `service.go:171-196`: if `ManualCode` provided, validates reason length (≥10 chars) and calls `docs.CodeExists`; queues `numbering.override` governance event.
5. **Auto-code path — tx open and authz** — `service.go:198-217`: begins `*sql.Tx` at `sql.LevelReadCommitted`; `setAuthzGUC` writes `metaldocs.tenant_id` and `metaldocs.actor_id` session GUCs; `authz.Require(ctx, tx, "controlled_documents.create", processAreaCode)` — area-scoped ADR 0022 Phase 7 check.
6. **Override template validation** — `service.go:219-238`: if `OverrideTemplateVersionID` provided, `tplCheck.GetTemplateVersionState` reads `templates_template_version`; `domain.Resolve` enforces published + profile-match; queues `template.override` event.
7. **Template artifact pre-check** — `service.go:240-243`: `ensureTemplateArtifact` calls `docInit.ResolveTemplateStorageKey` then `docInit.Exists` (S3 existence check via documents module). Fails closed before sequence allocation if template artifact missing.
8. **Sequence allocation** — `service.go:245-274`: `seq.NextAndIncrement(ctx, tx, ...)` — CTE with `FOR UPDATE` on `cd_sequence_counters` row, increments and returns; `ensureCounterViaExec` creates the counter row lazily if absent (INSERT ON CONFLICT DO NOTHING).
9. **Domain object construction** — `service.go:292-318`: `NewControlledDocument` validates trimmed inputs; `NewVisibility` builds area/user grant lists defaulting to `processAreaCode` when no areas given.
10. **Transactional persist** — `service.go:322-344`: `docs.CreateTx(ctx, tx, doc)` calls `repository.go:353-366` which asserts `authz.Require` again (defense-in-depth within repo layer), then INSERTs into `controlled_documents` RETURNING id, then INSERTs visibility grants. `docInit.CloneTemplate(ctx, tx, doc, cloneReq)` delegates to `documents.application.NewCDDocumentInitializer` which creates the first `documents` row inside the same tx.
11. **Commit** — `service.go:336-339`: `tx.Commit()`. Any error rolls back both CD row and document row atomically.
12. **Post-commit governance events** — `service.go:347-352`: `govLogger.Log(...)` best-effort (warn-on-error, does not fail the request).
13. **Idempotency completion** — `handler.go` middleware: `CompleteReplay` records the 201 response for future replays.

### Flow 2: Active document lookup (`GET /api/v1/controlled-documents/{id}/active-document`)

The read path consumed by the Documents flow to decide which revision is "current".

1. **Tenant extraction** — `routes.go:267-271`: `tenantIDFromRequest` → `tenant.FromContext` (set by auth middleware via `injectTenant`).
2. **Visibility read-check** — `routes.go:280-324`: inline SQL EXISTS query joining `controlled_documents`, `controlled_document_area_grants`, `user_process_areas`, `controlled_document_user_grants` — no `authz.Require`, no `metaldocs.assert_caps` (T-006 gap).
3. **FULL OUTER JOIN query** — `routes.go:336-359`: selects active lineage row (status IN draft/under_review/approved/rejected/scheduled) LEFT OUTER JOIN latest published row. Returns `docID`, `contentHash`, `revisionVer`, `approvalState`, `publishedDocID`. If both sides NULL → 404.
4. **Approval instance enrichment** — `routes.go:406-436`: only if `approvalState = 'under_review'`, issues a second query on `approval_instances WHERE status = 'in_progress'` to populate `approval_instance_id`.
5. **Response** — `routes.go:438`: `ActiveDocumentResponse` with all fields optional.

### Flow 3: Lifecycle transition — Obsolete / Supersede (`PUT .../obsolete` or `.../supersede`)

1. **Handler dispatch** — `routes.go:441-466`: extracts tenantID, calls `svc.Obsolete` or `svc.Supersede`.
2. **Visibility pre-check** — `service.go:492-503`: `docs.CanRead(ctx, tenantID, id, actorUserID)` — returns 404 (not 403) if actor cannot see the CD.
3. **Tx open + GUC set** — `service.go:504-512`: begins tx; `setAuthzGUC` primes session settings.
4. **Row lock + area code** — `service.go:520-532`: `SELECT status, process_area_code FROM controlled_documents WHERE ... FOR UPDATE`; area code needed for area-scoped capability check.
5. **Tier-2 authz** — `service.go:534-536`: `authz.Require(ctx, tx, cap, areaCode)` — area-scoped against the CD's own process area (ADR 0022 Phase 7).
6. **State guard** — `service.go:538-540`: if `currentStatus != active` → `ErrCDNotActive` → 409.
7. **Status update** — `service.go:542-544`: `docs.UpdateStatusTx(ctx, tx, ...)` — UPDATE `controlled_documents SET status = $1 WHERE tenant_id = $2 AND id = $3`; tier-3 Postgres `trg_require_cap_asserted` fires here.
8. **Commit** — `service.go:545-547`.
9. **Post-commit governance event** — `service.go:549-571`: emits `controlled_documents.cd.obsoleted` or `controlled_documents.cd.superseded`; best-effort.

### Flow 4: Create revision (`POST /api/v1/controlled-documents/{id}/revisions`)

1. **Idempotency** — `handler.go:83-90`: `idempotency.Require(h.idempRevision, actorOf)` middleware.
2. **Actor + visibility check** — `service.go:585-596`: `CanRead` check to prevent probing non-visible CDs.
3. **CD fetch + active guard** — `service.go:597-602`: `GetByID`; if `!cd.IsActive()` → `ErrCDNotActive` → 409.
4. **Tx open, GUC, area-scoped authz** — `service.go:608-625`: same pattern as Create; authz uses `cd.ProcessAreaCode`.
5. **Clone template** — `service.go:627-633`: `docInit.CloneTemplate(ctx, tx, cd, cloneReq)` — adds a new `documents` row inside the same tx. Postgres constraint `ux_documents_cd_active` enforces single-active-sibling; violation mapped to `ErrActiveRevisionExists` → 409.
6. **Commit** — `service.go:635-638`.

### Flow 5: Preview code (`GET /api/v1/controlled-documents/preview-code`)

1. **Handler** — `routes.go:172-194`: validates `profile_code` and `area_code` query params, calls `svc.PeekSeq`.
2. **Profile/area validation** — `service.go:434-448`: validates both are active via `validateSequenceSeries`.
3. **In-tx authz** — `service.go:412-424`: opens tx, sets GUCs, calls `authz.Require` with `CapControlledDocumentCreate` against target area (ADR 0022 Phase 7 — preview-code allocates "within" a future CD creation).
4. **Peek** — `service.go:426-430`: `seq.Peek(ctx, ...)` — `SELECT next_seq FROM cd_sequence_counters` without incrementing; returns 1 if row absent.
5. **Response** — `routes.go:189-194`: formatted code via `AutoCode`.

---

## 5. Dependencies

### Outbound (module imports)

| Package | File | Purpose |
|---|---|---|
| `metaldocs/internal/modules/iam/authz` | `application/service.go:17`, `infrastructure/repository.go:17` | `authz.Require` for tier-2 in-tx capability check |
| `metaldocs/internal/modules/iam/domain` | `application/service.go:18`, `infrastructure/repository.go:18` | `CapControlledDocumentCreate/Obsolete/Supersede` constants; `WithAuthContext` in tests |
| `metaldocs/internal/modules/auth/domain` | `application/service.go:15` | `CurrentUserFromContext` for actorID extraction in `changeStatus` post-commit event |
| `metaldocs/internal/modules/taxonomy/domain` | `application/service.go:19`, `infrastructure/repository.go:19` | `DocumentProfile`, `ProcessArea`, `GovernanceLogger`, `GovernanceEvent`, taxonomy errors |
| `metaldocs/internal/modules/taxonomy/application` | `module.go:13` | `NewAuditGovernanceAdapter` (preferred) or `NewDBGovernanceLogger` (fallback when `AuditWriter` nil) |
| `metaldocs/internal/modules/audit/domain` | `module.go:8` | `auditdomain.Writer` in `Dependencies` struct |
| `metaldocs/internal/platform/authn` | `application/service.go:20`, `delivery/http/handler.go:12` | `UserIDFromContext` |
| `metaldocs/internal/platform/idempotency` | `delivery/http/handler.go:14` | `Store`, `New`, `Require` — POST replay store |
| `metaldocs/internal/platform/httpresponse` | `delivery/http/handler.go:15`, `routes.go` | `WriteError`, `WriteJSON` |
| `metaldocs/internal/platform/problem` | `delivery/http/routes.go:23` | Direct `problem.Write` for `ErrTemplateProfileMismatch` 422 branch |
| `metaldocs/internal/platform/tenant` | `delivery/http/handler.go:15`, `routes.go` | `FromContext`, `DevTenantID` fallback |
| `metaldocs/internal/platform/pagination` | `infrastructure/repository.go:16`, `delivery/http/routes.go` | Cursor encode/decode, `ClampLimit`, `ErrInvalidCursor` |
| `github.com/jackc/pgx/v5/pgconn` | `application/service.go`, `infrastructure/repository.go` | `PgError` for constraint-name-based error mapping |
| `github.com/jackc/pgx/v5/pgtype` | `infrastructure/repository.go` | `pgtype.FlatArray` for `ANY($n)` array parameters |
| `github.com/google/uuid` | `delivery/http/routes.go` | UUID parse/format in response mapping |
| `github.com/oapi-codegen/runtime/types` | `delivery/http/routes.go` | `openapi_types.UUID` in handler signatures |

### Inbound (verified by grep)

| Importer | File | Symbols used |
|---|---|---|
| `apps/api/cmd/metaldocs-api` | `main.go:46-49` | `controlleddocuments.New/Dependencies`; `controlleddocumentsapp.ControlledDocumentService/CreateControlledDocumentCmd`; `controlleddocumentsdomain.ControlledDocument`; `controlleddocumentsinfra.NewPostgresControlledDocumentRepository` |
| `internal/modules/documents/application` | `service.go`, `cd_initializer.go`, `cd_initializer_test.go`, `service_test.go`, `service_cd_test.go`, `service_caps_test.go`, `snapshot_wire_test.go`, `create_document_snapshot_integration_test.go` | Domain types `ControlledDocument`, `DocumentInitializer`, `DocumentRef`, `CloneTemplateRequest`; `DocumentInitializer` interface implementation |
| `internal/modules/documents/delivery/http` | `handler.go` | `controlleddocumentsdomain.ControlledDocument` in response types |

---

## 6. Persistence

### Tables owned

**`public.controlled_documents`** — canonical CD identity table.

| Column | Type | Constraints |
|---|---|---|
| `id` | UUID PK | gen_random_uuid() |
| `tenant_id` | UUID NOT NULL | — |
| `profile_code` | TEXT NOT NULL | FK → `metaldocs.document_profiles(tenant_id, code)` |
| `process_area_code` | TEXT NOT NULL | FK → `metaldocs.document_process_areas(tenant_id, code)` |
| `department_code` | TEXT | nullable |
| `code` | TEXT NOT NULL | UNIQUE (tenant_id, profile_code, code); CHECK length 2..100 |
| `sequence_num` | INT | nullable (null for manual-code CDs) |
| `title` | TEXT NOT NULL | — |
| `owner_user_id` | TEXT NOT NULL | — |
| `override_template_version_id` | UUID | nullable FK → `templates_template_version(id)` |
| `visibility_scope` | TEXT NOT NULL | CHECK IN ('company', 'restricted'); default 'company' |
| `status` | TEXT NOT NULL | CHECK IN ('active','obsolete','superseded'); default 'active' |
| `created_at` | TIMESTAMPTZ NOT NULL | — |
| `updated_at` | TIMESTAMPTZ NOT NULL | — |

DB-level triggers: `trg_controlled_documents_code_immutable` calls `reject_code_update()` before any UPDATE that touches `code`. Tripwire `trg_require_cap_asserted` (migration `0231_db_hardening_tripwire_and_dead_schema.sql`) fires on INSERT requiring `controlled_documents.create` and on UPDATE requiring `controlled_documents.obsolete` OR `controlled_documents.supersede`.

**`public.cd_sequence_counters`** — per-area sequence counter.

| Column | Type | Constraints |
|---|---|---|
| `tenant_id` | UUID NOT NULL | PK component |
| `profile_code` | TEXT NOT NULL | PK component; FK → `document_profiles` |
| `process_area_code` | TEXT NOT NULL | PK component; FK → `document_process_areas` |
| `next_seq` | INT NOT NULL | DEFAULT 1 |

Protected by tripwire: INSERT requires `controlled_documents.create`.

**`public.controlled_document_area_grants`** — area-based visibility grants (migration `archive/0198_controlled_document_visibility.sql`).

PK: `(tenant_id, controlled_document_id, area_code)`. FK: `controlled_document_id → controlled_documents(id) ON DELETE CASCADE`; `(tenant_id, area_code) → document_process_areas`. Index: `ix_cd_area_grants_tenant_area (tenant_id, area_code, controlled_document_id)`.

**`public.controlled_document_user_grants`** — user-based visibility grants.

PK: `(tenant_id, controlled_document_id, user_id)`. FK: `controlled_document_id → controlled_documents(id) ON DELETE CASCADE`. Index: `ix_cd_user_grants_tenant_user (tenant_id, user_id, controlled_document_id)`.

### Tables read but not owned

| Table | Owner | Access pattern |
|---|---|---|
| `documents` | documents | Active-doc FULL OUTER JOIN in `routes.go:336-358`; `documents` module reads CD via cross-module SQL |
| `approval_instances` | approval | Point-read in active-doc handler `routes.go:408-417` |
| `document_revisions` | documents | Content-hash fallback subquery in active-doc `routes.go:339-342` |
| `templates_template_version` | templates | State read via `PostgresTemplateVersionChecker.GetTemplateVersionState` at `repository.go:600-621` |
| `templates_template` | templates | JOIN in same query to get `doc_type_code` |
| `metaldocs.document_profiles` | taxonomy | `TaxonomyProfileReader.GetByCode` at `repository.go:631-663` |
| `metaldocs.document_process_areas` | taxonomy | `TaxonomyAreaReader.GetByCode` at `repository.go:671-701` |
| `user_process_areas` | IAM | Subquery inside `CanRead` / visibility EXISTS checks in `repository.go:487-497`, `routes.go:295-316` |
| `idempotency_keys` | platform | Indirectly via `platform/idempotency` middleware |

### Query patterns

- All writes are parameterized via `database/sql` + pgx driver; no string interpolation of values.
- `List` builds the query dynamically with index-based placeholders (`$1`, `$2`, ...) — safe but verbose (`repository.go:88-181`).
- `NextAndIncrement` uses a CTE with `SELECT ... FOR UPDATE` then `UPDATE ... RETURNING` for atomic sequence bump (`repository.go:570-584`).
- `CanRead` is a correlated EXISTS with three join paths; also duplicated verbatim in `GetActiveDocument` handler (`routes.go:281-317`) — duplication flag below.

### Migrations affecting this module

Active/current migration path (db/migrations + archive):

| File | Effect |
|---|---|
| `archive/0124_registry_controlled_documents.sql` | CREATE `controlled_documents`, legacy `profile_sequence_counters`; code-immutability trigger |
| `archive/0126_documents_v2_bridge_columns.sql` | Adds FK from `documents_v2` → `controlled_documents` |
| `archive/0127_documents_v2_tenant_consistency_trigger.sql` | Tenant consistency trigger on `documents_v2` writes |
| `archive/0128_grants_new_tables.sql` | GRANT on `controlled_documents`, `profile_sequence_counters` |
| `archive/0167_documents_bridge_and_state_columns.sql` | Bridge/state repair; FK reference to `controlled_documents` |
| `archive/0182_cd_sequence_per_area.sql` | DROP `profile_sequence_counters`; TRUNCATE + recreate CD table; CREATE `cd_sequence_counters` |
| `archive/0183_documents_name_not_empty.sql` | Backfill `documents.name` from `controlled_documents.title` |
| `archive/0188_tripwire_extend.sql` | Attaches `trg_require_cap_asserted` to `controlled_documents` (INSERT + UPDATE) and `cd_sequence_counters` (legacy capability names `registry.*`) |
| `archive/0198_controlled_document_visibility.sql` | ADD `visibility_scope` column; CREATE `controlled_document_area_grants`, `controlled_document_user_grants` |
| `db/migrations/0203_rename_templates_v2_objects.sql` | Renames template objects referenced by FK |
| `db/migrations/0210_controlled_documents_capability_namespace.sql` | Renames capability values from `registry.*` to `controlled_documents.*` |
| `db/migrations/0225_authz_p2_document_lifecycle_grants.sql` | Grants lifecycle capabilities to `area_admin`, `qms_admin` roles |
| `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql` | Rewrites `enforce_capability_asserted` with fail-close logic; maps `controlled_documents` INSERT → `controlled_documents.create`, UPDATE → `controlled_documents.obsolete|supersede`; `cd_sequence_counters` → `controlled_documents.create` |

---

## 7. Config & environment

The module consumes no environment variables directly. `module.go:27` accepts a `Dependencies` struct with `*sql.DB`, `*slog.Logger`, and `auditdomain.Writer` — all resolved by the composition root.

Test files read `DATABASE_URL` / `METALDOCS_DATABASE_URL` from the environment (`domain/sequence_test.go:21-23`) but these are integration-test-only and never reach production paths.

---

## 8. Concurrency & async

**Sequence allocation concurrency** — `repository.go:559-590`: `NextAndIncrement` uses a CTE with `SELECT ... FOR UPDATE` on `cd_sequence_counters` to serialize concurrent sequence increments per `(tenant, profile, area)`. `ensureCounterViaExec` uses `INSERT ON CONFLICT DO NOTHING` for safe lazy initialization.

**Transaction ownership** — `service.go:198-212`, `504-508`, `608-613`: the service opens `*sql.Tx` and passes it to repository methods and to `DocumentInitializer.CloneTemplate`. Multiple callers sharing the same DB connection pool are isolated by standard `database/sql` pool semantics.

**No goroutines, channels, or outbox writes** — the module is fully synchronous from request to commit. The governance event logger (`govLogger.Log`) writes post-commit synchronously and best-effort; it is not an outbox. No River job enqueues. No background timers.

---

## 9. Error handling & observability

### Error handling

All error paths use `fmt.Errorf("...: %w", err)` for wrapping throughout `service.go` and `repository.go`. Domain sentinel errors (`ErrCDNotFound`, `ErrCDNotActive`, etc.) are defined in `domain/controlled_document.go:35-50` and `domain/resolution.go:22-27`.

`writeDomainError` in `routes.go:538-593` maps 18 distinct error values to RFC 9457 `application/problem+json` responses via `httpresponse.WriteError` (which delegates to `platform/problem.Write`). One branch calls `problem.Write` directly (`routes.go:561`) for `ErrTemplateProfileMismatch` → 422 `template_invalid`.

Unhandled errors fall to the `default` branch at `routes.go:586-591`: logs with `slog.Error` + `"route", "controlledDocuments.writeDomainError"`, returns 500 `INTERNAL_ERROR`.

Panic guards in `module.go:41-46`: `New` panics if service or handler construction returns nil. Constructor panics on nil required dependencies (`service.go:99-116`).

### Observability

`slog.WarnContext` is used post-commit for governance event failures (`service.go:349`, `service.go:568`). `slog.Error` used in `routes.go:418-426` (approval instance lookup failure) and `routes.go:569-574` (unexpected tenant error path).

No metrics instrumentation in module code. No distributed trace span creation. No request-ID propagation specific to this module (handled by platform middleware upstream). Health/readiness probes: none module-specific.

---

## 10. Legacy / duplication / smell flags

- **FLAG-1 (major): `CanRead` SQL duplicated verbatim in two places.** The 23-line visibility EXISTS query appears identically in `infrastructure/repository.go:469-510` (the `CanRead` method) and in `delivery/http/routes.go:281-317` (the `GetActiveDocument` handler). Any change to the visibility logic must be applied in both locations. The handler bypasses the domain's `CanRead` port entirely and issues the query directly against `h.db`, which also means `GetActiveDocument` imports `*sql.DB` into the delivery layer. `git log` shows both paths were added in the same feature commit (`86fd8885f`). RF-009 candidate.

- **FLAG-2 (major): `GetActiveDocument` handler issues two raw SQL queries directly against `h.db`** (`routes.go:281`, `routes.go:336`, `routes.go:408`) — no repository or service indirection. This is a domain/infrastructure concern leaked into the delivery layer. The handler holds a `*sql.DB` dependency solely for this one route. Confirmed in `handler.go:30-43` where `db *sql.DB` is a field and `NewHandler` accepts it alongside the service interface.

- **FLAG-3 (major — T-006 open): `GetActiveDocument` has no `authz.Require` / capability check.** Any authenticated user with tenant access can retrieve document content hashes, approval states, and published document IDs for any CD they can "read" (visibility check). The visibility check is enforced, but no IAM capability guards this endpoint. Acknowledged in `wiki/modules/controlled-documents-tech-debt.md` T-006. Severity: Major (defense-in-depth gap).

- **FLAG-4 (major — T-005 open): No GUC + RLS backstop on owned tables.** All tenant enforcement is query-predicate only (`WHERE tenant_id = $1`). A repository method that omits the predicate has no DB-level safeguard. Acknowledged as T-005 in tech-debt register. Severity: Major (multi-tenant defense gap).

- **FLAG-5 (minor): `reflect` usage for a single nullable field set** in `TaxonomyAreaReader.GetByCode` at `repository.go:698-790`. `setNullableStringPtrField` uses `reflect.ValueOf(...).Elem().FieldByName(...)` to set `area.ParentCode` from a `sql.NullString`. Every other nullable field in the same file is set directly via a helper function. The reflect pattern is used for one field in one function — inconsistent, fragile to rename, slower, and unnecessary.

- **FLAG-6 (minor — T-009 open): Latent nil-panic on `Create` when `WithDocumentInitializer` setter was never called.** `service.go:325` calls `s.docInit.CloneTemplate(...)` if `s.docInit != nil` — the guard exists — but `service.go:358` calls `s.docInit.ResolveTemplateStorageKey(...)` via `ensureTemplateArtifact` without the nil guard (line 357: `if s.docInit == nil { return ErrTemplateArtifactInvariantUnconfigured }`). The first path is guarded; the second correctly returns an error. However the order-of-construction invariant is undocumented in the `New` function and the `WithDocumentInitializer` call site in `main.go:507` is far from construction. Late injection via a public setter on a mutable struct is the live risk.

- **FLAG-7 (minor — T-010 open): Second standalone `PostgresControlledDocumentRepository` instance created outside the module boundary.** `main.go:357` calls `controlleddocumentsinfra.NewPostgresControlledDocumentRepository(deps.SQLDB)` directly, bypassing the module's DI. The module exposes no `Repository()` accessor so consumers reach into the `infrastructure` sub-package. This ties `main.go` to the internal package layout.

- **FLAG-8 (minor): `TODO(phase11)` comment for leading-wildcard ILIKE search** at `repository.go:128`. The `List` filter's free-text search uses `%query%` ILIKE which is non-indexable and will table-scan on large tenants. Explicit named tech backlog item.

- **FLAG-9 (minor): `DBExecutor = DBTX` type alias is redundant** at `domain/sequence.go:14`. Both names refer to the same interface; only `DBTX` is used throughout the module. `DBExecutor` is exported but never consumed outside the module itself — potential confusion for readers.

- **FLAG-10 (minor — T-012 open): 79/94 exported symbols lack Go doc comments.** Notable undocumented exports: `ControlledDocument`, `CDStatus`, `CDStatusActive/Obsolete/Superseded`, `IsActive`, `ControlledDocumentRepository`, `CDFilter`, `Module`, `New`, `Dependencies`, `RegisterRoutes`, all repository concrete types. Coverage stat from `wiki/modules/controlled-documents-tech-debt.md` §T-012.

- **FLAG-11 (minor): Module was renamed from `registry` to `controlleddocuments` in 2026-05-11** (`git log` commit `86fd8885f`: "refactor(controlled-documents): phase 2 rename from registry"). The old capability keys `registry.create`, `registry.obsolete`, `registry.supersede` were renamed to `controlled_documents.*` in migration `0210`. No residual `registry` package references confirmed in current production paths. The wiki artifact `03-deps.md` and `01-surface.md` still reference `internal/modules/registry/` paths in all file:line anchors — drift in those artifact files.

- **FLAG-12 (info): `application/migration.go` and `BackfillLegacyDocuments` referenced in wiki artifacts and changelog but no longer present in the module.** The file `application/migration.go` does not exist in the current codebase (`ls` confirmed: only `service.go`, `service_test.go`, `integration_test.go`, `tenant_isolation_test.go`). The wiki changelog (2026-05-15 entry) and `_artifacts/01-surface.md` reference it as a recovery-only hook. It was removed (confirmed by `main.go` search for `RunLegacyMaintenance` and `BackfillLegacyDocuments` returning no results). This is a wiki drift item.

- **FLAG-13 (info): Governance logger falls back to `taxonomyapp.NewDBGovernanceLogger`** when `deps.AuditWriter` is nil (`module.go:37`). This couples the module to taxonomy's DB-based logger as a fallback path. In production the `AuditWriter` is injected (`buildControlledDocumentsModule` passes `deps.AuditWriter`), so the fallback is a development/test convenience but it remains a live code path.

---

## 11. Wiki drift

1. **`wiki/modules/controlled-documents/_artifacts/01-surface.md` and `03-deps.md`**: All file:line anchors reference `internal/modules/registry/` — the pre-rename package path. Current code lives at `internal/modules/controlleddocuments/`. Additionally, `01-surface.md` lists `application/migration.go` and `BackfillLegacyDocuments` as exported symbols — these no longer exist in the codebase.

2. **`wiki/modules/controlled-documents/_artifacts/04-persistence.md`**: Section 5 ("Tripwire pairing audit") records 5 violations with "NO Authz.Require called" for `Create`, `CreateTx`, `UpdateStatus`, `EnsureCounter`, `NextAndIncrement`. The current `repository.go` has `authz.Require` wired in `Create` (`repository.go:341`) and `CreateTx` (`repository.go:362`). `UpdateStatus` still has no `authz.Require` call (uses `UpdateStatusTx` now in the service flow which relies on the tier-2 check done before calling it). The table also references line numbers from the old `registry` module layout.

3. **`wiki/modules/controlled-documents/_artifacts/05-industry.md`**: IP-001 states "NOT applied" for RFC 9457 — this was closed in Plan 7. The module now uses `httpresponse.WriteError` → `problem.Write` for all error responses. The artifact predates Plan 7 closure and was never updated.

4. **`wiki/modules/controlled-documents/_artifacts/04-persistence.md`**: Section 3 (Triggers) does not include the `trg_require_cap_asserted` entries for `controlled_documents` and `cd_sequence_counters` from `0231_db_hardening_tripwire_and_dead_schema.sql`, nor the visibility tables from `0198`. The table refers to `0142b_role_capabilities_v2_enforce.sql` for the tripwire but the current hardened trigger comes from `0231`.

5. **`wiki/modules/controlled-documents.md` section 5.2**: Public surface table references `infrastructure/repository.go:353` for `CreateTx` and `infrastructure/repository.go:432` for `UpdateStatus`. In the current file `CreateTx` is at line 353 (confirmed) and `UpdateStatus` is at line 432 (confirmed). These anchors appear current. However the same table references `application/service.go:293` for `Obsolete/Supersede` — in the current `service.go` `Obsolete` is at line 451 and `Supersede` at line 455. The line anchors for service.go are stale.

6. **`wiki/modules/controlled-documents.md` section 6.2**: Tripwire pairing note says "VIOLATION — no `authz.Require`" for `GetActiveDocument`. This is still accurate (T-006 remains open), but the note has become ambiguous since the rest of the module now has tier-2 in place; the note should be scoped explicitly to "capability authz" not "all authz" since the visibility check does exist.

---

## 12. Open questions

1. **[runtime-unverified] `UpdateStatus` (non-Tx variant) in `repository.go:432` is called from no live path in the current service code.** `service.go` uses `UpdateStatusTx` via `changeStatus`. `UpdateStatus` is defined in the `ControlledDocumentRepository` interface and implemented but callers are not apparent from static analysis. Possible dead export or reserved for external callers.

2. **[runtime-unverified] `Create` (non-Tx variant) in `repository.go:333` opens its own internal tx and asserts `authz.Require`.** In the normal flow, `service.go` calls `CreateTx` (not `Create`) when the service-owned tx is present. The `Create` path is invoked in the `else` branch at `service.go:341` when `createTx == nil` (manual-code path with `s.db == nil`). When would `s.db` be nil in production? Only in test fakes. Whether `Create` (non-Tx) is reachable in a well-configured production binary requires runtime verification.

3. **[runtime-unverified] Governance event delivery reliability.** `govLogger.Log` is called post-commit, best-effort. If the database is unavailable at that moment (post-commit network blip), governance events for `numbering.override`, `template.override`, `controlled_documents.cd.obsoleted`, and `controlled_documents.cd.superseded` are silently dropped (warn log only). There is no outbox row, no retry, no DLQ. The production impact depends on `AuditGovernanceAdapter` implementation details not fully verified here.

4. **[runtime-unverified] `injectTenant` returns `tenant.DevTenantID` fallback** in `tenantIDFromContext` when the context key is absent (`handler.go:62-66`). In production this fallback would return a dev sentinel for a request that somehow lacked the tenant injection — which would corrupt query results against production data. The outer `injectTenant` middleware errors with 500 on `tenant.FromContext` failure, so the fallback in `tenantIDFromContext` should be unreachable in the normal request path. However it remains a latent footgun if called outside the middleware context.

5. **[runtime-unverified] `GetActiveDocument` N+1 pattern.** The handler issues up to 3 sequential DB queries per request: (1) visibility EXISTS check, (2) FULL OUTER JOIN for active/published rows, (3) conditional approval instance lookup. No batching or single-query redesign. Performance impact under load is unverified without runtime profiling.
