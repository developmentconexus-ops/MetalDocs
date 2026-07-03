# Refactor Backlog — templates

> Actionable rows. One row = one PR. Pulled from `wiki/modules/templates-tech-debt.md`. Rows that lack a debt-id are blocked from grooming.

**Last verified:** 2026-07-02 (DOC-07d — R-100 closed as moot). Prior: 2026-06-21 (verify-and-archive sweep — solved rows pruned; see _cleanup-2026-06-21.md)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-007 | Introduce `Repository.WithTx` and wrap `PublishTemplateVersion` / `Approve` / `CreateTemplate` in a single `pgx.Tx`; emit `AuditObsoleted` for the obsolete side-effect | T-007 | M | Major | — | — | open | — |
| R-009 | Verify `internal/platform/idempotency` replay semantics on generated POST routes (`/templates`, `/publish`, `/submit`, `/review`, `/approve`) and classify remaining POST mutation surfaces | T-009 | M | Major | - | - | open (partial wrapper exists) | Plan 12.4 verified generated create path requires/sends `Idempotency-Key` and receives HTTP 201; same-key replay audit still pending |
| R-011 | Add cursor pagination (Plan 2 cursor primitive) to `ListTemplates`; default page size 50 | T-011 | S | Minor | — | — | open | — |
| R-014 | Add Go doc comments to every exported symbol under `internal/modules/templates/{domain,application,delivery,repository}/` | T-014 | S | Minor | — | — | open | — |
| R-100 | Retire predecessor frontend-heavy stub `wiki/modules/templates.md` (kebab) and repoint inbound links to `wiki/modules/templates.md` | maint:doc-cleanup | XS | Minor | — | — | closed 2026-07-02 (moot) | — |
| R-101 | Correct/retire this row — original rename intent lost (source==target); re-derive or close | maint:migration-cleanup | M | Minor | R-006 | — | open | — |

## R-100 closure evidence (DOC-07d)

- The row's premise ("predecessor frontend-heavy stub `wiki/modules/templates.md`... repoint to `wiki/modules/templates.md`") was itself a mojibake corruption artifact — both the source and target paths printed identically because a mojibake em-dash/arrow sequence had eaten the distinguishing `_v2` suffix from the historical text.
- Root-caused via git history: the actual predecessor was the **code** module `internal/modules/templates_v2/`, renamed to `internal/modules/templates/` in commit `801e8541` (2026-05-13). No separate `wiki/modules/templates_v2.md` (or any second templates wiki page) was ever committed — confirmed via `find wiki -iname "*templates_v2*"` (zero hits) and a full listing of `wiki/modules/templates*` (single `templates.md` + `templates-tech-debt.md` + `templates/_artifacts/`).
- There is nothing to retire: no orphan file exists on disk. Closed as moot rather than "done" since no wiki-cleanup action was actually required beyond correcting the corrupted self-referential text.
- Fixed the two corrupted lines in `wiki/modules/templates.md` (§ cross-refs "Predecessor doc" line, and the 2026-05-10 changelog entry) to state the real history in valid UTF-8, replacing the self-referential "templates.md supersedes templates.md" artifact.
- `wiki/modules/frontend/templates.md` was checked separately and is a live, current, correctly-scoped frontend module doc (not a stub) — it is not related to R-100's target and was left untouched.

## Notes

- 2026-05-17 product/API note: creator-scoped template-use `visibility`, `areas`, and `specific_areas` were removed from runtime/API selection behavior. The database columns remain inert compatibility fields until a coordinated baseline/reference-data cleanup is planned.
- R-101 deferred until `/api/v1/` flip + dir rename can be done atomically (touches frontend `lib/api-types/`); Plan 12.4 removed the R-006 contract coverage blocker.
