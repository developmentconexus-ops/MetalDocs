# Refactor Backlog — templates_v2

> Actionable rows. One row = one PR. Pulled from `wiki/modules/templates_v2-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** 2026-05-11 (Plan 6a)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Wire `CapabilityService` into `tv2http.New` and call `authz.Require(cap, area)` in every `Service` mutation | T-001 | M | Critical | — | — | merged | Plan 5 (2026-05-11): `WithDB` builder + `authz.Require` in all 6 mutation paths + tripwire on both templates_v2 tables |
| R-002 | Add `tenant_id` argument to `Repository.GetVersion` / `GetVersionByID`; make `CreateNextVersion` pass tenant when cloning from `PublishedVersionID` | T-002 | S | Critical | — | — | merged (partial) | Plan 5 (2026-05-11): CreateNextVersion guard added; repo getter signatures unchanged (residual) |
| R-003 | Replace `tenantIDFromReq` header trust with subject-claim derived tenant; remove `DevTenantID` fallback in non-dev environments | T-003 | S | Critical | R-001 | — | **done** | Plan 3 (2026-05-11) |
| R-004 | Make `PublishTemplateVersion` route to `Service.Approve` (or fold the path) so SoD + role + content_hash gates always run on publish | T-004 | M | Critical | — | — | merged (partial) | Plan 5 (2026-05-11): content_hash + SoD + authz.Require added; role-binding check against pending_approver_role still absent |
| R-005 | Migrate templates_v2 error responses to `internal/platform/problem` (RFC 9457) and update `MapErr` to emit Problem documents | T-005 | S | Major | — | — | open | — |
| R-006 | Add the 12 hand-rolled routes to the OpenAPI spec; regenerate `api.gen.go`; replace hand-rolled handlers with strict-server impls | T-006 | L | Major | — | — | open | — |
| R-007 | Introduce `Repository.WithTx` and wrap `PublishTemplateVersion` / `Approve` / `CreateTemplate` in a single `pgx.Tx`; emit `AuditObsoleted` for the obsolete side-effect | T-007 | M | Major | — | — | open | — |
| R-008 | Inject `ResolverRegistryReader` at `tv2app.New` in composition root; have `ValidatePlaceholders` reject unknown `resolver_key` for `PHComputed` | T-008 | S | Major | — | — | open | — |
| R-009 | Wire `internal/platform/idempotency` middleware into POST routes (`/templates`, `/publish`, `/submit`, `/approve`); persist key + first-response within 24h | T-009 | M | Major | — | — | open | — |
| R-010 | Enforce `ExpectedLockVersion` in `SaveTemplateDraft` — UPDATE … WHERE lock_version = $X; return 412 on mismatch | T-010 | S | Major | — | — | open | — |
| R-011 | Add cursor pagination (Plan 2 cursor primitive) to `ListTemplates`; default page size 50 | T-011 | S | Minor | — | — | open | — |
| R-012 | Drop `templates_v2_template_version.editable_zones` column (verify 0157 effect; emit corrective migration if needed) | T-012 | XS | Minor | — | — | open | — |
| R-013 | Replace `templates_v2_audit_log` writes with canonical `metaldocs.audit_events` sink; backfill historical rows; drop the local table | T-013 | M | Minor | — | — | merged | Plan 6a (2026-05-11, commit 71a2dc53) |
| R-014 | Add Go doc comments to every exported symbol under `internal/modules/templates_v2/{domain,application,delivery,repository}/` | T-014 | S | Minor | — | — | open | — |
| R-100 | Retire predecessor frontend-heavy stub `wiki/modules/templates-v2.md` (kebab) and repoint inbound links to `wiki/modules/templates_v2.md` | maint:doc-cleanup | XS | Minor | — | — | open | — |
| R-101 | Rename module dir `internal/modules/templates_v2/` → `internal/modules/templates/`, flip routes `/api/v2/templates` → `/api/v1/templates`, rename wiki doc to `templates.md` (single follow-up commit) | maint:migration-cleanup | M | Minor | R-006 | — | open | — |

## Notes

- R-006 is `L` due to spec authoring + handler migration; if grooming pulls it in, split per route group (lifecycle / autosave / catalog / archive).
- R-001 + R-003 + R-004 should land before public exposure — these are the multi-tenant cutover blockers.
- R-101 deferred until `/api/v1/` flip + dir rename can be done atomically (touches frontend `lib/api-types/`); blocked on R-006 so the spec captures the right URL shape.
