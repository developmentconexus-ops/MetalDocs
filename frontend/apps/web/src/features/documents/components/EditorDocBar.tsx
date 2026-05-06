import styles from './EditorDocBar.module.css';
import { CodeChip } from '../../../components/ui/CodeChip';

type EditorDocBarProps = {
  code?: string;
  documentName?: string;
  revisionVersion?: number;
  docStatus?: string;
  autosaveStatus?: 'idle' | 'saving' | 'saved' | 'error';
  isEditable?: boolean;
  onBack?: () => void;
  onCheckpoints?: () => void;
  exportButton?: React.ReactNode;
  onFinalize?: () => void;
};

export function EditorDocBar({
  code,
  documentName,
  revisionVersion,
  docStatus,
  autosaveStatus = 'idle',
  isEditable = false,
  onBack,
  onCheckpoints,
  exportButton,
  onFinalize,
}: EditorDocBarProps) {
  void documentName;
  void revisionVersion;
  void docStatus;
  void autosaveStatus;

  return (
    <div className={styles.bar}>
      <button type="button" className={styles.backBtn} onClick={onBack} aria-label="Voltar">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="15 18 9 12 15 6" />
        </svg>
      </button>
      {code && <CodeChip>{code}</CodeChip>}
      <span className={styles.docName}>{"Admissão e Onboarding de Colaboradores"}</span>
      <span className={styles.versionMeta}>{"v5 · rascunho"}</span>
      <span className={styles.spacer} />
      <span className={styles.autosave}>
        <span className={styles.autosaveDot} />
        {"Salvo · há 12s"}
      </span>
      <button type="button" className={styles.ghostBtn} onClick={onCheckpoints}>
        Revisões
      </button>
      {exportButton}
      <button
        type="button"
        className={styles.primaryBtn}
        onClick={onFinalize}
        disabled={!isEditable}
      >
        Submeter para revisão
      </button>
    </div>
  );
}
