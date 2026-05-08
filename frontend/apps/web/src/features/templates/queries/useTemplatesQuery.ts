import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../api/templatesV2';
import { QK } from '../../../lib/queryKeys';

export function useTemplatesQuery() {
  return useQuery({
    queryKey: QK.templates.list(),
    queryFn: () => listTemplates(),
  });
}
