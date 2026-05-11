# Refactor Backlog — documents

> One row = one PR. Pulled from `wiki/modules/documents-tech-debt.md`.

**Last verified:** 2026-05-11

## Rows

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Migrate documents handlers to RFC 9457 Problem+JSON via `httpresponse.WriteProblem` | T-001 | M | major | R-002 | — | open | — |
| R-002 | Reconcile spec ↔ handlers: add ops for renameDocument, duplicateDocument, archiveDocument, comments CRUD; set `operationId: finalizeDocument`; rename spec `listDocumentsV2` → `listDocuments` | T-002 | L | critical | — | — | open | — |
| R-003 | Add `enforce_capability_asserted` trigger to `documents` table; wire `authz.Require("doc.write", areaCode)` into CreateDocumentTx / UpdateDocumentName / UpdateDocumentStatus / MarkArchived / Unarchive | T-003 | M | major | — | — | open | — |
| R-004 | Remove duplicate `PATCH /api/v2/documents/{id}` registration at `handler.go:86` | T-004 | XS | minor | — | — | open | — |
| R-005 | Wrap `RenameDocument` UPDATE + `audit.Write` in single `BeginTx` transaction | T-005 | S | major | — | — | open | — |
| R-006 | Wire `Idempotency-Key` middleware on `POST /api/v2/documents/{id}/finalize` against `metaldocs.idempotency_keys` store | T-006 | M | major | — | — | open | — |
| R-008 | Migrate `submit_service.go:85` `authz.Require("doc.submit", …)` call site to typed `iamdomain.Capability` (paired with iam T-001 closure) | T-008 | S | minor | — | — | merged | Plan 4 (2026-05-11, commit 3a227642) |
| R-009 | Fix `document_placeholder_values.revision_id` FK to point at `document_revisions(id)`; new migration, no edits to 0152 | T-009 | S | major | — | — | open | — |
| R-100 | Retire `wiki/modules/documents-v2.md` deprecated stub (post-rename 0167/0168) | maint:doc-cleanup | XS | minor | — | — | done | Phase 7 wiki-curator 2026-05-10 |

## Notes

- T-007 and T-010 have no backlog row yet — they are latent / gated on prior closures (audit-domain decoupling, codegen migration).
- R-002 (Critical) is the gating row for R-001 (RFC 9457 migration cannot land cleanly while spec ↔ handler drift exists).
