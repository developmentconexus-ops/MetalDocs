# Refactor Backlog â€” templates

> Actionable rows. One row = one PR. Pulled from `wiki/modules/templates-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** 2026-05-12 (Plan 7)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Wire `CapabilityService` into `tv2http.New` and call `authz.Require(cap, area)` in every `Service` mutation | T-001 | M | Critical | â€” | â€” | merged | Plan 5 (2026-05-11): `WithDB` builder + `authz.Require` in all 6 mutation paths + tripwire on both templates tables |
| R-002 | Add `tenant_id` argument to `Repository.GetVersion` / `GetVersionByID`; make `CreateNextVersion` pass tenant when cloning from `PublishedVersionID` | T-002 | S | Critical | â€” | â€” | merged (partial) | Plan 5 (2026-05-11): CreateNextVersion guard added; repo getter signatures unchanged (residual) |
| R-003 | Replace `tenantIDFromReq` header trust with subject-claim derived tenant; remove `DevTenantID` fallback in non-dev environments | T-003 | S | Critical | R-001 | â€” | **done** | Plan 3 (2026-05-11) |
| R-004 | Make `PublishTemplateVersion` route to `Service.Approve` (or fold the path) so SoD + role + content_hash gates always run on publish | T-004 | M | Critical | â€” | â€” | merged (partial) | Plan 5 (2026-05-11): content_hash + SoD + authz.Require added; role-binding check against pending_approver_role still absent |
| R-005 | Migrate templates error responses to `internal/platform/problem` (RFC 9457) and update `MapErr` to emit Problem documents | T-005 | S | Major | â€” | â€” | merged | Plan 7 (2026-05-11, commit bbe3933b) |
| R-006 | Add the 12 hand-rolled routes to the OpenAPI spec; regenerate `api.gen.go`; replace hand-rolled handlers with strict-server impls | T-006 | L | Major | â€” | â€” | open | â€” |
| R-007 | Introduce `Repository.WithTx` and wrap `PublishTemplateVersion` / `Approve` / `CreateTemplate` in a single `pgx.Tx`; emit `AuditObsoleted` for the obsolete side-effect | T-007 | M | Major | â€” | â€” | open | â€” |
| R-008 | Inject `ResolverRegistryReader` at `tv2app.New` in composition root; have `ValidatePlaceholders` reject unknown `resolver_key` for `PHComputed` | T-008 | S | Major | â€” | â€” | open | â€” |
| R-009 | Verify `internal/platform/idempotency` replay semantics on generated POST routes (`/templates`, `/publish`, `/submit`, `/review`, `/approve`) and classify remaining POST mutation surfaces | T-009 | M | Major | - | - | open (partial wrapper exists) | Plan 12.4 verified create path sends `Idempotency-Key` and receives HTTP 201; replay audit still pending |
| R-010 | Enforce `ExpectedLockVersion` in `SaveTemplateDraft` â€” UPDATE â€¦ WHERE lock_version = $X; return 412 on mismatch | T-010 | S | Major | â€” | â€” | open | â€” |
| R-011 | Add cursor pagination (Plan 2 cursor primitive) to `ListTemplates`; default page size 50 | T-011 | S | Minor | â€” | â€” | open | â€” |
| R-012 | Drop `templates_template_version.editable_zones` column (verify 0157 effect; emit corrective migration if needed) | T-012 | XS | Minor | â€” | â€” | open | â€” |
| R-013 | Replace `templates_audit_log` writes with canonical `metaldocs.audit_events` sink; backfill historical rows; drop the local table | T-013 | M | Minor | â€” | â€” | merged | Plan 6a (2026-05-11, commit 71a2dc53) |
| R-014 | Add Go doc comments to every exported symbol under `internal/modules/templates/{domain,application,delivery,repository}/` | T-014 | S | Minor | â€” | â€” | open | â€” |
| R-100 | Retire predecessor frontend-heavy stub `wiki/modules/templates.md` (kebab) and repoint inbound links to `wiki/modules/templates.md` | maint:doc-cleanup | XS | Minor | â€” | â€” | open | â€” |
| R-101 | Rename module dir `internal/modules/templates/` â†’ `internal/modules/templates/`, flip routes `/api/v1/templates` â†’ `/api/v1/templates`, rename wiki doc to `templates.md` (single follow-up commit) | maint:migration-cleanup | M | Minor | R-006 | â€” | open | â€” |

## Notes

- R-006 is `L` due to spec authoring + handler migration; if grooming pulls it in, split per route group (lifecycle / autosave / catalog / archive).
- R-001 + R-003 + R-004 should land before public exposure â€” these are the multi-tenant cutover blockers.
- R-101 deferred until `/api/v1/` flip + dir rename can be done atomically (touches frontend `lib/api-types/`); blocked on R-006 so the spec captures the right URL shape.
