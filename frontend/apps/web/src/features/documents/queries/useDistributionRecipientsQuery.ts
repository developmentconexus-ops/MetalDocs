import { useInfiniteQuery } from '@tanstack/react-query';
import { listDistributionRecipients } from '../api/distribution';
import { QK } from '../../../lib/queryKeys';

export function useDistributionRecipientsQuery(
  documentId: string,
  opts: { limit?: number } = {},
) {
  return useInfiniteQuery({
    queryKey: QK.documents.distribution.recipients(documentId, opts),
    queryFn: ({ pageParam }) =>
      listDistributionRecipients(documentId, {
        cursor: pageParam as string | undefined,
        limit: opts.limit,
      }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) =>
      lastPage.page.has_more ? (lastPage.page.next_cursor ?? undefined) : undefined,
    enabled: Boolean(documentId),
  });
}
