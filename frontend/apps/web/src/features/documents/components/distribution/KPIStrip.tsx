import styles from './KPIStrip.module.css';

const KPIS = [
  { label: 'Alvos totais',  value: 248, sub: '8 áreas · 1 doc',            bar: false, tone: ''        },
  { label: 'Reconheceram',  value: 156, sub: '62.9% da base',               bar: true,  tone: 'success' },
  { label: 'Apenas leram',  value: 26,  sub: '10.5% leram sem assinar',     bar: false, tone: ''        },
  { label: 'Pendentes',     value: 66,  sub: '6 áreas com pendência',       bar: false, tone: ''        },
  { label: 'Em atraso',     value: 12,  sub: '+2 vs. ontem',                bar: false, tone: 'danger'  },
] as const;

export function KPIStrip() {
  return (
    <div className={styles.strip}>
      {KPIS.map((k, i) => (
        <div
          key={k.label}
          className={`${styles.cell} ${i < KPIS.length - 1 ? styles.cellBorder : ''}`}
        >
          <div className={styles.kicker}>{k.label}</div>
          <div className={`${styles.value} ${k.tone === 'danger' ? styles.valueDanger : ''}`}>
            {k.value}
          </div>
          {k.bar && (
            <div className={styles.progressTrack}>
              <div className={styles.progressFill} style={{ width: '62.9%' }} />
            </div>
          )}
          <div className={styles.sub}>{k.sub}</div>
        </div>
      ))}
    </div>
  );
}
