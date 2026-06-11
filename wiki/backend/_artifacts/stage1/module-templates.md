# Stage-1 Audit Artifact — module-templates

> **Produced:** 2026-06-10
> **Scope:** `internal/modules/templates/` — complete read-only structural map.
> **Read-only policy:** no source files were modified.

---

## 1. Identity & purpose

`templates` owns the full lifecycle of DOCX-based document templates: creation, versioning, DOCX storage (presigned MinIO PUT/GET), placeholder-schema authoring, a two-stage ISO-segregated approval chain (draft → in_review → approved → published), automatic obsolescence of the prior published version on every publish, and archival. It is the upstream supplier of published template state consumed by `documents` (placeholder schema snapshot at instantiation), `taxonomy` (published-version existence check), `controlled-documents` (FK to `template_version_id`), and the docx-renderer (raw SQL read of DOCX storage key). The module is organized in a hexagonal layout: `domain/` (pure entities and invariants), `application/` (use-case layer with ports), `delivery/http/` (HTTP handler + route files), `repository/` (Postgres adapter), and `api/` (oapi-codegen generated surface). It exposes 21 HTTP routes under `/api/v1/templates` plus `GET /api/v1/signed`, all mounted via a generated `ServerInterfaceWrapper`. Authz is two-tier: a Tier 1 `AuthzFunc` (real `CapabilityService` wired at composition root) and Tier 2 `authz.Require` inside every mutating `*sql.Tx` backed by the Postgres `trg_require_cap_asserted` tripwire on both owned tables.

---

## 2. File inventory

### api/
| File | Role |
|---|---|
| `internal/modules/templates/api/api.gen.go` (4 487 lines) | oapi-codegen generated surface: `ServerInterface`, `StrictServerInterface`, `ServerInterfaceWrapper`, all `*RequestObject`/`*ResponseObject` types, `GetSwagger`, `GetSpec` |
| `internal/modules/templates/api/cfg.yaml` | oapi-codegen generator config |
| `internal/modules/templates/api/gen.go` | `//go:generate` directive |

### domain/
| File | Role |
|---|---|
| `internal/modules/templates/domain/template.go` | `Template` aggregate root struct + `IsArchived()` helper + `ErrNotFound`, `ErrKeyConflict`, `ErrArchived`, `ErrSystemTemplateImmutable` |
| `internal/modules/templates/domain/version.go` | `TemplateVersion` entity + `VersionStatus*` constants + `CanTransition()` state-machine guard + `RoleBindingFor()` helper + `ErrInvalidStateTransition`, `ErrContentHashMismatch`, `ErrStaleBase`, `ErrStaleLockVersion` |
| `internal/modules/templates/domain/schemas.go` | `MetadataSchema`, `PlaceholderType` (`PHText`/`PHDate`/`PHNumber`/`PHSelect`/`PHUser`/`PHPicture`/`PHComputed`) + `VisibilityOp` + `VisibilityCondition` + `Placeholder` + `CompositionConfig` (deprecated) |
| `internal/modules/templates/domain/approval.go` | `ApprovalConfig` struct + `NewApprovalConfig` constructor + `HasReviewer()` + `CheckSegregation()` ISO SoD enforcer |
| `internal/modules/templates/domain/audit.go` | `AuditAction` constants + `AuditEvent` struct + `NewAuditEvent` constructor |
| `internal/modules/templates/domain/errors.go` | All domain error sentinels not declared in version.go or template.go (19 vars total) |

### application/
| File | Role |
|---|---|
| `internal/modules/templates/application/service.go` | `Service` struct + `New` constructor + `WithDB` / `DB` builder — sole owner of the `*sql.DB` reference |
| `internal/modules/templates/application/ports.go` | `Repository` interface (40 methods), `Presigner` interface, `Clock`, `UUIDGen`, `ResolverRegistryReader` interfaces, `ListFilter` struct |
| `internal/modules/templates/application/create.go` | `CreateTemplateCmd`/`CreateTemplateResult` + `Service.CreateTemplate` (tx-backed) + `CreateVersionCmd` + `Service.CreateNextVersion` (tx-backed) + clone helpers |
| `internal/modules/templates/application/schema.go` | `UpdateSchemasCmd` + `Service.UpdateSchemas` (tx-backed CAS) + `ValidatePlaceholders` catalog gate (exported) + `placeholderCatalogSet` |
| `internal/modules/templates/application/lifecycle.go` (573 lines) | All lifecycle commands + `Service.SubmitForReview`, `Service.Review`, `Service.Approve` (tx-backed), `Service.PublishTemplateVersion` (tx-backed, mandatory), `Service.ArchiveTemplate` (tx-backed), `ApproveResult`, `PublishTemplateVersionResult`, `updateVersionWithAuthz` helper, `buildNextDraftVersion` helper, `containsRole` helper |
| `internal/modules/templates/application/autosave.go` | DOCX upload path: `PresignAutosaveCmd/Result`, `PresignTemplateUploadCmd`, `CommitAutosaveCmd`, `SaveTemplateDraftCmd` + their `Service` methods (presign, commit-autosave with content-hash gate, save-draft with lock-version CAS) |
| `internal/modules/templates/application/approval_config.go` | `UpsertApprovalConfigCmd` + `Service.UpsertApprovalConfig` (tx-backed; creator-or-admin gate) |
| `internal/modules/templates/application/queries.go` | `Service.GetTemplate`, `Service.GetVersion` (tenant-gated), `Service.ListTemplates`, `Service.ListAudit`, `Service.GetDocxURL`, `Service.PresignStoredObject`; `GetDocxURLCmd` |
| `internal/modules/templates/application/audit.go` | `newAuditEvent` internal helper (delegates to `domain.NewAuditEvent`) |
| `internal/modules/templates/application/authz_guc.go` | `setAuthzGUC` — sets `metaldocs.tenant_id` / `metaldocs.actor_id` GUC on a `*sql.Tx` |
| `internal/modules/templates/application/errors.go` | `wrapAppErr` + `isDomainErr` — error pass-through logic |
| `internal/modules/templates/application/visibility_graph.go` | `DetectVisibilityCycle` — DFS cycle check across `VisibilityCondition` DAG (exported) |
| `internal/modules/templates/application/*_test.go` | Unit tests (approval_config, autosave, create, lifecycle, lifecycle_publish_role, queries, schema, visibility_graph, fakes_test) |

