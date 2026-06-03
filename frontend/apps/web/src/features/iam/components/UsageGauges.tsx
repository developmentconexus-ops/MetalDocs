import { useQueryClient } from "@tanstack/react-query";
import { useUsageQuery } from "../queries/useUsageQuery";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { QK } from "../../../lib/queryKeys";
import ErrorBanner from "./ErrorBanner";
import styles from "./UsageGauges.module.css";

const integerFormatter = new Intl.NumberFormat("pt-BR");
const percentFormatter = new Intl.NumberFormat("pt-BR", {
  style: "percent",
  maximumFractionDigits: 1,
});
const BYTES_PER_KB = 1024;
const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB", "PB"] as const;

const WARN_THRESHOLD = 0.75;
const DANGER_THRESHOLD = 0.9;

function formatBytes(bytes: number): { value: string; unit: string } {
  if (!Number.isFinite(bytes) || bytes <= 0) {
    return { value: "0", unit: "B" };
  }
  const exp = Math.min(
    BYTE_UNITS.length - 1,
    Math.floor(Math.log(bytes) / Math.log(BYTES_PER_KB)),
  );
  const value = bytes / Math.pow(BYTES_PER_KB, exp);
  const fractionDigits = value >= 100 ? 0 : value >= 10 ? 1 : 2;
  return {
    value: new Intl.NumberFormat("pt-BR", {
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    }).format(value),
    unit: BYTE_UNITS[exp],
  };
}

function ratio(used: number, allocated: number): number {
  if (!allocated || allocated <= 0) return 0;
  return Math.max(0, Math.min(1, used / allocated));
}

function toneFor(ratioValue: number): "ok" | "warn" | "danger" {
  if (ratioValue >= DANGER_THRESHOLD) return "danger";
  if (ratioValue >= WARN_THRESHOLD) return "warn";
  return "ok";
}

interface GaugeCardProps {
  id: string;
  title: string;
  subtitle: string;
  used: number;
  allocated: number;
  isBytes?: boolean;
  hint?: string;
}

function GaugeCard({ id, title, subtitle, used, allocated, isBytes, hint }: GaugeCardProps) {
  const r = ratio(used, allocated);
  const tone = toneFor(r);
  const fillClass =
    tone === "danger" ? styles.fillDanger : tone === "warn" ? styles.fillWarn : styles.fillOk;
  const percentClass =
    tone === "danger" ? styles.percentDanger : tone === "warn" ? styles.percentWarn : styles.percentOk;
  const percentText = percentFormatter.format(r);

  const usedDisplay = isBytes ? formatBytes(used) : { value: integerFormatter.format(used), unit: "" };
  const allocatedDisplay = isBytes
    ? formatBytes(allocated)
    : { value: integerFormatter.format(allocated), unit: "" };

  const ariaText = `${title}: ${usedDisplay.value}${usedDisplay.unit ? " " + usedDisplay.unit : ""} de ${allocatedDisplay.value}${allocatedDisplay.unit ? " " + allocatedDisplay.unit : ""}, ${percentText}.`;

  return (
    <article className={styles.card} data-testid={`usage-gauge-${id}`}>
      <header className={styles.cardHeader}>
        <h3 className={styles.title}>{title}</h3>
        <span className={styles.subtitle}>{subtitle}</span>
      </header>

      <div className={styles.gaugeBlock}>
        <div className={styles.values}>
          <span className={styles.valueUsed}>
            {usedDisplay.value}
            {usedDisplay.unit ? <span className={styles.valueAllocated}> {usedDisplay.unit}</span> : null}
          </span>
          <span className={styles.valueAllocated}>
            / {allocatedDisplay.value}
            {allocatedDisplay.unit ? ` ${allocatedDisplay.unit}` : ""}
          </span>
          <span className={`${styles.percent} ${percentClass}`} aria-hidden="true">
            {percentText}
          </span>
        </div>

        <div
          className={styles.track}
          role="progressbar"
          aria-label={title}
          aria-valuenow={Math.round(r * 100)}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuetext={ariaText}
        >
          <div
            className={`${styles.fill} ${fillClass}`}
            style={{ width: `${Math.round(r * 1000) / 10}%` }}
          />
        </div>

        {hint ? <p className={styles.hint}>{hint}</p> : null}
      </div>
    </article>
  );
}

interface WindowStats {
  last24h: number;
  last7d: number;
  last30d: number;
}

interface StatsCardProps {
  id: string;
  title: string;
  subtitle: string;
  data: WindowStats;
  hint?: string;
}

