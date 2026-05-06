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
        <div className="kicker">Documentos / Novo</div>
        <h1 className="display-1">Novo documento controlado</h1>
        <p className="body">
          Cada documento recebe um código único{' '}
          <span className="mono">{`{perfil}-{área}-{seq}`}</span> e clona o template publicado para o perfil.
        </p>
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

export type WizardFooterProps = {
  stepLabel: string;
  primaryLabel?: string;
  primaryDisabled?: boolean;
  primaryVariant?: 'advance' | 'submit';
  showBack?: boolean;
  onCancel?: () => void;
  onBack?: () => void;
  onAdvance?: () => void;
};

export function WizardFooter({
  stepLabel,
  primaryLabel = 'Avançar →',
  primaryDisabled = false,
  primaryVariant = 'advance',
  showBack = true,
  onCancel,
  onBack,
  onAdvance,
}: WizardFooterProps): JSX.Element {
  return (
    <>
      <div className="divider" />
      <div className={styles.footerRow}>
        {showBack ? (
          <button type="button" className="btn" onClick={onBack}>
            ← Voltar
          </button>
        ) : (
          <button type="button" className="btn" onClick={onCancel}>
            Cancelar
          </button>
        )}
        <span className="spacer" />
        <span className="caption">{stepLabel}</span>
        <button
          type="button"
          className={primaryVariant === 'submit' ? 'btn btn-primary' : 'btn btn-primary'}
          disabled={primaryDisabled}
          onClick={onAdvance}
        >
          {primaryLabel}
        </button>
      </div>
    </>
  );
}

export default WizardShell;