### delivery/http/
| File | Role |
|---|---|
| `internal/modules/templates/delivery/http/handler.go` | `AuthzFunc` type + `Handler` struct + `New` constructor + `Register` (mounts all 22 routes) + `idempotent` middleware wiring + `tenantIDFromReq`, `userIDFromReq`, `writeErr`, `writeMappedErr` helpers |
| `internal/modules/templates/delivery/http/routes_generated.go` (304 lines) | `var _ ServerInterface = (*Handler)(nil)` compile-time check + thin generated wrappers delegating to internal handler bodies: `ListTemplates`, `CreateTemplate`, `GetTemplateVersion`, `PresignTemplateDocxUploadUrl`, `PresignTemplateSchemaUploadUrl`, `presignTemplateUpload`, `RedirectSignedUrl`, `SaveTemplateDraft`, `PublishTemplateVersion`, `CreateTemplateVersion`, `UpdateTemplateSchema`, `PresignTemplateAutosave`, `CommitTemplateAutosave`, `SubmitTemplateVersion`, `ReviewTemplateVersion`, `ApproveTemplateVersion`, `ArchiveTemplate`, `UpsertTemplateApprovalConfig`, `GetTemplate`, `GetTemplateDocxUrl`, `ListTemplateAudit`, `ListTemplatePlaceholderCatalog` |
| `internal/modules/templates/delivery/http/routes_create.go` | `createNextVersion` handler body + `toTemplateResponse` + `toVersionResponse` + `timePtrRFC3339` response mappers |
| `internal/modules/templates/delivery/http/routes_query.go` | `listTemplates`, `GetSystemBlankTemplate`, `getTemplate`, `getVersion`, `getDocxURL`, `listAudit` handler bodies + `readQueryInt` helper; hardcoded system-blank template IDs |
| `internal/modules/templates/delivery/http/routes_lifecycle.go` | `submitForReview`, `review`, `approve`, `archiveTemplate`, `upsertApprovalConfig` handler bodies + `actorRolesFromReq` |
| `internal/modules/templates/delivery/http/routes_schema.go` | `updateSchemas` handler body |
| `internal/modules/templates/delivery/http/routes_autosave.go` | `presignAutosave`, `commitAutosave` handler bodies |
| `internal/modules/templates/delivery/http/routes_catalog.go` | `placeholderCatalog` static slice + `listPlaceholderCatalog` handler body |
| `internal/modules/templates/delivery/http/errors.go` | `MapErr` domain-error → HTTP status+code map + all `codeTpl*` constants |
| `internal/modules/templates/delivery/http/*_test.go` | HTTP-layer tests (errors_test, handler_problem_test, routes_autosave_test, routes_catalog_test, routes_contract_test, routes_create_test, routes_lifecycle_test, routes_query_test, routes_schema_test) |

### repository/
| File | Role |
|---|---|
| `internal/modules/templates/repository/postgres.go` (715 lines) | `Repository` struct + `New` + `WithAudit(auditdomain.Writer)` builder + all 40 `Repository` interface methods: CRUD on `templates_template`, `templates_template_version`, CAS variants for draft and schema, `ObsoletePreviousPublished[Tx]`, CRUD on `templates_approval_config`, `AppendAudit[Tx]`/`ListAudit` (routes through canonical `audit.Writer` if wired; reads from `templates_audit_log` for `ListAudit`), plus `isInvalidUUID` and `rowsAffected` helpers |
| `internal/modules/templates/repository/mappers.go` | `scanTemplate`, `scanTemplateVersion`, `marshalVersionSchemas`, `unmarshalAuditDetails` |
| `internal/modules/templates/repository/postgres_test.go` | Unit tests against fakes |
| `internal/modules/templates/repository/postgres_integration_test.go` | Postgres integration tests |

---

## 3. Public surface

### Exported types consumed outside the module

| Symbol | Package | Consumed by |
|---|---|---|
| `domain.Template` | `templates/domain` | `documents/application` (snapshot), `taxonomy/infrastructure` (version checker) |
| `domain.TemplateVersion` | `templates/domain` | `documents/application` (fillin, snapshot, freeze), `render/fanout` (client_test) |
| `domain.Placeholder`, `domain.PlaceholderType`, `PHText`…`PHComputed` | `templates/domain` | `documents/application/fillin_service.go`, `documents/application/freeze_service.go` |
| `domain.MetadataSchema` | `templates/domain` | `documents/application` |
| `domain.ErrNotFound` | `templates/domain` | `documents/application` |
| `application.ValidatePlaceholders` | `templates/application` | Not imported outside module (grep-verified); called only within `application/schema.go:114` |
| `application.DetectVisibilityCycle` | `templates/application` | Not imported outside module (grep-verified) |
| `api.gen.go` surface | `templates/api` | Consumed only by `delivery/http/routes_generated.go` within the module |

