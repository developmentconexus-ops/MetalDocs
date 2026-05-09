import styles from './CoverageByArea.module.css';
import { MOCK_DISTRIBUTION } from '../../lib/distributionMeta';

// sorted by lowest ack coverage first (mirrors design sort order)
const SORTED_AREAS = [...MOCK_DISTRIBUTION.byArea].sort(
  (a, b) => a.ack / a.total - b.ack / b.total,
);

export function CoverageByArea() {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.headerNote}>
          Áreas ordenadas por menor cobertura primeiro · linha tracejada = meta de 92%
        </span>
        <div className={styles.legend}>
          <span className={styles.legendChip}>
            <span className={`${styles.legendSwatch} ${styles.legendSwatchAck}`} />
            Reconheceu
          </span>
          <span className={styles.legendChip}>
            <span className={`${styles.legendSwatch} ${styles.legendSwatchRead}`} />
            Apenas leu
          </span>
          <span className={styles.legendChip}>
            <span className={`${styles.legendSwatch} ${styles.legendSwatchGoal}`} />
            Meta 92%
          </span>
        </div>
      </div>
      <div className={styles.rows}>
        {SORTED_AREAS.map((a) => {
          const pctAck  = Math.round((a.ack  / a.total) * 100);
          const pctRead = Math.round((a.read / a.total) * 100);
          return (
            <div key={a.area} className={styles.areaRow}>
              <div className={styles.areaLabel}>
                <div className={styles.areaName}>{a.area}</div>
                <div className={styles.areaSub}>{a.ack}/{a.total} reconheceram</div>
              </div>
              <div className={styles.barWrap}>
                <div className={styles.barTrack}>
                  <div
                    className={styles.barRead}
                    style={{ width: `${pctRead}%` }}
                  />
                  <div
                    className={styles.barAck}
                    style={{ width: `${pctAck}%` }}
                  />
                </div>
                <div className={styles.goalMarker} />
              </div>
              <div className={styles.pctWrap}>
                <span className={styles.pctValue}>{pctAck}%</span>
                <span className={styles.pctLabel}>ack</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
