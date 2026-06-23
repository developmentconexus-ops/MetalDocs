import { Dialog } from '../../../components/ui/Dialog';
import { resolveErrorMessage } from '../../../lib/api';
import { useNotificationsQuery } from '../queries/useNotificationQueries';
import {
  useMarkAllNotificationsReadMutation,
  useMarkNotificationReadMutation,
} from '../queries/useNotificationMutations';
import { NotificationRow } from './NotificationRow';
import styles from './NotificationsSpotlight.module.css';

interface NotificationsSpotlightProps {
  open: boolean;
  onClose: () => void;
}

// Caller's full inbox shown in the spotlight. 50 is the clamp ceiling on the
// list endpoint, so this is "everything recent" without paging.
const SPOTLIGHT_LIMIT = 50;

// NotificationsSpotlight is the full-list overlay opened from the bell popover's
// "Mostrar todas" action: the complete recent list plus a bulk mark-all-read.
export function NotificationsSpotlight({ open, onClose }: NotificationsSpotlightProps) {
  const list = useNotificationsQuery({ limit: SPOTLIGHT_LIMIT });
  const markRead = useMarkNotificationReadMutation();
  const markAll = useMarkAllNotificationsReadMutation();

  const items = list.data?.items ?? [];
  const hasUnread = items.some((n) => n.status !== 'READ');

  const footer = (
    <div className={styles.footer}>
      <button
        type="button"
        className={styles.markAll}
        onClick={() => markAll.mutate()}
        disabled={!hasUnread || markAll.isPending}
      >
        {markAll.isPending ? 'Marcando…' : 'Marcar todas como lidas'}
      </button>
    </div>
  );

  return (
    <Dialog open={open} title="Notificações" onClose={onClose} footer={footer}>
      {list.isLoading ? (
        <p className={styles.state}>Carregando notificações…</p>
      ) : list.isError ? (
        <div className={styles.state} role="alert">
          <p>{resolveErrorMessage(list.error)}</p>
          <button type="button" className={styles.retry} onClick={() => list.refetch()}>
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
    </Dialog>
  );
}