### HTTP routes

All routes mount under `/api/v1/templates` unless noted. Tier 1 capability is the `AuthzFunc` check at handler entry; Tier 2 is `authz.Require` inside the `*sql.Tx`.

| Method | Path | Tier 1 cap | Tier 2 cap | Idempotency |
|---|---|---|---|---|
| GET | `/api/v1/signed` | `template.view` (via presign) | none | no |
| GET | `/api/v1/templates` | `template.view` | none | no |
| POST | `/api/v1/templates` | `template.create` | `CapTemplateCreate` | `h.idempotent` |
| GET | `/api/v1/templates/{id}` | `template.view` | none | no |
| GET | `/api/v1/templates/system/blank` | `template.view` | none | no |
| GET | `/api/v1/templates/{id}/versions/{n}` | `template.view` | none | no |
| POST | `/api/v1/templates/{id}/versions` | `template.create` | `CapTemplateEdit` | no |
| PUT | `/api/v1/templates/{id}/versions/{n}/draft` | `template.edit` | `CapTemplateEdit` | no |
| PUT | `/api/v1/templates/{id}/versions/{n}/schema` | `template.edit` | `CapTemplateEdit` | no |
| POST | `/api/v1/templates/{id}/versions/{n}/docx-upload-url` | `template.edit` | none | no |
| POST | `/api/v1/templates/{id}/versions/{n}/schema-upload-url` | `template.edit` | none | no |
| POST | `/api/v1/templates/{id}/versions/{n}/autosave/presign` | `template.edit` | none | no |
| POST | `/api/v1/templates/{id}/versions/{n}/autosave/commit` | `template.edit` | `CapTemplateEdit` | no |
| POST | `/api/v1/templates/{id}/versions/{n}/submit` | `template.submit` | `CapTemplateSubmit` | `h.idempotent` |
| POST | `/api/v1/templates/{id}/versions/{n}/review` | `template.review` | `CapTemplateReview` | `h.idempotent` |
| POST | `/api/v1/templates/{id}/versions/{n}/approve` | `template.approve` | `CapTemplateApprove` | `h.idempotent` |
| POST | `/api/v1/templates/{id}/versions/{n}/publish` | `template.approve` (**not** `template.publish`) | `CapTemplatePublish` | `h.idempotent` |
| POST | `/api/v1/templates/{id}/archive` | `template.archive` | `CapTemplateArchive` | no |
| PUT | `/api/v1/templates/{id}/approval-config` | `template.admin` (**undeclared cap**) | `CapTemplateEdit` | no |
| GET | `/api/v1/templates/{id}/versions/{n}/docx-url` | `template.view` | none | no |
| GET | `/api/v1/templates/{id}/audit` | `template.view` | none | no |
| GET | `/api/v1/templates/placeholder-catalog` | `template.view` | none | no |

---

## 4. Logic flows

### Flow 1 — Template creation (`POST /api/v1/templates`)

1. `handler.go:44` — request arrives; `h.idempotent` middleware checks `idempotency_keys` table for prior result keyed on `(tenantID, userID, route_template, idempotency_key_header)`.
2. `routes_generated.go:20` — `CreateTemplate`: extracts `tenantID` via `tenant.FromContext`, calls `h.authz(r, tenantID, "*", "template.create")` (Tier 1 real capability check).
3. `routes_generated.go:31` — strict-JSON decode of `CreateTemplateJSONRequestBody`; missing `key`/`name` returns 400.
4. `create.go:30` — `Service.CreateTemplate`: key uniqueness guard via `GetTemplateByKey` (returns 409 on collision).
5. `create.go:62` — opens `*sql.Tx`; calls `setAuthzGUC` (GUC `metaldocs.tenant_id` + `metaldocs.actor_id`); calls `authz.Require(CapTemplateCreate, "tenant")` — tripwire on `templates_template` must see matching GUC.
6. `create.go:74` — `CreateTemplateTx`: inserts into `templates_template` (hard-codes `areas='{}'`, `visibility='public'`, `specific_areas='{}'`).
7. `create.go:77` — `CreateVersionTx`: inserts into `templates_template_version` with `revision_number = COALESCE(MAX+1, 0)` subquery; RETURNING clause back-fills `v.RevisionNumber`.
8. `create.go:80` — `UpsertApprovalConfigTx`: inserts into `templates_approval_config`.
9. `create.go:87` — `AppendAuditTx`: delegates to `auditdomain.Writer.RecordTx` → canonical `metaldocs.audit_events`; routes to `templates_audit_log` via `ListAudit` only.
10. `create.go:98` — `tx.Commit()`.
11. `routes_generated.go:63` — writes HTTP 201 with `{id, version_id, data:{template, version}}`.

**Failure modes:** key conflict → 409; authz denied → 403; Postgres error → 500.

---

### Flow 2 — DOCX upload (presign → PUT → commit) for a draft version

