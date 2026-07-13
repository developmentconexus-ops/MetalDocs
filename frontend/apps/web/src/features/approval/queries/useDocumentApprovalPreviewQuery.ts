import { useQuery } from '@tanstack/react-query';
import { api } from '../../../lib/api/client';
import { QK } from '../../../lib/queryKeys';
import type { components } from '../../../lib/api-types';

export type ApprovalRoutePreview = components['schemas']['ApprovalRoutePreview'];

/**
 * SLICE 8b (unit 3.2): submit-time preview of the resolved active approval
 * route for a document, GET /documents/{id}/approval-preview. Read-only —
 * lets the submit dialog build a chosen_actors picker for any submit_choice
 * stage before the caller actually POSTs /submit. Enabled only once a
 * document id is known and the caller opts in (the submit dialog is open).
 */
export function useDocumentApprovalPreviewQuery(documentId: string, enabled: boolean) {
  return useQuery<ApprovalRoutePreview>({
    queryKey: QK.approval.documentPreview(documentId),
    queryFn: async () => {
      const { data, error } = await api.GET('/documents/{id}/approval-preview', {
        params: { path: { id: documentId } },
      });
      if (error) throw error;
      return data;
    },
    enabled: Boolean(documentId) && enabled,
  });
}
