import type { WizardExternalConfig } from '../../../../state/wizard.reducer';
import styles from './StepAreaCodeVisibility.module.css';

export type ExternalSubcontrolsProps = {
  external: WizardExternalConfig;
  /* TODO(novo-documento:sharing): wire when /share endpoint lands.
     wiki/backlog/novo-documento.md#sharing */
  onSetExternal?: (patch: Partial<WizardExternalConfig>) => void;
};

export function ExternalSubcontrols({
  external,
  onSetExternal: _onSetExternal,
}: ExternalSubcontrolsProps): JSX.Element {
  // Defer per NOTES.md audit — external-share model not yet defined. Rendered
  // disabled. See wiki/backlog/novo-documento.md#sharing.
  return (
    <div
      className={`card ${styles.subcontrolsCard}`}
      role="group"
      aria-disabled="true"
      title="Em breve"
    >
      <div className="kicker">Compartilhamento externo</div>
      <div className={styles.subcontrolsCol}>
        <label>
          <input type="checkbox" checked={external.passwordRequired} onChange={() => {}} disabled aria-disabled="true" />{' '}
          Exigir senha (em breve)
        </label>
        <label>
          <input type="checkbox" checked={external.watermark} onChange={() => {}} disabled aria-disabled="true" />{' '}
          Aplicar marca d'água (em breve)
        </label>
        <label>
          Expira em (dias)
          <input
            className="input"
            type="number"
            min={0}
            value={external.expiresInDays ?? ''}
            disabled
            aria-disabled="true"
          />
        </label>
      </div>
    </div>
  );
}