1. `routes_generated.go:90` — `PresignTemplateDocxUploadUrl`: authz `template.edit`; calls `Service.PresignTemplateUpload` → `autosave.go:33` → validates draft status → `presign.PresignPUT(key, 10m)` → returns `{url, storage_key}`.
2. Client performs a direct `PUT` to MinIO using the presigned URL (no API involvement).
3. `routes_autosave.go:50` — `commitAutosave`: authz `template.edit`; reads `expected_content_hash` from JSON body.
4. `autosave.go:157` — `CommitAutosave`: calls `presign.HeadContentHash(key)` to verify the object exists and its `ETag` matches; mismatch triggers `presign.Delete` + returns 409.
5. Opens `*sql.Tx`; `setAuthzGUC`; `authz.Require(CapTemplateEdit)`; `UpdateVersionTx` (full version update with `lock_version` CAS); `AppendAuditTx(AuditSaved)`.
6. Returns the updated `TemplateVersion` to the caller.

**Failure modes:** upload missing → 409 `upload_missing`; hash mismatch + delete error → joined error; stale lock → 412 `concurrent_modification`; tripwire abort → 500.

---

### Flow 3 — Schema update with OCC (`PUT /api/v1/templates/{id}/versions/{n}/schema`)

1. `routes_schema.go:11` — `updateSchemas`: authz `template.edit`; reads `{metadata_schema, placeholder_schema, expected_lock_version}` from body; rejects missing `expected_lock_version` with 400.
2. `schema.go:22` — `Service.UpdateSchemas`: calls `s.GetTemplate` (tenant gate) then `s.GetVersion`; rejects non-draft status.
3. `schema.go:34` — `ValidatePlaceholders`: enforces 7-token `PHType` enum, duplicate-ID/name detection, regex compile check, numeric/date range validation, `VisibleIf` op validity, DFS cycle check (`DetectVisibilityCycle`). For `PHComputed` with a named placeholder in `placeholderCatalogSet`, also validates the `resolver_key` matches the name. If `s.resolvers != nil` (currently `nil` at wiring), also validates `resolver_key` against `Known()`.
4. Opens `*sql.Tx`; `setAuthzGUC`; `authz.Require(CapTemplateEdit)`; `UpdateVersionSchemaCASTx(v, expectedLockVersion)` — UPDATE with `AND lock_version = $2`, increments `lock_version`, returns `ErrStaleLockVersion` if row affected = 0 and exists; `AppendAuditTx(AuditSaved)`.
5. Returns updated version; caller sees `lock_version` bumped.

**Failure modes:** missing `expected_lock_version` → 400; non-catalog PHType → 422; stale lock → 412 `concurrent_modification`; non-draft version → 409.

---

### Flow 4 — Publish via approval chain (`POST /approve`)

This is the ISO-compliant path (Approve with hasReviewer=true: draft → in_review → approved → published).

1. `routes_lifecycle.go:97` — `approve`: authz `template.approve`; reads `{accept, reason}`.
2. `lifecycle.go:209` — `Service.Approve`: loads template + version. Checks `version.Status == VersionStatusApproved` (with reviewer) or `VersionStatusInReview` (without reviewer).
3. `lifecycle.go:231` — `version.RoleBindingFor(VersionStatusPublished)` returns `PendingApproverRole`; if non-empty, verifies `actorRoles` contains it; returns `ErrForbiddenRole` (→ 403) if not.
4. `lifecycle.go:234` — `CheckSegregation(approver, actorID, authorID, reviewerID)`: verifies actor ≠ author and actor ≠ reviewer; returns `ErrISOSegregationViolation` (→ 403) if violated.
5. `lifecycle.go:239` — Accept path: `version.CanTransition(Published)` guard; sets `Status=published`, `ApproverID`, `ApprovedAt`, `PublishedAt`; updates `template.PublishedVersionID`, `PublishedVersionNumber`, `CurrentRevisionNumber`.
6. `lifecycle.go:265` — opens `*sql.Tx`; `setAuthzGUC`; `authz.Require(CapTemplateApprove)`; `ObsoletePreviousPublishedTx`; `UpdateTemplateTx`; `UpdateVersionTx`; `CreateVersionTx` (auto-spawns draft `v(n+1)` in same tx); `AppendAuditTx(AuditPublished)`.
7. `lifecycle.go:296` — `tx.Commit()`.
8. `routes_lifecycle.go:139` — writes 200 `{data:{version, next_draft:{id, version_number}|null}}`.

**Failure modes:** wrong state → 409; SoD violation → 403 `iso_segregation_violation`; forbidden role → 403 `forbidden_capability`; authz cap denied → 403; tripwire abort → 500.

---

### Flow 5 — Direct publish (`POST /publish`)

Alternate path that bypasses the review/approve chain; intended for `system_admin` with `template.publish` capability who also holds the `pending_approver_role`.

1. `routes_generated.go:196` — `PublishTemplateVersion`: authz **`template.approve`** (Tier 1 — note mismatch with route intent, should be `template.publish`); reads `{docx_key, schema_key}`.
2. `lifecycle.go:373` — `Service.PublishTemplateVersion`: loads template + version; checks `status == draft`; checks `content_hash != ""`.
3. `lifecycle.go:395` — `CheckSegregation(approver, actor, author, reviewer)`.
4. `lifecycle.go:401` — `version.RoleBindingFor(Published)`: if non-empty and actor lacks the role, emits `AppendAudit(AuditPublishForbiddenRole)` (best-effort; errors logged, not surfaced) then returns `ErrForbiddenRole`.
5. `lifecycle.go:426` — returns `domain.ErrTransactionRequired` if `s.db == nil` (guards non-tx code path).
6. Opens `*sql.Tx`; `setAuthzGUC`; `authz.Require(CapTemplatePublish)` (Tier 2 — correct cap); `ObsoletePreviousPublishedTx`; `UpdateTemplateTx`; `UpdateVersionTx`; `CreateVersionTx` (auto-spawns `v(n+1)`); second `UpdateTemplateTx` (bumps `latest_version`); `AppendAuditTx(AuditPublished)`.
7. `tx.Commit()`.
8. Returns 200 with flat body `{published_version_id, next_draft_id, next_draft_version_num, published_version_number}` — different shape from the approve path.

