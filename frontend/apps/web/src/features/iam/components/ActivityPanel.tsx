import { Link } from "react-router-dom";
import { useAuditEventsQuery } from "../queries/useAuditEventsQuery";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import ActivityEventRow from "./ActivityEventRow";
import styles from "./ActivityPanel.module.css";

const SKELETON_ROWS = 6;
const LIMIT = 20;

export default function ActivityPanel() {
  const { data, isLoading, error, refetch } = useAuditEventsQuery({ limit: LIMIT });
  const events = data?.pages[0]?.items ?? [];

  return (
    <section
      className={styles.panel}
      aria-labelledby="activity-panel-title"
      data-testid="activity-panel"
    >
      <header className={styles.header}>
        <div>
          <h3 id="activity-panel-title" className={styles.title}>
            Atividade recente
          </h3>
          <p className={styles.subtitle}>Últimos {LIMIT} eventos auditados</p>
        </div>
        <Link to="/admin/audit" className={styles.seeAll}>
          Ver tudo →
        </Link>
      </header>

      {error ? (
        <div role="alert" className={styles.errorBanner}>
          <span>{resolveQueryError(error, "Não foi possível carregar a atividade.")}</span>
          <button
            type="button"
            className={styles.retryButton}
            onClick={() => refetch()}
          >
            Tentar novamente
          </button>
        </div>
      ) : null}

      {isLoading ? (
        <div className={styles.skeleton} aria-hidden="true">
          {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
            <div key={i} className={styles.skeletonRow}>
              <div className={styles.skeletonBar} style={{ maxWidth: `${50 + ((i * 11) % 40)}%` }} />
              <div className={`${styles.skeletonBar} ${styles.skeletonShort}`} />
            </div>
          ))}
        </div>
      ) : null}

      {!isLoading && !error && events.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyTitle}>Nenhuma atividade registrada</span>
          <span>Eventos auditados aparecerão aqui assim que ocorrerem.</span>
        </div>
      ) : null}

      {!isLoading && !error && events.length > 0 ? (
        <ul className={styles.list} role="list">
          {events.map((ev, i) => {
            const prev = events[i - 1];
            const groupedWithPrev = !!prev && prev.actorId === ev.actorId;
            return (
              <ActivityEventRow
                key={ev.id}
                event={ev}
                groupedWithPrev={groupedWithPrev}
              />
            );
          })}
        </ul>
      ) : null}
    </section>
  );
}
