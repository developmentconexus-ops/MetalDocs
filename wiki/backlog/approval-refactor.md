# Refactor Backlog — approval

> Actionable rows. One row = one PR. Pulled from `wiki/modules/approval-tech-debt.md`.

**Last verified:** 2026-05-25 (5c + 5d medium sweep sync)

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Migrate approval HTTP error envelope to RFC 9457 Problem+JSON | T-001 | M | critical | — | — | merged | Plan 7 (2026-05-11, commit b8747d6a + c4f5535f) |
| R-002 | Add OpenAPI operationIds + schemas for v2 doc-action approval routes | T-002 | M | critical | — | — | open | — |
| R-003 | Replace `looksLikeValidationError` substring classifier with typed sentinel matching | T-003 | S | major | — | — | merged | Plan 7 (2026-05-11, commit b8747d6a) |
| R-004 | Remove deprecated post-commit `PDFDispatchInvoker` path; require outbox | T-004 | S | major | — | — | open | — |
| R-005 | Convert inbox list+count to single tx (or single window-function query) | T-005 | S | major | — | — | open | — |
| R-006 | Audit cancel & cutover service paths for `authz.Require` pairing | T-006 | S | major | 2026-05-25 | 5c high hardening sync | closed | `cancel_service.go` tenant-scopes cancel stages and keeps system bypass internal; `cutover_service.go` uses system-bypassed tx preflight |
| R-007 | Consolidate `infra/signature/` and `infrastructure/` packages | T-007 | XS | minor | — | — | open | — |
| R-008 | Rename `approval_instances.document_v2_id` → `document_id` (migration) | T-008 | S | minor | — | — | merged | migration 0194 |
| R-009 | Validate `NOT VALID` FKs on `submitted_by` and `actor_user_id` | T-009 | XS | minor | — | — | open | — |
| R-010 | Add Go doc comments to exported approval symbols | T-010 | M | minor | — | — | open | — |
| R-011 | Remove `WithMembershipContext` in favor of `setAuthzGUC` | T-011 | XS | minor | — | — | open | — |
| R-012 | Close approval-tech-debt T-012 — iam `AuthorizationService` deleted Plan 4; update cross-module debt row; no adoption path needed | T-012 | XS | minor | — | — | open | — |

## Notes

- R-001 + R-002 are both `M` and depend on a shared envelope/spec decision — sequence them: R-002 first (lock the schema), then R-001 (server emits matching shape).
- R-006 is gated by extending `tally_check.sh` audit to non-test service files — keep diff small, no new lint dependencies.
- R-008 completed in migration 0194; repository column surface is now `document_id`.
