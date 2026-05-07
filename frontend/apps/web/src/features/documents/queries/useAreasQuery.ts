import { useQuery } from '@tanstack/react-query';
import { fetchAreas } from '../../taxonomy/api/taxonomy';
import { QK } from '../../../lib/queryKeys';

const FIVE_MINUTES = 5 * 60 * 1000;

export function useAreasQuery() {
  return useQuery({
    queryKey: QK.taxonomy.areas(),
    queryFn: fetchAreas,
    staleTime: FIVE_MINUTES,
  });
}
