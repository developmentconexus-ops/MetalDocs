import type { components } from "../../../lib/api-types";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { useSecuritySignalsQuery } from "../queries/useSecuritySignalsQuery";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import styles from "./SecuritySignalsList.module.css";

type Signal = components["schemas"]["SecuritySignal"];
type Kind = Signal["kind"];
type Severity = Signal["severity"];

const KIND_LABEL: Record<Kind, string> = {
  "repeated-failed-login": "Logins falhos repetidos",
  "new-device-login": "Login de novo dispositivo",
  "lockout-spike": "Pico de bloqueios",
  "off-hours-admin-action": "Ação admin fora de horário",
};

const MITIGATION: Record<Kind, string> = {
  "repeated-failed-login":
    "Revisar tentativas, considerar bloqueio temporário ou rotação de senha.",
  "new-device-login":
    "Confirmar com o usuário e exigir MFA caso ainda não esteja ativo.",
  "lockout-spike":
    "Investigar origem (possível brute-force) e ampliar throttle no IP.",
  "off-hours-admin-action":
    "Validar com o operador e abrir registro de auditoria detalhado.",
};

const SEVERITY_LABEL: Record<Severity, string> = {
  info: "Info",
  warn: "Atenção",
  critical: "Crítico",
};

function severityClass(sev: Severity): { signal: string; badge: string } {
  if (sev === "critical")
    return { signal: styles.signalCritical, badge: styles.badgeCritical };
  if (sev === "warn")
    return { signal: styles.signalWarn, badge: styles.badgeWarn };
  return { signal: styles.signalInfo, badge: styles.badgeInfo };
}

function readCount(evidence: Signal["evidence"]): number | undefined {
  if (!evidence) return undefined;
  for (const key of ["count", "total", "attempts", "occurrences"]) {
    const v = (evidence as Record<string, unknown>)[key];
    if (typeof v === "number") return v;
  }
  return undefined;
}

export default function SecuritySignalsList() {
  const { data, isLoading, error } = useSecuritySignalsQuery("7d");

  return (
    <section
      className={styles.panel}
      aria-labelledby="security-signals-title"
      data-testid="security-signals-list"
    >
      <header className={styles.header}>
        <h3 id="security-signals-title" className={styles.title}>
          Sinais de segurança · últimos 7 dias
        </h3>
        <p className={styles.subtitle}>
          Anomalias detectadas. Cada métrica vem acompanhada da mitigação
          recomendada.
        </p>
      </header>

      {isLoading ? (
        <div className={styles.empty} aria-busy="true">
          Carregando…
        </div>
      ) : null}

      {error ? (
        <div role="alert" className={styles.error}>
          {resolveQueryError(error, "Não foi possível carregar os sinais.")}
        </div>
      ) : null}

      {data && data.items.length === 0 ? (
        <div className={styles.empty}>
          Nenhum sinal anômalo nos últimos 7 dias.
        </div>
      ) : null}

      {data && data.items.length > 0 ? (
        <ul className={styles.list} role="list">
          {data.items.map((s) => {
            const sev = severityClass(s.severity);
            const count = readCount(s.evidence);
            return (
              <li
                key={s.signalId}
                className={`${styles.signal} ${sev.signal}`}
                role="listitem"
              >
                <span className={`${styles.badge} ${sev.badge}`}>
                  {SEVERITY_LABEL[s.severity]}
                </span>
                <div className={styles.body}>
                  <span className={styles.summary}>
                    {KIND_LABEL[s.kind]} — {s.summary}
                  </span>
                  <span className={styles.mitigation}>{MITIGATION[s.kind]}</span>
                </div>
                <div className={styles.meta}>
                  {count !== undefined ? (
                    <span className={styles.metric}>{count}</span>
                  ) : null}
                  <time
                    className={styles.time}
                    dateTime={s.detectedAt}
                    title={SP_DATE_TIME_FORMATTER.format(new Date(s.detectedAt))}
                  >
                    {getRelativeTime(s.detectedAt)}
                  </time>
                </div>
              </li>
            );
          })}
        </ul>
      ) : null}
    </section>
  );
}
