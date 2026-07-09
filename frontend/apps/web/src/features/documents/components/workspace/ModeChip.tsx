import type { ApprovalInstance } from '../../../approval/api/approvalTypes';
import type { WorkspaceMode } from '../../../approval/lib/workspaceMode';
import styles from './ModeChip.module.css';

/**
 * Server-derived viewer facts the chip reads to build its tooltip (ADR 0078 —
 * the frontend never derives eligibility itself, only explains the mode the
 * selector already resolved). Reuses the instance's `viewer` shape verbatim.
 */
export type ModeChipViewer = ApprovalInstance['viewer'] | null;

export interface ModeChipProps {
  mode: WorkspaceMode;
  viewer: ModeChipViewer;
}

// F2d.5 S2a — PT-BR label per mode (spec table). author-editing and
// author-changes-requested share "Editando": both are writable-canvas states
// for the author; S2b differentiates them by canvas content, not by label.
const MODE_LABEL: Record<WorkspaceMode, string> = {
  'author-editing': 'Editando',
  'author-changes-requested': 'Editando',
  'author-waiting': 'Aguardando revisão',
  reviewing: 'Revisando',
  approving: 'Aprovando',
  observing: 'Visualizando',
  lifecycle: 'Visualizando',
};

function describeMode(mode: WorkspaceMode, viewer: ModeChipViewer): string {
  const delegatedFrom = viewer?.via_delegation_from?.display_name ?? null;

  switch (mode) {
    case 'author-editing':
      return 'Você é a autora deste documento e pode editá-lo.';
    case 'author-changes-requested':
      return 'Mudanças foram solicitadas nesta etapa — edite o documento para responder.';
    case 'author-waiting':
      return 'Você é a autora; o documento está em revisão por outra pessoa no momento.';
    case 'reviewing':
      return delegatedFrom
        ? `Você é elegível para revisar a etapa ativa, por delegação de ${delegatedFrom}.`
        : 'Você é elegível para revisar a etapa ativa.';
    case 'approving':
      return delegatedFrom
        ? `Você é elegível para aprovar a etapa ativa, por delegação de ${delegatedFrom}.`
        : 'Você é elegível para aprovar a etapa ativa.';
    case 'lifecycle':
      return 'Este documento está em uma etapa de ciclo de vida (publicação, vigência ou substituição).';
    case 'observing':
    default:
      return 'Você está acompanhando este documento, sem uma ação pendente no momento.';
  }
}

/**
 * Header chip surfacing the F2d.3 workspace mode as PT-BR copy, with a
 * tooltip explaining WHY (from the server-derived viewer facts) — never a
 * client-side eligibility claim, only a rendering of facts already resolved
 * by `deriveWorkspaceMode`.
 */
export function ModeChip({ mode, viewer }: ModeChipProps) {
  const label = MODE_LABEL[mode];
  const title = describeMode(mode, viewer);
  return (
    <span className={styles.chip} data-mode={mode} title={title}>
      {label}
    </span>
  );
}
