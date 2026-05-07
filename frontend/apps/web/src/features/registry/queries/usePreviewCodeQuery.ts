import { useQuery } from '@tanstack/react-query';
import { previewCode, type PreviewCodeResponse } from '../api/controlledDocuments';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from '../../documents/queries/_constants';

export function usePreviewCodeQuery(profileCode: string | null, areaCode: string | null) {
  return useQuery<PreviewCodeResponse>({
    queryKey: QK.controlledDocuments.preview(profileCode ?? '', areaCode ?? ''),
    queryFn: () => previewCode(profileCode!, areaCode!),
    enabled: Boolean(profileCode && areaCode),
    staleTime: STALE_FIVE_MINUTES,
  });
}
