import styles from './ActivityPanel.module.css';

// Real activity feed requires two endpoints the backend does not yet expose:
//   - GET /api/v1/users/me/inbox     (pending approvals for current user)
//   - GET /api/v1/audit?scope=documents&since=… (audit trail surfaced via codegen)
// Until both land, render the panel as an explicit placeholder rather than ship
// fabricated activity entries that look authoritative.

type Props = {
  onClose: () => void;
};

export function ActivityPanel({ onClose }: Props): JSX.Element {
  return (
    <aside className={styles.root} aria-label="Atividade e auditoria">
      <div className={styles.header}>
        <div>
          <span className={styles.title}>Atividade &amp; auditoria</span>
          <p className={styles.sub}>Em breve</p>
        </div>
        <button type="button" className={styles.close} aria-label="Fechar" onClick={onClose}>
          ✕
        </button>
      </div>

      <section className={styles.section}>
        <h3 className={styles.sectionLabel}>SUA CAIXA</h3>
        <p className={styles.sub}>
          A caixa de aprovação aparecerá aqui assim que o endpoint estiver disponível.
        </p>
        <button type="button" className={styles.inboxButton} disabled aria-disabled="true" title="Em breve">
          Abrir caixa de aprovação
        </button>
      </section>

      <div className={styles.divider} />

      <section className={styles.section}>
        <h3 className={styles.sectionLabel}>TRILHA DE AUDITORIA</h3>
        <p className={styles.sub}>
          A trilha de auditoria do tenant aparecerá aqui quando o endpoint for exposto.
        </p>
        <button type="button" className={styles.viewAll} disabled aria-disabled="true" title="Em breve">Ver tudo →</button>
      </section>
    </aside>
  );
}
