// Phase 3a: structure mirror only — no state, no handlers, no queries.
// Phase 3c wires state management and navigation.
import { WizardShell, type WizardShellStep } from '../../../shared/components/wizard/WizardShell';
import { StepScope } from '../components/wizard/steps/StepScope';

const TPL_STEPS: WizardShellStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Identidade' },
  { id: '3', label: 'Estrutura' },
  { id: '4', label: 'Permissões' },
  { id: '5', label: 'Confirmação' },
];

export function TemplateWizardPage(): JSX.Element {
  return (
    <WizardShell
      kicker="Templates / Novo"
      title="Novo template reutilizável"
      description={
        <>
          Templates publicados ficam disponíveis para autores criarem novos documentos. Use placeholders{' '}
          <span className="mono">{'{{campo}}'}</span> para campos dinâmicos.
        </>
      }
      steps={TPL_STEPS}
      currentStep="1"
    >
      <StepScope />
    </WizardShell>
  );
}

export { TemplateWizardPage as Component };

export default TemplateWizardPage;
