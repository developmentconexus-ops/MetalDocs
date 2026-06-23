import type { Notification } from '../api/notifications';
import { eventTypeLabel } from '../lib/eventTypeLabel';
import { formatNotificationTimestamp } from '../lib/formatNotificationTimestamp';
import styles from './NotificationRow.module.css';

interface NotificationRowProps {
  notification: Notification;
  /** Invoked with the row id when an unread row is activated. */
  onMarkRead?: (id: string) => void;
}

// A notification is unread until its status reaches READ (PENDING/SENT are the
// two unread states the backend emits).
function isUnread(notification: Notification): boolean {
  return notification.status !== 'READ';
}

// NotificationRow is a presentational row: event chip, title, message, relative
// time, and an unread marker. Clicking an unread row marks it read; read rows
// are inert. Data fetching and mutation live in the parent.
export function NotificationRow({ notification, onMarkRead }: NotificationRowProps) {
  const unread = isUnread(notification);

  const handleActivate = () => {
    if (unread) onMarkRead?.(notification.id);
  };

  return (
    <li
      className={`${styles.row} ${unread ? styles.unread : ''}`}
      data-status={notification.status}
    >
      {unread ? (
        <button
          type="button"
          className={styles.hit}
          onClick={handleActivate}
          aria-label={`Marcar como lida: ${notification.title}`}
        />
      ) : null}

      <span className={styles.dot} aria-hidden="true" data-on={unread} />

      <div className={styles.body}>
        <div className={styles.header}>
          <span className={styles.chip}>{eventTypeLabel(notification.event_type)}</span>
          <time className={styles.time} dateTime={notification.created_at}>
            {formatNotificationTimestamp(notification.created_at)}
          </time>
        </div>
        <p className={styles.title}>{notification.title}</p>
        {notification.message ? <p className={styles.message}>{notification.message}</p> : null}
      </div>
    </li>
  );
}
