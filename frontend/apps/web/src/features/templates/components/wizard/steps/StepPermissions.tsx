import { WizardFooter } from '../../../../shared/components/wizard/WizardFooter';
import { Icon } from '../../../../../components/ui/Icon';
import styles from './StepPermissions.module.css';

export type StepPermissionsProps = {
  onAdvance: () => void;
  onBack: () => void;
  advanceDisabled: boolean;
};

export function StepPermissions({
  onAdvance,
  onBack,
  advanceDisabled,
}: StepPermissionsProps): JSX.Element {
  return (
    <div className="card">
      <div className="kicker">Etapa 4 de 5</div>
      <h2 className={`h2 ${styles.stepTitle}`}>Visibilidade deste template</h2>
      <p className={`caption ${styles.intro}`}>
        A criacao atual do backend publica o rascunho como visivel para toda a empresa. Regras por area ou funcao
        precisam de contrato proprio antes de aparecerem neste wizard.
      </p>

      <div className={styles.allBanner}>
        <span className={styles.allIcon} aria-hidden="true">
          <Icon name="home" size={20} />
        </span>
        <div className={styles.allBody}>
          <div className={styles.allTitle}>Toda a empresa</div>
          <p className="caption">
            Estado real deste PR: o template sera criado com visibilidade publica. Seletor por areas, funcoes e
            contagens de usuarios permanecem no backlog ate a API expor esses campos.
          </p>
        </div>
      </div>

      <div className={styles.coverageSummary}>
        <div className={`kicker ${styles.coverageKicker}`}>Itens diferidos</div>
        <div className={styles.deferList}>
          <span>Permissoes por funcao</span>
          <span>Permissoes por area</span>
          <span>Contagens de usuarios</span>
        </div>
      </div>

      <WizardFooter
        stepLabel="Etapa 4 de 5 - Visibilidade publica atual confirmada"
        showBack
        onBack={onBack}
        onAdvance={onAdvance}
        primaryDisabled={advanceDisabled}
      />
    </div>
  );
}

export default StepPermissions;
