import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useMfaCoverageQuery } from "../queries/useMfaCoverageQuery";
import { getRoleLabel } from "../presenters/role-label-presenter";
import styles from "./MfaCoverageCard.module.css";

const CRITICAL_ROLES: ReadonlySet<string> = new Set([
  "signer",
  "approver",
  "system_admin",
  "qms_admin",
]);

const pctFmt = new Intl.NumberFormat("pt-BR", {
  maximumFractionDigits: 1,
});

function formatPct(value: number): string {
  return `${pctFmt.format(value)}%`;
}

function fillClass(pct: number, isCritical: boolean): string {
  if (isCritical && pct < 100) return styles.barFillDanger;
  if (pct < 70) return styles.barFillWarn;
  return "";
}

export default function MfaCoverageCard() {
  const { data, isLoading, error } = useMfaCoverageQuery();

  return (
    <section
      className={styles.card}
      aria-labelledby="mfa-coverage-title"
      data-testid="mfa-coverage-card"
    >
      <header className={styles.header}>
        <h3 id="mfa-coverage-title" className={styles.title}>
          Cobertura de MFA
        </h3>
        <p className={styles.subtitle}>
          Usuários com segundo fator ativo, por papel. Funções sensíveis
          destacadas.
        </p>
      </header>

      {isLoading ? (
        <div className={styles.empty} aria-busy="true">
          Carregando…
        </div>
      ) : null}

      {error ? (
        <div role="alert" className={styles.error}>
          {resolveQueryError(error, "Não foi possível carregar a cobertura de MFA.")}
        </div>
      ) : null}

      {data ? (
        <>
          <div className={styles.heroRow}>
            <span className={styles.heroPct}>{formatPct(data.mfaEnabledPct)}</span>
            <span className={styles.heroSuffix}>
              {data.mfaEnabled} de {data.totalUsers} usuários
            </span>
          </div>

          <div className={styles.roles} role="list">
            {data.byRole.map((slice) => {
              const isCritical = CRITICAL_ROLES.has(slice.role);
              const labelClass = isCritical
                ? `${styles.roleLabel} ${styles.critical}`
                : styles.roleLabel;
              return (
                <div key={slice.role} className={styles.roleRow} role="listitem">
                  <span className={labelClass}>
                    {isCritical ? (
                      <span className={styles.criticalDot} aria-hidden="true" />
                    ) : null}
                    {getRoleLabel(slice.role)}
                  </span>
                  <span className={styles.pct}>{formatPct(slice.pct)}</span>
                  <div
                    className={styles.bar}
                    role="progressbar"
                    aria-valuenow={Math.round(slice.pct)}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-label={`${getRoleLabel(slice.role)}: ${formatPct(slice.pct)} com MFA`}
                  >
                    <div
                      className={`${styles.barFill} ${fillClass(slice.pct, isCritical)}`}
                      style={{ width: `${Math.min(100, Math.max(0, slice.pct))}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </>
      ) : null}
    </section>
  );
}
