import { formatDueRelative } from '../../../../lib/format/dates';
import type { ApprovalInstance, StageInstance } from '../../api/approvalTypes';
import { LockBadge } from '../LockBadge';
import styles from './StageContextHeader.module.css';

interface StageContextHeaderProps {
  instance: ApprovalInstance;
  activeStage: StageInstance | undefined;
  lockedByInstanceId?: string;
}

/** "Ana, Bruno +2" — truncates the actor pool to the first two names. */
function formatPool(actors: StageInstance['actors']): string | null {
  if (!actors || actors.length === 0) return null;
  const names = actors.map((a) => a.display_name);
  if (names.length <= 2) return names.join(', ');
  return `${names.slice(0, 2).join(', ')} +${names.length - 2}`;
}

/**
 * Derived quorum progress (spec §4) — NOT a DTO field. Approval stage: "N de M
 * aprovaram" from actors where status === 'approved'. Review stage: pool count
 * only (verdicts aren't in the DTO). Hidden entirely when actors is empty —
 * never fabricate "0 de 0".
 */
function formatQuorum(stage: StageInstance): string | null {
  const actors = stage.actors ?? [];
  if (actors.length === 0) return null;

  if (stage.stage_kind === 'approval') {
    const approved = actors.filter((a) => a.status === 'approved').length;
    return `${approved} de ${actors.length} aprovaram`;
  }

  return `${actors.length} revisor${actors.length === 1 ? '' : 'es'}`;
}

export function StageContextHeader({ instance, activeStage, lockedByInstanceId }: StageContextHeaderProps) {
  const totalStages = instance.stages.length;
  const stageNumber = activeStage != null ? activeStage.stage_index + 1 : null;
  const pool = activeStage ? formatPool(activeStage.actors) : null;
  const quorum = activeStage ? formatQuorum(activeStage) : null;
  const due = activeStage ? formatDueRelative(activeStage.due_at) : null;

  return (
    <header className={styles.header}>
      <LockBadge lockedByInstanceId={lockedByInstanceId} />

      {stageNumber != null && (
        <div className={styles.stageMeta}>
          <span className={styles.stageIndex}>
            Etapa {stageNumber} de {totalStages}
          </span>
          {activeStage != null && <span className={styles.stageLabel}>{activeStage.label}</span>}
        </div>
      )}

      <div className={styles.chips}>
        {pool != null && <span className={styles.chip}>{pool}</span>}
        {quorum != null && <span className={styles.chip}>{quorum}</span>}
        {due != null && (
          <span className={`${styles.chip} ${due.overdue ? styles.chipOverdue : ''}`}>{due.label}</span>
        )}
      </div>
    </header>
  );
}
