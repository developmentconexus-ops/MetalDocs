import { useQuery } from '@tanstack/react-query';
import { getDistributionSummary } from '../api/distribution';
import { QK } from '../../../lib/queryKeys';

export function useDistributionSummaryQuery(documentId: string) {
  return useQuery({
    queryKey: QK.documents.distribution.summary(documentId),
    queryFn: () => getDistributionSummary(documentId),
    enabled: Boolean(documentId),
  });
}
