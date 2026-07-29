// PT-BR labels for route admin enums and error codes.
//
// All UI surfaces in this feature MUST go through these helpers so future
// i18n work has one swap point.

import type { components } from '../../../../lib/api-types';

export type QuorumKind = components['schemas']['QuorumKind'];
export type DriftPolicy = components['schemas']['DriftPolicy'];
export type ActorSelectorKind = components['schemas']['ActorSelector']['kind'];

export const ACTOR_SELECTOR_KIND_OPTIONS: ReadonlyArray<ActorSelectorKind> = [
  'named_user',
  'role_in_fixed_area',
  'role_in_document_area',
  'submit_choice',
];

const ACTOR_SELECTOR_KIND_LABELS: Record<ActorSelectorKind, string> = {
  named_user: 'Usuário específico',
  role_in_fixed_area: 'Papel em área fixa',
  role_in_document_area: 'Papel na área do documento',
  submit_choice: 'Escolha no envio',
};

export function labelForSelectorKind(kind: ActorSelectorKind): string {
  return ACTOR_SELECTOR_KIND_LABELS[kind] ?? kind;
}

/**
 * What a route governs (ADR 0082 subject kinds, ADR 0086 keying). Both kinds
 * are keyed by the profile, so the labels name the governed population, not a
 * scope: "Documentos" = every document of the profile, "Templates" = every
 * template of the profile.
 */
export type RouteSubjectKindLabelKey = 'document' | 'template';

export const SUBJECT_KIND_OPTIONS: ReadonlyArray<RouteSubjectKindLabelKey> = ['document', 'template'];

const SUBJECT_KIND_LABELS: Record<RouteSubjectKindLabelKey, string> = {
  document: 'Documentos',
  template: 'Templates',
};

const SUBJECT_KIND_DESCRIPTIONS: Record<RouteSubjectKindLabelKey, string> = {
  document: 'Governa a aprovação dos documentos criados sob o perfil.',
  template: 'Governa a aprovação dos templates criados sob o perfil.',
};

export function labelForSubjectKind(kind: RouteSubjectKindLabelKey | undefined): string {
  // Routes written before subject kinds existed omit the field; they are all
  // document routes, which is also the server-side default.
  return SUBJECT_KIND_LABELS[kind ?? 'document'] ?? kind ?? '';
}

export function describeSubjectKind(kind: RouteSubjectKindLabelKey): string {
  return SUBJECT_KIND_DESCRIPTIONS[kind] ?? '';
}

export const QUORUM_KIND_OPTIONS: ReadonlyArray<QuorumKind> = [
  'any_1_of',
  'all_of',
  'm_of_n',
];

export const DRIFT_POLICY_OPTIONS: ReadonlyArray<DriftPolicy> = [
  'reduce_quorum',
  'fail_stage',
  'keep_snapshot',
];

const QUORUM_KIND_LABELS: Record<QuorumKind, string> = {
  any_1_of: 'Qualquer um (1 de N)',
  all_of: 'Todos (N de N)',
  m_of_n: 'M de N',
};

const QUORUM_KIND_DESCRIPTIONS: Record<QuorumKind, string> = {
  any_1_of: 'Aprovado quando qualquer aprovador da etapa assina.',
  all_of: 'Aprovado somente quando todos os aprovadores assinam.',
  m_of_n: 'Aprovado quando ao menos M dos N aprovadores assinam.',
};

const DRIFT_POLICY_LABELS: Record<DriftPolicy, string> = {
  reduce_quorum: 'Reduzir quórum',
  fail_stage: 'Falhar etapa',
  keep_snapshot: 'Manter snapshot',
};

const DRIFT_POLICY_DESCRIPTIONS: Record<DriftPolicy, string> = {
  reduce_quorum: 'Se um aprovador perde o papel, o quórum exigido cai.',
  fail_stage: 'Se um aprovador perde o papel, a etapa falha.',
  keep_snapshot: 'A etapa preserva os aprovadores válidos no momento da submissão.',
};

export function labelForQuorumKind(kind: QuorumKind): string {
  return QUORUM_KIND_LABELS[kind] ?? kind;
}

export function describeQuorumKind(kind: QuorumKind): string {
  return QUORUM_KIND_DESCRIPTIONS[kind] ?? '';
}

export function labelForDriftPolicy(policy: DriftPolicy): string {
  return DRIFT_POLICY_LABELS[policy] ?? policy;
}

export function describeDriftPolicy(policy: DriftPolicy): string {
  return DRIFT_POLICY_DESCRIPTIONS[policy] ?? '';
}

/**
 * Reason copy for the disabled-edit tooltip on the route list. Cause-based,
 * never the legacy "referenced by active instance" copy.
 */
export function tooltipForDisabledEdit(active: boolean): string | undefined {
  if (!active) {
    return 'Rota inativa — desativada e somente leitura.';
  }
  return undefined;
}
