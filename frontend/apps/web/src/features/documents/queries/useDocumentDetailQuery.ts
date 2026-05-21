import { useQuery } from '@tanstack/react-query';
import { getDocument } from '../api/documents';
import { QK } from '../../../lib/queryKeys';

type DocumentDetailQueryOptions = {
  pollScheduledLifecycle?: boolean;
  refetchInterval?: number | false;
};

export function useDocumentDetailQuery(id: string, options: DocumentDetailQueryOptions = {}) {
  return useQuery({
    queryKey: QK.documents.detail(id),
    queryFn: () => getDocument(id),
    enabled: Boolean(id),
    refetchInterval: (query) => {
      if (typeof options.refetchInterval !== 'undefined') {
        return options.refetchInterval;
      }
      if (!options.pollScheduledLifecycle) {
        return false;
      }
      return query.state.data?.Status === 'scheduled' ? 5_000 : false;
    },
  });
}
