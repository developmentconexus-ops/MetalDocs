# Refactor Backlog — templates

> Actionable rows. One row = one PR. Pulled from `wiki/modules/templates-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** 2026-05-31 (fix/templates-publish-role-binding — R-004 closed)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Wire `CapabilityService` into `tv2http.New` and call `authz.Require(cap, area)` in every `Service` mutation | T-001 | M | Critical | — | — | merged | Plan 5 + 2026-05-17 extension: `WithDB` builder + `authz.Require`; DOCX import/autosave commit now asserts `template.edit`; tripwire on both templates tables |
| R-002 | Add `tenant_id` argument to `Repository.GetVersion` / `GetVersionByID`; make `CreateNextVersion` pass tenant when cloning from `PublishedVersionID` | T-002 | S | Critical | — | — | merged (partial) | Plan 5 (2026-05-11): CreateNextVersion guard added; repo getter signatures unchanged (residual) |
| R-003 | Replace `tenantIDFromReq` header trust with subject-claim derived tenant; remove `DevTenantID` fallback in non-dev environments | T-003 | S | Critical | R-001 | — | **done** | Plan 3 (2026-05-11) |
| R-004 | Make `PublishTemplateVersion` route to `Service.Approve` (or fold the path) so SoD + role + content_hash gates always run on publish | T-004 | M | Critical | — | — | **done** | Plan 5 (2026-05-11): content_hash + SoD + authz.Require added. `fix/templates-publish-role-binding` (2026-05-31): role-binding gate added via `domain.TemplateVersion.RoleBindingFor`; `Service.Approve` refactored to share the helper; denied attempts audited as `publish_forbidden_role`; contract test covers RFC 9457 `forbidden_role` 403 |
| R-005 | Migrate templates error responses to `internal/platform/problem` (RFC 9457) and update `MapErr` to emit Problem documents | T-005 | S | Major | — | — | merged | Plan 7 (2026-05-11, commit bbe3933b) |
| R-006 | Add the 12 hand-rolled routes to the OpenAPI spec; regenerate `api.gen.go`; replace hand-rolled handlers with strict-server impls | T-006 | L | Major | — | — | merged (contract coverage) | Plan 12.4 (2026-05-16): bundled/partial OpenAPI, backend generated API, and frontend generated API types now cover the mounted route set; strict-server cleanup can be tracked separately |
| R-007 | Introduce `Repository.WithTx` and wrap `PublishTemplateVersion` / `Approve` / `CreateTemplate` in a single `pgx.Tx`; emit `AuditObsoleted` for the obsolete side-effect | T-007 | M | Major | — | — | open | — |
| R-008 | Inject `ResolverRegistryReader` at `tv2app.New` in composition root; have `ValidatePlaceholders` reject unknown `resolver_key` for `PHComputed` | T-008 | S | Major | — | — | open | — |
| R-009 | Verify `internal/platform/idempotency` replay semantics on generated POST routes (`/templates`, `/publish`, `/submit`, `/review`, `/approve`) and classify remaining POST mutation surfaces | T-009 | M | Major | - | - | open (partial wrapper exists) | Plan 12.4 verified generated create path requires/sends `Idempotency-Key` and receives HTTP 201; same-key replay audit still pending |
| R-010 | Enforce `ExpectedLockVersion` in `SaveTemplateDraft` — UPDATE … WHERE lock_version = $X; return 412 on mismatch | T-010 | S | Major | — | fix/templates-schema-occ-lock | closed | 2026-05-17: `SaveTemplateDraft` CAS landed; 2026-05-31: PUT `/schema` now also lock-version gated (request body requires `expected_lock_version`; `UpdateVersionSchemaCAS`; FE surfaces 412 `stale_lock_version` instead of silent overwrite). Legacy `/autosave/commit` still hash-gated — tracked separately. |
| R-011 | Add cursor pagination (Plan 2 cursor primitive) to `ListTemplates`; default page size 50 | T-011 | S | Minor | — | — | open | — |
| R-012 | Drop `templates_template_version.editable_zones` column (verify 0157 effect; emit corrective migration if needed) | T-012 | XS | Minor | — | — | open | — |
| R-013 | Replace `templates_audit_log` writes with canonical `metaldocs.audit_events` sink; backfill historical rows; drop the local table | T-013 | M | Minor | — | — | merged | Plan 6a (2026-05-11, commit 71a2dc53) |
| R-014 | Add Go doc comments to every exported symbol under `internal/modules/templates/{domain,application,delivery,repository}/` | T-014 | S | Minor | — | — | open | — |
| R-100 | Retire predecessor frontend-heavy stub `wiki/modules/templates.md` (kebab) and repoint inbound links to `wiki/modules/templates.md` | maint:doc-cleanup | XS | Minor | — | — | open | — |
| R-101 | Rename module dir `internal/modules/templates/` → `internal/modules/templates/`, flip routes `/api/v1/templates` → `/api/v1/templates`, rename wiki doc to `templates.md` (single follow-up commit) | maint:migration-cleanup | M | Minor | R-006 | — | open | — |

## Notes

- 2026-05-17 product/API note: creator-scoped template-use `visibility`, `areas`, and `specific_areas` were removed from runtime/API selection behavior. The database columns remain inert compatibility fields until a coordinated baseline/reference-data cleanup is planned.
- R-006 is closed for route/spec/generated coverage in Plan 12.4. A future strict-server cleanup can be opened as a narrower follow-up if desired.
- R-001 + R-003 + R-004 should land before public exposure — these are the multi-tenant cutover blockers.
- R-101 deferred until `/api/v1/` flip + dir rename can be done atomically (touches frontend `lib/api-types/`); Plan 12.4 removed the R-006 contract coverage blocker.
