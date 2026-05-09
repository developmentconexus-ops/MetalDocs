import { MOCK_DISTRIBUTION } from '../../lib/distributionMeta';
import styles from './TimelineCard.module.css';

const DATA = MOCK_DISTRIBUTION.dailyReads;
const TOTAL = MOCK_DISTRIBUTION.totalTargets;
const GOAL = Math.round(TOTAL * 0.92); // 228

const W = 1100, H = 180, PAD_L = 50, PAD_R = 30, PAD_T = 20, PAD_B = 30;
const INNER_W = W - PAD_L - PAD_R;
const INNER_H = H - PAD_T - PAD_B;

function xs(i: number) {
  return PAD_L + (i / (DATA.length - 1)) * INNER_W;
}
function ys(v: number) {
  return PAD_T + INNER_H - (v / TOTAL) * INNER_H;
}

const linePath = DATA.map((d, i) => `${i === 0 ? 'M' : 'L'} ${xs(i)} ${ys(d.cumAck)}`).join(' ');
const areaPath = `${linePath} L ${xs(DATA.length - 1)} ${PAD_T + INNER_H} L ${xs(0)} ${PAD_T + INNER_H} Z`;

const lastPoint = DATA[DATA.length - 1];
const secondToLast = DATA[DATA.length - 2];

export function TimelineCard() {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.subtitle}>
          Reconhecimentos acumulados desde a publicação
        </div>
        <div className={styles.legend}>
          <span className={styles.legendChip}>
            <span className={styles.legendLine} />
            Reconhecimentos
          </span>
          <span className={styles.legendChip}>
            <span className={styles.legendDash} />
            Meta · 228 (92%)
          </span>
        </div>
      </div>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        className={styles.svg}
        style={{ height: H }}
      >
        <defs>
          <linearGradient id="distTrendGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--brand)" stopOpacity="0.18" />
            <stop offset="100%" stopColor="var(--brand)" stopOpacity="0" />
          </linearGradient>
        </defs>

        {/* Y grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map((p, i) => (
          <line
            key={`yg-${i}`}
            x1={PAD_L} x2={W - PAD_R}
            y1={PAD_T + INNER_H * (1 - p)} y2={PAD_T + INNER_H * (1 - p)}
            stroke="var(--border)" strokeWidth="1"
            strokeDasharray={i === 0 ? '' : '2 4'}
          />
        ))}
        {/* Y labels */}
        {[0, 0.25, 0.5, 0.75, 1].map((p, i) => (
          <text
            key={`yl-${i}`}
            x={PAD_L - 8} y={PAD_T + INNER_H * (1 - p) + 3}
            fontSize="9.5" fill="var(--text-muted)"
            fontFamily="var(--font-mono)" textAnchor="end"
          >
            {Math.round(TOTAL * p)}
          </text>
        ))}

        {/* Goal line */}
        <line
          x1={PAD_L} x2={W - PAD_R}
          y1={ys(GOAL)} y2={ys(GOAL)}
          stroke="var(--text-faint)" strokeWidth="1.5" strokeDasharray="4 4"
        />
        <text
          x={W - PAD_R - 4} y={ys(GOAL) - 4}
          fontSize="9.5" fill="var(--text-muted)"
          fontFamily="var(--font-mono)" textAnchor="end"
        >
          meta · {GOAL}
        </text>

        {/* Area fill */}
        <path d={areaPath} fill="url(#distTrendGrad)" />
        {/* Line */}
        <path d={linePath} fill="none" stroke="var(--brand)" strokeWidth="2" />
        {/* Data points */}
        {DATA.map((d, i) => (
          <circle
            key={`pt-${i}`}
            cx={xs(i)} cy={ys(d.cumAck)}
            r={i === DATA.length - 1 ? 4.5 : 2.8}
            fill="var(--brand)"
            stroke="var(--surface)"
            strokeWidth={i === DATA.length - 1 ? 3 : 1.5}
          />
        ))}

        {/* X labels */}
        {DATA.map((d, i) => (
          <text
            key={`xl-${i}`}
            x={xs(i)} y={H - 10}
            fontSize="9.5" fill="var(--text-muted)"
            fontFamily="var(--font-mono)" textAnchor="middle"
          >
            {d.day}
          </text>
        ))}

        {/* "hoje" pill on last point */}
        <g transform={`translate(${xs(DATA.length - 1) - 60}, ${ys(lastPoint.cumAck) - 28})`}>
          <rect width="56" height="20" rx="4" fill="var(--brand)" />
          <text
            x="28" y="13" textAnchor="middle"
            fontSize="11" fontWeight="600" fill="#fff"
            fontFamily="var(--font-mono)"
          >
            +{lastPoint.cumAck - secondToLast.cumAck} hoje
          </text>
        </g>
      </svg>
    </div>
  );
}
