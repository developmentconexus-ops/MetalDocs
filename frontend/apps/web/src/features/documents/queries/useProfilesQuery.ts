import { useQuery } from '@tanstack/react-query';
import { fetchProfiles } from '../../taxonomy/api/taxonomy';
import { QK } from '../../../lib/queryKeys';

const FIVE_MINUTES = 5 * 60 * 1000;

export function useProfilesQuery() {
  return useQuery({
    queryKey: QK.taxonomy.profiles(),
    queryFn: () => fetchProfiles(),
    staleTime: FIVE_MINUTES,
  });
}
