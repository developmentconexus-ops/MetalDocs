import styles from './ActivityPanel.module.css';

type InboxItem = {
  code: string;
  due: string;
  dueTone: 'danger' | 'warning' | 'muted';
};

type AuditItem = {
  id: string;
  initials: string;
  who: string;
  action: string;
  target: string;
  time: string;
  isSystem?: boolean;
};

const INBOX: InboxItem[] = [
  { code: 'POP-QUA-118', due: 'Vence em 6h', dueTone: 'danger' },
  { code: 'POP-QUA-119', due: 'Vence em 2d', dueTone: 'warning' },
  { code: 'POL-RH-003', due: 'Sem prazo', dueTone: 'muted' },
];

const AUDIT: AuditItem[] = [
  { id: 'a1', initials: 'JP', who: 'Juliana P.', action: 'aprovou', target: 'POP-FIN-019', time: '14:32' },
  { id: 'a2', initials: 'S',  who: 'Sistema',    action: 'gerou PDF', target: 'IT-PROD-204', time: '13:51', isSystem: true },
  { id: 'a3', initials: 'RC', who: 'Rafael C.',  action: 'submeteu', target: 'POP-QUA-119', time: '12:04' },
  { id: 'a4', initials: 'AL', who: 'André L.',   action: 'editou',   target: 'POP-TI-007',  time: '11:24' },
  { id: 'a5', initials: 'MS', who: 'Marina S.',  action: 'arquivou', target: 'POP-RH-099',  time: '10:17' },
  { id: 'a6', initials: 'BM', who: 'Bruno M.',   action: 'congelou', target: 'IT-PROD-203', time: '09:44' },
  { id: 'a7', initials: 'CT', who: 'Camila T.',  action: 'criou rascunho', target: 'DC-RH-042', time: '09:12' },
];

type Props = {
  onClose: () => void;
};

export function ActivityPanel({ onClose }: Props): JSX.Element {
  return (
    <aside className={styles.root} aria-label="Atividade e auditoria">
      <div className={styles.header}>
        <div>
          <span className={styles.title}>Atividade &amp; auditoria</span>
          <p className={styles.sub}>Tempo real · últimas 8h</p>
        </div>
        <button type="button" className={styles.close} aria-label="Fechar" onClick={onClose}>
          ✕
        </button>
      </div>

      <section className={styles.section}>
        <h3 className={styles.sectionLabel}>SUA CAIXA · {INBOX.length} PENDENTES</h3>
        <ul className={styles.inboxList}>
          {INBOX.map((it) => (
            <li key={it.code} className={styles.inboxItem}>
              <span className={styles.inboxCode}>{it.code}</span>
              <span className={`${styles.inboxDue} ${styles[`tone_${it.dueTone}`]}`}>{it.due}</span>
            </li>
          ))}
        </ul>
        <button type="button" className={styles.inboxButton}>
          Abrir caixa de aprovação
        </button>
      </section>

      <div className={styles.divider} />

      <section className={styles.section}>
        <h3 className={styles.sectionLabel}>TRILHA DE AUDITORIA</h3>
        <ul className={styles.auditList}>
          {AUDIT.map((it) => (
            <li key={it.id} className={styles.auditItem}>
              <span
                className={`${styles.avatar} ${it.isSystem ? styles.avatarSystem : ''}`}
                aria-hidden="true"
              >
                {it.initials}
              </span>
              <span className={styles.auditBody}>
                <span className={styles.auditWho}>{it.who}</span>{' '}
                <span className={styles.auditAction}>{it.action}</span>{' '}
                <span className={styles.auditTarget}>{it.target}</span>
              </span>
              <span className={styles.auditTime}>{it.time}</span>
            </li>
          ))}
        </ul>
        <button type="button" className={styles.viewAll}>Ver tudo →</button>
      </section>
    </aside>
  );
}
