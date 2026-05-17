# Editor Review Comments Lifecycle Design

> Date: 2026-05-17
> Mode: brainstorming design, no implementation
> Scope: Documents-first review comments and revision workflow, then Template Editor parity for Plan 12.5

## Decision

MetalDocs comments are active review feedback, not user-facing comments attached forever to a released version. Reviewers use inline comments and suggestions during review. If the item is rejected, the same comments remain visible when it returns to draft so the author can make revisions. Released output must be clean: approved/published PDF views do not show editor comments or unresolved suggestions.

The database may retain comment history for audit and review evidence, but the editor experience treats comments as part of the current review/revision cycle until they are resolved, deleted, archived, or superseded by release.

## Product Lifecycle

The target lifecycle is:

1. Author creates a document or template draft.
2. Author submits it for review.
3. Reviewer opens the editor, adds comments or suggestions where revision is needed.
4. Reviewer rejects or requests revision when comments require author action.
5. The item returns to draft with the comments still embedded in the editable editor experience.
6. Author fixes the content and resolves, removes, or leaves comments according to the review process.
7. Author submits again.
8. Reviewer approves when no blocking feedback remains.
9. Approved/released output is clean and rendered as a final view or PDF without active editor comments.

## Core Semantics

### Active Comments

Active comments belong to the current editable or reviewable work item:

- Document active comments belong to the current document/revision workflow.
- Template active comments belong to the current template draft/review workflow.
- Active comments are visible in draft, in review, and rejected/revision-requested states.
- Active comments are hidden from clean released output.

### Review Cycle History

MetalDocs may preserve comment rows after release for audit and traceability. That history is not the same as showing comments on the released editor/PDF surface.

Persisted comment rows should be able to answer:

- who created the comment
- who replied or resolved it
- which work item it belonged to
- which review/revision cycle it was created in when that capability exists
- whether it was unresolved at rejection time
- whether it was resolved before approval

### Release Rule

The default approval rule should be:

- unresolved comments block approval
- pending suggestions block approval
- resolved comments may remain in audit history
- final released PDF/export does not render editor comments

Any future exception, such as approving with resolved comment history visible to auditors, must be a separate product decision and not part of Plan 12.5 screen finalization.

## Eigenpal Boundary

Eigenpal owns the editor interaction:

- inline comment anchors
- replies
- resolved state
- suggestion UI if exposed by the adapter
- DOCX marker round-trip

MetalDocs owns the SaaS truth:

- persistence
- tenant scoping
- actor identity
- review-state permissions
- approval blocking rules
- audit/history retention
- frontend cache/query ownership

The existing document comments contract remains correct on the most important technical invariant: comment anchors live inside the DOCX, not in SQL position offsets. SQL stores thread metadata and workflow state.

## Documents First

Documents are the reference implementation because they already have:

- `DocumentEditorPage`
- `MetalDocsEditor` comment props
- `useDocumentComments`
- document comments API wrappers
- `document_comments` persistence
- editor revision/sidebar structure

Before Template Editor parity, run a Documents Editor integration audit and classify every visible comment/revision widget and action against:

- runtime routes
- OpenAPI/generated types
- frontend API wrappers
- TanStack Query ownership
- documents module wiki truth
- database dictionary truth
- Eigenpal adapter capability

Likely known issues from discovery:

- comments exist but still use local effect/state patterns rather than canonical TanStack Query hooks
- some document comment paths are runtime-backed but not fully represented in module route truth
- the product rule for unresolved comments blocking approval needs explicit audit
- the sidebar/revision model should become the template editor baseline

## Template Parity

Template Editor should follow the hardened Documents Editor behavior, with one additional required template-specific surface:

- left placeholder sidebar remains part of the template editor

Template comments must not be faked. Today the template editor audit found no real template comments model or endpoint. Therefore template comments require backend/database/API prerequisite work before the UI is implemented.

The likely endpoint family should mirror the document comments shape while using template-owned runtime paths:

```text
GET    /api/v1/templates/{templateId}/comments
POST   /api/v1/templates/{templateId}/comments
PATCH  /api/v1/templates/{templateId}/comments/{libraryCommentId}
DELETE /api/v1/templates/{templateId}/comments/{libraryCommentId}
```

If implementation discovery proves comments must be scoped by active draft/version row for database integrity, that scope should remain an internal storage detail unless the user-facing workflow requires exposing it. The product language remains "comments on the template under review", not "comments on version N".

## Data Model Direction

Recommended first implementation:

- keep `document_comments` for documents
- add a template-owned comments table, likely `template_comments` or `template_review_comments`
- use the same Eigenpal library-comment mapping as documents
- include tenant, template, actor, content JSON, parent library id, resolved fields, timestamps
- add review-cycle linkage only if the current approval workflow has a real review-cycle entity or a stable lifecycle event ID to attach to

Do not start Plan 12.5 with a generic polymorphic comments table. That would force broad refactoring across documents, templates, approvals, audit, OpenAPI, and frontend caches before the current editor screens are stabilized.

## API and Frontend Query Rules

Any new or normalized comment surface must follow the MetalDocs contract rules:

- OpenAPI first for public routes
- generated backend and frontend types
- feature API wrappers using the canonical API client
- TanStack Query hooks for server state
- query keys in `frontend/apps/web/src/lib/queryKeys.ts`
- mutation invalidation for editor detail, comments, review state, and affected inbox/approval surfaces
- no direct `fetch` in page components
- no hand-written DTOs when OpenAPI can own the shape

Approval, rejection, publish, and release mutations should not use optimistic updates that could misrepresent regulated workflow state. Prefer pending UI plus server response and targeted invalidation.

## Implementation Order

This design changes Plan 12.5 execution order:

1. Audit Documents Editor review/comments/revision behavior first.
2. Write a Documents Editor hardening plan.
3. Implement only approved, real-capability document hardening slices.
4. Re-audit Template Editor against the hardened document baseline.
5. Split template work into:
   - frontend parity work that already has backend capability
   - backend/database/API prerequisites for template comments
   - deferred version-history or comments UI where capability is still missing

Subagents should only be used after the audit defines independent slices.

## Verification Gates

Before any Template Editor implementation continues, keep the existing Plan 12 gates:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-system-runnable.ps1 -TargetRoute /api/v1/templates
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-module-contract-sync.ps1 -Module templates
```

Before Documents Editor implementation, define and run equivalent route/contract gates for `/api/v1/documents` and the affected document comments routes.

For frontend changes:

```powershell
cd frontend/apps/web
pnpm.cmd tsc --noEmit -p tsconfig.build.json
pnpm test
```

For backend/API changes:

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/<module>/api/...
go test ./internal/modules/<module>/... -count=1
```

For database changes:

- add forward-only post-baseline migrations only
- write `public.schema_migrations`
- update dictionary pages
- verify runtime startup applies the migration cleanly

## Out of Scope

- no mock template comments UI
- no released PDF comment rendering
- no generic cross-module comments platform in the first pass
- no custom editor toolbar replacing Eigenpal
- no version-history UI until real list capability exists
- no implementation before the Documents Editor audit and follow-up plan are approved

## Success Criteria

- The user-facing workflow treats comments as active review feedback.
- Rejected work returns to draft with comments visible for author revision.
- Approval/release produces a clean final output with no active comments.
- Audit/history can preserve comment evidence without polluting released views.
- Documents Editor becomes the baseline for Template Editor.
- Template Editor keeps its left placeholder sidebar while gaining only real, backend-supported document-editor parity features.
- Every visible widget/action/state is classified before implementation.
- Missing backend/database/API capability is documented as prerequisite or defer, not faked.

