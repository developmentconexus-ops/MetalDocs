import { useQuery } from '@tanstack/react-query';
import { getTemplate } from '../api/templates';
import { QK } from '../../../lib/queryKeys';

export function useTemplateDetailQuery(templateId: string) {
  return useQuery({
    queryKey: QK.templates.detail(templateId),
    queryFn: () => getTemplate(templateId),
    enabled: Boolean(templateId),
    staleTime: 60_000,
  });
}
