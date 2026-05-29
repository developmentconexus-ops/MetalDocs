import styles from './DocRefCard.module.css';

interface DocRefCardProps {
  areaLabel: string;
  code: string;
  typeLabel: string;
  versionLabel: string;
}

export function DocRefCard({ areaLabel, code, typeLabel, versionLabel }: DocRefCardProps) {
  return (
    <div className={styles.card}>
      <div className={styles.headerStrip}>{areaLabel}</div>
      <div className={styles.body}>
        <div className={styles.code}>{code}</div>
        <div className={styles.typeShort}>{typeLabel}</div>
        <div className={styles.spacer} />
        <div className={styles.divider} />
        <div className={styles.footer}>
          <span className={styles.version}>{versionLabel}</span>
          <span className={styles.statusDot} />
        </div>
      </div>
    </div>
  );
}
