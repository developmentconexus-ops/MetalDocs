import { useQuery } from '@tanstack/react-query';
import { getInstance } from '../api/approvalApi';
import type { ApprovalInstance } from '../api/approvalTypes';
import { QK } from '../../../lib/queryKeys';

export function useApprovalInstanceQuery(documentId: string, enabled: boolean) {
  return useQuery<ApprovalInstance | null>({
    queryKey: QK.approval.instance(documentId),
    queryFn: async () => {
      try {
        return await getInstance(documentId);
      } catch (err) {
        if ((err as { status?: number }).status === 404) return null; // no active instance
        throw err;
      }
    },
    enabled: Boolean(documentId) && enabled,
  });
}
