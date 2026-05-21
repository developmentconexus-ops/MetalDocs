import { useQuery } from '@tanstack/react-query';

import { fetchActiveDocumentInstance } from '../../controlled-documents/api/controlledDocuments';
import { QK } from '../../../lib/queryKeys';

type ControlledDocumentActiveDocumentQueryOptions = {
  refetchInterval?: number | false;
};

export function useControlledDocumentActiveDocumentQuery(
  controlledDocumentId: string | null | undefined,
  options: ControlledDocumentActiveDocumentQueryOptions = {},
) {
  return useQuery({
    queryKey: QK.controlledDocuments.activeDocument(controlledDocumentId ?? ''),
    queryFn: () => fetchActiveDocumentInstance(controlledDocumentId ?? ''),
    enabled: Boolean(controlledDocumentId),
    refetchInterval: options.refetchInterval,
  });
}

