import styles from './LibraryStatCards.module.css';

const CARDS = [
  { statusKey: 'under_review', label: 'Em Revisão', mod: 'review' },
  { statusKey: 'draft', label: 'Rascunhos', mod: 'draft' },
  { statusKey: 'published', label: 'Publicados', mod: 'published' },
  { statusKey: 'approved', label: 'Aprovados', mod: 'approved' },
] as const;

type Props = {
  total: number;
  statsByStatus: Record<string, number>;
};

export function LibraryStatCards({ total, statsByStatus }: Props): JSX.Element {
  return (
    <div className={styles.root}>
      <div className={`${styles.card} ${styles.total}`}>
        <span className={styles.count}>{total.toLocaleString('pt-BR')}</span>
        <span className={styles.label}>Total</span>
      </div>
      {CARDS.map(({ statusKey, label, mod }) => (
        <div key={statusKey} className={`${styles.card} ${styles[mod]}`}>
          <span className={styles.count}>{(statsByStatus[statusKey] ?? 0).toLocaleString('pt-BR')}</span>
          <span className={styles.label}>{label}</span>
        </div>
      ))}
    </div>
  );
}
