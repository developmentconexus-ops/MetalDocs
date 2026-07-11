// @ts-expect-error uuid package exists in workspace, no typings exposed in this app.
import { v4 as uuidv4 } from 'uuid';

import type { RouteSummary, StageRequest } from '../../api/routeAdminApi';
import { validateSignaturePolicy, validateStageOrder, type GovernanceClass } from './routeGovernance';
import type { StageDraft } from './StageCard';

export interface RouteDraft {
  name: string;
  profileCode: string;
  stages: StageDraft[];
}

export function defaultStage(): StageDraft {
  return {
    uid: uuidv4() as string,
    label: '',
    requiredRole: '',
    requiredCapability: 'doc.signoff',
    areaCode: '',
    quorumKind: 'any_1_of',
    m: '1',
    driftPolicy: 'reduce_quorum',
    stageKind: 'approval',
  };
}

export function toDraft(route: RouteSummary | null): RouteDraft {
  if (!route) {
    return { name: '', profileCode: '', stages: [defaultStage()] };
  }
  return {
    name: route.name,
    profileCode: route.profile_code,
    stages: route.stages.map((stage) => ({
      uid: uuidv4() as string,
      label: stage.name,
      requiredRole: stage.required_role ?? '',
      requiredCapability: stage.required_capability || 'doc.signoff',
      areaCode: stage.area_code ?? '',
      quorumKind: stage.quorum,
      m: String(stage.quorum_m ?? 1),
      driftPolicy: stage.drift_policy,
      stageKind: stage.stage_kind ?? 'approval',
    })),
  };
}

export function toStageRequests(draft: RouteDraft): StageRequest[] {
  return draft.stages.map((stage, index): StageRequest => {
    const payload: StageRequest = {
      order: index + 1,
      name: stage.label.trim(),
      required_role: stage.requiredRole.trim(),
      required_capability: stage.requiredCapability.trim() || 'doc.signoff',
      area_code: stage.areaCode.trim(),
      quorum: stage.quorumKind,
      drift_policy: stage.driftPolicy,
      stage_kind: stage.stageKind,
    };
    if (stage.quorumKind === 'm_of_n') {
      payload.quorum_m = Number(stage.m);
    }
    return payload;
  });
}

export function validateDraft(
  draft: RouteDraft,
  isEdit: boolean,
  governanceClass: GovernanceClass | null,
): string | null {
  if (!draft.name.trim()) {
    return 'Informe o nome da rota.';
  }
  if (!isEdit && !draft.profileCode.trim()) {
    return 'Informe o código do perfil.';
  }
  if (draft.stages.length === 0) {
    return 'A rota deve possuir ao menos uma etapa.';
  }

  const labels = new Set<string>();
  for (const stage of draft.stages) {
    const label = stage.label.trim();
    if (!label) {
      return 'Toda etapa deve ter nome.';
    }
    const normalized = label.toLocaleLowerCase('pt-BR');
    if (labels.has(normalized)) {
      return 'Nomes de etapa devem ser distintos.';
    }
    labels.add(normalized);

    if (!stage.requiredRole) {
      return `A etapa "${label}" deve ter uma role definida.`;
    }
    if (!stage.areaCode) {
      return `A etapa "${label}" deve ter uma área definida.`;
    }
    if (stage.quorumKind === 'm_of_n') {
      const mValue = Number(stage.m);
      if (!Number.isFinite(mValue) || mValue < 1) {
        return `Na etapa "${label}", informe um valor de M válido.`;
      }
    }
  }

  const stageKinds = draft.stages.map((stage) => stage.stageKind);
  const orderError = validateStageOrder(stageKinds);
  if (orderError) {
    return orderError;
  }

  // null governance class (archived profile on edit) → backend is sole enforcer.
  if (governanceClass) {
    const policyError = validateSignaturePolicy(governanceClass, stageKinds);
    if (policyError) {
      return policyError;
    }
  }
  return null;
}
