import styles from './DocRefCard.module.css';

export function DocRefCard() {
  return (
    <div className={styles.card}>
      <div className={styles.headerStrip}>
        SSMA
      </div>
      <div className={styles.body}>
        <div className={styles.code}>PR-EHS-014</div>
        <div className={styles.typeShort}>Procedimento</div>
        <div className={styles.spacer} />
        <div className={styles.divider} />
        <div className={styles.footer}>
          <span className={styles.version}>v3.2</span>
          <span className={styles.statusDot} />
        </div>
      </div>
    </div>
  );
}
