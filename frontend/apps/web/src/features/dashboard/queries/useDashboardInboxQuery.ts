import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { fetchDashboardInbox } from '../api/dashboard';

export function useDashboardInboxQuery() {
  return useQuery({
    queryKey: QK.inbox({ limit: 6 }),
    queryFn: fetchDashboardInbox,
    staleTime: 30_000,
  });
}
