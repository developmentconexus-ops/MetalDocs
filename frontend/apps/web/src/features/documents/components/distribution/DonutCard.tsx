import { TrackingPendingNote } from './TrackingPendingNote';
import styles from './DonutCard.module.css';

/**
 * Read/ack coverage donut — numerator surface.
 * Replaced by honest tracking-pending state until the backend producer ships.
 */
export function DonutCard() {
  return (
    <div className={styles.card}>
      <div className={styles.kicker}>Cobertura geral</div>
      <TrackingPendingNote />
    </div>
  );
}