function StatsCard({ id, title, subtitle, data, hint }: StatsCardProps) {
  return (
    <article className={styles.card} data-testid={`usage-stats-${id}`}>
      <header className={styles.cardHeader}>
        <h3 className={styles.title}>{title}</h3>
        <span className={styles.subtitle}>{subtitle}</span>
      </header>

      <dl className={styles.statsRow}>
        <div className={styles.statCell}>
          <dt className={styles.statLabel}>24 horas</dt>
          <dd className={styles.statValue}>{integerFormatter.format(data.last24h)}</dd>
        </div>
        <div className={styles.statCell}>
          <dt className={styles.statLabel}>7 dias</dt>
          <dd className={styles.statValue}>{integerFormatter.format(data.last7d)}</dd>
        </div>
        <div className={styles.statCell}>
          <dt className={styles.statLabel}>30 dias</dt>
          <dd className={styles.statValue}>{integerFormatter.format(data.last30d)}</dd>
        </div>
      </dl>

      {hint ? <p className={styles.hint}>{hint}</p> : null}
    </article>
  );
}

interface UnavailableCardProps {
  id: string;
  title: string;
  subtitle: string;
  body: string;
}

function UnavailableCard({ id, title, subtitle, body }: UnavailableCardProps) {
  return (
    <article className={styles.card} data-testid={`usage-unavailable-${id}`}>
      <header className={styles.cardHeader}>
        <h3 className={styles.title}>{title}</h3>
        <span className={styles.unavailable} aria-label="Não disponível">
          Não disponível
        </span>
      </header>
      <p className={styles.description}>{subtitle}</p>
      <div className={styles.unavailableBlock}>
        <span className={styles.unavailableTitle}>Dados ainda não expostos pela API</span>
        <span>{body}</span>
      </div>
    </article>
  );
}

function SkeletonCard({ id }: { id: string }) {
  return (
    <article className={styles.card} data-testid={`usage-skeleton-${id}`} aria-hidden="true">
      <header className={styles.cardHeader}>
        <span className={`${styles.skeletonValue} ${styles.skeletonTitle}`} />
        <span className={`${styles.skeletonValue} ${styles.skeletonSubtitle}`} />
      </header>
      <div className={styles.gaugeBlock}>
        <span className={styles.skeletonValue} />
        <div className={styles.skeleton} />
      </div>
    </article>
  );
}

export default function UsageGauges() {
  const { data, isLoading, error } = useUsageQuery();
  const queryClient = useQueryClient();

  if (error) {
    return (
      <div data-testid="usage-error">
        <ErrorBanner
          message={resolveQueryError(error, "Não foi possível carregar o consumo.")}
          onRetry={() => queryClient.invalidateQueries({ queryKey: QK.iam.usage() })}
        />
      </div>
    );
  }

  if (isLoading || !data) {
    return (
      <div className={styles.grid} data-testid="usage-gauges-loading">
        <SkeletonCard id="a" />
        <SkeletonCard id="b" />
        <SkeletonCard id="c" />
        <SkeletonCard id="d" />
      </div>
    );
  }

  return (
    <div className={styles.grid} data-testid="usage-gauges">
      <GaugeCard
        id="seats"
        title="Assentos licenciados"
        subtitle="Capacidade"
        used={data.seats.used}
        allocated={data.seats.allocated}
        hint={`${integerFormatter.format(Math.max(0, data.seats.allocated - data.seats.used))} assentos disponíveis.`}
      />

      <GaugeCard
        id="storage"
        title="Armazenamento"
        subtitle="Capacidade"
        used={data.storage.usedBytes}
        allocated={data.storage.allocatedBytes}
        isBytes
        hint={`${formatBytes(Math.max(0, data.storage.allocatedBytes - data.storage.usedBytes)).value} ${formatBytes(Math.max(0, data.storage.allocatedBytes - data.storage.usedBytes)).unit} livres.`}
      />

      <StatsCard
        id="active-users"
        title="Usuários ativos"
        subtitle="Janelas"
        data={data.activeUsers}
        hint={
          data.seats.allocated <= 0
            ? `Ativos vs. licenciados em 30 dias: ${integerFormatter.format(data.activeUsers.last30d)} (sem assentos provisionados).`
            : `Ativos vs. licenciados em 30 dias: ${integerFormatter.format(data.activeUsers.last30d)} de ${integerFormatter.format(data.seats.allocated)}.`
        }
      />

      <StatsCard
        id="api-calls"
        title="Chamadas de API"
        subtitle="Janelas"
        data={data.apiCalls}
        hint="Volume agregado de requisições ao tenant."
      />

      <UnavailableCard
        id="growth-trend"
        title="Tendência (12 semanas)"
        subtitle="Histórico de crescimento de assentos e armazenamento"
        body="A API ainda não expõe a série semanal. Será integrada quando o endpoint de tendência estiver disponível."
      />

      <UnavailableCard
        id="plan-tier"
        title="Plano e direitos"
        subtitle="Tier contratado e matriz de entitlements"
        body="A API ainda não expõe o plano e a matriz de direitos. Será integrada quando o endpoint de plano estiver disponível."
      />
    </div>
  );
}
