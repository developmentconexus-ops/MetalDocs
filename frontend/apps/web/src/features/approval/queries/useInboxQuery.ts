import { useQuery } from '@tanstack/react-query';
import { listInbox } from '../api/approvalApi';
import type { ListInboxParams } from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export function useInboxQuery(params: ListInboxParams = {}) {
  return useQuery({
    queryKey: QK.inbox(params),
    queryFn: () => listInbox(params),
  });
}
