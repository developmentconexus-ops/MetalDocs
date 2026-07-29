import { useQuery } from '@tanstack/react-query';
import { getDocument } from '../api/documents';
import { QK } from '../../../lib/queryKeys';
import { LIFECYCLE_POLL_INTERVAL_MS, isDocumentLifecycleSettling } from '../lib/documentReleasePresentation';

type DocumentDetailQueryOptions = {
  /**
   * Poll while the document's lifecycle is still settling on its own — the legacy
   * `status === 'scheduled'` wait AND (ADR 0085 Stage C) a release generation held
   * in a transient coordinator state. `isDocumentLifecycleSettling` owns the split.
   */
  pollLifecycleUntilSettled?: boolean;
  refetchInterval?: number | false;
};

export function useDocumentDetailQuery(id: string, options: DocumentDetailQueryOptions = {}) {
  return useQuery({
    queryKey: QK.documents.detail(id),
    queryFn: () => getDocument(id),
    enabled: Boolean(id),
    refetchInterval: (query) => {
      if (typeof options.refetchInterval !== 'undefined') {
        return options.refetchInterval;
      }
      if (!options.pollLifecycleUntilSettled) {
        return false;
      }
      return isDocumentLifecycleSettling(query.state.data) ? LIFECYCLE_POLL_INTERVAL_MS : false;
    },
  });
}
