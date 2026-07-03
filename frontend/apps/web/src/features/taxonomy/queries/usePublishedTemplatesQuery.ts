import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../../templates/api/templates';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './constants';

/**
 * Published templates, for the profile default-template picker. Filters to
 * published_version_id != null client-side — the templates list endpoint has
 * no "published only" server filter.
 */
export function usePublishedTemplatesQuery(enabled: boolean) {
  return useQuery({
    queryKey: QK.templates.list(),
    queryFn: async () => {
      const { templates } = await listTemplates();
      return templates.filter((t) => t.published_version_id != null);
    },
    enabled,
    staleTime: STALE_FIVE_MINUTES,
  });
}
