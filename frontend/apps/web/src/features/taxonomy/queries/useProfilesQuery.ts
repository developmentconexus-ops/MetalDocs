import { useQuery } from '@tanstack/react-query';
import { fetchProfiles } from '../api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

/**
 * Active (non-archived) taxonomy profiles. This is the shared, cross-feature
 * hook consumed by documents/templates/approval pickers — keep its no-arg,
 * active-only default stable. The taxonomy admin page's archived toggle uses
 * `useAdminProfilesQuery` instead so this shared cache entry never carries
 * archived rows into unrelated pickers.
 */
export function useProfilesQuery() {
  return useQuery({
    queryKey: QK.taxonomy.profiles(),
    queryFn: () => fetchProfiles(),
    staleTime: STALE_FIVE_MINUTES,
  });
}
