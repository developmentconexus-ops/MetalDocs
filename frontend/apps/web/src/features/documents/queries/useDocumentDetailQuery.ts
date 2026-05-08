import { useQuery } from '@tanstack/react-query';
import { getDocument } from '../api/documentsV2';
import { QK } from '../../../lib/queryKeys';

export function useDocumentDetailQuery(id: string) {
  return useQuery({
    queryKey: QK.documents.detail(id),
    queryFn: () => getDocument(id),
    enabled: Boolean(id),
  });
}
