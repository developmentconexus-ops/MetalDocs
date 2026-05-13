import styles from './LibraryStatCards.module.css';

// ─────────────────────────────────────────────────────────────────────────────
// TODO(library/stat-cards): three of four cards are mocked.
//   - "Em revisão"         → REAL: wired to statsByStatus.under_review.
//   - "Aprovação pendente" → MOCK: needs GET /api/v1/documents/stats extended
//     with pending_my_approval count (docs awaiting the current user's
//     signature, distinct from system-wide under_review).
//   - "Frozen este mês"    → MOCK: needs stats endpoint to return frozen_this_month
//     count + delta vs prior month. Backend issue not yet filed.
//   - "Próx. revisão"      → MOCK: needs GET /api/v1/documents?nextReviewWithin=60d
//     count. No endpoint yet.
// Trend strings are illustrative only. Replace all MOCK values once endpoints land.
// ─────────────────────────────────────────────────────────────────────────────

const CARDS = [
  { key: 'review',   label: 'Em revisão',         mod: 'review',   trend: '+3 hoje',         value: (s: Record<string, number>) => s['under_review'] ?? 0 },
  { key: 'pending',  label: 'Aprovação pendente',  mod: 'pending',  trend: '2 vencendo',      value: (_: Record<string, number>) => 3 },     // MOCK
  { key: 'frozen',   label: 'Frozen este mês',     mod: 'frozen',   trend: '+12% vs anterior', value: (_: Record<string, number>) => 47 },   // MOCK
  { key: 'upcoming', label: 'Próx. revisão',       mod: 'upcoming', trend: 'Maio · Jun',      value: (_: Record<string, number>) => 23 },     // MOCK
] as const;

type Props = {
  total: number;
  statsByStatus: Record<string, number>;
};

export function LibraryStatCards({ statsByStatus }: Props): JSX.Element {
  return (
    <div className={styles.root}>
      {CARDS.map(({ key, label, mod, trend, value }) => (
        <div key={key} className={`${styles.card} ${styles[mod]}`}>
          <span className={styles.label}>{label}</span>
          <div className={styles.valueRow}>
            <span className={styles.count}>{value(statsByStatus).toLocaleString('pt-BR')}</span>
            <span className={`${styles.trend} ${styles[`trend_${mod}`]}`}>{trend}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
