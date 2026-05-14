import type { WizardInvitee } from '../../../../state/wizard.reducer';
import styles from './StepAreaCodeVisibility.module.css';

export type PeopleSubcontrolsProps = {
  invitees: ReadonlyArray<WizardInvitee>;
  /* TODO(novo-documento:sharing): wire when /share endpoint lands.
     wiki/backlog/novo-documento.md#sharing */
  onAddInvitee?: (invitee: WizardInvitee) => void;
  onRemoveInvitee?: (id: string) => void;
};

export function PeopleSubcontrols({
  invitees,
  onAddInvitee: _onAddInvitee,
  onRemoveInvitee: _onRemoveInvitee,
}: PeopleSubcontrolsProps): JSX.Element {
  // Keep disabled until a real invite/search endpoint is available so the UI
  // stays truthful about current capabilities.
  return (
    <div
      className={`card ${styles.subcontrolsCard}`}
      role="group"
      aria-disabled="true"
      title="Em breve"
    >
      <div className="kicker">Pessoas convidadas</div>
      <div className={styles.subcontrolsRow}>
        {invitees.length === 0 ? (
          <span className="caption">Em breve — convite por pessoa.</span>
        ) : (
          invitees.map((row) => (
            <span key={row.id} className="pill">
              {row.label}{' '}
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                aria-label={`Remover ${row.label}`}
                disabled
                aria-disabled="true"
              >
                <span aria-hidden="true">×</span>
              </button>
            </span>
          ))
        )}
      </div>
      <button
        type="button"
        className={`btn btn-sm ${styles.subcontrolsAddBtn}`}
        disabled
        aria-disabled="true"
        title="Em breve"
      >
        Adicionar pessoa
      </button>
    </div>
  );
}
