import styles from './AutosaveStatus.module.css';

export type AutosaveState = 'idle' | 'saving' | 'saved' | 'error';

type AutosaveStatusProps = {
  status: AutosaveState;
  /** Optional override labels (defaults are pt-BR). */
  labels?: {
    idle?: string;
    saving?: string;
    saved?: string;
    error?: string;
  };
  className?: string;
};

const DEFAULT_LABELS = {
  idle: 'Salvo',
  saving: 'Salvando…',
  saved: 'Salvo',
  error: 'Erro ao salvar',
};

/**
 * Editor autosave indicator. Shows pulsing dot while saving,
 * green check when saved, red label on error, neutral idle.
 */
export function AutosaveStatus({ status, labels, className }: AutosaveStatusProps) {
  const lbl = { ...DEFAULT_LABELS, ...(labels ?? {}) };
  const isError = status === 'error';
  const wrapperClass = `${styles.status}${isError ? ` ${styles.statusError}` : ''}${className ? ` ${className}` : ''}`;

  if (status === 'saving') {
    return (
      <span className={wrapperClass}>
        <span className={`${styles.dot} ${styles.dotSaving}`} aria-hidden="true" />
        {lbl.saving}
      </span>
    );
  }
  if (status === 'error') {
    return (
      <span className={wrapperClass}>
        <span className={`${styles.dot} ${styles.dotError}`} aria-hidden="true" />
        {lbl.error}
      </span>
    );
  }
  if (status === 'saved') {
    return (
      <span className={wrapperClass}>
        <CheckIcon className={styles.check} />
        {lbl.saved}
      </span>
    );
  }
  return (
    <span className={wrapperClass}>
      <span className={`${styles.dot} ${styles.dotIdle}`} aria-hidden="true" />
      {lbl.idle}
    </span>
  );
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M20 6 9 17l-5-5" />
    </svg>
  );
}
