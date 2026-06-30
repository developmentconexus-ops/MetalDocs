// Approval-workflow policy shared by the document approval adapter
// (useDocumentApprovalArtifact) and the document approval extras component
// (DocumentApprovalExtras). Single source of truth for the per-state transition
// policy (which workflow actions appear, the disabled-edit reason, and the
// read-only flag) and the status→ApprovalState coercion. No React import — pure
// data + functions only.
//
// Lifted verbatim from the former ControlledDocumentDetailPanel so behavior is
// preserved byte-for-byte. Do NOT fork a second policy.

import type { ApprovalState } from '../../approval/api/approvalTypes';

export interface TransitionPolicy {
  disabledReason?: string;
  readOnly?: boolean;
  actions: {
    submit: boolean;
    signoff: boolean;
    cancelInstance: boolean;
    publishOrSchedule: boolean;
  };
}

export const TRANSITION_POLICY: Record<ApprovalState, TransitionPolicy> = {
  draft: {
    actions: { submit: true, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
  under_review: {
    disabledReason: 'Documento em revisão — edição bloqueada',
    actions: { submit: false, signoff: true, cancelInstance: true, publishOrSchedule: false },
  },
  approved: {
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: true },
  },
  scheduled: {
    disabledReason: 'Aguardando data de vigência agendada',
    readOnly: true,
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
  published: {
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: true },
  },
  superseded: {
    disabledReason: 'Versão substituída — somente leitura',
    readOnly: true,
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
  rejected: {
    disabledReason: 'Documento rejeitado — edite e submeta novamente',
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
  obsolete: {
    disabledReason: 'Documento obsoleto — somente leitura',
    readOnly: true,
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
  cancelled: {
    disabledReason: 'Aprovação cancelada',
    readOnly: true,
    actions: { submit: false, signoff: false, cancelInstance: false, publishOrSchedule: false },
  },
};

export function toApprovalState(status: string): ApprovalState {
  const allowed: ApprovalState[] = [
    'draft',
    'under_review',
    'approved',
    'scheduled',
    'published',
    'superseded',
    'rejected',
    'obsolete',
    'cancelled',
  ];
  if (allowed.includes(status as ApprovalState)) {
    return status as ApprovalState;
  }
  return 'draft';
}
