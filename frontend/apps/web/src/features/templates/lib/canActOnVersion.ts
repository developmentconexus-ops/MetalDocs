// Defense-in-depth gating for template lifecycle actions.
//
// These helpers are a UX hint only — they disable buttons and surface a
// Portuguese tooltip explaining what the user is missing (capability vs
// status). The backend remains the sole authorization boundary (see
// wiki/concepts/authz-tiers.md). Gating is capability + status ONLY — the
// kernel approval flow (ADR: unified submit-for-approval → signoff → publish)
// carries no role binding, so there is no role param anywhere in this module.

import type { VersionDTO } from '../api/templates';

export type ActorContext = {
  capabilities: readonly string[];
};

export type VersionActionGate =
  | { allowed: true }
  | { allowed: false; reason: string };

const REASON_STATUS = 'Ação indisponível neste status.';
const reasonMissingCapability = (cap: string) =>
  `Sua sessão não inclui a capacidade ${cap}.`;

function evaluate(args: {
  statusOk: boolean;
  capability: string;
  capabilities: readonly string[];
}): VersionActionGate {
  if (!args.statusOk) {
    return { allowed: false, reason: REASON_STATUS };
  }
  if (!args.capabilities.includes(args.capability)) {
    return { allowed: false, reason: reasonMissingCapability(args.capability) };
  }
  return { allowed: true };
}

// canSubmit gates "Submeter para aprovação" on a draft version.
// Backend cap: template.submit (POST .../submit-for-approval).
export function canSubmit(version: VersionDTO, actor: ActorContext): VersionActionGate {
  return evaluate({
    statusOk: version.status === 'draft',
    capability: 'template.submit',
    capabilities: actor.capabilities,
  });
}

// canSignoff gates the e-signature decision (Assinar aprovação / Rejeitar) on
// an under_review version. Backend cap: template.approve (POST .../signoff).
export function canSignoff(version: VersionDTO, actor: ActorContext): VersionActionGate {
  return evaluate({
    statusOk: version.status === 'under_review',
    capability: 'template.approve',
    capabilities: actor.capabilities,
  });
}

// canPublish gates "Publicar" on an approved version. Backend cap:
// template.publish (POST .../publish) — distinct from the signoff capability.
export function canPublish(version: VersionDTO, actor: ActorContext): VersionActionGate {
  return evaluate({
    statusOk: version.status === 'approved',
    capability: 'template.publish',
    capabilities: actor.capabilities,
  });
}
