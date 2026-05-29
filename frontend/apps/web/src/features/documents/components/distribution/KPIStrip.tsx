import { MOCK_DISTRIBUTION } from '../../lib/distributionMeta';
import styles from './KPIStrip.module.css';

const { totalTargets, acknowledged, read, pending, overdue, byArea } = MOCK_DISTRIBUTION;
const readOnly = read - acknowledged;
const ackPct = ((acknowledged / totalTargets) * 100).toFixed(1);
const readOnlyPct = ((readOnly / totalTargets) * 100).toFixed(1);
const areasWithPending = byArea.filter(a => a.ack < a.total).length;

const KPIS = [
  { label: 'Alvos totais',  value: totalTargets, sub: `${byArea.length} áreas · 1 doc`,              bar: false, tone: ''        },
  { label: 'Reconheceram',  value: acknowledged,  sub: `${ackPct}% da base`,                          bar: true,  tone: 'success', barPct: `${ackPct}%` },
  { label: 'Apenas leram',  value: readOnly,      sub: `${readOnlyPct}% leram sem assinar`,           bar: false, tone: ''        },
  { label: 'Pendentes',     value: pending,       sub: `${areasWithPending} áreas com pendência`,     bar: false, tone: ''        },
  { label: 'Em atraso',     value: overdue,       sub: '+2 vs. ontem',                                bar: false, tone: 'danger'  },
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
              <div className={styles.progressFill} style={{ width: ('barPct' in k ? k.barPct : '0%') }} />
            </div>
          )}
          <div className={styles.sub}>{k.sub}</div>
        </div>
      ))}
    </div>
  );
}
