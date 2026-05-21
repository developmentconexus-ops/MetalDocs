import { useQuery } from '@tanstack/react-query';
import { getDocumentRevisionHistory } from '../api/documents';
import { QK } from '../../../lib/queryKeys';

type DocumentRevisionHistoryQueryOptions = {
  refetchInterval?: number | false;
};

export function useDocumentRevisionHistoryQuery(id: string, options: DocumentRevisionHistoryQueryOptions = {}) {
  return useQuery({
    queryKey: QK.documents.revisionHistory(id),
    queryFn: () => getDocumentRevisionHistory(id),
    enabled: Boolean(id),
    refetchInterval: options.refetchInterval,
  });
}
