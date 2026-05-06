import { WizardShell, type WizardStepNumber } from '../components/wizard/WizardShell';
import { StepProfile } from '../components/wizard/steps/StepProfile';
import { StepAreaCodeVisibility } from '../components/wizard/steps/StepAreaCodeVisibility';
import { StepTemplate } from '../components/wizard/steps/StepTemplate';
import { StepConfirm } from '../components/wizard/steps/StepConfirm';

export type NewDocumentWizardPageProps = Record<string, never>;

// TODO(novo-documento): wire URL ?step=N + reducer dispatches in Phase 3c.
// Switch this constant manually for Phase 3b screenshot review.
const CURRENT_STEP = 1 as WizardStepNumber;

export function NewDocumentWizardPage(): JSX.Element {
  const step = CURRENT_STEP;

  return (
    <WizardShell currentStep={step}>
      {step === 1 && <StepProfile />}
      {step === 2 && <StepAreaCodeVisibility />}
      {step === 3 && <StepTemplate />}
      {step === 4 && <StepConfirm />}
    </WizardShell>
  );
}

export default NewDocumentWizardPage;
