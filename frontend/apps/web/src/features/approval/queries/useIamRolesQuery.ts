import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import {
  getIamRolesFallback,
  type IamRole,
  type IamRolesResponse,
} from '../../../lib/iam/roles';

const STALE_TEN_MINUTES = 10 * 60 * 1000;

/**
 * Subscribes to the IAM role catalogue.
 *
 * ADR 0018 Option B — until `GET /api/v1/iam/roles` ships, the data source is
 * the local fallback in `lib/iam/roles.ts`. Swapping to the live endpoint is a
 * one-line change in `queryFn`; the hook signature and query key stay stable.
 */
export function useIamRolesQuery() {
  return useQuery<IamRolesResponse, Error, IamRole[]>({
    queryKey: QK.iam.roles(),
    queryFn: async () => getIamRolesFallback(),
    staleTime: STALE_TEN_MINUTES,
    select: (response) => response.roles,
  });
}
