import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../../templates';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './_constants';

export function useTemplatesByProfileQuery(profileCode: string | null) {
  return useQuery({
    // Use sentinel key when disabled so this observer never collides with a real templates.list() cache.
    queryKey: QK.templates.byProfile(profileCode ?? '__disabled__'),
    // Safe: `enabled` blocks queryFn while profileCode is null.
    queryFn: () => listTemplates({ doc_type: profileCode! }),
    enabled: profileCode !== null,
    staleTime: STALE_FIVE_MINUTES,
  });
}
