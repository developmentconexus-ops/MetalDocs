import { MOCK_DISTRIBUTION } from '../../lib/distributionMeta';
import styles from './DonutCard.module.css';

const { acknowledged, read, pending, totalTargets, deadline } = MOCK_DISTRIBUTION;
const readOnly = read - acknowledged; // "apenas leu" = read but not ack
const C = 2 * Math.PI * 42;
const ackPct = Math.round((acknowledged / totalTargets) * 100);
const readPct = Math.round((read / totalTargets) * 100);
const pendingPct = Math.round((pending / totalTargets) * 100);
const distantFromGoal = totalTargets - acknowledged;

export function DonutCard() {
  return (
    <div className={styles.card}>
      <div className={styles.kicker}>Cobertura geral</div>
      <div className={styles.body}>
        <div className={styles.donutWrap}>
          <svg viewBox="0 0 100 100" className={styles.donutSvg}>
            <circle
              cx="50" cy="50" r="42"
              fill="none" stroke="var(--surface-3)" strokeWidth="9"
            />
            <circle
              cx="50" cy="50" r="42"
              fill="none" stroke="var(--brand)" strokeWidth="9" strokeLinecap="round"
              opacity="0.32"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - readPct / 100)}
            />
            <circle
              cx="50" cy="50" r="42"
              fill="none" stroke="var(--success)" strokeWidth="9" strokeLinecap="round"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - ackPct / 100)}
            />
          </svg>
          <div className={styles.donutCenter}>
            <div className={styles.donutPct}>
              {ackPct}<span className={styles.donutPctSymbol}>%</span>
            </div>
            <div className={styles.donutLabel}>reconheceu</div>
          </div>
        </div>
        <div className={styles.legend}>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotAck}`} />
            <span className={styles.legendName}>Reconheceu (assinou)</span>
            <span className={styles.legendCount}>{acknowledged}</span>
            <span className={styles.legendPct}>{ackPct}%</span>
          </div>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotRead}`} />
            <span className={styles.legendName}>Apenas leu</span>
            <span className={styles.legendCount}>{readOnly}</span>
            <span className={styles.legendPct}>{Math.round((readOnly / totalTargets) * 100)}%</span>
          </div>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotPending}`} />
            <span className={styles.legendName}>Pendente</span>
            <span className={styles.legendCount}>{pending}</span>
            <span className={styles.legendPct}>{pendingPct}%</span>
          </div>
        </div>
      </div>
      <div className={styles.footer}>
        <span className={styles.footerMeta}>
          Meta de cobertura: <strong>92%</strong> até {deadline}
        </span>
        <span className={styles.footerWarning}>
          <span className={styles.warningDot} />
          {distantFromGoal} pessoas distantes da meta
        </span>
      </div>
    </div>
  );
}
