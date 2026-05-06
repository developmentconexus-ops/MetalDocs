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

export function LibraryStatCards({ statsByStatus }: Props): JSX.Element {
  return (
    <div className={styles.root}>
      {CARDS.map(({ statusKey, label, mod }) => {
        const count = statsByStatus[statusKey] ?? 0;
        return (
          <div key={statusKey} className={`${styles.card} ${styles[mod]}`}>
            <span className={styles.label}>{label}</span>
            <span className={styles.count}>{count.toLocaleString('pt-BR')}</span>
          </div>
        );
      })}
    </div>
  );
}
