import type { ReactNode } from 'react';
import { Stepper } from '../../../../components/ui/Stepper';
import styles from './WizardShell.module.css';

export type WizardShellStep = { id: string; label: string };

export type WizardShellProps = {
  kicker: string;
  title: string;
  description?: ReactNode;
  steps: WizardShellStep[];
  currentStep: string;
  onStepClick?: (id: string) => void;
  children: ReactNode;
};

export function WizardShell({
  kicker,
  title,
  description,
  steps,
  currentStep,
  onStepClick,
  children,
}: WizardShellProps): JSX.Element {
  return (
    <div className={styles.scrollWrapper}>
      <div className={styles.container}>
        <div className={styles.header}>
          <div className="kicker">{kicker}</div>
          <h1 className="display-1">{title}</h1>
          {description && <p className={styles.description}>{description}</p>}
        </div>
        <Stepper
          steps={steps}
          current={currentStep}
          onStepClick={onStepClick}
        />
        {children}
      </div>
    </div>
  );
}

export default WizardShell;
