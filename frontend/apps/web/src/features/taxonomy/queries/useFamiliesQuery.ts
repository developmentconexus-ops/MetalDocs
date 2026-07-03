import { useQuery } from '@tanstack/react-query';
import { fetchFamilies } from '../api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

export function useFamiliesQuery(includeInactive = false) {
  return useQuery({
    queryKey: QK.taxonomy.families({ includeInactive }),
    queryFn: () => fetchFamilies(includeInactive),
    staleTime: STALE_FIVE_MINUTES,
  });
}
