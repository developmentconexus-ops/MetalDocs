import { useCallback, useRef } from "react";
import * as api from './api/notifications';
import type { OperationsStreamSnapshot } from './api/notifications';
import { useUiStore } from "../../store/ui.store";
import { asMessage } from "../shared/errors";

// Notifications are now fetched via TanStack Query (QK.notifications.unreadCount).
// This hook retains only the SSE operations stream subscription used by AppShell.
export function useNotifications() {
  const { setError } = useUiStore();

  const handleMarkNotificationRead = useCallback(
    async (notificationId: string) => {
      try {
        await api.markNotificationRead(notificationId);
      } catch (err) {
        setError(asMessage(err));
      }
    },
    [setError],
  );

  const lastSnapshotRef = useRef<OperationsStreamSnapshot | null>(null);
  const lastRefreshRef = useRef(0);

  const subscribeOperations = useCallback(
    (onRefresh: () => void) =>
      api.subscribeOperationsStream(
        (snapshot) => {
          const now = Date.now();
          const previous = lastSnapshotRef.current;
          const hasChanges =
            !previous ||
            previous.pendingNotifications !== snapshot.pendingNotifications ||
            previous.pendingApprovals !== snapshot.pendingApprovals ||
            previous.documentsInReview !== snapshot.documentsInReview ||
            previous.totalDocuments !== snapshot.totalDocuments;
          const enoughTimePassed = now - lastRefreshRef.current >= 15000;

          lastSnapshotRef.current = snapshot;
          if (!hasChanges && !enoughTimePassed) return;
          lastRefreshRef.current = now;
          onRefresh();
        },
        () => {
          // Stream keeps retrying in browser; UI fallback remains available.
        },
      ),
    [],
  );

  return {
    handleMarkNotificationRead,
    subscribeOperations,
  };
}
