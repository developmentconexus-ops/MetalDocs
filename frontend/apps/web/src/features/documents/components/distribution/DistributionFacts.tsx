import { TrackingPendingNote } from './TrackingPendingNote';
import styles from './DistributionFacts.module.css';

/**
 * Distribution facts / policy details card.
 * The policy fields (deadline, channel, reminders) are numerator-adjacent
 * and not served by the current backend endpoints. Replaced by honest state.
 */
export function DistributionFacts() {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.kicker}>Detalhes da distribuição</div>
      </div>
      <div className={styles.list}>
        <TrackingPendingNote />
      </div>
    </div>
  );
}
