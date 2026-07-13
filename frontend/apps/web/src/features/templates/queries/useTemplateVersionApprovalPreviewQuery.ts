import { useQuery } from '@tanstack/react-query';
import { api } from '../../../lib/api/client';
import { QK } from '../../../lib/queryKeys';
import type { components } from '../../../lib/api-types';

export type ApprovalRoutePreview = components['schemas']['ApprovalRoutePreview'];

/**
 * SLICE 8b (unit 3.2): submit-time preview of the resolved active approval
 * route for a template version, GET
 * /templates/{id}/versions/{n}/approval-preview. Mirrors
 * useDocumentApprovalPreviewQuery (features/approval/queries) — read-only,
 * lets the submit-for-approval action build a chosen_actors picker for any
 * submit_choice stage before POSTing submit-for-approval. Enabled only once
 * both ids are known and the caller opts in (the submit dialog is open).
 */
export function useTemplateVersionApprovalPreviewQuery(
  templateId: string,
  versionNumber: number,
  enabled: boolean,
) {
  return useQuery<ApprovalRoutePreview>({
    queryKey: QK.templates.versionPreview(templateId, versionNumber),
    queryFn: async () => {
      const { data, error } = await api.GET('/templates/{id}/versions/{n}/approval-preview', {
        params: { path: { id: templateId, n: versionNumber } },
      });
      if (error) throw error;
      return data;
    },
    enabled: Boolean(templateId) && Number.isFinite(versionNumber) && enabled,
  });
}
