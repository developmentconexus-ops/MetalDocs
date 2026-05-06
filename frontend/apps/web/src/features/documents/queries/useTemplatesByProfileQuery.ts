import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../../templates/api/templatesV2';

export function useTemplatesByProfileQuery(profileCode: string | null) {
  return useQuery({
    queryKey: ['templates', { docType: profileCode }] as const,
    queryFn: () => listTemplates({ doc_type: profileCode ?? undefined }),
    enabled: profileCode !== null,
  });
}
