import { useQuery } from '@tanstack/react-query';
import { fetchLibraryStats } from '../api/library';
import { QK } from '../../../lib/queryKeys';

export function useLibraryStatsQuery() {
  return useQuery({
    queryKey: QK.documents.stats(),
    queryFn: fetchLibraryStats,
    // Stats rarely change; don't refetch on every focus/mount.
    staleTime: 30_000,
  });
}
