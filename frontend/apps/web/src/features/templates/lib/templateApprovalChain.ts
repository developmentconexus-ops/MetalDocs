// Template approval-chain builder.
//
// Templates are NOT signoff-less. A template `VersionDTO` carries the same
// submit → review → approve actors + timestamps that a document stores in its
// `ApprovalInstance.stages[]`, only inline on the version row. This module maps
// those inline fields to the SAME kind-agnostic `ApprovalChainItem[]` the shared
// flow-viz renders for documents — so the approval cockpit timeline looks
// identical for both kinds, with zero fabricated names or timestamps.
//
// Actor display echoes the raw actor id (parity with the document mapper, whose
// `actorDisplay` also echoes `actor_user_id` pending an IAM name-lookup); pending
// stages surface the real `pending_*_role` instead. No React import — pure data.

import type { ApprovalChainItem, ApprovalFlowState } from '../../shared/controlled-artifact/types';
import type { VersionDTO } from '../api/templates';

// Lifecycle position of each template status. Drives the fallback flow-state for
// a stage that has no timestamp yet (e.g. the current under_review stage).
const STATUS_ORDER: Record<string, number> = {
  draft: 0,
  under_review: 1,
  approved: 2,
  published: 3,
  obsolete: 3,
};

/**
 * Resolve one stage's flow state: a recorded timestamp (or a later lifecycle
 * position) means the stage is done; the stage at the current lifecycle position
 * is the pending-decision step; everything else is a not-yet-reached future step.
 */
function flowState(done: boolean, isCurrent: boolean): ApprovalFlowState {
  if (done) return 'approved';
  return isCurrent ? 'current' : 'pending';
}

/**
 * Map a template `VersionDTO` to the shared submit → review → approve chain.
 * Always three steps, so the timeline is stable across the whole lifecycle.
 */
export function buildTemplateApprovalChain(version: VersionDTO): ApprovalChainItem[] {
  const order = STATUS_ORDER[version.status] ?? 0;

  const submitted = version.submitted_at != null;
  const reviewDone = version.reviewed_at != null || order >= 2;
  const approveDone = version.approved_at != null || order >= 3;

  const submit: ApprovalChainItem = {
    stageIndex: 0,
    label: 'Autoria',
    status: submitted ? 'sent' : 'active',
    roleLabel: 'Autoria',
    // Submit is a hand-off, not a signoff: 'sent' once submitted, otherwise the
    // current step while still in draft.
    flowState: submitted ? 'sent' : order === 0 ? 'current' : 'pending',
    actorUserId: version.author_id ?? null,
    actorDisplay: version.author_id ?? null,
    decision: submitted ? 'submitted' : null,
    signedAt: version.submitted_at ?? null,
  };

  const review: ApprovalChainItem = {
    stageIndex: 1,
    label: 'Revisão técnica',
    status: reviewDone ? 'approved' : order === 1 ? 'active' : 'pending',
    roleLabel: 'Revisão técnica',
    flowState: flowState(reviewDone, order === 1),
    actorUserId: version.reviewer_id ?? null,
    actorDisplay: version.reviewer_id ?? version.pending_reviewer_role ?? null,
    decision: reviewDone ? 'approved' : null,
    signedAt: version.reviewed_at ?? null,
  };

  const approve: ApprovalChainItem = {
    stageIndex: 2,
    label: 'Aprovação',
    status: approveDone ? 'approved' : order === 2 ? 'active' : 'pending',
    roleLabel: 'Aprovação',
    flowState: flowState(approveDone, order === 2),
    actorUserId: version.approver_id ?? null,
    actorDisplay: version.approver_id ?? version.pending_approver_role ?? null,
    decision: approveDone ? 'approved' : null,
    signedAt: version.approved_at ?? null,
  };

  return [submit, review, approve];
}
