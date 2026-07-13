// Template approval-chain builder.
//
// Kernel-truth rebuild (ADR 0082 / unit 3.1a S3): the chain is always exactly
// three steps mirroring the kernel's own lifecycle, Submissão → Aprovação →
// Publicação (draft->under_review, under_review->approved, approved->published),
// driven solely by `VersionDTO.status` + its `submitted_at`/`approved_at`/
// `published_at` timestamps. Maps to the SAME kind-agnostic `ApprovalChainItem[]`
// the shared flow-viz renders for documents, so the cockpit timeline looks
// identical for both kinds.
//
// Actor identity: `author_id` is a live kernel-truth column (echoed raw, parity
// with the document mapper's `actorDisplay`, pending an IAM name-lookup) so the
// Submissão step shows it. `reviewer_id`/`approver_id`, by contrast, are NEVER
// written by the kernel completion path — `ApprovalCompletionWriter.
// MarkTemplateVersionApproved` (internal/modules/templates/infrastructure/
// approval_completion_writer.go) sets only `status`+`approved_at`, and publish
// only backfills `approved_at`/`published_at`. Displaying `reviewer_id`/
// `approver_id` for the Aprovação/Publicação steps would surface a dead legacy
// column, not the real signer (that identity lives in the approval-kernel's own
// module, not on the template version row) — so those steps honest-omit the
// actor (null) rather than fabricate one. No React import — pure data.

import type { ApprovalChainItem, ApprovalFlowState } from '../../shared/controlled-artifact/types';
import type { VersionDTO } from '../api/templates';

// Lifecycle position of each template status. Drives the fallback flow-state for
// the stage that has no timestamp yet (the current pending-decision step).
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
 * Map a template `VersionDTO` to the shared Submissão → Aprovação → Publicação
 * chain. Always three steps, so the timeline is stable across the whole
 * lifecycle.
 */
export function buildTemplateApprovalChain(version: VersionDTO): ApprovalChainItem[] {
  const order = STATUS_ORDER[version.status] ?? 0;

  const submitted = version.submitted_at != null;
  const approveDone = version.approved_at != null || order >= 2;
  const publishDone = version.published_at != null || order >= 3;

  const submit: ApprovalChainItem = {
    stageIndex: 0,
    label: 'Submissão',
    status: submitted ? 'sent' : 'active',
    roleLabel: 'Submissão',
    // Submit is a hand-off, not a signoff: 'sent' once submitted, otherwise the
    // current step while still in draft.
    flowState: submitted ? 'sent' : order === 0 ? 'current' : 'pending',
    actorUserId: version.author_id ?? null,
    actorDisplay: version.author_id ?? null,
    decision: submitted ? 'submitted' : null,
    signedAt: version.submitted_at ?? null,
  };

  const approve: ApprovalChainItem = {
    stageIndex: 1,
    label: 'Aprovação',
    status: approveDone ? 'approved' : order === 1 ? 'active' : 'waiting',
    roleLabel: 'Aprovação',
    flowState: flowState(approveDone, order === 1),
    // Honest-omit: the kernel completion writer never populates reviewer_id/
    // approver_id (see module header) — no real signer identity to show here.
    actorUserId: null,
    actorDisplay: null,
    decision: approveDone ? 'approved' : null,
    signedAt: version.approved_at ?? null,
  };

  const publish: ApprovalChainItem = {
    stageIndex: 2,
    label: 'Publicação',
    status: publishDone ? 'approved' : order === 2 ? 'active' : 'waiting',
    roleLabel: 'Publicação',
    flowState: flowState(publishDone, order === 2),
    // VersionDTO carries no publisher-id field — honest-omit, same as above.
    actorUserId: null,
    actorDisplay: null,
    decision: publishDone ? 'published' : null,
    signedAt: version.published_at ?? null,
  };

  return [submit, approve, publish];
}
