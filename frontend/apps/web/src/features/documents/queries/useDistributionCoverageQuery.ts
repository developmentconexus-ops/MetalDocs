import { useQuery } from '@tanstack/react-query';
import { getDistributionCoverage } from '../api/distribution';
import { QK } from '../../../lib/queryKeys';

export function useDistributionCoverageQuery(documentId: string) {
  return useQuery({
    queryKey: QK.documents.distribution.coverage(documentId),
    queryFn: () => getDistributionCoverage(documentId),
    enabled: Boolean(documentId),
  });
}
