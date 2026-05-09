import styles from './TemplateWizardShell.module.css';

export type TemplateWizardFooterProps = {
  stepLabel: string;
  primaryLabel?: string;
  primaryDisabled?: boolean;
  showBack?: boolean;
  onCancel?: () => void;
  onBack?: () => void;
  onAdvance?: () => void;
};

export function TemplateWizardFooter({
  stepLabel,
  primaryLabel = 'Avançar →',
  primaryDisabled = false,
  showBack = true,
  onCancel,
  onBack,
  onAdvance,
}: TemplateWizardFooterProps): JSX.Element {
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
          className="btn btn-primary"
          disabled={primaryDisabled}
          onClick={onAdvance}
        >
          {primaryLabel}
        </button>
      </div>
    </>
  );
}
