import { useQuery } from '@tanstack/react-query';
import { fetchAreas } from '../../taxonomy/api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './_constants';

export function useAreasQuery() {
  return useQuery({
    queryKey: QK.taxonomy.areas(),
    queryFn: () => fetchAreas(),
    staleTime: STALE_FIVE_MINUTES,
  });
}
