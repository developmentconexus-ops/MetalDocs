// Document-workflow helpers shared by the detail adapter (useDocumentArtifact) and
// the detail route (DocumentDetailRoute). Single source of truth for the active-sibling
// state vocabulary and the sibling-CTA label/destination mapping. No React import —
// pure functions only.
//
// FE-11: the revision-initiate UI gate used to live here as a role check
// (`canInitiateRevision`), which violates ADR 0022 (capabilities, never roles).
// The backend gates POST /controlled-documents/{id}/revisions on the
// `document.edit` capability (CapDocumentEdit) at both tiers — see
// apps/api/cmd/metaldocs-api/permissions.go:203 (tier-1 route→capability) and
// internal/modules/documents/approval/application/submit_service.go /
// publish_service.go / internal/modules/documents/repository/repository.go
// (tier-2 authz.Require(ctx, tx, CapDocumentEdit, areaCode) in-tx checks).
// Callers now gate on `useHasCapability('document.edit')` at the component/adapter
// level instead of calling a pure role-based function here.

export const ACTIVE_SIBLING_STATES = ['draft', 'under_review', 'approved', 'scheduled', 'rejected'] as const;
export type ActiveSiblingState = (typeof ACTIVE_SIBLING_STATES)[number];

export function getActiveSiblingCtaLabel(state: ActiveSiblingState): string {
  if (state === 'draft') return 'Continuar rascunho';
  if (state === 'under_review') return 'Acompanhar revisão';
  if (state === 'approved') return 'Publicar revisão aprovada';
  if (state === 'scheduled') return 'Ver publicação agendada';
  return 'Retomar revisão rejeitada';
}

export function getActiveSiblingDestination(documentId: string): string {
  return `/documents/${documentId}`;
}
