import { useQuery } from '@tanstack/react-query';
import { fetchProfiles } from '../../taxonomy/api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './_constants';

export function useProfilesQuery() {
  return useQuery({
    queryKey: QK.taxonomy.profiles(),
    queryFn: () => fetchProfiles(),
    staleTime: STALE_FIVE_MINUTES,
  });
}
