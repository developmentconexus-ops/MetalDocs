import { useQuery } from '@tanstack/react-query';
import { api } from '../../../lib/api/client';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './_constants';

export function useBlankTemplateQuery() {
  return useQuery({
    queryKey: QK.templates.blank(),
    queryFn: async () => {
      const { data, error } = await api.GET('/api/v1/templates/system/blank');
      if (error) throw error;
      return data;
    },
    staleTime: STALE_FIVE_MINUTES,
  });
}