**Failure modes:** not draft → 409; content_hash empty → 409; SoD violation → 403; forbidden role → 403; cap denied → 403; transaction required (s.db nil) → internal 500 (unreachable at production wiring).

---

## 5. Dependencies

### Outbound imports (verified by reading source)

| Dependency | Import path | Why |
|---|---|---|
| `iam/authz` | `metaldocs/internal/modules/iam/authz` | `authz.Require` inside every mutating tx |
| `iam/domain` | `metaldocs/internal/modules/iam/domain` | `CapTemplate*` constants; `UserIDFromContext`; `RolesFromContext` |
| `audit/domain` | `metaldocs/internal/modules/audit/domain` | `auditdomain.Writer` interface; `auditdomain.Event` struct wired in repository |
| `platform/tenant` | `metaldocs/internal/platform/tenant` | `tenant.FromContext` in handler and repository fallback |
| `platform/problem` | `metaldocs/internal/platform/problem` | `problem.Write`, `problem.New`, `problem.Code` constants in delivery layer |
| `platform/httpresponse` | `metaldocs/internal/platform/httpresponse` | `WriteJSON`, `ReadJSON` |
| `platform/idempotency` | `metaldocs/internal/platform/idempotency` | `idempotency.New`, `idempotency.Require` for idempotent POST routes |
| `database/sql` | stdlib | `*sql.DB`, `*sql.Tx` throughout |
| `github.com/jackc/pgx/v5/pgconn` | external | `pgconn.PgError` for SQLSTATE 23505 / 22P02 mapping |
| `github.com/google/uuid` | external | `uuid.NewString()` in `auditEvent` helper |
| `github.com/oapi-codegen/runtime/types` | external | `openapi_types.UUID` in generated wrapper signatures |

### Inbound consumers (grep-verified)

| Consumer | Import | What it uses |
|---|---|---|
| `documents/application` | `templates/domain` | `domain.Template`, `domain.TemplateVersion`, `domain.Placeholder`, `domain.MetadataSchema`, `domain.PlaceholderType`, `domain.ErrNotFound`, `PHText`…`PHComputed` constants |
| `taxonomy/infrastructure/template_version_checker.go` | raw SQL (no Go import) | Queries `templates_template_version` + `templates_template` directly to check `IsPublished` |
| `controlleddocuments/infrastructure/repository.go` | raw SQL (no Go import) | Queries `templates_template_version` + `templates_template` |
| `documents/application/fillin_service.go` | `templates/domain` (Go import) + raw SQL | Joins `templates_template_version` in SQL; imports domain types |
| `render/fanout` (test only) | `templates/domain` (test import) | Used in integration/fanout tests |
| `apps/api/cmd/metaldocs-api/main.go` | `templates/application`, `templates/delivery/http`, `templates/repository` | Full wiring: repo + presigner + service + handler |

---

## 6. Persistence

### Tables owned

| Table | PK | Key columns | Notes |
|---|---|---|---|
| `templates_template` | `id uuid` | `tenant_id`, `key` (unique per tenant), `doc_type_code`, `latest_version`, `published_version_id`, `created_by`, `system_owned`, `archived_at` | Also carries `areas text[]`, `visibility text`, `specific_areas text[]` — inert legacy columns written with hard-coded empty values, never read back into Go types |
| `templates_template_version` | `id uuid` | `template_id`, `version_number` (unique per template), `revision_number` (unique per template, ADR 0013), `status`, `docx_storage_key`, `content_hash`, `metadata_schema jsonb`, `placeholder_schema jsonb`, `lock_version int` | `lock_version` CAS enforced on every UPDATE |
| `templates_approval_config` | `template_id uuid` | `reviewer_role`, `approver_role` | One row per template, upserted on create and on config change |
| `templates_audit_log` | `id bigserial` | `tenant_id`, `template_id`, `version_id`, `actor_id`, `action`, `details jsonb`, `occurred_at` | Legacy module-local sink; `ListAudit` reads from here. `AppendAudit[Tx]` now delegates to canonical `audit.Writer` if wired (always wired at production via `WithAudit(deps.AuditWriter)`). The local table still exists and is not written by current production code paths — it is the sole source for the `GET /{id}/audit` response |

### Key query patterns

