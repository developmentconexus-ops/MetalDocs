import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { QK } from '../../../lib/queryKeys';
import { resolveErrorMessage } from '../../../lib/api';
import { markAllNotificationsRead, markNotificationRead } from '../api/notifications';

// Both the unread badge and any open list reflect a read transition, so every
// mutation invalidates both keys on settle.
function invalidateNotifications(queryClient: ReturnType<typeof useQueryClient>): void {
  void queryClient.invalidateQueries({ queryKey: QK.notifications.unreadCount() });
  void queryClient.invalidateQueries({ queryKey: QK.notifications.list() });
}

// useMarkNotificationReadMutation marks a single row read. Silent on success
// (the row click is its own feedback); only failures surface a toast.
export function useMarkNotificationReadMutation() {
  const queryClient = useQueryClient();

  return useMutation<void, Error, string>({
    mutationFn: (id) => markNotificationRead(id),
    onError: (err) => {
      toast.error(resolveErrorMessage(err));
    },
    onSettled: () => {
      invalidateNotifications(queryClient);
    },
  });
}

// useMarkAllNotificationsReadMutation drives the spotlight "marcar todas como
// lidas" action. Reports how many rows were transitioned (idempotent: 0 when
// nothing was unread).
export function useMarkAllNotificationsReadMutation() {
  const queryClient = useQueryClient();

  return useMutation<number, Error, void>({
    mutationFn: () => markAllNotificationsRead(),
    onSuccess: (updated) => {
      toast.success(
        updated > 0
          ? `${updated} ${updated === 1 ? 'notificação marcada' : 'notificações marcadas'} como lida${updated === 1 ? '' : 's'}.`
          : 'Nenhuma notificação pendente.',
      );
    },
    onError: (err) => {
      toast.error(resolveErrorMessage(err));
    },
    onSettled: () => {
      invalidateNotifications(queryClient);
    },
  });
}
