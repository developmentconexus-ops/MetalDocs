import { InboxApprovalCard } from './InboxApprovalCard';
import styles from './InboxStack.module.css';

export function InboxStack() {
  return (
    <div className={styles.root}>
      <aside className={styles.queueRail}>
        <div className={styles.queueHeader}>
          <div className="kicker">SUA FILA</div>
          <div className={styles.queueCount}>
            <span className={styles.queueCountNum}>4</span>
            <span className="caption">decisões pendentes</span>
          </div>
          <div className="caption">
            <span className={styles.urgentToday}>2 vencem hoje</span>
          </div>
        </div>

        <button type="button" className={`${styles.queueItem} ${styles.queueItemActive}`}>
          <div className={styles.queueItemNumber}>01</div>
          <div className={styles.queueItemMeta}>
            <div className={styles.queueItemTop}>
              <span className={`${styles.queueItemCode} mono`}>POP-QUA-0148</span>
              <span className={styles.urgentDot} />
            </div>
            <div className={styles.queueItemTitle}>Inspeção de recebimento — solda MIG/MAG</div>
            <div className={styles.queueItemSub}>
              <span className={styles.queueItemDeadline}>3h 28min</span>
              <span className={styles.dot}>·</span>
              <span>Qualidade</span>
            </div>
          </div>
        </button>

        <button type="button" className={styles.queueItem}>
          <div className={styles.queueItemNumber}>02</div>
          <div className={styles.queueItemMeta}>
            <div className={styles.queueItemTop}>
              <span className={`${styles.queueItemCode} mono`}>IT-PROD-0072</span>
              <span className={styles.urgentDot} />
            </div>
            <div className={styles.queueItemTitle}>Setup da prensa hidráulica P-440</div>
            <div className={styles.queueItemSub}>
              <span className={styles.queueItemDeadline}>9h</span>
              <span className={styles.dot}>·</span>
              <span>Produção</span>
            </div>
          </div>
        </button>

        <button type="button" className={styles.queueItem}>
          <div className={styles.queueItemNumber}>03</div>
          <div className={styles.queueItemMeta}>
            <div className={styles.queueItemTop}>
              <span className={`${styles.queueItemCode} mono`}>POL-RH-0011</span>
            </div>
            <div className={styles.queueItemTitle}>Política de afastamento parcial</div>
            <div className={styles.queueItemSub}>
              <span className={styles.queueItemDeadline}>1 dia</span>
              <span className={styles.dot}>·</span>
              <span>RH</span>
            </div>
          </div>
        </button>

        <button type="button" className={styles.queueItem}>
          <div className={styles.queueItemNumber}>04</div>
          <div className={styles.queueItemMeta}>
            <div className={styles.queueItemTop}>
              <span className={`${styles.queueItemCode} mono`}>DC-TI-0203</span>
            </div>
            <div className={styles.queueItemTitle}>Diagrama de rede — DMZ planta 2</div>
            <div className={styles.queueItemSub}>
              <span className={styles.queueItemDeadline}>2 dias</span>
              <span className={styles.dot}>·</span>
              <span>TI</span>
            </div>
          </div>
        </button>
      </aside>

      <main className={styles.cardArea}>
        <div className={styles.cardNav}>
          <span className={`${styles.cardCounter} mono`}>01 / 04</span>
          <span className={styles.spacer} />
          <button type="button" className="btn btn-sm">← anterior</button>
          <button type="button" className="btn btn-sm">próximo →</button>
        </div>

        <div className={styles.cardStack}>
          <div className={`${styles.ghostCard} ${styles.ghostCard2}`} aria-hidden="true" />
          <div className={`${styles.ghostCard} ${styles.ghostCard1}`} aria-hidden="true" />
          <InboxApprovalCard />
        </div>

        <div className={styles.keyboardHints}>
          <kbd>A</kbd> aprovar · <kbd>D</kbd> devolver · <kbd>←/→</kbd> navegar
        </div>
      </main>
    </div>
  );
}