| Pattern | Location | Notes |
|---|---|---|
| `GetTemplate` — tenant-scoped point lookup with LEFT JOINs to published/latest version for revision metadata | `repository/postgres.go:72` | Safe: `WHERE t.id = $1 AND t.tenant_id = $2::uuid` |
| `GetVersion` — template_id + version_number + tenant join | `repository/postgres.go:258` | `JOIN templates_template t ON t.id = v.template_id … AND t.tenant_id = $3::uuid` — tenant enforced via join |
| `GetVersionByID` — version UUID + tenant join | `repository/postgres.go:279` | Same join pattern; safe. T-002 residual risk is gone from this query but callers bear responsibility for template-ID cross-check (done in `CreateNextVersion`) |
| `ListTemplates` — tenant-scoped, `system_owned=false` filter, `doc_type_code` optional | `repository/postgres.go:114` | Has `LIMIT $3 OFFSET $4` — pagination parameters flow in from handler |
| `CreateVersion` — allocates `revision_number` atomically via `COALESCE(MAX+1, 0)` subquery | `repository/postgres.go:184` | Same subquery in Tx variant. No explicit row lock — race condition between two concurrent creates of the same template's next version is resolved by the `UNIQUE (template_id, revision_number)` index |
| `UpdateVersion` / `UpdateVersionTx` — full OCC with `lock_version` CAS | `repository/postgres.go:332` | `WHERE id = $1 AND … lock_version = $16`; discriminates not-found vs. stale via follow-up existence SELECT |
| `UpdateVersionDraftCAS` / schema CAS | `repository/postgres.go:445`, `:496` | Narrower field sets; same CAS + discriminate pattern |
| `ObsoletePreviousPublished[Tx]` | `repository/postgres.go:544` | Bulk UPDATE of `status='obsolete'` — no per-row `lock_version` check |

### Migration files (templates-relevant)

| Migration | Effect |
|---|---|
| `db/migrations/0203_rename_templates_v2_objects.sql` | Renames templates tables from v2 naming; updates authz tripwire function for renamed tables |
| `db/migrations/0213_templates_tenant_id_uuid.sql` | Casts `templates_template.tenant_id` to uuid type |
| `db/migrations/0233_templates_template_version_revision_number.sql` | Adds `revision_number int NOT NULL DEFAULT 0` column; backfills `revision_number = version_number - 1`; creates `UNIQUE INDEX ux_templates_version_revision (template_id, revision_number)` (ADR 0013) |
| `db/baseline/0001_current_schema.sql:2191` | Canonical DDL for all four owned tables; tripwire function references at line 477–482 |

---

## 7. Config & environment

The templates module itself reads **no environment variables or config keys** directly. All config is injected at the composition root (`apps/api/cmd/metaldocs-api/main.go:700-710`):

| Config point | Where resolved | What it controls |
|---|---|---|
| `deps.MinioClient` | `main.go:704` | MinIO client used by `objectstore.NewTemplatesPresigner` |
| `deps.MinioPublicClient` | `main.go:704` | Public-endpoint MinIO client for presigned URL generation |
| `deps.MinioBucket` | `main.go:704` | Bucket name |
| Max object size `25 * 1024 * 1024` bytes | `main.go:704` | Hardcoded constant passed to presigner constructor; enforced at presign time |
| `deps.SQLDB` | `main.go:705` | `*sql.DB`; injected twice — once into `repository.New`, once via `service.WithDB` |
| `deps.AuditWriter` | `main.go:705` | `audit.Writer` injected via `repository.WithAudit` — enables canonical audit sink |
| `capabilityService` | `main.go:706-709` | `AuthzFunc` closure capturing `capabilityService.CheckCapability` for Tier 1 authz |

`autosaveUploadTTL = 10 * time.Minute` and `docxDownloadTTL = 15 * time.Minute` are hardcoded constants in `application/autosave.go:14` and `application/queries.go:39` respectively.

---

## 8. Concurrency & async

- **No goroutines or channels** are spawned within the templates module.
- **No outbox writes, job enqueues, or timers** are present.
- **OCC via `lock_version`:** every mutation through `UpdateVersion[Tx]`, `UpdateVersionDraftCAS[Tx]`, and `UpdateVersionSchemaCAS[Tx]` uses `WHERE lock_version = $N` with auto-increment `lock_version + 1` on success. A zero-rows-affected result is disambiguated by a follow-up `EXISTS` query to distinguish not-found from stale-lock (`ErrStaleLockVersion`).
- **Multi-step atomicity:** create, approve, and publish paths all open a `*sql.Tx` when `s.db != nil`, executing all mutations (obsolete + update template + update version + create next-version + audit) inside a single transaction. `SubmitForReview` and `ArchiveTemplate` execute their single state-mutation + audit as separate statements (the state UPDATE is inside the tx; the `AppendAudit` for `SubmitForReview` is **outside** the tx at `lifecycle.go:92` — a partial-failure gap).
- **Revision number allocation race:** `CreateVersion` uses a subquery `COALESCE(MAX(rn.revision_number)+1, 0)` without an explicit `SELECT … FOR UPDATE`. Two concurrent creates of the same template's next version would both read the same MAX and attempt identical `revision_number` values. The `UNIQUE INDEX ux_templates_version_revision (template_id, revision_number)` on `templates_template_version` makes one fail, but no retry logic exists in the service — the caller sees a raw Postgres uniqueness error, not a domain error. [runtime-unverified: whether this race is observable in practice under current single-writer create patterns]

---

## 9. Error handling & observability

### Error handling

- Domain errors are `errors.New` sentinels defined in `domain/version.go`, `domain/template.go`, `domain/errors.go`. They are all named in `application/errors.go:isDomainErr` and passed through unwrapped by `wrapAppErr`.
- `MapErr` in `delivery/http/errors.go:35` maps the full set of domain sentinels to RFC 9457 HTTP status + problem code pairs. The mapping is exhaustive for all known domain errors; the default arm returns 500 `internal_error`.
- All non-2xx responses use `problem.Write(w, problem.New(status, code, message))` (`delivery/http/handler.go:108`) — RFC 9457 `application/problem+json` is fully adopted. T-005 is confirmed closed.
- `PublishTemplateVersion` uses `log.Printf` (stdlib, no structured logger) for best-effort audit-append failures (`lifecycle.go:422`). This is the only `log.*` call in the module; all other errors are returned up the call stack.

