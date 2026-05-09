// Phase 3a: structure mirror only — no state, no handlers, no queries.
// Phase 3c wires state management and navigation.
import { Stepper, type StepperStep } from '../../../components/ui/Stepper';
import { StepScope } from '../components/wizard/steps/StepScope';
import styles from './TemplateWizardPage.module.css';

const TPL_STEPS: StepperStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Identidade' },
  { id: '3', label: 'Estrutura' },
  { id: '4', label: 'Permissões' },
  { id: '5', label: 'Confirmação' },
];

export function TemplateWizardPage(): JSX.Element {
  return (
    <div className={styles.scrollWrapper}>
      <div className={styles.container}>
        <div className={styles.header}>
          <div className="kicker">Templates / Novo</div>
          <h1 className="display-1">Novo template reutilizável</h1>
          <p className={styles.description}>
            Templates publicados ficam disponíveis para autores criarem novos documentos. Use placeholders{' '}
            <span className="mono">{'{{campo}}'}</span> para campos dinâmicos.
          </p>
        </div>
        <Stepper
          steps={TPL_STEPS}
          current="1"
        />
        <StepScope />
      </div>
    </div>
  );
}

export default TemplateWizardPage;
