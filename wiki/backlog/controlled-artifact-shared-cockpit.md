# Backlog: controlled-artifact shared cockpit — T11 query ownership

> Last updated: 2026-07-01 (opened from the `feat/template-document-parity` senior code review — frontend findings M-2 and M-6)

The shared controlled-artifact view layer (ADR 0053) landed with the detail/approval
adapters owning the model, but the **document detail route still re-runs several
queries the adapter already owns**, and the **document approval adapter builds a
reduced view-model in isolation**. Both were reviewed and consciously deferred to a
follow-up (labelled T11) to keep the parity PR scoped. Neither is a correctness bug;
both are structural debt to retire once `model.actions` carries fully wired `run()`
handlers.

---

## M-2 — `DocumentDetailRoute` re-runs adapter-owned queries

`frontend/apps/web/src/features/documents/pages/DocumentDetailRoute.tsx:40`

The route calls `useDocumentDetailQuery`, `useApprovalInstanceQuery`,
`useControlledDocumentActiveDocumentQuery`, and `useDistributionSummaryQuery`
directly, then reads their `.data` alongside the adapter's computed `model`. React
Query dedupes the network requests (same keys), so there is no runtime cost, but any
stale-time / option change must be applied in two places, and the coverage aside reads
the raw `total_targets` while the adapter converts it to a display string.

**Done when:** the route consumes `model.actions` (with wired `run()` handlers) and
`model.kpis` for coverage, and the duplicate `use*Query` calls are removed from the
route body.

## M-6 — document approval adapter builds a reduced model in isolation

`frontend/apps/web/src/features/documents/adapters/useDocumentApprovalArtifact.ts`

The approval cockpit constructs a reduced `ArtifactViewModel` (Aprovações breadcrumb,
single code chip, empty profile/area/visibility meta) rather than composing
`useDocumentArtifact` the way `useTemplateApprovalArtifact` composes
`useTemplateArtifact`. The intent is now documented inline at the model literal. If the
cockpit ever needs the full identity block, compose the detail model and override
hero/meta instead of widening the literal — this keeps a single source of truth for the
document hero/meta mapping.

**Done when:** the two document adapters share one hero/meta mapping (compose-then-strip),
matching the template adapter pattern.
