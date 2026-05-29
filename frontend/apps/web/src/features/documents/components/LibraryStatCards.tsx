import styles from './LibraryStatCards.module.css';

// Backend currently exposes only statsByStatus on GET /api/v1/documents/stats.
// "Aprovação pendente / Frozen este mês / Próx. revisão" require dedicated counts
// that the backend does not yet emit; render them as inert placeholders rather
// than display fabricated numbers.
type CardConfig = {
  key: string;
  label: string;
  mod: 'review' | 'pending' | 'frozen' | 'upcoming';
  value: (s: Record<string, number>) => number | null;
  hint: string;
};

const CARDS: CardConfig[] = [
  {
    key: 'review',
    label: 'Em revisão',
    mod: 'review',
    value: (s) => s['under_review'] ?? 0,
    hint: 'documentos',
  },
  {
    key: 'pending',
    label: 'Aprovação pendente',
    mod: 'pending',
    value: () => null,
    hint: 'sem fonte de dados',
  },
  {
    key: 'frozen',
    label: 'Frozen este mês',
    mod: 'frozen',
    value: () => null,
    hint: 'sem fonte de dados',
  },
  {
    key: 'upcoming',
    label: 'Próx. revisão',
    mod: 'upcoming',
    value: () => null,
    hint: 'sem fonte de dados',
  },
];

type Props = {
  total: number;
  statsByStatus: Record<string, number>;
};

export function LibraryStatCards({ statsByStatus }: Props): JSX.Element {
  return (
    <div className={styles.root}>
      {CARDS.map(({ key, label, mod, value, hint }) => {
        const raw = value(statsByStatus);
        const display = raw === null ? '—' : raw.toLocaleString('pt-BR');
        return (
          <div key={key} className={`${styles.card} ${styles[mod]}`}>
            <span className={styles.label}>{label}</span>
            <div className={styles.valueRow}>
              <span className={styles.count}>{display}</span>
              <span className={`${styles.trend} ${styles[`trend_${mod}`]}`}>{hint}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
