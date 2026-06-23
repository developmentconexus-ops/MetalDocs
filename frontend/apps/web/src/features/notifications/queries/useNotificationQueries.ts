import { useQuery } from '@tanstack/react-query';

import { QK } from '../../../lib/queryKeys';
import {
  fetchNotifications,
  fetchUnreadCount,
  type NotificationsListParams,
} from '../api/notifications';

// useNotificationsQuery lists the caller's notifications (self-scope, newest
// first). Mirrors the dashboard query shape: centralized key + modest staleTime.
export function useNotificationsQuery(params: NotificationsListParams = {}) {
  return useQuery({
    queryKey: QK.notifications.list(params),
    queryFn: () => fetchNotifications(params),
    staleTime: 30_000,
  });
}

// useUnreadCountQuery powers the bell badge. There is no live stream (the SSE
// path was removed), so the badge stays current by polling on a slow interval
// and on window refocus.
export function useUnreadCountQuery() {
  return useQuery({
    queryKey: QK.notifications.unreadCount(),
    queryFn: fetchUnreadCount,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}
