import styles from './AutosaveStatus.module.css';

export type AutosaveState =
  | 'idle' | 'dirty' | 'saving' | 'saved' | 'stale' | 'session_lost' | 'error';

type AutosaveStatusProps = {
  status: AutosaveState;
  /** Optional override labels (defaults are pt-BR). */
  labels?: Partial<Record<AutosaveState, string>>;
  className?: string;
};

const DEFAULT_LABELS: Record<AutosaveState, string> = {
  idle: 'Salvo',
  dirty: 'Editado',
  saving: 'Salvando…',
  saved: 'Salvo',
  stale: 'Atualização disponível',
  session_lost: 'Sessão perdida',
  error: 'Erro ao salvar',
};

/**
 * Editor autosave indicator. Mirrors the 7-state union from useDocumentAutosave.
 * - Announces errors assertively; all other transitions are polite.
 * - role="status" + aria-live keep screen-reader users informed without interrupting.
 */
export function AutosaveStatus({ status, labels, className }: AutosaveStatusProps) {
  const lbl = { ...DEFAULT_LABELS, ...(labels ?? {}) };
  const isError = status === 'error' || status === 'session_lost';
  const isWarn = status === 'stale';
  const wrapperClass =
    `${styles.status}` +
    (isError ? ` ${styles.statusError}` : '') +
    (isWarn ? ` ${styles.statusWarn}` : '') +
    (className ? ` ${className}` : '');
  const ariaLive: 'polite' | 'assertive' = isError ? 'assertive' : 'polite';

  return (
    <span className={wrapperClass} role="status" aria-live={ariaLive}>
      {renderIcon(status)}
      {lbl[status]}
    </span>
  );
}

function renderIcon(status: AutosaveState) {
  switch (status) {
    case 'saving':
      return <span className={`${styles.dot} ${styles.dotSaving}`} aria-hidden="true" />;
    case 'error':
    case 'session_lost':
      return <span className={`${styles.dot} ${styles.dotError}`} aria-hidden="true" />;
    case 'stale':
      return <span className={`${styles.dot} ${styles.dotWarn}`} aria-hidden="true" />;
    case 'dirty':
      return <span className={`${styles.dot} ${styles.dotDirty}`} aria-hidden="true" />;
    case 'saved':
    case 'idle':
    default:
      return <span className={`${styles.dot} ${styles.dotIdle}`} aria-hidden="true" />;
  }
}
