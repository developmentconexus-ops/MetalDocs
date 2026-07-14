# ADR 0056 — Template comments: defer a generic platform, reuse the document-comments table pattern

- **Status:** Accepted
- **Last verified:** 2026-07-02
- **Date:** 2026-07-02
- **Scope:** Resolves the 3-option decision point at `wiki/backlog/template-editor.md:21-29` ("comments" section) for template review-feedback threads. Does not implement anything — no `template_comments` table, no route, no UI exists after this ADR. Governs the shape any future implementation MUST take.
- **Depends on:** none. References the existing `document_comments` implementation as the pattern to reuse.

---

## Context

Templates have no comment thread today. Reviewer feedback during a template version's `in_review`/`under_review` status flows only through the `VersionActionPanel` reason field (a single free-text field per decision, not a threaded discussion) — `wiki/backlog/template-editor.md:23`. Documents, by contrast, already have a working threaded-comment feature.

The backlog listed three undecided options (`wiki/backlog/template-editor.md:26-28`):
1. Reuse the document-comments table with a `template_version_id` column.
2. A separate `template_review_comments` table.
3. Inline-only annotations via eigenpal native comments (different UX).

### Verified runtime facts — the document-comments pattern

- **Table:** `public.document_comments` (`db/baseline/0001_current_schema.sql:2136-2149`) — columns `id, tenant_id, document_id, library_comment_id, parent_library_id, author_id, author_display, content_json, resolved_at, resolved_by, created_at, updated_at`. `library_comment_id` + `parent_library_id` are eigenpal's own comment/thread identifiers (the editor engine owns comment numbering; MetalDocs persists them). RLS is forced (`:2151`), tenant-isolated by policy (`:4816-4819`), `document_id` FK cascades on delete (`:4312-4313`), and a unique index enforces one row per `(document_id, library_comment_id)` (`:3491-3494`).
- **Domain contract:** `internal/modules/documents/delivery/http/handler.go:107-112` — a narrow `documentComments` port: `ListDocumentComments`, `AddDocumentComment`, `UpdateDocumentComment`, `DeleteDocumentComment`, each scoped by `(tenantID, documentID)` plus `userID`/`libraryID` as needed.
- **HTTP surface:** `handler.go:1294-1401` — `listComments`, `createComment`, `updateComment`, `deleteComment` under `/documents/{id}/comments`; content crosses the wire as `documentsapi.DocumentCommentContentNode[]` (decoded from `content_json`, `handler.go:1403-1432`), never as raw eigenpal node types — consistent with ADR 0046's opaque-content-at-the-seam rule for everything except comments, which cross as the narrower `EditorComment` DTO per ADR 0046 §Decision.3.
- **Ownership:** the comments table is a child of `documents`, scoped and cascaded by `document_id`; it is not a generic cross-module "comments platform" — it is a documents-module concern with one bound parent type.

## Decision

**Defer a generic, cross-module comments platform. It is YAGNI today** — there is exactly one live consumer (documents) and zero committed demand for a second. Building an abstraction for a currently-unknown second shape (template comments) before that shape is confirmed would guess at requirements neither consumer has stated (per-thread resolution workflow? approver-only visibility? cross-version carryover on `CreateNextVersion`? none of this is specified for templates today).

**When product asks for template comments, implement option 1: reuse the same table pattern as `document_comments`**, not a shared/generic table and not option 3 (inline-only eigenpal-native, which sacrifices the discussion-thread UX the reason field already lacks). Concretely, at that time:

1. Add `public.template_comments` — same column shape as `document_comments` (`id, tenant_id, template_version_id, library_comment_id, parent_library_id, author_id, author_display, content_json, resolved_at, resolved_by, created_at, updated_at`), scoped to `template_version_id` (not `template_id` — comments are version-scoped, matching how document comments scope to a specific `document_id` snapshot, not a document family). RLS forced + tenant policy, matching `document_comments`.
2. Add a `templateComments` port on the templates module mirroring `documentComments`'s four methods, and the matching four HTTP routes under `/templates/{id}/versions/{n}/comments`.
3. Reuse `documentsapi.DocumentCommentContentNode` shape (or a templates-namespaced copy with identical fields) for content — do not invent a new content schema.
4. Do **not** extract a shared `comments` module or generic port at that time either, unless a *third* consumer appears. Two consumers (documents, templates) with an identical, copy-pasted shape is an acceptable near-term duplication — see the `frontend-structure.md` "promote on second caller" rule for the analogous frontend threshold; the backend analog here is "promote on **third** caller" because the first duplication (copy the table pattern once) is cheaper and more legible than a premature cross-module comments platform that would need to arbitrate two different parent-entity FK shapes (`document_id` vs `template_version_id`) from day one.

## Consequences

- No code, schema, or contract changes ship from this ADR. `wiki/backlog/template-editor.md` "comments" section is resolved (option 1 chosen) but remains unimplemented until product asks.
- The next implementer has an unambiguous pattern to copy from (`document_comments` + the four-method port + four routes) instead of re-litigating the 3-option choice.
- If a third comment-bearing entity emerges before template comments ship, revisit this ADR before generalizing — do not let a third consumer silently motivate a platform without a new decision record.

## References

- `wiki/backlog/template-editor.md` — "comments" section, 3-option decision point this ADR resolves.
- `db/baseline/0001_current_schema.sql:2136-2149,3491-3494,4312-4313,4816-4819` — `document_comments` table, indexes, FK, RLS policy.
- `internal/modules/documents/delivery/http/handler.go:107-112,1294-1432` — comments port + HTTP handlers (the pattern to reuse).
- ADR [`0046-eigenpal-anti-corruption-layer.md`](0046-eigenpal-anti-corruption-layer.md) — `EditorComment` DTO / opaque-content seam rule that any template-comments implementation must also follow.
