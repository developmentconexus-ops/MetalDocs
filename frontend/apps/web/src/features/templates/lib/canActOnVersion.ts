// Defense-in-depth gating for template lifecycle actions.
//
// These helpers are a UX hint only — they disable buttons and surface a
// Portuguese tooltip explaining what the user is missing (capability vs role
// binding vs status). The backend remains the sole authorization boundary
// (see wiki/concepts/authz-tiers.md).

import type { VersionDTO } from '../api/templates';

export type ActorContext = {
  roles: readonly string[];
  capabilities: readonly string[];
};

export type VersionActionGate =
  | { allowed: true }
  | { allowed: false; reason: string };

const REASON_STATUS = 'Ação indisponível neste status.';
const reasonMissingCapability = (cap: string) =>
  `Sua sessão não inclui a capacidade ${cap}.`;
const reasonRoleMismatch = (role: string) =>
  `Você não possui o papel necessário (${role}).`;

function evaluate(args: {
  statusOk: boolean;
  capability: string;
  capabilities: readonly string[];
  requiredRole: string | null;
  roles: readonly string[];
}): VersionActionGate {
  if (!args.statusOk) {
    return { allowed: false, reason: REASON_STATUS };
  }
  if (!args.capabilities.includes(args.capability)) {
    return { allowed: false, reason: reasonMissingCapability(args.capability) };
  }
  if (args.requiredRole !== null && !args.roles.includes(args.requiredRole)) {
    return { allowed: false, reason: reasonRoleMismatch(args.requiredRole) };
  }
  return { allowed: true };
}

// canSubmit gates "Submeter para revisão" on a draft version.
// Backend cap: template.submit. No role binding at submit time.
export function canSubmit(version: VersionDTO, actor: ActorContext): VersionActionGate {
  return evaluate({
    statusOk: version.status === 'draft',
    capability: 'template.submit',
    capabilities: actor.capabilities,
    requiredRole: null,
    roles: actor.roles,
  });
}

// canReview gates "Approve Review" / "Reject" on an in_review version that
// carries a pending_reviewer_role (reviewer-required flow).
// Backend cap: template.review. Role binding: pending_reviewer_role.
export function canReview(version: VersionDTO, actor: ActorContext): VersionActionGate {
  const hasReviewerBinding = version.pending_reviewer_role != null;
  return evaluate({
    statusOk: version.status === 'in_review' && hasReviewerBinding,
    capability: 'template.review',
    capabilities: actor.capabilities,
    requiredRole: version.pending_reviewer_role,
    roles: actor.roles,
  });
}

// canApprove gates "Publish" (the approved → published transition) on an
// approved version. Backend cap: template.approve.
// Role binding: pending_approver_role (null means no role gate).
export function canApprove(version: VersionDTO, actor: ActorContext): VersionActionGate {
  return evaluate({
    statusOk: version.status === 'approved',
    capability: 'template.approve',
    capabilities: actor.capabilities,
    requiredRole: version.pending_approver_role,
    roles: actor.roles,
  });
}

// canPublish gates the no-reviewer "Publish" path: in_review version with
// pending_reviewer_role == null transitions straight to published.
// Backend cap: template.publish. Role binding: pending_approver_role.
export function canPublish(version: VersionDTO, actor: ActorContext): VersionActionGate {
  const noReviewerFlow = version.pending_reviewer_role == null;
  return evaluate({
    statusOk: version.status === 'in_review' && noReviewerFlow,
    capability: 'template.publish',
    capabilities: actor.capabilities,
    requiredRole: version.pending_approver_role,
    roles: actor.roles,
  });
}
