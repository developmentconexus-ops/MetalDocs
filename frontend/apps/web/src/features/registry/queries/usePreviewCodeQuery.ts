import { useQuery } from '@tanstack/react-query';
import { previewCode, type PreviewCodeResponse } from '../api/controlledDocuments';
import { QK } from '../../../lib/queryKeys';
import { STALE_FIVE_MINUTES } from '../../documents/queries/_constants';

export function usePreviewCodeQuery(profileCode: string | null, areaCode: string | null) {
  const ready = Boolean(profileCode && areaCode);

  return useQuery<PreviewCodeResponse>({
    queryKey: ready
      ? QK.controlledDocuments.preview(profileCode!, areaCode!)
      : QK.controlledDocuments.list(),
    queryFn: () => previewCode(profileCode!, areaCode!),
    enabled: ready,
    staleTime: STALE_FIVE_MINUTES,
  });
}
