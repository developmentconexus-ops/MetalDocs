import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { fetchLibraryStats } from '../../documents/api/library';

export function useDashboardStatsQuery() {
  return useQuery({
    queryKey: QK.documents.stats(),
    queryFn: fetchLibraryStats,
    staleTime: 30_000,
  });
}
