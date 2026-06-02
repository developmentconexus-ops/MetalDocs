import { useQuery } from '@tanstack/react-query';

import { QK } from '../../../lib/queryKeys';
import { listRoutes, type ListRoutesResponse, type RouteSummary } from '../api/routeAdminApi';

const STALE_THIRTY_SECONDS = 30 * 1000;

/**
 * Subscribes to the full route catalogue. Mirrors `useInboxQuery` shape.
 *
 * ETag seeding happens inside the `listRoutes()` transport layer (so every
 * caller benefits), not here — the cached `version` of each row is seeded into
 * the ETag cache on each successful list so the next PUT / DELETE on that route
 * carries the right `If-Match` header without an explicit re-read.
 */
export function useRoutesQuery() {
  return useQuery<ListRoutesResponse, Error, RouteSummary[]>({
    queryKey: QK.approval.routes.list(),
    queryFn: () => listRoutes(),
    staleTime: STALE_THIRTY_SECONDS,
    select: (response) => response.routes,
  });
}
