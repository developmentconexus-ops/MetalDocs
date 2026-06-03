import { useState } from "react";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { SP_DATE_TIME_FORMATTER } from "../constants";
import {
  getAuditEventLabel,
  getAuditEventSeverity,
  type AuditEventSeverity,
} from "../presenters/audit-event-presenter";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import type { AuditEventItem } from "../queries/useAuditEventsQuery";
import styles from "./AuditEventsTable.module.css";

export type { AuditEventItem };

interface AuditEventsTableProps {
  events: ReadonlyArray<AuditEventItem>;
  isLoading: boolean;
  error: unknown;
  onRetry: () => void;
  hasMore: boolean;
  isFetchingMore: boolean;
  onLoadMore: () => void;
}

const SEVERITY_CLASS: Readonly<Record<AuditEventSeverity, string>> = {
  info: styles.severityInfo,
  success: styles.severitySuccess,
  warn: styles.severityWarn,
  error: styles.severityError,
};

const SKELETON_ROWS = 8;

function initials(actorId: string): string {
  const trimmed = actorId.trim();
  if (!trimmed) return "?";
  const compact = trimmed.replace(/[-_]/g, " ").split(/\s+/);
  if (compact.length >= 2) {
    return (compact[0][0] + compact[1][0]).toUpperCase();
  }
  return trimmed.slice(0, 2).toUpperCase();
}

function shortResource(resourceType: string, resourceId: string): string {
  if (!resourceType && !resourceId) return "—";
  if (!resourceId) return resourceType;
  const id = resourceId.length > 12 ? `${resourceId.slice(0, 8)}…` : resourceId;
  return `${resourceType || "?"}/${id}`;
}

function isProblemPayload(
  payload: Record<string, unknown>,
): payload is Record<string, unknown> & { type: string; title: string } {
  return (
    typeof payload?.["type"] === "string" &&
    typeof payload?.["title"] === "string"
  );
}

export default function AuditEventsTable({
  events,
  isLoading,
  error,
  onRetry,
  hasMore,
  isFetchingMore,
  onLoadMore,
}: AuditEventsTableProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  return (
    <section
      className={styles.panel}
      data-testid="audit-events-table"
      aria-label="Eventos de auditoria"
    >
      <header className={styles.header}>
        <span className={styles.headerTitle}>Linha do tempo</span>
        <span className={styles.headerCount} aria-live="polite">
          {events.length} evento(s) carregado(s)
        </span>
      </header>

      {error ? (
        <div role="alert" className={styles.errorBanner}>
          <span>
            {resolveQueryError(error, "Não foi possível carregar a auditoria.")}
          </span>
          <button
            type="button"
            className={styles.retryButton}
            onClick={onRetry}
          >
            Tentar novamente
          </button>
        </div>
      ) : null}

      {isLoading ? (
        <div className={styles.skeleton} aria-hidden="true">
          {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
            <div key={i} className={styles.skeletonRow}>
              <div className={styles.skeletonBar} />
              <div
                className={styles.skeletonBar}
                style={{ maxWidth: `${50 + ((i * 13) % 40)}%` }}
              />
              <div className={styles.skeletonBar} />
            </div>
          ))}
        </div>
      ) : null}

      {!isLoading && !error && events.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyTitle}>Nenhum evento encontrado</span>
          <span>Ajuste os filtros ou amplie a janela de tempo.</span>
        </div>
      ) : null}

      {!isLoading && events.length > 0 ? (
        <ul className={styles.list} role="list">
          {events.map((ev, i) => {
            const prev = events[i - 1];
            const grouped = !!prev && prev.actorId === ev.actorId;
            const severity = getAuditEventSeverity(ev.action);
            const label = getAuditEventLabel(ev.action);
            const expanded = expandedId === ev.id;
            const absolute = SP_DATE_TIME_FORMATTER.format(
              new Date(ev.occurredAt),
            );
            return (
              <li key={ev.id}>
                <div
                  className={`${styles.row} ${grouped ? styles.grouped : ""} ${
                    expanded ? styles.expanded : ""
                  }`}
                  data-testid="audit-event-row"
                >
                  <span
                    className={`${styles.severityBar} ${SEVERITY_CLASS[severity]}`}
                    aria-hidden="true"
                  />
                  <time
                    className={styles.time}
                    dateTime={ev.occurredAt}
                    title={absolute}
                  >
                    {getRelativeTime(ev.occurredAt)}
                  </time>
                  <div className={styles.body}>
                    <span className={styles.label}>{label}</span>
                    <span className={styles.actionCode}>{ev.action}</span>
                  </div>
                  <div className={styles.actor}>
                    {grouped ? (
                      <span className={styles.groupedHint}>↳ mesmo ator</span>
                    ) : (
                      <>
                        <span className={styles.avatar} aria-hidden="true">
                          {initials(ev.actorId)}
                        </span>
                        <span className={styles.actorName} title={ev.actorId}>
                          {ev.actorId || "—"}
                        </span>
                      </>
                    )}
                  </div>
                  <span
                    className={styles.resource}
                    title={`${ev.resourceType}/${ev.resourceId}`}
                  >
                    {shortResource(ev.resourceType, ev.resourceId)}
                  </span>
                  <button
                    type="button"
                    aria-label={
                      expanded ? "Recolher payload" : "Expandir payload"
                    }
                    aria-expanded={expanded}
                    aria-controls={`audit-payload-${ev.id}`}
                    className={`${styles.expandBtn} ${expanded ? styles.open : ""}`}
                    onClick={() =>
                      setExpandedId((cur) =>
                        cur === ev.id ? null : ev.id,
                      )
                    }
                  >
                    <svg
                      width="14"
                      height="14"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M6 9l6 6 6-6" />
                    </svg>
                  </button>
                </div>

                {expanded ? (
                  <div
                    id={`audit-payload-${ev.id}`}
                    className={styles.payloadRow}
                  >
                    {isProblemPayload(ev.payload) ? (
                      <div className={styles.problemBanner} role="note">
                        <span className={styles.problemTitle}>
                          {String(ev.payload.title)}
                        </span>
                        {typeof ev.payload.detail === "string" ? (
                          <span className={styles.problemDetail}>
                            {ev.payload.detail}
                          </span>
                        ) : null}
                      </div>
                    ) : null}
                    <div className={styles.payloadHeader}>
                      <span>
                        <span className={styles.payloadKey}>Ocorrido em:</span>{" "}
                        {absolute}
                      </span>
                      {ev.traceId ? (
                        <span>
                          <span className={styles.payloadKey}>Trace:</span>{" "}
                          {ev.traceId}
                        </span>
                      ) : null}
                    </div>
                    <pre className={styles.payload}>
                      {JSON.stringify(ev.payload, null, 2)}
                    </pre>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ul>
      ) : null}

      {!isLoading && events.length > 0 ? (
        <div className={styles.footer}>
          <span className={styles.footerCount}>
            {events.length} evento(s)
          </span>
          {hasMore ? (
            <button
              type="button"
              className={styles.loadMore}
              onClick={onLoadMore}
              disabled={isFetchingMore}
            >
              {isFetchingMore ? "Carregando…" : "Carregar mais"}
            </button>
          ) : (
            <span className={styles.footerCount}>Fim da lista</span>
          )}
        </div>
      ) : null}
    </section>
  );
}
