import { useQuery } from '@tanstack/react-query';
import { getDocumentRevisionHistory } from '../api/documents';
import { QK } from '../../../lib/queryKeys';

export function useDocumentRevisionHistoryQuery(id: string) {
  return useQuery({
    queryKey: QK.documents.revisionHistory(id),
    queryFn: () => getDocumentRevisionHistory(id),
    enabled: Boolean(id),
  });
}
