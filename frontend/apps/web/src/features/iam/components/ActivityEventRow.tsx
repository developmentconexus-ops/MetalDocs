import {
  getAuditEventLabel,
  getAuditEventSeverity,
} from "../presenters/audit-event-presenter";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import type { AuditEventItem } from "../queries/useAuditEventsQuery";
import styles from "./ActivityEventRow.module.css";

interface ActivityEventRowProps {
  event: AuditEventItem;
  groupedWithPrev?: boolean;
}

const severityClass: Record<string, string> = {
  info: styles.severityInfo,
  warn: styles.severityWarn,
  error: styles.severityError,
  success: styles.severitySuccess,
};

const absoluteFormatter = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "medium",
  timeStyle: "short",
});

export default function ActivityEventRow({
  event,
  groupedWithPrev,
}: ActivityEventRowProps) {
  const label = getAuditEventLabel(event.action);
  const severity = getAuditEventSeverity(event.action);
  const absolute = absoluteFormatter.format(new Date(event.occurred_at));

  return (
    <li
      className={`${styles.row} ${groupedWithPrev ? styles.grouped : ""}`}
      data-testid="activity-event-row"
    >
      <span
        className={`${styles.severityBar} ${severityClass[severity] ?? ""}`}
        aria-hidden="true"
      />
      <span className={styles.label}>{label}</span>
      <time
        className={styles.time}
        dateTime={event.occurred_at}
        title={absolute}
      >
        {getRelativeTime(event.occurred_at)}
      </time>
      <span className={styles.meta}>
        {groupedWithPrev ? (
          <span className={styles.groupedHint}>↳ mesmo ator</span>
        ) : (
          <span className={styles.actor}>{event.actor_id}</span>
        )}
        <span className={styles.resource}>
          {event.resource_type}/{event.resource_id.slice(0, 8)}
        </span>
      </span>
    </li>
  );
}
