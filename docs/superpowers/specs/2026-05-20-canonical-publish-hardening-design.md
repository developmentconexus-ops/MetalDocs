# Canonical Publish Hardening Design

Date: 2026-05-20
Scope: modern documents + approval canonical `/documents/:id` publish/revision behavior
Status: proposed

## Goal

Repair the remaining approved/published lifecycle gaps in the modern canonical document surface so the workflow behaves like a senior, regulated SaaS product:

- no ambiguous publish behavior
- no invalid revision-create affordance while a sibling active revision exists
- no runtime/contract mismatch on schedule publish
- no fallback to legacy surfaces

This design preserves the current domain rule that `approved` is not yet `published`, while leaving a clean backend seam for a future system policy that may auto-publish on approval.

## Confirmed Findings Driving This Design

1. `Agendar publicação` is wired to the wrong backend identifier today.
2. The canonical page only treats sibling `draft` as active when deciding whether to allow a new revision.
3. Superseding the previously published revision is still optional in the publish dialog.
4. Approval wiki memory still drifts on the supersede capability name.

## Product Rules

### Lifecycle

- `approved` remains a pre-publication state.
- `published` remains an explicit later transition.
- Future support for `auto_publish_on_approval` is a domain-policy concern, not a screen-local rule.

### Publish Family

- If a published lineage head already exists, `Publish now` must always supersede it.
- If a published lineage head already exists, `Schedule publish` must always schedule supersession of it.
- The canonical UI must not present “publish without supersede” as an operator choice in that situation.

### Active Sibling Revision

- The canonical page must treat any runtime active sibling revision as blocking new revision creation.
- Active means any `active-document` state that still occupies the lineage slot:
  - `draft`
  - `under_review`
  - `approved`
  - `scheduled`
  - `rejected`
- When a sibling active revision exists, the page must expose a state-aware continuation CTA instead of `Iniciar revisão`.

## Recommended UX

### Canonical CTA Mapping

For a published document page with an active sibling revision:

- `draft` -> `Continuar rascunho`
- `under_review` -> `Acompanhar revisão`
- `approved` -> `Publicar revisão aprovada`
- `scheduled` -> `Ver publicação agendada`
- `rejected` -> `Retomar revisão rejeitada`

This keeps the canonical page truthful and action-oriented while preventing invalid lineage branching.

### Publish Dialog

When a published lineage head already exists:

- remove the optional supersede checkbox
- present copy that clearly states the current published revision will be replaced now or on the scheduled date
- keep both timing choices:
  - `Publicar agora`
  - `Agendar publicação`

When no prior published lineage head exists:

- plain publish/schedule may remain available

## Backend / Contract Design

### Schedule Publish Identifier

The schedule-publish HTTP path is document-scoped, but the service currently expects an approval-instance identity.

Canonical repair:

- the handler must resolve the active approval instance from the document ID before calling the service
- the service contract may keep using approval-instance identity internally if that remains the stable application boundary
- handler tests must verify the correct document-scoped runtime behavior instead of locking in the wrong identifier

### If-Match

The publish family should not parse `If-Match` and silently discard it.

Recommended target:

- schedule publish must enforce the same precondition discipline as the rest of the canonical publish family
- if true enforcement is not implemented in this slice, the contract must be explicitly narrowed instead of pretending the precondition is honored

Preferred direction: real enforcement.

## Frontend Design

### Canonical Page

`DocumentPublishedPage` should:

- consume `active-document` as the runtime source of sibling revision truth
- recognize all active sibling states, not only `draft`
- derive state-aware CTA label + navigation target from `approvalState`
- suppress new revision creation whenever an active sibling exists

### Publish Dialog Wiring

The dialog should:

- operate in required-supersede mode whenever `publishedDocumentId` exists
- remove the operator choice that leads to two simultaneous published lineage heads
- keep explicit success/error handling aligned with the current TanStack Query invalidation approach

## Wiki / Memory Sync Requirements

If implemented, sync:

- `wiki/modules/documents.md`
- `wiki/modules/approval.md`
- `wiki/modules/registry.md`

Required updates:

- approved vs published remains the current operating rule
- canonical `/documents/:id` active-sibling behavior
- supersede capability naming and runtime truth
- future auto-publish policy noted as a deferred domain seam, not current behavior

## Out of Scope

- implementing the future system configuration for auto-publish on approval
- redesigning the entire approval policy model
- legacy registry/product surfaces as primary UX
- unrelated browser tooling repairs, except documenting them as workflow gaps

## Implementation Boundaries

Expected touched areas:

- frontend canonical document page
- frontend publish dialog wiring/tests
- backend approval publish/schedule HTTP boundary
- backend schedule-publish tests
- bounded wiki sync

Avoid:

- broad refactors in unrelated document/editor flows
- new legacy lookups or fallback navigation
- speculative abstractions for future configuration

## Acceptance Criteria

1. `Agendar publicação` works on the canonical approved page using the correct runtime identity path.
2. When a prior published revision exists, canonical publish/schedule always supersede it.
3. The canonical page never offers `Iniciar revisão` while any active sibling revision exists.
4. The canonical page shows a state-aware CTA for existing active sibling revisions.
5. Runtime/API/browser verification confirms:
   - `approved -> published` or `approved -> scheduled`
   - clean `active-document` truth after publish-only state
   - next revision can only be created after the lineage returns to publish-only state
6. Wiki memory matches runtime and contract truth for this slice.

## Deferred Follow-Up

- system policy seam for:
  - `manual_publish_after_approval`
  - `auto_publish_on_approval`
- create-revision UX alignment so governed `revision_title` remains official only at submit-for-review time
