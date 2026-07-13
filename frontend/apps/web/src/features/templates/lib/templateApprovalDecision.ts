// Template approval decision-model builder.
//
// Single construction path for the template `ArtifactDecisionModel` (FE-02).
// Kernel model: the decision exists ONLY for the under_review → signoff step
// (POST .../signoff, backend cap template.approve). Publish (approved →
// published) has no signature ceremony — it is a plain `actions[]` entry
// (see templateApprovalActions.ts), not a decision. Signoff is password-gated
// like the document sign-off (`documentSignoffDecision.ts`), so this now
// mirrors that idiom's password + signer fields; unlike documents there is no
// `legal` MP 2.200-2/2001 checkbox — not requested, no backend field.

import type { ArtifactDecisionModel } from '../../shared/controlled-artifact/types';
import type { VersionStatus } from '../api/templates';

export interface TemplateApprovalSigner {
  displayName: string;
  email: string | null;
  username: string;
}

/** Awaitable decision submit — the route supplies the actual signoff call + invalidation. */
export type TemplateApprovalDecisionSubmit = (
  accept: boolean,
  reason: string,
  password: string,
) => Promise<void>;

export interface BuildTemplateApprovalDecisionArgs {
  status: VersionStatus | null;
  /** Gate computed by the adapter via `canSignoff` — status + template.approve capability. */
  signoffAvailable: boolean;
  signer: TemplateApprovalSigner | null;
  submit: TemplateApprovalDecisionSubmit;
}

/**
 * Builds the decision model offered on the template approval cockpit while the
 * version is under_review and the actor may sign off. Returns undefined
 * otherwise — the cockpit then falls back to the plain `actions[]` band
 * (draft: no actions; approved: Publicar).
 */
export function buildTemplateApprovalDecision({
  status,
  signoffAvailable,
  signer,
  submit,
}: BuildTemplateApprovalDecisionArgs): ArtifactDecisionModel | undefined {
  if (status !== 'under_review' || !signoffAvailable) {
    return undefined;
  }

  return {
    kicker: 'Decisão requerida',
    heading: 'Registrar decisão',
    description: 'Revise o conteúdo do modelo e confirme sua senha para registrar a decisão.',
    options: [
      {
        key: 'accept',
        label: 'Assinar aprovação',
        description: 'Aprova a versão e avança para publicação.',
        tone: 'approve',
        submitLabel: 'Assinar aprovação',
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
    password: { label: 'Senha' },
    legal: null,
    signer: signer
      ? {
          name: signer.displayName,
          detail: signer.email ?? signer.username,
          note: 'Assinatura digital gerada no ato da confirmação.',
        }
      : null,
    submit: async ({ optionKey, reason, password }) => {
      await submit(optionKey === 'accept', reason, password);
    },
  };
}
