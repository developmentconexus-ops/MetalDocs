// Document-workflow helpers shared by the detail adapter (useDocumentArtifact) and
// the detail route (DocumentDetailRoute). Single source of truth for the active-sibling
// state vocabulary, the sibling-CTA label/destination mapping, and the revision-initiate
// UI gate. No React import — pure functions only.

export const ACTIVE_SIBLING_STATES = ['draft', 'under_review', 'approved', 'scheduled', 'rejected'] as const;
export type ActiveSiblingState = (typeof ACTIVE_SIBLING_STATES)[number];

export function getActiveSiblingCtaLabel(state: ActiveSiblingState): string {
  if (state === 'draft') return 'Continuar rascunho';
  if (state === 'under_review') return 'Acompanhar revisão';
  if (state === 'approved') return 'Publicar revisão aprovada';
  if (state === 'scheduled') return 'Ver publicação agendada';
  return 'Retomar revisão rejeitada';
}

export function getActiveSiblingDestination(documentId: string, state: ActiveSiblingState): string {
  if (state === 'draft' || state === 'rejected') {
    return `/documents/${documentId}/edit`;
  }
  return `/documents/${documentId}`;
}

// TODO(authz): migrate this UI gate from roles to useHasCapability once the revision-initiate capability key exists (ADR 0022 — capabilities, never roles). Tracked follow-up.
export function canInitiateRevision(user: { roles: string[] } | null | undefined): boolean {
  return user != null && user.roles.some((r) => ['system_admin', 'editor', 'qms_admin', 'area_admin'].includes(r));
}
