import { useQuery } from '@tanstack/react-query';
import { getApprovalInstance } from '../api/documents';
import { QK } from '../../../lib/queryKeys';

export function useApprovalInstanceQuery(documentId: string) {
  return useQuery({
    queryKey: QK.approval.instance(documentId),
    queryFn: () => getApprovalInstance(documentId),
    enabled: Boolean(documentId),
    // 404 = no approval instance; treat as null not error
    retry: (failureCount: number, error: unknown) => {
      if ((error as { status?: number }).status === 404) return false;
      return failureCount < 2;
    },
  });
}
