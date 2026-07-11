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

/** R1: a route is `review* → approval*` — once an approval stage appears, no
 * review stage may follow it. */
export function stageOrderViolationMessage(): string {
  return 'A rota deve seguir revisão(ões) antes de assinatura(s) — nenhuma etapa de revisão pode vir depois de uma etapa de assinatura.';
}

export function validateStageOrder(stageKinds: StageKind[]): string | null {
  let sawApproval = false;
  for (const kind of stageKinds) {
    if (kind === 'approval') {
      sawApproval = true;
    } else if (sawApproval) {
      return stageOrderViolationMessage();
    }
  }
  return null;
}

/** R4: SoD — the document author is excluded from every stage automatically. */
export function authorExcludedNote(): string {
  return 'O autor do documento é automaticamente excluído de todas as etapas.';
}

/** R5: overlap between review and approval is expected; "Aprovar já" collapses
 * both ceremonies into one gesture when eligible. */
export function approveNowOverlapNote(): string {
  return 'Quem revisa e também assina pode aprovar já — uma única cerimônia registra os dois eventos.';
}
