import styles from './DonutCard.module.css';

const C = 2 * Math.PI * 42;
const ACK_PCT = 63;
const READ_PCT = 73; // read includes ack portion

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
              strokeDashoffset={C * (1 - READ_PCT / 100)}
            />
            <circle
              cx="50" cy="50" r="42"
              fill="none" stroke="var(--success)" strokeWidth="9" strokeLinecap="round"
              strokeDasharray={C}
              strokeDashoffset={C * (1 - ACK_PCT / 100)}
            />
          </svg>
          <div className={styles.donutCenter}>
            <div className={styles.donutPct}>
              63<span className={styles.donutPctSymbol}>%</span>
            </div>
            <div className={styles.donutLabel}>reconheceu</div>
          </div>
        </div>
        <div className={styles.legend}>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotAck}`} />
            <span className={styles.legendName}>Reconheceu (assinou)</span>
            <span className={styles.legendCount}>156</span>
            <span className={styles.legendPct}>63%</span>
          </div>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotRead}`} />
            <span className={styles.legendName}>Apenas leu</span>
            <span className={styles.legendCount}>26</span>
            <span className={styles.legendPct}>10%</span>
          </div>
          <div className={styles.legendRow}>
            <span className={`${styles.legendDot} ${styles.legendDotPending}`} />
            <span className={styles.legendName}>Pendente</span>
            <span className={styles.legendCount}>66</span>
            <span className={styles.legendPct}>27%</span>
          </div>
        </div>
      </div>
      <div className={styles.footer}>
        <span className={styles.footerMeta}>
          Meta de cobertura: <strong>92%</strong> até 22 mar 2026
        </span>
        <span className={styles.footerWarning}>
          <span className={styles.warningDot} />
          60 pessoas distantes da meta
        </span>
      </div>
    </div>
  );
}
