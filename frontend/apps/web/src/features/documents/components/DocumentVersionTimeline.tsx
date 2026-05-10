import { useState } from 'react';
import styles from './DocumentVersionTimeline.module.css';

export type VersionEntry = {
  v: string;
  when: string;   // "18 mar 2025"
  author: string;
  summary: string;
  current: boolean;
};

type Props = {
  versions: VersionEntry[];
};

export function DocumentVersionTimeline({ versions }: Props) {
  const currentIdx = versions.findIndex((v) => v.current);
  const [hovered, setHovered] = useState(currentIdx >= 0 ? currentIdx : versions.length - 1);

  const hov = versions[hovered];

  return (
    <div className={styles.card}>
      {/* Track */}
      <div className={styles.track}>
        <div className={styles.line} />
        <div
          className={styles.pins}
          style={{ gridTemplateColumns: `repeat(${versions.length}, 1fr)` }}
        >
          {versions.map((v, i) => {
            const active = i === hovered;
            return (
              <button
                key={v.v}
                type="button"
                className={styles.pin}
                onMouseEnter={() => setHovered(i)}
                onFocus={() => setHovered(i)}
                aria-pressed={active}
                aria-label={`${v.v} — ${v.when}`}
              >
                <span
                  className={`${styles.pinDate} ${active ? styles.pinDateActive : ''}`}
                >
                  {v.when.split(' ').slice(0, 2).join(' ')}
                </span>
                <span
                  className={`${styles.pinDot} ${active ? styles.pinDotActive : ''} ${v.current ? styles.pinDotCurrent : ''}`}
                />
                <span
                  className={`${styles.pinLabel} ${active ? styles.pinLabelActive : ''} ${v.current ? styles.pinLabelCurrent : ''}`}
                >
                  {v.v}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Detail panel — key swap triggers fadeInUp animation */}
      <div key={hovered} className={styles.detail}>
        <div className={styles.detailHeader}>
          <span className={styles.detailVersion}>{hov.v}</span>
          {hov.current && (
            <span className={styles.detailBadge}>atual</span>
          )}
          <span className={styles.detailMeta}>
            · {hov.when} · {hov.author}
          </span>
        </div>
        <div className={styles.detailSummary}>{hov.summary}</div>
      </div>
    </div>
  );
}
