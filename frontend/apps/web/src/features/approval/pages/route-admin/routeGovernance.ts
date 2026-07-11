import type { components } from '../../../../lib/api-types';

export type GovernanceClass = components['schemas']['DocumentProfileItem']['governance_class'];
export type StageKind = NonNullable<components['schemas']['StageRequest']['stage_kind']>;

export interface RoutePolicy {
  routeAllowed: boolean;
  signatureRequired: boolean;
  badgeLabel: string;
  badgeTone: 'required' | 'optional' | 'blocked';
}

const ROUTE_POLICIES: Record<GovernanceClass, RoutePolicy> = {
  controlado: {
    routeAllowed: true,
    signatureRequired: true,
    badgeLabel: 'Obrigatório ≥1 assinatura',
    badgeTone: 'required',
  },
  simples: {
    routeAllowed: true,
    signatureRequired: false,
    badgeLabel: 'Assinatura opcional',
    badgeTone: 'optional',
  },
  livre: {
    routeAllowed: false,
    signatureRequired: false,
    badgeLabel: 'Perfil livre — sem rota de aprovação',
    badgeTone: 'blocked',
  },
};

export function routePolicyFor(gc: GovernanceClass): RoutePolicy {
  return ROUTE_POLICIES[gc];
}

const STAGE_KIND_LABELS: Record<StageKind, string> = {
  review: 'Revisão',
  approval: 'Assinatura',
};

const STAGE_KIND_DESCRIPTIONS: Record<StageKind, string> = {
  review: 'Revisores comentam e conferem; não assinam.',
  approval: 'Aprovadores assinam (e podem conversar).',
};

export const STAGE_KIND_OPTIONS: ReadonlyArray<StageKind> = ['review', 'approval'];

export function labelForStageKind(k: StageKind): string {
  return STAGE_KIND_LABELS[k];
}

export function describeStageKind(k: StageKind): string {
  return STAGE_KIND_DESCRIPTIONS[k];
}

export function flowPreviewEmptyLabel(): string {
  return 'Adicione etapas para ver o fluxo.';
}

export function stageActorSlotDefaultHeading(): string {
  return 'Quem atua nesta etapa';
}

export function livreBlockedMessage(): string {
  return 'Perfil livre não permite rota de aprovação — reservado a rascunhos e material não governado.';
}

export function validateSignaturePolicy(
  gc: GovernanceClass,
  stageKinds: StageKind[],
): string | null {
  const policy = routePolicyFor(gc);
  if (!policy.routeAllowed) {
    return livreBlockedMessage();
  }
  if (policy.signatureRequired && !stageKinds.includes('approval')) {
    return 'Perfil controlado exige ao menos uma etapa de assinatura (aprovação).';
  }
  return null;
}
