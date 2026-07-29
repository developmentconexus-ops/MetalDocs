// @ts-expect-error uuid package exists in workspace, no typings exposed in this app.
import { v4 as uuidv4 } from 'uuid';

import type { CreateRouteRequest, RouteSummary, StageRequest } from '../../api/routeAdminApi';
import type { components } from '../../../../lib/api-types';
import { SIGNOFF_CAPABILITY } from './capabilities';
import { validateSignaturePolicy, validateStageOrder, type GovernanceClass } from './routeGovernance';
import { defaultSelector, type SelectorDraft, type StageDraft } from './StageCard';

type ActorSelector = components['schemas']['ActorSelector'];

/** What a route governs. Immutable after creation (ADR 0082 / ADR 0086). */
export type RouteSubjectKind = NonNullable<CreateRouteRequest['subject_kind']>;

export const DEFAULT_SUBJECT_KIND: RouteSubjectKind = 'document';

export interface RouteDraft {
  name: string;
  subjectKind: RouteSubjectKind;
  profileCode: string;
  stages: StageDraft[];
}

export function defaultStage(): StageDraft {
  return {
    uid: uuidv4() as string,
    label: '',
    requiredCapability: SIGNOFF_CAPABILITY,
    selectors: [defaultSelector()],
    quorumKind: 'any_1_of',
    m: '1',
    driftPolicy: 'reduce_quorum',
    stageKind: 'approval',
  };
}

function toSelectorDraft(selector: ActorSelector): SelectorDraft {
  return {
    uid: uuidv4() as string,
    kind: selector.kind,
    userId: selector.user_id ?? '',
    role: selector.role ?? '',
    areaCode: selector.area_code ?? '',
  };
}

export function toDraft(route: RouteSummary | null): RouteDraft {
  if (!route) {
    return {
      name: '',
      subjectKind: DEFAULT_SUBJECT_KIND,
      profileCode: '',
      stages: [defaultStage()],
    };
  }
  return {
    name: route.name,
    // `subject_kind` is optional on the wire (routes predating ADR 0082 were
    // all document routes and the server may omit it); absent means `document`,
    // which is also the server-side default.
    subjectKind: route.subject_kind ?? DEFAULT_SUBJECT_KIND,
    // Post-ADR-0086 every route — document and template alike — carries a
    // profile_code (a template route is keyed by the profile it governs, with
    // subject_key = profile_code), and the contract now types it non-nullable.
    // The `??` stays as a runtime guard only: the Go wire struct still holds a
    // *string so a malformed (NULL profile_code) row would surface as JSON
    // null, and RouteDraft.profileCode backs a controlled select that cannot
    // hold null.
    profileCode: route.profile_code ?? '',
    stages: route.stages.map((stage) => ({
      uid: uuidv4() as string,
      label: stage.name,
      requiredCapability: stage.required_capability,
      selectors:
        stage.selectors && stage.selectors.length > 0
          ? stage.selectors.map(toSelectorDraft)
          : [defaultSelector()],
      quorumKind: stage.quorum,
      m: String(stage.quorum_m ?? 1),
      driftPolicy: stage.drift_policy,
      stageKind: stage.stage_kind ?? 'approval',
    })),
  };
}

function toWireSelector(selector: SelectorDraft): ActorSelector {
  switch (selector.kind) {
    case 'named_user':
      return { kind: 'named_user', user_id: selector.userId.trim() };
    case 'role_in_fixed_area':
    case 'submit_choice':
      return {
        kind: selector.kind,
        role: selector.role.trim().toLocaleLowerCase('pt-BR'),
        area_code: selector.areaCode.trim().toLocaleLowerCase('pt-BR'),
      };
    case 'role_in_document_area':
      return { kind: 'role_in_document_area', role: selector.role.trim().toLocaleLowerCase('pt-BR') };
    default:
      return { kind: selector.kind };
  }
}

export function toStageRequests(draft: RouteDraft): StageRequest[] {
  return draft.stages.map((stage, index): StageRequest => {
    const payload: StageRequest = {
      order: index + 1,
      name: stage.label.trim(),
      required_capability: stage.requiredCapability.trim(),
      quorum: stage.quorumKind,
      drift_policy: stage.driftPolicy,
      stage_kind: stage.stageKind,
      selectors: stage.selectors.map(toWireSelector),
    };
    if (stage.quorumKind === 'm_of_n') {
      payload.quorum_m = Number(stage.m);
    }
    return payload;
  });
}

/**
 * Builds the create-route wire payload from a validated draft.
 *
 * ADR 0086: a route is keyed by configuration, never by an instance, so
 * `profile_code` is sent for BOTH subject kinds. `subject_key` is deliberately
 * omitted — the server derives it from `profile_code`, and sending a divergent
 * key is a 422 (`validation.template_subject_key_mismatch`).
 */
export function toCreateRequest(draft: RouteDraft): CreateRouteRequest {
  return {
    name: draft.name.trim(),
    subject_kind: draft.subjectKind,
    profile_code: draft.profileCode.trim(),
    stages: toStageRequests(draft),
  };
}

export function validateDraft(
  draft: RouteDraft,
  isEdit: boolean,
  governanceClass: GovernanceClass | null,
): string | null {
  if (!draft.name.trim()) {
    return 'Informe o nome da rota.';
  }
  // ADR 0086: the profile keys BOTH subject kinds — a template route governs
  // every template of its profile, exactly as a document route governs every
  // document of it. Hence the requirement is unconditional on subjectKind.
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

    if (stage.selectors.length === 0) {
      return `A etapa "${label}" deve ter ao menos um seletor.`;
    }
    for (const selector of stage.selectors) {
      if (selector.kind === 'named_user' && !selector.userId.trim()) {
        return `Na etapa "${label}", selecione um usuário.`;
      }
      if (selector.kind === 'role_in_fixed_area' || selector.kind === 'submit_choice') {
        if (!selector.role.trim()) {
          return `Na etapa "${label}", selecione uma role.`;
        }
        if (!selector.areaCode.trim()) {
          return `Na etapa "${label}", selecione uma área.`;
        }
      }
      if (selector.kind === 'role_in_document_area' && !selector.role.trim()) {
        return `Na etapa "${label}", selecione uma role.`;
      }
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
