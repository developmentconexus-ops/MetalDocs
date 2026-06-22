import { TrackingPendingNote } from './TrackingPendingNote';
import styles from './TimelineCard.module.css';

/**
 * Adoption timeline chart — numerator surface.
 * Replaced by honest tracking-pending state until the backend producer ships.
 */
export function TimelineCard() {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.subtitle}>Curva de adoção desde a publicação</div>
      </div>
      <TrackingPendingNote />
    </div>
  );
}
