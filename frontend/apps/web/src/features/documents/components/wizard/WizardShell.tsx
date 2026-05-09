// Thin wrapper that supplies hardcoded doc-wizard strings to the shared WizardShell.
// NewDocumentWizardPage uses numeric WizardStepNumber; this adapter converts for Stepper.
import { WizardShell as SharedWizardShell } from '../../../shared/components/wizard/WizardShell';

export type WizardStepNumber = 1 | 2 | 3 | 4;

const DOC_WIZARD_STEPS = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Área & Código' },
  { id: '3', label: 'Template' },
  { id: '4', label: 'Confirmação' },
];

export type WizardShellProps = {
  currentStep: WizardStepNumber;
  onStepClick?: (step: WizardStepNumber) => void;
  children: React.ReactNode;
};

export function WizardShell({ currentStep, onStepClick, children }: WizardShellProps): JSX.Element {
  return (
    <SharedWizardShell
      kicker="Documentos / Novo"
      title="Novo documento controlado"
      description={
        <>
          Cada documento recebe um código único{' '}
          <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
        </>
      }
      steps={DOC_WIZARD_STEPS}
      currentStep={String(currentStep)}
      onStepClick={onStepClick ? (id) => onStepClick(Number(id) as WizardStepNumber) : undefined}
    >
      {children}
    </SharedWizardShell>
  );
}

export default WizardShell;
