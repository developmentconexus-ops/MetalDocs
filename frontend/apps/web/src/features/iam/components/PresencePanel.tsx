import { useQueryClient } from "@tanstack/react-query";
import { usePresenceStream } from "../queries/usePresenceStream";
import { Avatar } from "../../../components/ui/Avatar";
import { getRelativeTime } from "../presenters/relative-time-presenter";
import { resolveQueryError } from "../../../lib/api/resolveQueryError";
import { QK } from "../../../lib/queryKeys";
import ErrorBanner from "./ErrorBanner";
import styles from "./PresencePanel.module.css";

const SKELETON_ROWS = 5;

export default function PresencePanel() {
  const { data, isLoading, error } = usePresenceStream();
  const queryClient = useQueryClient();
  const items = data?.items ?? [];
  const count = items.length;

  return (
    <section
      className={styles.panel}
      aria-labelledby="presence-panel-title"
      data-testid="presence-panel"
    >
      <header className={styles.header}>
        <div>
          <h3 id="presence-panel-title" className={styles.title}>
            Quem está online
          </h3>
          <p className={styles.subtitle}>Snapshot atualizado a cada 30s</p>
        </div>
        {!isLoading && !error ? (
          <span className={styles.countBadge} aria-live="polite">
            <span className={styles.dot} aria-hidden="true" />
            {count}
          </span>
        ) : null}
      </header>

      {error ? (
        <ErrorBanner
          message={resolveQueryError(error, "Não foi possível carregar a presença.")}
          onRetry={() =>
            queryClient.invalidateQueries({ queryKey: QK.iam.presenceSnapshot() })
          }
        />
      ) : null}

      {isLoading ? (
        <div className={styles.skeleton} aria-hidden="true">
          {Array.from({ length: SKELETON_ROWS }).map((_, i) => (
            <div key={i} className={styles.skeletonRow}>
              <div className={styles.skeletonAvatar} />
              <div className={styles.skeletonBar} style={{ maxWidth: `${60 + ((i * 13) % 30)}%` }} />
            </div>
          ))}
        </div>
      ) : null}

      {!isLoading && !error && count === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyTitle}>Ninguém online no momento</span>
          <span>A presença aparece em tempo real quando usuários se conectam.</span>
        </div>
      ) : null}

      {!isLoading && !error && count > 0 ? (
        <ul className={styles.list} role="list">
          {items.map((u) => (
            <li key={u.userId} className={styles.row}>
              <Avatar name={u.displayName || u.username} size="md" />
              <span className={styles.name}>{u.displayName || u.username}</span>
              <span className={styles.username}>@{u.username}</span>
              <time
                className={styles.lastSeen}
                dateTime={u.lastSeenAt}
                title={new Date(u.lastSeenAt).toLocaleString("pt-BR", {
                  timeZone: "America/Sao_Paulo",
                })}
              >
                {getRelativeTime(u.lastSeenAt)}
              </time>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}
