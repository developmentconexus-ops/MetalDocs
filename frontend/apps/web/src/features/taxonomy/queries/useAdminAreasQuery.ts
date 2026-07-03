import { useQuery } from '@tanstack/react-query';
import { fetchAreas } from '../api/taxonomy';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

/**
 * Taxonomy admin page areas list, with an archived-inclusion toggle. Kept
 * separate from `documents/queries/useAreasQuery` (the shared active-only
 * picker hook) for the same reason as `useAdminProfilesQuery`.
 */
export function useAdminAreasQuery(includeArchived: boolean) {
  return useQuery({
    queryKey: QK.taxonomy.areas({ includeArchived }),
    queryFn: () => fetchAreas(includeArchived),
    staleTime: STALE_FIVE_MINUTES,
  });
}
