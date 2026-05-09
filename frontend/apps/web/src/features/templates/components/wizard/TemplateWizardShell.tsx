import type { ReactNode } from 'react';
import { Stepper, type StepperStep } from '../../../../components/ui/Stepper';
import styles from './TemplateWizardShell.module.css';

export type TemplateWizardShellProps = {
  steps: StepperStep[];
  currentStep: string;
  onStepClick?: (id: string) => void;
  children: ReactNode;
};

export function TemplateWizardShell({
  steps,
  currentStep,
  onStepClick,
  children,
}: TemplateWizardShellProps): JSX.Element {
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
        <Stepper steps={steps} current={currentStep} onStepClick={onStepClick} />
        {children}
      </div>
    </div>
  );
}
