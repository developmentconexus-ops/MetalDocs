// PT-BR labels for route admin enums and error codes.
//
// All UI surfaces in this feature MUST go through these helpers so future
// i18n work has one swap point.

import type { components } from '../../../../lib/api-types';

export type QuorumKind = components['schemas']['QuorumKind'];
export type DriftPolicy = components['schemas']['DriftPolicy'];

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
