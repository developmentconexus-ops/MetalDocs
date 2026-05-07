import type { ReactNode } from 'react';
import { Stepper, type StepperStep } from '../../../../components/ui/Stepper';
import styles from './WizardShell.module.css';

const WIZARD_STEPS: StepperStep[] = [
  { id: '1', label: 'Perfil' },
  { id: '2', label: 'Área & Código' },
  { id: '3', label: 'Template' },
  { id: '4', label: 'Confirmação' },
];

export type WizardStepNumber = 1 | 2 | 3 | 4;

export type WizardShellProps = {
  currentStep: WizardStepNumber;
  onStepClick?: (step: WizardStepNumber) => void;
  children: ReactNode;
};

export function WizardShell({ currentStep, onStepClick, children }: WizardShellProps): JSX.Element {
  return (
    <div className={styles.scrollWrapper}>
      <div className={styles.container}>
        <div className={styles.header}>
          <div className="kicker">Documentos / Novo</div>
          <h1 className="display-1">Novo documento controlado</h1>
          <p className={styles.description}>
            Cada documento recebe um código único{' '}
            <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
          </p>
        </div>
        <Stepper
          steps={WIZARD_STEPS}
          current={String(currentStep)}
          onStepClick={onStepClick ? (id) => onStepClick(Number(id) as WizardStepNumber) : undefined}
        />
        {children}
      </div>
    </div>
  );
}

export default WizardShell;