### Observability

- **No structured logging** inside module code. The `slog.Warn` call in `handler.go:110` fires only when `problem.Write` itself fails (infrastructure-level edge case).
- **No metrics** (no Prometheus counters, histograms, or gauges).
- **No distributed traces** (no OpenTelemetry spans).
- **Audit trail** for template events: `AppendAudit[Tx]` delegates to `auditdomain.Writer` (canonical `metaldocs.audit_events`). `ListAudit` reads from the legacy `templates_audit_log` table. These two sinks are now partially disconnected: write path goes to canonical sink; read path (the `GET /{id}/audit` route) reads the legacy table. Events written via the current code would not appear in `GET /{id}/audit` responses unless the canonical sink and the legacy table are kept in sync externally.

---

## 10. Legacy / duplication / smell flags

- **[SMELL-01] `template.admin` is an undeclared capability string** — `delivery/http/routes_lifecycle.go:192` passes `"template.admin"` to `h.authz` for the `PUT /approval-config` route. No `CapTemplateAdmin` constant exists in `internal/modules/iam/domain/model.go` (verified: only 8 `CapTemplate*` constants declared). The capability check will call `CapabilityService.CheckCapability` with an unknown string, meaning the behavior depends on how the capability service handles unrecognized codes. Severity: **high** — authorization boundary is ambiguous.

- **[SMELL-02] Tier 1 cap mismatch on `POST /publish`** — `delivery/http/routes_generated.go:203` uses `h.authz(r, tenantID, "*", "template.approve")` for the publish route, but the route's Tier 2 check (inside the tx) uses `CapTemplatePublish`. The IAM role-capability matrix maps `CapTemplatePublish` and `CapTemplateApprove` differently. An actor with `template.approve` but not `template.publish` clears Tier 1 but fails Tier 2; an actor with `template.publish` but not `template.approve` is blocked by Tier 1. The net effect is that only actors with both pass. Severity: **medium** — intent is obscured; documentation and code disagree. RF-AUTH-01 (backend-target-architecture refactor register).

- **[SMELL-03] `SubmitForReview` audit row outside transaction** — `lifecycle.go:77-94`: the state-update (`UpdateVersionTx`) is inside a `*sql.Tx`, but the subsequent `AppendAudit` call (line 92) is outside the committed transaction. A failure between commit and audit append leaves the version in `in_review` status with no audit row in the canonical sink. Severity: **medium** — partial-failure gap.

- **[SMELL-04] `ObsoletePreviousPublished` has no per-row `lock_version` guard** — `repository/postgres.go:544-556`: bulk UPDATE sets all previously-published versions to `obsolete` based on `WHERE template_id = $1 AND status = 'published' AND id <> $2`. No `lock_version` comparison. Concurrent publish of two drafts could have one obsoleting a version the other just published. The outer tx's `UpdateVersionTx` CAS would catch a stale loser on the new version, but the obsolete UPDATE runs before that check. Severity: **medium** — RF-CONCURRENCY.

- **[SMELL-05] `templates_audit_log` read/write sink divergence** — `repository/postgres.go:631-674`: `AppendAudit[Tx]` writes to the canonical `metaldocs.audit_events` via `auditdomain.Writer`. `ListAudit` reads from `templates_audit_log` (`repository/postgres.go:676`). These two sinks are now disconnected: `GET /{id}/audit` returns an empty or stale list unless `templates_audit_log` is populated by other means. T-013 was closed (write path migrated) but the read path still points at the legacy table. Severity: **high** — audit trail is effectively broken for the list endpoint.

- **[SMELL-06] Hard-coded `approverRole: "approver"` in `CreateTemplate`** — `routes_generated.go:56`: the `CreateTemplate` handler passes `ApproverRole: "approver"` regardless of the request body. The request body's `approver_role` field (if any) is silently ignored. The caller cannot specify the approver role at template creation time via the API; it must use `PUT /approval-config` afterward. Severity: **medium** — undocumented API behavior; callers may expect otherwise.

- **[SMELL-07] Legacy `areas`, `visibility`, `specific_areas` columns written but never read** — `repository/postgres.go:52-53`: INSERT hard-codes `areas='{}'::text[]`, `visibility='public'`, `specific_areas='{}'::text[]`. `scanTemplate` does not scan these columns. The domain `Template` struct has no corresponding fields. These columns date from the pre-Plan-5 visibility-as-creator-scoped-permission era. They are still present in `db/baseline/0001_current_schema.sql:2198-2200` and consume storage on every row. Severity: **low** — dead schema surface; RF-SCHEMA-01.

- **[SMELL-08] `CompositionConfig` struct retained for backward compat with no active consumer** — `domain/schemas.go:81`: the ADR 0008 comment notes composition was removed 2026-04-27. The struct is exported and has no inbound callers (grep-verified). Severity: **low** — dead exported type.

- **[SMELL-09] `revision_number` subquery allocation is not race-free under concurrent `CreateVersion`** — `repository/postgres.go:202`: `COALESCE(MAX(rn.revision_number)+1, 0)` is a non-locked read inside a tx. Two concurrent create-version transactions for the same template can read the same MAX and attempt the same `revision_number`. The UNIQUE index (`ux_templates_version_revision`) makes one fail with a raw Postgres constraint error that is not mapped to a domain error — the service returns an unwrapped `pq` error, not `ErrKeyConflict` or `ErrStaleLockVersion`. Severity: **medium** — RF-CONCURRENCY.

