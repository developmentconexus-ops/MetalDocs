import { useId, useRef } from 'react';
import { useRovingRadioGroup } from '../../../../../components/ui/useRovingRadioGroup';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import type { StartingPoint } from '../../../state/templateWizard.reducer';
import styles from './StepStructure.module.css';

export type StepStructureProps = {
  startingPoint: StartingPoint | null;
  selectedDocxName: string | null;
  selectedDocxSize: number | null;
  onSelectStartingPoint: (value: StartingPoint) => void;
  onSelectDocx: (name: string, size: number) => void;
  onClearDocx: () => void;
  onAdvance: () => void;
  onBack: () => void;
  advanceDisabled: boolean;
};

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function StepStructure({
  startingPoint,
  selectedDocxName,
  selectedDocxSize,
  onSelectStartingPoint,
  onSelectDocx,
  onClearDocx,
  onAdvance,
  onBack,
  advanceDisabled,
}: StepStructureProps): JSX.Element {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const fileInputId = useId();
  const startIndex = startingPoint === 'docx' ? 0 : startingPoint === 'blank' ? 1 : -1;
  const startGroup = useRovingRadioGroup({
    count: 2,
    selectedIndex: startIndex,
    onSelect: (newIndex) => onSelectStartingPoint(newIndex === 0 ? 'docx' : 'blank'),
    orientation: 'horizontal',
    ariaLabel: 'Ponto de partida do template',
  });

  // TODO(novo-template-wizard:step3-docx-upload): replace with real presigned upload
  // when wizard ordering allows mid-flow create. Backlog: wiki/backlog/novo-template-wizard.md.
  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    onSelectDocx(file.name, file.size);
    // reset native input so picking the same file twice still triggers change
    e.target.value = '';
  }

  function pickDocx() {
    onSelectStartingPoint('docx');
    fileInputRef.current?.click();
  }

  function pickBlank() {
    onSelectStartingPoint('blank');
  }

  return (
    <div className="card">
      <div className="kicker">Etapa 3 de 5</div>
      <h2 className="h2">Estrutura do template</h2>
      <p className={`caption ${styles.intro}`}>
        Suba um <span className="mono">.docx</span> base ou comece em branco. Placeholders no formato{' '}
        <span className="mono">{'{{campo}}'}</span> serão detectados automaticamente no editor.
      </p>

      {/* Two starting points — radiogroup (mutex selection) */}
      <div className={styles.startingPointGrid} {...startGroup.groupProps}>
        <button
          {...startGroup.getItemProps(0)}
          type="button"
          role="radio"
          aria-checked={startingPoint === 'docx'}
          className={`${styles.startingPointCard} ${startingPoint === 'docx' ? styles.selected : ''}`}
          onClick={pickDocx}
        >
          <div className={styles.thumbnail} aria-hidden="true">
            <span className={`mono ${styles.thumbnailDocxLabel}`}>DOCX</span>
          </div>
          <div className={styles.cardBody}>
            <div className={styles.cardTitleRow}>
              <span className={styles.cardTitle}>A partir de um .docx</span>
              {startingPoint === 'docx' && (
                <span className={styles.cardCheck} aria-hidden="true">
                  ✓
                </span>
              )}
            </div>
            <p className={styles.cardDesc}>
              Importa formatação. Recomendado quando você já tem um documento de referência.
            </p>
          </div>
        </button>

        <button
          {...startGroup.getItemProps(1)}
          type="button"
          role="radio"
          aria-checked={startingPoint === 'blank'}
          className={`${styles.startingPointCard} ${startingPoint === 'blank' ? styles.selected : ''}`}
          onClick={pickBlank}
        >
          <div className={styles.thumbnail} aria-hidden="true">
            +
          </div>
          <div className={styles.cardBody}>
            <div className={styles.cardTitleRow}>
              <span className={styles.cardTitle}>Em branco</span>
              {startingPoint === 'blank' && (
                <span className={styles.cardCheck} aria-hidden="true">
                  ✓
                </span>
              )}
            </div>
            <p className={styles.cardDesc}>
              Cria estrutura no editor visual a partir do zero. Para templates novos sem documento de referência.
            </p>
          </div>
        </button>
      </div>

      {/* Hidden native file input — triggered by .docx card click. */}
      <input
        ref={fileInputRef}
        id={fileInputId}
        type="file"
        accept=".docx,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
        className={styles.fileInput}
        onChange={handleFileChange}
        aria-hidden="true"
        tabIndex={-1}
      />

      {/* Selected docx row — only when docx chosen AND a file is picked. */}
      {startingPoint === 'docx' && selectedDocxName && (
        <div className={styles.fileRow}>
          <div className={styles.fileIcon} aria-hidden="true">
            DOCX
          </div>
          <div className={styles.fileMeta}>
            <div className={styles.fileName} title={selectedDocxName}>
              {selectedDocxName}
            </div>
            <div className={`mono ${styles.fileSize}`}>
              {selectedDocxSize !== null ? formatBytes(selectedDocxSize) : ''}
            </div>
          </div>
          <button
            type="button"
            className="btn btn-sm btn-ghost"
            onClick={() => {
              onClearDocx();
              fileInputRef.current?.click();
            }}
          >
            Substituir
          </button>
        </div>
      )}

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? startingPoint === 'docx'
              ? 'Etapa 3 de 5 · Selecione um .docx para continuar'
              : 'Etapa 3 de 5 · Escolha um ponto de partida'
            : 'Etapa 3 de 5 · Pronto para avançar'
        }
        showBack
        onBack={onBack}
        onAdvance={onAdvance}
        primaryDisabled={advanceDisabled}
      />
    </div>
  );
}

export default StepStructure;
