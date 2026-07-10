// Template approval decision-model builder.
//
// Single construction path for the template `ArtifactDecisionModel` (FE-02) — was
// previously built inline in `TemplateApprovalRoute`. Templates carry no legal
// e-signature (no password/legal fields, unlike the document sign-off model built
// in `useDocumentArtifact`), so this mirrors that adapter's shape minus
// those two fields.

import type { ArtifactDecisionModel } from '../../shared/controlled-artifact/types';
import type { VersionDTO, VersionStatus } from '../api/templates';

/** Awaitable decision submit — the route supplies the actual API call + invalidation. */
export type TemplateApprovalDecisionSubmit = (accept: boolean, reason: string) => Promise<void>;

export interface BuildTemplateApprovalDecisionArgs {
  status: VersionStatus | null;
  version: VersionDTO | null;
  /** Whether the actor can act on this version's accept action (reuses the gate the
   *  adapter already computed via `buildTemplateApprovalActions`, rather than
   *  re-deriving capability logic). */
  acceptAvailable: boolean;
  submit: TemplateApprovalDecisionSubmit;
}

/**
 * Builds the decision model offered on the template approval cockpit for the
 * review (under_review + pending reviewer) and publish (approved, or under_review
 * with no reviewer) states. Returns undefined when no decision applies — the
 * cockpit then falls back to the plain `actions[]` band.
 */
export function buildTemplateApprovalDecision({
  status,
  version,
  acceptAvailable,
  submit,
}: BuildTemplateApprovalDecisionArgs): ArtifactDecisionModel | undefined {
  const underReview = status === 'under_review';
  const hasReviewer = version?.pending_reviewer_role != null;
  const isReview = underReview && hasReviewer;
  const offerDecision = (underReview || status === 'approved') && acceptAvailable;

  if (!offerDecision) {
    return undefined;
  }

  const acceptLabel = isReview ? 'Aprovar revisão' : 'Publicar';

  return {
    kicker: 'Decisão requerida',
    heading: 'Registrar decisão',
    description: isReview
      ? 'Revise o conteúdo do modelo e registre sua decisão de revisão.'
      : 'Confirme para publicar esta versão do modelo.',
    options: [
      {
        key: 'accept',
        label: acceptLabel,
        description: isReview
          ? 'Aprova a revisão técnica e avança o fluxo.'
          : 'Publica esta versão do modelo.',
        tone: 'approve',
        submitLabel: acceptLabel,
        requiresReason: false,
      },
      {
        key: 'reject',
        label: 'Rejeitar',
        description: 'Devolve o modelo para rascunho · requer motivo.',
        tone: 'reject',
        submitLabel: 'Rejeitar',
        requiresReason: true,
      },
    ],
    reasonLabel: 'Motivo',
    reasonPlaceholder: 'Comentário registrado na trilha do modelo…',
    password: null,
    legal: null,
    signer: null,
    submit: async ({ optionKey, reason }) => {
      await submit(optionKey === 'accept', reason);
    },
  };
}
