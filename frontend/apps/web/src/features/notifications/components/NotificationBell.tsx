import { useEffect, useRef, useState } from 'react';

import { Icon } from '../../../components/ui/Icon';
import { resolveErrorMessage } from '../../../lib/api';
import { useNotificationsQuery, useUnreadCountQuery } from '../queries/useNotificationQueries';
import { useMarkNotificationReadMutation } from '../queries/useNotificationMutations';
import { NotificationRow } from './NotificationRow';
import { NotificationsSpotlight } from './NotificationsSpotlight';
import styles from './NotificationBell.module.css';

// Recent slice shown inline in the popover; the full list lives in the spotlight.
const POPOVER_LIMIT = 6;
const BADGE_CAP = 9;

// NotificationBell is the top-bar entry point: a bell with an unread-count
// badge, a popover of recent notifications, and a "Mostrar todas" action that
// opens the full-list spotlight.
export function NotificationBell() {
  const [open, setOpen] = useState(false);
  const [spotlightOpen, setSpotlightOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const unread = useUnreadCountQuery();
  const recent = useNotificationsQuery({ limit: POPOVER_LIMIT });
  const markRead = useMarkNotificationReadMutation();

  const count = unread.data ?? 0;
  const items = recent.data?.items ?? [];

  // Close the popover on outside click or Escape.
  useEffect(() => {
    if (!open) return;
    const onPointer = (event: PointerEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('pointerdown', onPointer);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('pointerdown', onPointer);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  const openSpotlight = () => {
    setOpen(false);
    setSpotlightOpen(true);
  };

  const badgeLabel = count > BADGE_CAP ? `${BADGE_CAP}+` : String(count);

  return (
    <div className={styles.container} ref={containerRef}>
      <button
        type="button"
        className={styles.bell}
        aria-label={count > 0 ? `Notificações (${count} não lidas)` : 'Notificações'}
        aria-haspopup="dialog"
        aria-expanded={open}
        title="Notificações"
        onClick={() => setOpen((v) => !v)}
      >
        <Icon name="bell" size={16} />
        {count > 0 ? (
          <span className={styles.badge} aria-hidden="true">
            {badgeLabel}
          </span>
        ) : null}
      </button>

      {open ? (
        <div className={styles.popover} role="dialog" aria-label="Notificações recentes">
          <header className={styles.header}>
            <span className={styles.heading}>Notificações</span>
            {count > 0 ? <span className={styles.headerCount}>{count} não lidas</span> : null}
          </header>

          <div className={styles.scroll}>
            {recent.isLoading ? (
              <p className={styles.state}>Carregando…</p>
            ) : recent.isError ? (
              <div className={styles.state} role="alert">
                <p>{resolveErrorMessage(recent.error)}</p>
                <button type="button" className={styles.retry} onClick={() => recent.refetch()}>
                  Tentar novamente
                </button>
              </div>
            ) : items.length === 0 ? (
              <p className={styles.state}>Nenhuma notificação.</p>
            ) : (
              <ul className={styles.list}>
                {items.map((n) => (
                  <NotificationRow key={n.id} notification={n} onMarkRead={markRead.mutate} />
                ))}
              </ul>
            )}
          </div>

          <footer className={styles.footer}>
            <button type="button" className={styles.showAll} onClick={openSpotlight}>
              Mostrar todas
            </button>
          </footer>
        </div>
      ) : null}

      <NotificationsSpotlight open={spotlightOpen} onClose={() => setSpotlightOpen(false)} />
    </div>
  );
}
