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
  void revisionVersion;
  void docStatus;

  const autosaveLabel = (() => {
    if (autosaveStatus === 'saving') return 'Salvando…';
    if (autosaveStatus === 'error') return 'Erro ao salvar';
    if (autosaveStatus === 'saved') return 'Salvo';
    return 'Salvo · há 12s';
  })();

  return (
    <>
      {/* Left: back button */}
      <div className={styles.overlayLeft}>
        <button type="button" className={styles.backBtn} onClick={onBack} aria-label="Voltar">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>
      </div>

      {/* Center: code chip + doc name + version (floats over eigenpal title bar) */}
      <div className={styles.overlayCenter}>
        {code && <CodeChip>{code}</CodeChip>}
        {documentName && <span className={styles.docName}>{documentName}</span>}
        <span className={styles.versionMeta}>{"v5 · rascunho"}</span>
      </div>

      {/* Right: autosave + actions */}
      <div className={styles.overlayRight}>
        <span className={`${styles.autosave} ${autosaveStatus === 'error' ? styles.autosaveError : ''}`}>
          <span className={`${styles.autosaveDot} ${autosaveStatus === 'saving' ? styles.autosaveDotSaving : autosaveStatus === 'error' ? styles.autosaveDotError : ''}`} />
          {autosaveLabel}
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
    </>
  );
}
