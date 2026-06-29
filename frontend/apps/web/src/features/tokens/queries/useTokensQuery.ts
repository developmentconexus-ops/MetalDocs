import { useQuery } from '@tanstack/react-query';
import { QK } from '../../../lib/queryKeys';
import { listTokens } from '../api/tokens';

const STALE_ONE_MINUTE = 60 * 1000;

export function useTokensQuery() {
  return useQuery({
    queryKey: QK.tokens.list(),
    queryFn: listTokens,
    staleTime: STALE_ONE_MINUTE,
  });
}
