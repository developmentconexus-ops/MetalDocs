import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../api/templates';
import { QK } from '../../../lib/queryKeys';

export function useTemplatesQuery() {
  return useQuery({
    queryKey: QK.templates.list(),
    queryFn: () => listTemplates(),
    staleTime: 60_000,
  });
}
