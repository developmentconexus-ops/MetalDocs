# Refactor Backlog — taxonomy

> Actionable rows. One row = one PR. Pulled from `wiki/modules/taxonomy-tech-debt.md`.

**Last verified:** 2026-05-11 (Plan 5)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Verify authenticated user belongs to claimed tenant before honouring X-Tenant-ID; reject on mismatch | T-001 | M | Critical | — | — | **done** | Plan 3 (2026-05-11) |
| R-002 | Add ADR + migration: backfill tenant_id on document_families OR document the global-by-design choice with a threat model and lock-down policy | T-002 | L | Critical | — | — | open | — |
| R-003 | Add MethodPatch to the families branch of permissions.go path-prefix dispatcher (line 174-180) | T-003 | XS | Critical | — | — | merged | Plan 5 (2026-05-11) |
| R-004 | Wire DBGovernanceLogger into FamilyService; emit on Create/Update/Deactivate | T-004 | S | Critical | — | — | open | — |
| R-005 | Emit governance events on ProfileService.Create/Update and AreaService.Create/Update | T-005 | S | Critical | — | — | open | — |
| R-006 | Add tier-2 authz.Require + DB tripwire (assert_caps) on document_profiles, document_process_areas, document_families | T-006 | L | Major | R-001, R-002 | — | merged (partial) | Plan 5 (2026-05-11): Create+Update methods + tripwire on all 3 tables done; archive/deactivate paths residual |
| R-007 | Wrap FamilyService.Deactivate in a single tx with row lock; add tenant_id predicate to HasActiveProfiles | T-007 | M | Major | — | — | open | — |
| R-008 | Migrate taxonomy error responses to RFC 9457 Problem+JSON | T-008 | M | Major | — | — | open | — |
| R-009 | Author OpenAPI spec for /api/v2/taxonomy/* and re-mount routes via oapi-codegen | T-009 | L | Major | — | — | open | — |
| R-010 | Unify governance_events under audit.Writer or write an ADR justifying the parallel sink | T-010 | L | Major | — | — | open | — |
| R-011 | Add 23505 → 409 RESOURCE_CONFLICT mapping in writeProfileError / writeFamilyError / writeAreaError | T-011 | XS | Minor | R-008 | — | open | — |
| R-012 | Add cursor pagination to listProfiles / listAreas / listFamilies | T-012 | M | Minor | R-009 | — | open | — |
| R-013 | Add DB BEFORE-UPDATE trigger rejecting `NEW.code <> OLD.code` on document_families | T-013 | XS | Minor | — | — | merged | Plan 5 (2026-05-11) |
| R-014 | Add Go doc comments to all 80 exported symbols under internal/modules/taxonomy/ | T-014 | M | Minor | — | — | open | — |
| R-015 | Drop redundant PK on `code` alone; promote `(tenant_id, code)` to PK on document_profiles + document_process_areas | T-015 | M | Minor | R-002 | — | open | — |
| R-016 | Author ADR for area hierarchy: self-FK + application-layer cycle prevention | T-016 | S | Minor | — | — | open | — |
| R-017 | Retire the 2026-05-02 taxonomy stub references in cross-link search-results; verify the new doc renders correctly in the wiki index | maint:docs-link | XS | Minor | — | — | open | — |

## Notes

- R-002 / R-006 / R-009 / R-010 are `L`-effort — split before opening PRs.
- R-003 is the cheapest Critical fix in the register; recommend it ships first as a standalone PR.
