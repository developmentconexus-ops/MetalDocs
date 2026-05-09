// Shared WizardFooter — used by document and template wizards.
// Extracted from features/documents/components/wizard/WizardFooter.tsx (Phase 2 promotion).
import styles from './WizardFooter.module.css';

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
      <div className={`divider ${styles.footerDivider}`} />
      <div className={styles.footerRow}>
        {showBack ? (
          <button type="button" className="btn" onClick={onBack}>
            <span aria-hidden="true">←</span> Voltar
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
          className={primaryVariant === 'submit' ? 'btn btn-success' : 'btn btn-primary'}
          disabled={primaryDisabled}
          onClick={onAdvance}
        >
          {primaryLabel}
        </button>
      </div>
    </>
  );
}
