// Approval-workflow policy shared by the document approval adapter
// (useDocumentApprovalArtifact) and the approval cockpit sidebar. Single
// source of truth for the per-state transition
// policy (which workflow actions appear, the disabled-edit reason, and the
// read-only flag) and the status→ApprovalState coercion. No React import — pure
// data + functions only.
//
// Lifted verbatim from the former ControlledDocumentDetailPanel so behavior is
// preserved byte-for-byte. Do NOT fork a second policy.

import type { ApprovalState } from '../../approval/api/approvalTypes';
import type { ApprovalChainItem, ApprovalFlowState } from '../../shared/controlled-artifact/types';

/**
 * Minimal structural interface for a stage passed to `mapApprovalChain`.
 * Intentionally loose on `signature_method` (string vs. SignatureMethod) so
 * both the hand-rolled ApprovalInstance.stages and the codegen
 * ApprovalInstanceByDocumentResponse.stages satisfy this parameter.
 * Only the fields actually consumed by the mapping are required.
 */
interface MappableStage {
  stage_index: number;
  label: string;
  status: string;
  signoffs: Array<{
    actor_user_id: string;
    decision: string;
    signed_at: string;
  }>;
}

export interface TransitionPolicy {
  disabledReason?: string;
  readOnly?: boolean;
  actions: {
    // F2d.3: `signoff` (a stage action) left this policy — stage eligibility is now
    // derived from the workspace-mode selector (viewer truth), not document status.
    cancelInstance: boolean;
    publishOrSchedule: boolean;
  };
}

export const TRANSITION_POLICY: Record<ApprovalState, TransitionPolicy> = {
  draft: {
    actions: { cancelInstance: false, publishOrSchedule: false },
  },
  under_review: {
    disabledReason: 'Documento em revisão — edição bloqueada',
    actions: { cancelInstance: true, publishOrSchedule: false },
  },
  approved: {
    actions: { cancelInstance: false, publishOrSchedule: true },
  },
  scheduled: {
    disabledReason: 'Aguardando data de vigência agendada',
    readOnly: true,
    actions: { cancelInstance: false, publishOrSchedule: false },
  },
  published: {
    actions: { cancelInstance: false, publishOrSchedule: true },
  },
  superseded: {
    disabledReason: 'Versão substituída — somente leitura',
    readOnly: true,
    actions: { cancelInstance: false, publishOrSchedule: false },
  },
  rejected: {
    disabledReason: 'Documento rejeitado — edite e submeta novamente',
    actions: { cancelInstance: false, publishOrSchedule: false },
  },
  obsolete: {
    disabledReason: 'Documento obsoleto — somente leitura',
    readOnly: true,
    actions: { cancelInstance: false, publishOrSchedule: false },
  },
  cancelled: {
    disabledReason: 'Aprovação cancelada',
    readOnly: true,
    actions: { cancelInstance: false, publishOrSchedule: false },
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

/**
 * Map an ApprovalInstance's stages array to the kind-agnostic ApprovalChainItem[]
 * used by the shared controlled-artifact view layer.
 *
 * Deduplicated from useDocumentArtifact and useDocumentApprovalArtifact — both
 * adapters map the same StageInstance shape. Callers retain their own null-guards
 * so absence of an instance still yields null.
 */
export function mapApprovalChain(stages: MappableStage[]): ApprovalChainItem[] {
  return stages.map((stage) => {
    const signoff = stage.signoffs[0] ?? null;
    return {
      stageIndex: stage.stage_index,
      label: stage.label,
      status: stage.status,
      // The stage label doubles as the role descriptor in the flow-viz secondary
      // line (documents have no separate role field distinct from the stage).
      roleLabel: stage.label,
      flowState: stageFlowState(stage.status, signoff?.decision ?? null),
      actorUserId: signoff?.actor_user_id ?? null,
      // TODO(iam): resolve actorDisplay to a display name via user-lookup query when available; currently echoes actor_user_id.
      actorDisplay: signoff?.actor_user_id ?? null,
      decision: signoff?.decision ?? null,
      signedAt: signoff?.signed_at ?? null,
    };
  });
}

/**
 * Normalize a document stage's raw `status` + signoff `decision` into the
 * kind-agnostic `ApprovalFlowState` the shared flow-viz renders. A recorded
 * decision wins; otherwise an active/current stage is the pending-decision step
 * and anything else is a not-yet-reached future step.
 */
function stageFlowState(status: string, decision: string | null): ApprovalFlowState {
  // `decision` follows the Signoff contract verbs ('approve' | 'reject'); map the
  // recorded outcome onto the past-tense flow-viz adjective.
  if (decision === 'approve') return 'approved';
  if (decision === 'reject') return 'rejected';
  if (status === 'active' || status === 'in_progress' || status === 'current') return 'current';
  return 'pending';
}

/**
 * Normalize a document editor approval-instance actor `status` into the shared
 * `ApprovalFlowState`. Distinct from `stageFlowState`: that editor DTO resolves the
 * outcome onto the actor directly ('approved' | 'rejected' | 'active' | 'waiting'),
 * so there is no separate signoff-decision tense to reconcile. Single source of
 * truth for the actor-status → flow-state mapping used by the editor sidebar chain.
 */
export function actorStatusToFlowState(status: string): ApprovalFlowState {
  if (status === 'approved') return 'approved';
  if (status === 'rejected') return 'rejected';
  if (status === 'active') return 'current';
  return 'pending';
}
