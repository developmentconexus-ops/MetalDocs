import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../../templates/api/templatesV2';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from './_constants';

export function useTemplatesByProfileQuery(profileCode: string | null) {
  return useQuery({
    queryKey: profileCode ? QK.templates.byProfile(profileCode) : QK.templates.list(),
    // Safe: `enabled` blocks queryFn while profileCode is null.
    queryFn: () => listTemplates({ doc_type: profileCode! }),
    enabled: profileCode !== null,
    staleTime: STALE_FIVE_MINUTES,
  });
}
