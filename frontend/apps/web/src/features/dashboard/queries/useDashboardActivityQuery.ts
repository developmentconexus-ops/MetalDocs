import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { fetchRecentAuditEvents } from '../api/audit';

const ACTIVITY_LIMIT = 8;

export function useDashboardActivityQuery() {
  return useQuery({
    queryKey: QK.audit.recent(ACTIVITY_LIMIT),
    queryFn: () => fetchRecentAuditEvents(ACTIVITY_LIMIT),
    staleTime: 30_000,
  });
}
