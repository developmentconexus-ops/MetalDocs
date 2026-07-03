import { useQuery } from '@tanstack/react-query';
import { fetchProfiles } from '../api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

/**
 * Taxonomy admin page profiles list, with an archived-inclusion toggle. Kept
 * separate from `useProfilesQuery` (the shared active-only picker hook) so
 * an admin viewing archived rows never leaks them into a document/template
 * profile picker's cache entry.
 */
export function useAdminProfilesQuery(includeArchived: boolean) {
  return useQuery({
    queryKey: QK.taxonomy.profiles({ includeArchived }),
    queryFn: () => fetchProfiles(includeArchived),
    staleTime: STALE_FIVE_MINUTES,
  });
}
