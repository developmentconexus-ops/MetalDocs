import { useRovingRadioGroup } from '../../../../../components/ui/useRovingRadioGroup';
import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import type { StartingPoint } from '../../../state/templateWizard.reducer';
import styles from './StepStructure.module.css';

export type StepStructureProps = {
  startingPoint: StartingPoint | null;
  onSelectStartingPoint: (value: StartingPoint) => void;
  onAdvance: () => void;
  onBack: () => void;
  advanceDisabled: boolean;
};

export function StepStructure({
  startingPoint,
  onSelectStartingPoint,
  onAdvance,
  onBack,
  advanceDisabled,
}: StepStructureProps): JSX.Element {
  const startIndex = startingPoint === 'blank' ? 0 : -1;
  const startGroup = useRovingRadioGroup({
    count: 1,
    selectedIndex: startIndex,
    onSelect: (newIndex) => {
      if (newIndex === 0) onSelectStartingPoint('blank');
    },
    orientation: 'horizontal',
    ariaLabel: 'Ponto de partida do template',
  });

  function pickBlank() {
    onSelectStartingPoint('blank');
  }

  return (
    <div className="card">
      <div className="kicker">Etapa 3 de 5</div>
      <h2 className="h2">Estrutura do template</h2>
      <p className={`caption ${styles.intro}`}>
        Comece em branco e monte a estrutura no editor visual. Importacao de DOCX fica desabilitada ate existir um
        handoff real de upload para o editor.
      </p>

      <div className={styles.startingPointGrid}>
        <div
          className={`${styles.startingPointCard} ${styles.disabled}`}
          aria-disabled="true"
          title="Em breve: importacao DOCX depende de upload apos criar o template."
        >
          <div className={styles.thumbnail} aria-hidden="true">
            <span className={`mono ${styles.thumbnailDocxLabel}`}>DOCX</span>
          </div>
          <div className={styles.cardBody}>
            <div className={styles.cardTitleRow}>
              <span className={styles.cardTitle}>A partir de um .docx</span>
              <span className={styles.soonBadge}>em breve</span>
            </div>
            <p className={styles.cardDesc}>
              Requer contrato de upload e handoff para o editor. Sem mock de arquivo neste wizard.
            </p>
          </div>
        </div>

        <div {...startGroup.groupProps} className={styles.radioGroupReset}>
          <button
            {...startGroup.getItemProps(0)}
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
                    OK
                  </span>
                )}
              </div>
              <p className={styles.cardDesc}>
                Cria uma versao inicial em rascunho para editar a estrutura no editor visual.
              </p>
            </div>
          </button>
        </div>
      </div>

      <WizardFooter
        stepLabel={
          advanceDisabled
            ? 'Etapa 3 de 5 - Escolha o ponto de partida em branco'
            : 'Etapa 3 de 5 - Pronto para avancar'
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
