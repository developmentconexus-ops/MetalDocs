// Document sign-off decision-model builder.
//
// Single construction path for the document `ArtifactDecisionModel` (FE-02) — was
// previously built inline in `SignoffDetailPage`. Unlike the template decision
// (`templateApprovalDecision.ts`), the document sign-off carries legal e-signature
// weight (MP 2.200-2/2001): a password re-auth + a legal-effect confirmation +
// signer identity, offered only while the instance is locked to the signing actor.

import type { ArtifactDecisionModel } from '../../shared/controlled-artifact/types';

export interface DocumentSignoffSigner {
  displayName: string;
  email: string | null;
  username: string;
}

export interface BuildDocumentSignoffDecisionArgs {
  /** Gate computed by the route/adapter: sidebar ready + policy allows signoff +
   *  the instance is locked + content hash resolved. */
  offered: boolean;
  signer: DocumentSignoffSigner | null;
  /** Preselects a decision option from the `?decision=` query param. */
  defaultOptionKey: 'approve' | 'reject' | null;
  /** Route-owned: flush the canvas' pending edits, call the signoff mutation, then
   *  refetch the instance. The adapter only decides WHETHER to offer the decision;
   *  the actual signing sequence stays route-owned (canvas ref, signoff mutation). */
  submit: (input: { optionKey: string; reason: string; password: string }) => Promise<void>;
}

/**
 * Builds the decision model offered on the document sign-off cockpit while the
 * approval instance is locked to the current actor. Returns undefined when signing
 * isn't currently offered — the cockpit then falls back to the read-only sidebar
 * states (loading / error / no-context) or the plain `actions[]` band.
 */
export function buildDocumentSignoffDecision({
  offered,
  signer,
  defaultOptionKey,
  submit,
}: BuildDocumentSignoffDecisionArgs): ArtifactDecisionModel | undefined {
  if (!offered) {
    return undefined;
  }

  return {
    kicker: 'Assinatura requerida',
    heading: 'Registrar decisão',
    description:
      'Sua decisão é registrada com efeito de assinatura eletrônica na trilha de auditoria do documento.',
    options: [
      {
        key: 'approve',
        label: 'Assinar e aprovar',
        description: 'Aprova o documento e avança o fluxo de aprovação.',
        tone: 'approve',
        submitLabel: 'Assinar e aprovar',
        requiresReason: false,
      },
      {
        key: 'reject',
        label: 'Assinar e devolver',
        description: 'Devolve o documento ao autor · requer justificativa.',
        tone: 'reject',
        submitLabel: 'Assinar e devolver',
        requiresReason: true,
      },
    ],
    reasonLabel: 'Justificativa',
    reasonPlaceholder: 'Comentário visível na trilha de auditoria…',
    password: { label: 'Senha' },
    legal: {
      text: 'Confirmo que revisei o conteúdo integralmente e que esta decisão tem efeito de assinatura eletrônica conforme a MP 2.200-2/2001.',
    },
    signer: signer
      ? {
          name: signer.displayName,
          detail: signer.email ?? signer.username,
          note: 'Assinatura digital gerada no ato da confirmação.',
        }
      : null,
    defaultOptionKey: defaultOptionKey ?? null,
    submit,
  };
}
