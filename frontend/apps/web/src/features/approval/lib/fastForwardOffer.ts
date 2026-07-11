// G3 (unit 2.4, Slice D) — pure derivation of the opportunistic "Aprovar já"
// offer from the instance payload. This is a DISPLAY hint only: the backend
// (authz.Require + CheckEligibility/CheckSoD, fast-forward endpoint) is the
// sole authority. We do NOT gate on ReviewVerdictResponse.fast_forward_eligible
// because that flag only exists on the verdict MUTATION response, and recording
// a normal `ready` verdict first COMPLETES the review stage — which would make a
// subsequent fast-forward call fail (ErrFastForwardNotEligible / already-signed).
// Fast-forward must be a single standalone gesture, never preceded by a normal
// verdict, so the offer is derived opportunistically from the instance's current
// state instead of from that response flag (spec R5, hint-only).
import type { ApprovalInstance } from '../api/approvalTypes';

export interface FastForwardOffer {
  reviewStageId: string;
  nextApprovalStageId: string;
}

export function deriveFastForwardOffer(
  instance: ApprovalInstance,
  currentUserId: string,
): FastForwardOffer | null {
  if (!currentUserId) return null;
  if (instance.status !== 'in_progress') return null;

  const stages = instance.stages ?? [];
  const active = stages.find((s) => s.status === 'active');
  if (!active) return null;
  if (active.stage_kind !== 'review') return null;

  const viewer = instance.viewer;
  if (viewer?.eligible_for_active_stage !== true) return null;
  if (viewer?.has_signed_active_stage !== false) return null;

  const next = stages.find((s) => s.stage_index === active.stage_index + 1);
  if (!next) return null;
  if (next.stage_kind !== 'approval') return null;
  if (!(next.actors ?? []).some((a) => a.user_id === currentUserId)) return null;

  return { reviewStageId: active.id, nextApprovalStageId: next.id };
}
