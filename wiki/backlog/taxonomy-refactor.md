# Refactor Backlog — taxonomy

> Actionable rows. One row = one PR. Pulled from `wiki/modules/taxonomy-tech-debt.md`.

**Last verified:** 2026-07-16 (ROADMAP unit 4.5 — R-015 closed, `document_profiles` PK promoted by migration 0308). Prior: 2026-07-02 (DOC-07c — R-017 closed; CON-08 closure — R-009 closed; R-012 unblocked). Prior: 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-002 | Add ADR + migration: backfill tenant_id on document_families OR document the global-by-design choice with a threat model and lock-down policy | T-002 | L | Critical | — | — | open | — |
| R-006 | Add tier-2 authz.Require + DB tripwire (assert_caps) on document_profiles, document_process_areas, document_families | T-006 | L | Major | R-001, R-002 | — | merged (partial) | Plan 5 (2026-05-11): Create+Update methods + tripwire on all 3 tables done; archive/deactivate paths residual |
| R-009 | Author OpenAPI spec for /api/v1/taxonomy/* and re-mount routes via oapi-codegen | T-009 | L | Major | — | — | closed | CON-08 (2026-07-02): spec/codegen/mount were already landed pre-existing; this pass closed the residual wire-truth drift — `include_archived` query param + `TaxonomyProfileUpsertRequest`/`TaxonomyAreaUpsertRequest`/`TaxonomyFamilyUpsertRequest`/`SetTaxonomyProfileDefaultTemplateRequest` request bodies + full `ProcessAreaItem`/`DocumentFamilyItem` field sets were undeclared; regenerated `api.gen.go`; added `router_test.go` registration pin test |
| R-012 | Add cursor pagination to listProfiles / listAreas / listFamilies | T-012 | M | Minor | — | — | open | R-009 closed 2026-07-02; no longer blocked |
| R-014 | Add Go doc comments to all 80 exported symbols under internal/modules/taxonomy/ | T-014 | M | Minor | — | — | open | — |
| R-015 | Drop redundant PK on `code` alone; promote `(tenant_id, code)` to PK on document_profiles + document_process_areas | T-015 | M | Minor | R-002 | — | closed 2026-07-16 | `document_process_areas` closed 2026-07-02 (migration 0264); `document_profiles` closed 2026-07-16 (migration 0308, ROADMAP unit 4.5 — also dropped 4 dead FK-blocking tables, see `wiki/modules/taxonomy-tech-debt.md` T-015) |
| R-016 | Author ADR for area hierarchy: self-FK + application-layer cycle prevention | T-016 | S | Minor | — | — | open | — |
| R-017 | Retire the 2026-05-02 taxonomy stub references in cross-link search-results; verify the new doc renders correctly in the wiki index | maint:docs-link | XS | Minor | — | — | closed 2026-07-02 | — |

## R-017 closure evidence (DOC-07c)

- `wiki/modules/taxonomy.md` verified as the current living doc — Last verified 2026-06-12 (Wave 2.12), own changelog (`taxonomy.md:469`) correctly records it superseded the 2026-05-02 stub on 2026-05-11.
- `wiki/modules/index.md:14` verified pointing to the live `taxonomy.md`/`taxonomy-tech-debt.md` (not any stub path) — renders correctly.
- The stale cross-link was `wiki/modules/taxonomy/_artifacts/00-context.md:5` ("Stub: `wiki/modules/taxonomy.md` (Last verified 2026-05-02 — STALE)") — a Phase-0 historical artifact whose "STALE" label could be misread as describing current `taxonomy.md` state. Added a retirement/pointer note at the top of that file directing readers to the live doc; did not rewrite the historical Phase-0 content itself (artifact is a point-in-time record, same convention as `iam/_artifacts/06-selfreview.md`).
- No other live cross-links to the 2026-05-02 stub found (`wiki/backend/roadmap.md` and `wiki/modules/taxonomy/_artifacts/05-industry.md` hits were unrelated "stub" mentions, not the 2026-05-02 date).

## Notes

- R-002 / R-006 are `L`-effort — split before opening PRs.
