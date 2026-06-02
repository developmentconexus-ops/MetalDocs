import { useQuery } from '@tanstack/react-query';

import { QK } from '../../../lib/queryKeys';
import {
  listRoutes,
  seedRouteEtag,
  type ListRoutesResponse,
  type RouteSummary,
} from '../api/routeAdminApi';

const STALE_THIRTY_SECONDS = 30 * 1000;

/**
 * Subscribes to the full route catalogue. Mirrors `useInboxQuery` shape.
 *
 * On every successful refetch, the cached `version` of each row is seeded into
 * the ETag cache so the next PUT / DELETE on that route carries the right
 * `If-Match` header without an explicit re-read.
 */
export function useRoutesQuery() {
  return useQuery<ListRoutesResponse, Error, RouteSummary[]>({
    queryKey: QK.approval.routes.list(),
    queryFn: async () => {
      const response = await listRoutes();
      for (const route of response.routes) {
        seedRouteEtag(route.id, route.version);
      }
      return response;
    },
    staleTime: STALE_THIRTY_SECONDS,
    select: (response) => response.routes,
  });
}
