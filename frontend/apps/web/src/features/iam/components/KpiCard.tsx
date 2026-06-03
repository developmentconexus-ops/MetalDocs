import { Link } from "react-router-dom";
import styles from "./KpiCard.module.css";

type KpiTone = "neutral" | "success" | "warning" | "danger" | "info";

interface KpiCardProps {
  label: string;
  value: number | string | null | undefined;
  unit?: string;
  tone?: KpiTone;
  isLoading?: boolean;
  to: string;
  ctaLabel?: string;
}

const toneClass: Record<KpiTone, string> = {
  neutral: styles.tone,
  success: `${styles.tone} ${styles.toneSuccess}`,
  warning: `${styles.tone} ${styles.toneWarning}`,
  danger: `${styles.tone} ${styles.toneDanger}`,
  info: `${styles.tone} ${styles.toneInfo}`,
};

const accentVar: Record<KpiTone, string> = {
  neutral: "var(--accent)",
  success: "var(--success)",
  warning: "var(--warning)",
  danger: "var(--danger)",
  info: "var(--info)",
};

const numberFormatter = new Intl.NumberFormat("pt-BR");

function formatValue(value: KpiCardProps["value"]): string {
  if (value === null || value === undefined) return "—";
  if (typeof value === "number") return numberFormatter.format(value);
  return value;
}

export default function KpiCard({
  label,
  value,
  unit,
  tone = "neutral",
  isLoading,
  to,
  ctaLabel = "Ver detalhes",
}: KpiCardProps) {
  return (
    <Link
      to={to}
      className={styles.card}
      data-testid="kpi-card"
      aria-label={`${label}: ${isLoading ? "carregando" : formatValue(value)}${unit ?? ""}`}
      style={{ "--accent-bar": accentVar[tone] } as React.CSSProperties}
    >
      <div className={styles.kicker}>
        <span>{label}</span>
        <span className={toneClass[tone]} aria-hidden="true" />
      </div>
      <div className={styles.value}>
        {isLoading ? (
          <span className={styles.valueLoading} aria-hidden="true" />
        ) : (
          <>
            {formatValue(value)}
            {unit ? <span className={styles.unit}>{unit}</span> : null}
          </>
        )}
      </div>
      <div className={styles.hint}>
        <span className={styles.cta}>{ctaLabel} →</span>
      </div>
    </Link>
  );
}