- **[SMELL-10] `PublishTemplateVersion` updates `UpdateTemplateTx` twice** — `lifecycle.go:456-468`: first `UpdateTemplateTx` sets `PublishedVersionID`/`PublishedVersionNumber`/`CurrentRevisionNumber`; second `UpdateTemplateTx` (line 468) sets `LatestVersion`. Both are inside the same transaction so there is no inconsistency, but the duplicate UPDATE is unnecessary and makes the code harder to follow. Severity: **low** — code clarity.

- **[SMELL-11] `TODO(pagination)` in `listAudit`** — `delivery/http/routes_query.go:222`: explicit TODO noting audit listing should migrate to keyset pagination. Current implementation uses `LIMIT/OFFSET` which is correct for bounded admin queries but inconsistent with the cursor pattern available elsewhere (T-011 analogue for audit). Severity: **low**.

- **[SMELL-12] All exported symbols lack Go doc comments** — T-014 (open). Every exported type, function, method, and constant in `domain/`, `application/`, `delivery/http/`, and `repository/` has no Go doc comment. `golint`/`revive` would flag the entire module. Severity: **low**.

- **[SMELL-13] `delivery/http/routes_query.go:13-14` hardcoded system-blank template UUIDs** — `systemBlankTemplateTenantID = "ffffffff-ffff-ffff-ffff-ffffffffffff"` and `systemBlankTemplateID = "00000000-0000-0000-0000-000000000101"` are string literals in handler code. These should be config or seed-driven constants. Severity: **low**.

---

## 11. Wiki drift

1. **`wiki/modules/templates.md §3.2` claims `GetVersion` has no tenant argument**: the doc (line 111) states "20 HTTP routes mounted at `/api/v1/templates/*`" — the code actually mounts 22 routes (the `Register` method in `handler.go:34-65` lists 22 `mux.Handle*` calls, including `GET /api/v1/templates/system/blank` and `GET /api/v1/signed`). The route count in the wiki is stale.

2. **`wiki/modules/templates.md §6.1 ListTemplates flow diagram` shows "no LIMIT"**: the comment and sequence diagram in the wiki (line 256) claims `SELECT FROM templates_template WHERE tenant_id = $1 AND ... (no LIMIT)`. The actual query in `repository/postgres.go:114-127` has `LIMIT $3 OFFSET $4` and the handler passes `limit`/`offset` from query parameters (capped at 200). T-011 listed in tech-debt as an open gap is actually closed in the repository layer; the wiki diagram is wrong.

3. **`wiki/modules/templates-tech-debt.md §T-013` status**: the tech-debt register marks T-013 as "CLOSED 2026-05-11 (Plan 6a)" and states `AppendAudit` now delegates to the canonical sink. This is confirmed correct in the code. However, it does not note that `ListAudit` **still reads from `templates_audit_log`**, not from `metaldocs.audit_events`. The T-013 closure note is incomplete — the sink divergence (write to canonical, read from legacy) is a real remaining gap not called out in the register.

4. **`wiki/modules/templates.md §5.3 HTTP operations table` "Publish" authz note**: the table says `service CapTemplatePublish` for the publish route. The Tier 1 handler authz is actually `template.approve` (not `template.publish`) per `routes_generated.go:203`. The wiki's Tier 1 description is incorrect.

5. **`wiki/modules/templates.md §8.6` claims `SubmitForReview` audit is inside tx**: the section states multi-step publish operations emit statements as independent calls. However the more specific concern is that `SubmitForReview`'s `AppendAudit` call at `lifecycle.go:92` runs **after** `tx.Commit()` at line 70 — it is genuinely outside the transaction. The wiki does not call this out explicitly; it lists `AppendAudit` as always outside tx for all ops, which overstates the gap (approve and publish paths do use `AppendAuditTx`).

---

## 12. Open questions

1. **[runtime-unverified]** `ResolverRegistryReader` is `nil` at composition root (`main.go:705` passes no variadic). The resolver-key validation branch in `ValidatePlaceholders` is therefore always skipped at runtime. The template injection blast radius (arbitrary `resolver_key` propagates to every document from a published version) cannot be confirmed without a live environment, but the code path is clear from reading.

2. **[runtime-unverified]** Whether `templates_audit_log` still receives writes from any other code path (migration, seed, or legacy service instance). If not, `GET /{id}/audit` returns an empty list for all templates created after the Plan 6a migration. Verification requires a live DB query.

3. **[runtime-unverified]** Behavior of `CapabilityService.CheckCapability` when called with `"template.admin"` (undeclared capability). If it returns false/denied for unknown caps, `PUT /approval-config` is permanently inaccessible to all users. If it returns permitted (open-world assumption), the approval-config route is unprotected at Tier 1.

4. **[runtime-unverified]** The `ux_templates_version_revision` UNIQUE index enforces uniqueness of `(template_id, revision_number)`. Concurrent `CreateVersionTx` calls for the same template will have one fail with a raw pgx constraint error. Whether this surfaces as a 500 or is mapped to a friendlier error code requires a live concurrent test.

5. **[runtime-unverified]** The `GET /api/v1/templates/system/blank` route accesses a fixed tenant UUID (`ffffffff-ffff-ffff-ffff-ffffffffffff`) and template UUID (`00000000-0000-0000-0000-000000000101`). Whether the corresponding seed rows exist in production/dev and remain consistent across environments requires runtime verification.
