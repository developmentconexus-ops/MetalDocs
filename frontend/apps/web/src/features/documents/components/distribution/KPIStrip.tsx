import { TrackingPendingNote } from './TrackingPendingNote';
import styles from './KPIStrip.module.css';

export interface KPIStripProps {
  totalTargets: number | string;
  loading?: boolean;
}

export function KPIStrip({ totalTargets, loading = false }: KPIStripProps) {
  return (
    <div className={styles.strip}>
      {/* Only denominator cell: total_targets is real */}
      <div className={`${styles.cell} ${styles.cellBorder}`}>
        <div className={styles.kicker}>Alvos totais</div>
        <div className={styles.value}>
          {loading ? '…' : totalTargets}
        </div>
        <div className={styles.sub}>destinatários obrigatórios</div>
      </div>
      {/* Numerator cells — honest tracking-pending */}
      <div className={`${styles.cell} ${styles.cellSpan4}`}>
        <TrackingPendingNote />
      </div>
    </div>
  );
}
