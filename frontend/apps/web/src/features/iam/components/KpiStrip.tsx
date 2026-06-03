import { useKpiQuery } from "../queries/useKpiQuery";
import { useSessionsQuery } from "../queries/useSessionsQuery";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import KpiCard from "./KpiCard";
import styles from "./KpiStrip.module.css";

function sumRoleDistribution(
  roles: ReadonlyArray<{ count: number }> | undefined,
): number {
  if (!roles) return 0;
  return roles.reduce((acc, r) => acc + (r.count ?? 0), 0);
}

function mfaTone(pct: number | undefined): "success" | "warning" | "danger" {
  if (pct === undefined) return "warning";
  if (pct >= 90) return "success";
  if (pct >= 60) return "warning";
  return "danger";
}

export default function KpiStrip() {
  const kpi = useKpiQuery();
  const sessions = useSessionsQuery({ limit: 1 });

  const snapshot = kpi.data;
  const totalUsers = sumRoleDistribution(snapshot?.roleDistribution);
  const failedLogins = snapshot?.failedLogins24h;
  const lockedAccounts = snapshot?.lockedAccounts;
  const mfaPct = snapshot?.mfaCoveragePct;
  const sessionsTotal = sessions.data?.total ?? sessions.data?.items?.length;

  const isLoading = kpi.isLoading;
  const hasError = !!kpi.error;

  return (
    <section
      className={styles.strip}
      aria-label="Indicadores principais"
      data-testid="kpi-strip"
    >
      {hasError ? (
        <div role="alert" className={styles.errorBanner}>
          <span>
            Não foi possível carregar os indicadores. {resolveQueryError(kpi.error, "Tente novamente.")}
          </span>
          <button
            type="button"
            className={styles.retryButton}
            onClick={() => kpi.refetch()}
          >
            Tentar novamente
          </button>
        </div>
      ) : null}

      <KpiCard
        label="Usuários ativos"
        value={totalUsers}
        tone="info"
        isLoading={isLoading}
        to="/admin/people"
        ctaLabel="Ver pessoas"
      />
      <KpiCard
        label="Cobertura MFA"
        value={mfaPct !== undefined ? Math.round(mfaPct) : null}
        unit="%"
        tone={mfaTone(mfaPct)}
        isLoading={isLoading}
        to="/admin/people?filter=mfa-off"
        ctaLabel="Ver lacunas"
      />
      <KpiCard
        label="Tentativas falhas (24h)"
        value={failedLogins}
        tone={failedLogins && failedLogins > 50 ? "danger" : "warning"}
        isLoading={isLoading}
        to="/admin/audit?action=auth.login.failed"
        ctaLabel="Investigar"
      />
      <KpiCard
        label="Contas bloqueadas"
        value={lockedAccounts}
        tone={lockedAccounts && lockedAccounts > 0 ? "danger" : "success"}
        isLoading={isLoading}
        to="/admin/people?filter=locked"
        ctaLabel="Desbloquear"
      />
      <KpiCard
        label="Sessões ativas"
        value={sessionsTotal}
        tone="neutral"
        isLoading={sessions.isLoading}
        to="/admin/sessions"
        ctaLabel="Ver sessões"
      />
    </section>
  );
}
