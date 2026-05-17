import type { ChangeEvent } from 'react';
import { useRovingRadioGroup } from '../../../../../components/ui/useRovingRadioGroup';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import type { StartingPoint } from '../../../state/templateWizard.reducer';
import styles from './StepStructure.module.css';

export type StepStructureProps = {
  startingPoint: StartingPoint | null;
  selectedDocxName: string | null;
  selectedDocxSize: number | null;
  onSelectStartingPoint: (value: StartingPoint) => void;
  onSelectDocxFile: (file: File) => void;
  onAdvance: () => void;
  onBack: () => void;
  advanceDisabled: boolean;
};

function formatFileSize(size: number): string {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

export function StepStructure({
  startingPoint,
  selectedDocxName,
  selectedDocxSize,
  onSelectStartingPoint,
  onSelectDocxFile,
  onAdvance,
  onBack,
  advanceDisabled,
}: StepStructureProps): JSX.Element {
  const startIndex = startingPoint === 'docx' ? 0 : startingPoint === 'blank' ? 1 : -1;
  const startGroup = useRovingRadioGroup({
    count: 2,
    selectedIndex: startIndex,
    onSelect: (newIndex) => {
      onSelectStartingPoint(newIndex === 0 ? 'docx' : 'blank');
    },
    orientation: 'horizontal',
    ariaLabel: 'Ponto de partida do template',
  });

  function handleDocxChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.currentTarget.files?.[0];
    if (file) {
      onSelectDocxFile(file);
    }
  }

  return (
    <div className="card">
      <div className="kicker">Etapa 3 de 4</div>
      <h2 className="h2">Estrutura do template</h2>
      <p className={`caption ${styles.intro}`}>
        Comece em branco ou importe um arquivo .docx para abrir a primeira versão com conteúdo no editor.
      </p>

      <div {...startGroup.groupProps} className={styles.startingPointGrid}>
        <label
          {...startGroup.getItemProps(0)}
          role="radio"
          aria-checked={startingPoint === 'docx'}
          className={`${styles.startingPointCard} ${startingPoint === 'docx' ? styles.selected : ''}`}
          onClick={() => onSelectStartingPoint('docx')}
        >
          <input
            className={styles.fileInput}
            type="file"
            accept=".docx,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
            onChange={handleDocxChange}
          />
          <div className={styles.thumbnail} aria-hidden="true">
            <span className={`mono ${styles.thumbnailDocxLabel}`}>DOCX</span>
          </div>
          <div className={styles.cardBody}>
            <div className={styles.cardTitleRow}>
              <span className={styles.cardTitle}>A partir de um .docx</span>
              {startingPoint === 'docx' && (
                <span className={styles.cardCheck} aria-hidden="true">
                  OK
                </span>
              )}
            </div>
            <p className={styles.cardDesc}>
              Envia o arquivo após criar o template e abre a versão 1.0 com o DOCX importado.
            </p>
          </div>
        </label>

        <button
          {...startGroup.getItemProps(1)}
          type="button"
          role="radio"
          aria-checked={startingPoint === 'blank'}
          className={`${styles.startingPointCard} ${startingPoint === 'blank' ? styles.selected : ''}`}
          onClick={() => onSelectStartingPoint('blank')}
        >
          <div className={styles.thumbnail} aria-hidden="true">
            +
          </div>
          <div className={styles.cardBody}>
            <div className={styles.cardTitleRow}>
              <span className={styles.cardTitle}>Em branco</span>
              {startingPoint === 'blank' && (
                <span className={styles.cardCheck} aria-hidden="true">
                  OK
                </span>
              )}
            </div>
            <p className={styles.cardDesc}>
              Cria uma versão inicial em rascunho para editar a estrutura no editor visual.
            </p>
          </div>
        </button>
      </div>

      {selectedDocxName && selectedDocxSize !== null && (
        <div className={styles.fileRow}>
          <div className={styles.fileIcon} aria-hidden="true">
            DOCX
          </div>
          <div className={styles.fileMeta}>
            <div className={styles.fileName}>{selectedDocxName}</div>
            <div className={styles.fileSize}>{formatFileSize(selectedDocxSize)}</div>
          </div>
        </div>
      )}

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 3 de 4 - Escolha o ponto de partida'
            : 'Etapa 3 de 4 - Pronto para avançar'
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
