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
  children: ReactNode;
};

export function WizardShell({ currentStep, children }: WizardShellProps): JSX.Element {
  return (
    <div className={styles.scrollWrapper}>
      <div className={styles.container}>
        <div className="kicker">Documentos / Novo</div>
        <h1 className="display-1">Novo documento controlado</h1>
        <p className="body">
          Cada documento recebe um código único{' '}
          <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
        </p>
        <Stepper steps={WIZARD_STEPS} current={String(currentStep)} />
        {children}
      </div>
    </div>
  );
}

export type WizardFooterProps = {
  stepLabel: string;
  primaryLabel?: string;
  primaryDisabled?: boolean;
  showBack?: boolean;
};

export function WizardFooter({
  stepLabel,
  primaryLabel = 'Avançar →',
  primaryDisabled = false,
  showBack = true,
}: WizardFooterProps): JSX.Element {
  return (
    <>
      <div className="divider" />
      <div className={styles.footerRow}>
        {showBack ? (
          <button type="button" className="btn">
            ← Voltar
          </button>
        ) : (
          <span />
        )}
        <span className="spacer" />
        <span className="caption">{stepLabel}</span>
        <button type="button" className="btn btn-primary" disabled={primaryDisabled}>
          {primaryLabel}
        </button>
      </div>
    </>
  );
}

export default WizardShell;
